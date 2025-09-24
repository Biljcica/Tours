package service

import (
	"context"
	"errors"
	"time"
	"sort"
	"fmt"
	"log"

	"database-example/model"
	"database-example/repo"
	"database-example/util"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    //"github.com/andjelavukosav/Docker/common/saga/publish_tour"
)

type TourService struct {
	TourRepo *repo.TourRepository
    PublishTourOrchestrator *PublishTourOrchestrator
}

func NewTourService(tourRepo *repo.TourRepository, orchestrator *PublishTourOrchestrator) *TourService {
	return &TourService{
        TourRepo: tourRepo,
        PublishTourOrchestrator: orchestrator,
    }
}

func (s *TourService) CreateTour(ctx context.Context, authorID, name, description, difficulty string, tags []string) (*model.Tour, error) {
	if name == "" || description == "" || difficulty == "" || len(tags) == 0 {
		return nil, errors.New("tour must have name, description, difficulty and at least one tag as required data.")
	}

	tour := &model.Tour{
		ID:          uuid.New().String(),
		AuthorID:    authorID,
		Name:        name,
		Description: description,
		Difficulty:  difficulty,
		Tags:        tags,
		Status:      model.Draft,
		Price:       0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		KeyPoints:   []model.KeyPoint{}, // <- prazna lista
		Distance: 	 0,
		Durations:   []model.Duration{}, //prazna lista
	}

	err := s.TourRepo.CreateTour(ctx, tour)
	if err != nil {
		return nil, err
	}
	return tour, nil
}

func (s *TourService) GetTour(ctx context.Context, id string) (*model.Tour, error) {
	return s.TourRepo.GetTourByID(ctx, id)
}

func (s *TourService) GetToursByAuthor(ctx context.Context, authorID string) ([]model.Tour, error) {
	return s.TourRepo.GetToursByAuthor(ctx, authorID)
}

func (s *TourService) PublishTour(ctx context.Context, tour *model.Tour) (*model.Tour, error) {
	if err := s.validateForPublish(tour); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot publish tour: %v", err)
	}

	now := time.Now()
	update := map[string]interface{}{
        "status":      model.Published,
        "publishedAt": &now,
    }

	updatedTour, err := s.TourRepo.ChangeTourStatus(ctx, tour.ID, update)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to update tour: %v", err)
    }

    return updatedTour, nil
}

/*func (s *TourService) PublishTour(ctx context.Context, tour *model.Tour) error {
	// Validacija tura
	if err := s.validateForPublish(tour); err != nil {
		return status.Errorf(codes.InvalidArgument, "cannot publish tour: %v", err)
	}

	// Startujemo SAGA workflow, orchestrator će poslati komandu drugim servisima
	tourDetails := publish_tour.TourDetails{
		ID:          tour.ID,
		Name:        tour.Name,
		Description: tour.Description,
		Price:       tour.Price,
	}

	err := s.PublishTourOrchestrator.Start(tourDetails)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to start publish tour saga: %v", err)
	}

	// Status se neće menjati ovde – promena se dešava kada orchestrator primi success reply
	return nil
}
*/

func (s *TourService) ArchiveTour(ctx context.Context, tour *model.Tour) (*model.Tour, error) {

    if tour.Status != model.Published {
		return nil, status.Error(codes.FailedPrecondition, "tour must be published before it can be archived")
	}
    
	now := time.Now()
	update := map[string]interface{}{
		"status":     model.Archived,
		"archivedAt": &now,
	}

	updatedTour, err := s.TourRepo.ChangeTourStatus(ctx, tour.ID, update)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update tour: %v", err)
	}

	return updatedTour, nil
}


func (s *TourService) GetAllTours(ctx context.Context) ([]model.Tour, error) {
	return s.TourRepo.GetAllTours(ctx)
}

