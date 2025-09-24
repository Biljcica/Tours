package handler

import (
	"context"
	"fmt"

	"database-example/model"
	"database-example/mapper"
	pb "database-example/proto/tours"
	"database-example/service"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ToursHandler struct {
	pb.UnimplementedToursServiceServer
	TourService *service.TourService
}

func NewToursHandler(toursService *service.TourService) *ToursHandler {
	return &ToursHandler{
		TourService: toursService,
	}
}

// ------------------------------------------------
// Kreiranje ture
// ------------------------------------------------
func (h *ToursHandler) CreateTour(ctx context.Context, req *pb.CreateTourRequest) (*pb.TourResponse, error) {
	tour, err := h.TourService.CreateTour(ctx, req.AuthorId, req.Name, req.Description, req.Difficulty, req.Tags)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create tour: %v", err)
	}

	return mapper.MapTourToPb(tour), nil
}

// ------------------------------------------------
// Vraća sve ture autora
// ------------------------------------------------
func (h *ToursHandler) GetAuthorTours(ctx context.Context, req *pb.GetAuthorToursRequest) (*pb.GetAuthorToursResponse, error) {
	tours, err := h.TourService.GetToursByAuthor(ctx, req.AuthorId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get tours: %v", err)
	}

	var pbTours []*pb.TourResponse
	for _, t := range tours {
		pbTours = append(pbTours, mapper.MapTourToPb(&t))
	}

	return &pb.GetAuthorToursResponse{
		Tours: pbTours,
	}, nil
}

// Dobavi turu po id
func (h *ToursHandler) GetTourById(ctx context.Context, req *pb.GetTourByIdRequest) (*pb.TourResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "tour id is required")
	}

	// Pozovi service
	tour, err := h.TourService.GetTour(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get tour: %v", err)
	}

	// Pretvori model.Tour u pb.TourResponse
	resp := mapper.MapTourToPb(tour)

	return resp, nil
}

// ------------------------------------------------
// Dodavanje ključne tačke
// ------------------------------------------------
func (h *ToursHandler) AddKeyPoint(ctx context.Context, req *pb.AddKeyPointRequest) (*pb.KeyPointResponse, error) {
	keyPoint := &model.KeyPoint{
		ID:          uuid.New().String(),
		Name:        req.Point.Name,
		Description: req.Point.Description,
		Latitude:    req.Point.Latitude,
		Longitude:   req.Point.Longitude,
		ImageURL:    req.Point.ImageURL,
		Order:       req.Point.Order,
	}

	tour, err := h.TourService.AddKeyPointToTour(ctx, req.TourId, keyPoint)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add key point to tour: %v", err)
	}

	// pronađi novu KeyPoint u addedKP.KeyPoints
	var returnedKP *model.KeyPoint
	for _, kp := range tour.KeyPoints {
		if kp.ID == keyPoint.ID {
			returnedKP = &kp
			break
		}
	}

	return &pb.KeyPointResponse{
		Id:          returnedKP.ID,
		Name:        returnedKP.Name,
		Description: returnedKP.Description,
		Latitude:    returnedKP.Latitude,
		Longitude:   returnedKP.Longitude,
		ImageURL:    returnedKP.ImageURL,
		Order:       returnedKP.Order,
	}, nil

}

// Update kljucne tacke
func (h *ToursHandler) UpdateKeyPoint(ctx context.Context, req *pb.UpdateKeyPointRequest) (*pb.KeyPointResponse, error) {
	kp := req.KeyPoint
	tourId := req.TourId

	updatedKP := &model.KeyPoint{
		ID:          kp.Id,
		Name:        kp.Name,
		Description: kp.Description,
		Latitude:    kp.Latitude,
		Longitude:   kp.Longitude,
		ImageURL:    kp.ImageURL,
		Order:       kp.Order,
	}

	res, err := h.TourService.UpdateKeyPoint(tourId, updatedKP)
	if err != nil {
		return nil, err
	}

	// 3. Vrati KeyPointResponse
	return &pb.KeyPointResponse{
		Id:          res.ID,
		Name:        res.Name,
		Description: res.Description,
		Latitude:    res.Latitude,
		Longitude:   res.Longitude,
		ImageURL:    res.ImageURL,
		Order:       res.Order,
	}, nil
}

