package service

import (
	"context"
	"errors"
	"time"
	"sort"
	"fmt"

	"database-example/model"
	"database-example/repo"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type TourService struct {
	TourRepo *repo.TourRepository
}

func NewTourService(tourRepo *repo.TourRepository) *TourService {
	return &TourService{TourRepo: tourRepo}
}

func (s *TourService) CreateTour(ctx context.Context, authorID, name, description, difficulty string, tags []string) (*model.Tour, error) {
	if name == "" {
		return nil, errors.New("tour name is required")
	}

	tour := &model.Tour{
		ID:          uuid.New().String(),
		AuthorID:    authorID,
		Name:        name,
		Description: description,
		Difficulty:  difficulty,
		Tags:        tags,
		Status:      "draft",
		Price:       0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		KeyPoints:   []model.KeyPoint{}, // <- prazna lista

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

func (s *TourService) PublishTour(ctx context.Context, id string) error {
	return s.TourRepo.UpdateTourStatus(ctx, id, "published")
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
	} else{
		// Ubaci na traženo mesto i pomeri ostale
		tour.KeyPoints = adjustKeyPointOrder(tour.KeyPoints, keyPoint.ID, keyPoint.Order, keyPoint)		
	}

	// Sačuvaj izmene u bazi
	err = s.TourRepo.UpdateTour(ctx, tourID, tour)
	if err != nil {
		return nil, err
	}

	return tour, nil
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

    for i, kp := range tour.KeyPoints {
        if kp.ID == updatedKP.ID {
			//provjeri da li je promijenjen order
			oldOrder = kp.Order
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

	// 4. Sacuvaj izmenjenu turu u repo
    if err := s.TourRepo.UpdateTour(context.Background(), tourId, tour); err != nil {
        return nil, fmt.Errorf("failed to update tour: %w", err)
    }

    // 5. Vrati azurirani key point
    return updatedKP, nil
}

func (s *TourService) DeleteKeyPoint(ctx context.Context, tourId string, keyPointId string) error {
	return s.TourRepo.DeleteKeyPoint(ctx, tourId, keyPointId)
}