func (s *TourService) AddKeyPointToTour(ctx context.Context, tourID string, keyPoint *model.KeyPoint) (*model.Tour, error) {
	tour, err := s.TourRepo.GetTourByID(ctx, tourID)
	if err != nil {
		return nil, err
	}

	if tour == nil {
		return nil, status.Errorf(codes.NotFound, "tour not found: %s", tourID)
	}

	//Ako Order nije postavljen ili je <=0, postavi ga kao posljednji
	if keyPoint.Order <= 0 || keyPoint.Order > int32(len(tour.KeyPoints)) {
		keyPoint.Order = int32(len(tour.KeyPoints) + 1)
		tour.KeyPoints = append(tour.KeyPoints, *keyPoint)
		if err := s.addDistanceIncremental(tour, keyPoint); err != nil {
			return nil, err
		}
	} else{
		// Ubaci na traženo mesto i pomeri ostale
		tour.KeyPoints = adjustKeyPointOrder(tour.KeyPoints, keyPoint.ID, keyPoint.Order, keyPoint)
		s.recalculateDistance(tour)	
	}

	// Sačuvaj izmjene u bazi
	err = s.TourRepo.UpdateTour(ctx, tourID, tour)
	if err != nil {
		return nil, err
	}

	return tour, nil
}

func (s *TourService) addDistanceIncremental(tour *model.Tour, keyPoint *model.KeyPoint) error {
    if len(tour.KeyPoints) > 1 {
        prev := tour.KeyPoints[len(tour.KeyPoints)-2]
        log.Printf("Dodajem na kraj")
        log.Printf("Calling OSRM from (%.6f, %.6f) to (%.6f, %.6f)", 
            prev.Latitude, prev.Longitude, keyPoint.Latitude, keyPoint.Longitude)

        profile := "foot"
        distance, duration, err := util.GetDistanceAndDurationFromOSRM(prev.Longitude, prev.Latitude, keyPoint.Longitude, keyPoint.Latitude, profile)
        if err != nil {
            return err
        }

        tour.Distance += distance
        log.Printf("Distance from OSRM: %.2f km", distance)

		// Update Tour.Durations za walking
        found := false
        for i, d := range tour.Durations {
            if d.Mode == model.Walking {
                tour.Durations[i].Minutes += int(duration)
                found = true
                break
            }
        }
        if !found {
            tour.Durations = append(tour.Durations, model.Duration{
                Mode:    model.Walking,
                Minutes: int(duration),
            })
        }
    }
    return nil
}

func (s *TourService) recalculateDistance(tour *model.Tour) {
    if len(tour.KeyPoints) > 1 {
        totalDistance, walkingMinutes, _ := calculateTourDistanceAndDuration(tour.KeyPoints, "foot")
        tour.Distance = totalDistance

        // Update Tour.Durations za walking
        found := false
        for i, d := range tour.Durations {
            if d.Mode == model.Walking {
                tour.Durations[i].Minutes = walkingMinutes
                found = true
                break
            }
        }
        if !found {
            tour.Durations = append(tour.Durations, model.Duration{
                Mode:    model.Walking,
                Minutes: walkingMinutes,
            })
        }
    }
}



func adjustKeyPointOrder(kps []model.KeyPoint, kpID string, newOrder int32, newKP *model.KeyPoint) []model.KeyPoint {
    // Sortiraj po trenutnom order
    sort.SliceStable(kps, func(i, j int) bool {
        return kps[i].Order < kps[j].Order
    })

    newKps := make([]model.KeyPoint, 0, len(kps))
    var currentKP model.KeyPoint
    if newKP != nil {
        currentKP = *newKP
        currentKP.ID = kpID
    } else {
        currentKP = model.KeyPoint{ID: kpID}
    }

    inserted := false
    order := int32(1)

    for _, kp := range kps {
        if kp.ID == kpID {
            continue // preskoči staru verziju tačke
        }

        // Ubaci novu ili izmenjenu tačku na traženi order
        if !inserted && order == newOrder {
            currentKP.Order = order
            newKps = append(newKps, currentKP)
            order++
            inserted = true
        }

        kp.Order = order
        newKps = append(newKps, kp)
        order++
    }

    // Ako novi order ide na kraj
    if !inserted {
        currentKP.Order = order
        newKps = append(newKps, currentKP)
    }

    return newKps
}