// Brisanje kljucne tacke
func (h *ToursHandler) DeleteKeyPoint(ctx context.Context, req *pb.DeleteKeyPointRequest) (*pb.DeleteKeyPointResponse, error) {
	if req.TourId == "" || req.KeypointId == "" {
		return &pb.DeleteKeyPointResponse{Success: false}, fmt.Errorf("tourId and keypointId are required")
	}

	// Pozovi servis
	err := h.TourService.DeleteKeyPoint(ctx, req.TourId, req.KeypointId)
	if err != nil {
		return &pb.DeleteKeyPointResponse{Success: false}, err
	}

	return &pb.DeleteKeyPointResponse{Success: true}, nil
}

func (h *ToursHandler) GetPublishedTours(ctx context.Context, req *pb.GetPublishedToursRequest) (*pb.GetPublishedToursResponse, error) {
	tours, err := h.TourService.GetAllTours(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get tours: %v", err)
	}

	fmt.Println("=== DEBUG: Sve ture iz baze ===")
	for i, t := range tours {
		fmt.Printf("%d) Tour ID: %s, Name: %s, Status: '%s', KeyPoints len: %d\n",
			i+1, t.ID, t.Name, t.Status, len(t.KeyPoints))
	}
	fmt.Println("=== KRAJ DEBUG ===")

	var pbTours []*pb.PublishedTour
	
	for _, t := range tours {
		
		if t.Status != model.Published {
			continue
		}

		var pbKeyPoints []*pb.KeyPoint
		for _, kp := range t.KeyPoints {
			pbKeyPoints = append(pbKeyPoints, &pb.KeyPoint{
				Id:          kp.ID,
				Name:        kp.Name,
				Description: kp.Description,
				Latitude:    kp.Latitude,
				Longitude:   kp.Longitude,
				ImageURL:    kp.ImageURL,
				Order:       kp.Order,
			})
		}

		pbTours = append(pbTours, &pb.PublishedTour{
			Id:          t.ID,
			Name:        t.Name,
			Price:       t.Price,
			Description: t.Description,
			Length:      t.Distance,
			StartTime:   "2025-01-01T09:00:00Z",
			KeyPoints:   pbKeyPoints,
		})
	}

	fmt.Printf("=== DEBUG: Filtered published tours count: %d ===\n", len(pbTours))

	return &pb.GetPublishedToursResponse{
		Tours: pbTours,
	}, nil
}

func (h *ToursHandler) GetAllTours(ctx context.Context, req *pb.GetAllToursRequest) (*pb.GetAllToursResponse, error) {
	tours, err := h.TourService.GetAllTours(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get tours: %v", err)
	}

	fmt.Println("=== DEBUG: Sve ture iz baze ===")
	for i, t := range tours {
		fmt.Printf("%d) Tour ID: %s, Name: %s, Status: '%s', KeyPoints len: %d\n",
			i+1, t.ID, t.Name, t.Status, len(t.KeyPoints))
	}
	fmt.Println("=== KRAJ DEBUG ===")

	var pbTours []*pb.PublishedTour
	for _, t := range tours {

		var pbKeyPoints []*pb.KeyPoint
		for _, kp := range t.KeyPoints {
			pbKeyPoints = append(pbKeyPoints, &pb.KeyPoint{
				Id:          kp.ID,
				Name:        kp.Name,
				Description: kp.Description,
				Latitude:    kp.Latitude,
				Longitude:   kp.Longitude,
				ImageURL:    kp.ImageURL,
				Order:       kp.Order,
			})
		}

		pbTours = append(pbTours, &pb.PublishedTour{
			Id:          t.ID,
			Name:        t.Name,
			Price:       t.Price,
			Description: t.Description,
			Length:      10.0,
			StartTime:   "2025-01-01T09:00:00Z",
			KeyPoints:   pbKeyPoints,
		})
	}

	fmt.Printf("=== DEBUG: Filtered published tours count: %d ===\n", len(pbTours))
	return &pb.GetAllToursResponse{
		Tours: pbTours,
	}, nil
}

/*
func (h *ToursHandler) GetAllTours(ctx context.Context, req *pb.GetAllToursRequest) (*pb.GetAllToursResponse, error) {
	tours, err := h.TourService.GetAllTours(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get all tours: %v", err)
	}

	var pbTours []*pb.TourResponse
	for _, t := range tours {
		pbTours = append(pbTours, mapTourToPb(&t))
	}

}*/


func (h *ToursHandler) UpdateTourStatus(ctx context.Context, req *pb.UpdateTourStatusRequest) (*pb.UpdateTourStatusResponse, error) {
    if req.TourId == "" || req.AuthorId == "" {
        return nil, status.Error(codes.InvalidArgument, "tourId and authorId are required")
    }
	if req.NewStatus == pb.TourStatus_TOUR_STATUS_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "newStatus is required")
	}

    tour, err := h.TourService.GetTour(ctx, req.TourId)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "tour not found: %v", err)
    }

    // Autorizacija
    if tour.AuthorID != req.AuthorId {
        return nil, status.Error(codes.PermissionDenied, "not authorized to publish this tour")
    }

	var updatedTour *model.Tour
	switch req.NewStatus {
	case pb.TourStatus_PUBLISHED:
		updatedTour, err = h.TourService.PublishTour(ctx, tour)
	case pb.TourStatus_ARCHIVED:
		updatedTour, err = h.TourService.ArchiveTour(ctx, tour)
	default:
		return nil, status.Error(codes.InvalidArgument, "unsupported status")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update tour status: %v", err)
	}

	var updatedAt *timestamppb.Timestamp
	switch req.NewStatus {
	case pb.TourStatus_PUBLISHED:
		if updatedTour.PublishedAt != nil {
			updatedAt = timestamppb.New(*updatedTour.PublishedAt)
		}
	case pb.TourStatus_ARCHIVED:
		if updatedTour.ArchivedAt != nil {
			updatedAt = timestamppb.New(*updatedTour.ArchivedAt)
		}
	}

    return &pb.UpdateTourStatusResponse{
        TourId:      updatedTour.ID,
        Status:      req.NewStatus,
		UpdatedAt:   updatedAt,
	}, nil
}

/*func (h *ToursHandler) UpdateTourStatus(ctx context.Context, req *pb.UpdateTourStatusRequest) (*pb.UpdateTourStatusResponse, error) {
    if req.TourId == "" || req.AuthorId == "" {
        return nil, status.Error(codes.InvalidArgument, "tourId and authorId are required")
    }
    if req.NewStatus == pb.TourStatus_TOUR_STATUS_UNSPECIFIED {
        return nil, status.Error(codes.InvalidArgument, "newStatus is required")
    }

    tour, err := h.TourService.GetTour(ctx, req.TourId)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "tour not found: %v", err)
    }

    if tour.AuthorID != req.AuthorId {
        return nil, status.Error(codes.PermissionDenied, "not authorized to update this tour")
    }

    switch req.NewStatus {
    case pb.TourStatus_PUBLISHED:
        // Startuje SAGA workflow, asinhrono
        err = h.TourService.PublishTour(ctx, tour)
        if err != nil {
            return nil, status.Errorf(codes.Internal, "failed to start publish tour workflow: %v", err)
        }

        // Vraća inicijalni odgovor jer status još nije ažuriran u bazi
        response := &pb.UpdateTourStatusResponse{
            TourId: tour.ID,
            Status: pb.TourStatus_PENDING_PUBLISH, // ili custom "PENDING_PUBLISH"
            UpdatedAt: nil,               // još nema timestamp
        }
		log.Printf("[UpdateTourStatus] Returning response: %+v", response)
        return response, nil

    case pb.TourStatus_ARCHIVED:
        updatedTour, err := h.TourService.ArchiveTour(ctx, tour)
        if err != nil {
            return nil, status.Errorf(codes.Internal, "failed to archive tour: %v", err)
        }
        return &pb.UpdateTourStatusResponse{
            TourId:    updatedTour.ID,
            Status:    pb.TourStatus_ARCHIVED,
            UpdatedAt: timestamppb.New(*updatedTour.ArchivedAt),
        }, nil

    default:
        return nil, status.Error(codes.InvalidArgument, "unsupported status")
    }
}*/