func (s *TourService) UpdateKeyPoint(tourId string, updatedKP *model.KeyPoint) (*model.KeyPoint, error) {
	// Ucitaj turu 
	tour, err := s.TourRepo.GetTourByID(context.Background(), tourId)
    if err != nil {
        return nil, fmt.Errorf("failed to get tour: %w", err)
    }
    if tour == nil {
        return nil, fmt.Errorf("tour not found: %s", tourId)
    }

	// Nadji key point u listi
    var found bool
	var oldOrder int32
	var oldLat float64
	var oldLon float64

    for i, kp := range tour.KeyPoints {
        if kp.ID == updatedKP.ID {
			//provjeri da li je promijenjen order
			oldOrder = kp.Order
			oldLat = kp.Latitude
			oldLon = kp.Longitude
            // 3. Azuriraj polja
            tour.KeyPoints[i].Name = updatedKP.Name
            tour.KeyPoints[i].Description = updatedKP.Description
            tour.KeyPoints[i].Latitude = updatedKP.Latitude
            tour.KeyPoints[i].Longitude = updatedKP.Longitude
            tour.KeyPoints[i].ImageURL = updatedKP.ImageURL
            tour.KeyPoints[i].Order = updatedKP.Order
            found = true
            break
        }
    }

    if !found {
        return nil, fmt.Errorf("key point not found: %s", updatedKP.ID)
    }

	// 3. Reorder samo ako se order promenio
	if updatedKP.Order != oldOrder {
		tour.KeyPoints = adjustKeyPointOrder(tour.KeyPoints, updatedKP.ID, updatedKP.Order, updatedKP)
	}

	// Provera promena u koordinatama
	if updatedKP.Latitude != oldLat || updatedKP.Longitude != oldLon || updatedKP.Order != oldOrder {
		s.recalculateDistance(tour)
	}

	// 4. Sacuvaj izmenjenu turu u repo
    if err := s.TourRepo.UpdateTour(context.Background(), tourId, tour); err != nil {
        return nil, fmt.Errorf("failed to update tour: %w", err)
    }

    // 5. Vrati azurirani key point
    return updatedKP, nil
}

func (s *TourService) DeleteKeyPoint(ctx context.Context, tourId string, keyPointId string) error {
	tour, err := s.TourRepo.GetTourByID(ctx, tourId)
	if err != nil {
		return fmt.Errorf("failed to get tour: %w", err)
	}
	if tour == nil {
		return fmt.Errorf("tour not found: %s", tourId)
	}
 	newKeyPoints := make([]model.KeyPoint, 0)
    order := int32(1)

    // Sortiraj i ukloni
    sort.SliceStable(tour.KeyPoints, func(i, j int) bool {
        return tour.KeyPoints[i].Order < tour.KeyPoints[j].Order
    })

    for _, kp := range tour.KeyPoints {
        if kp.ID == keyPointId {
            continue
        }
        kp.Order = order
        newKeyPoints = append(newKeyPoints, kp)
        order++
    }

    tour.KeyPoints = newKeyPoints

		
	if len(tour.KeyPoints) > 1 {
		s.recalculateDistance(tour) // distance + walking duration
	} else {
		tour.Distance = 0
		// Ako nema više keyPoints, resetuj duration za walking
		for i, d := range tour.Durations {
			if d.Mode == model.Walking {
				tour.Durations[i].Minutes = 0
			}
		}
	}

	return s.TourRepo.UpdateTour(ctx, tourId, tour)
}

func (s *TourService) validateForPublish(tour *model.Tour) error {
	if tour.Name == "" || tour.Description == "" || tour.Difficulty == "" || len(tour.Tags) == 0 {
		return errors.New("tour must have name, description, difficulty and tags before publishing")
	}
	if len(tour.KeyPoints) < 2 {
		return errors.New("tour must have at least two key points before publishing")
	}
	if len(tour.Durations) == 0 {
		return errors.New("tour must have at least one duration before publishing")
	}
	return nil
}

func calculateTourDistanceAndDuration(keyPoints []model.KeyPoint, profile string) (float64, int, error) {
    if len(keyPoints) < 2 {
        return 0, 0, nil
    }

    totalDistance := 0.0
    totalDuration := 0

    for i := 0; i < len(keyPoints)-1; i++ {
        kp1 := keyPoints[i]
        kp2 := keyPoints[i+1]

        distance, duration, err := util.GetDistanceAndDurationFromOSRM(
            kp1.Longitude, kp1.Latitude, kp2.Longitude, kp2.Latitude, profile)
        if err != nil {
            fmt.Printf("error calculating route between %s and %s: %v\n", kp1.ID, kp2.ID, err)
            continue
        }

        totalDistance += distance
        totalDuration += int(duration)
    }

    return totalDistance, totalDuration, nil
}




