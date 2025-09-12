package handlers

import (
	"context"
	"database-example/model"
	pb "database-example/proto/tours"
	"database-example/service"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
// Pomocna funkcija: mapiranje model.Tour -> pb.TourResponse
// ------------------------------------------------
func mapTourToPb(t *model.Tour) *pb.TourResponse {
	var pbKeyPoints []*pb.KeyPoint
	for _, kp := range t.KeyPoints {
		pbKeyPoints = append(pbKeyPoints, &pb.KeyPoint{
			Id:          kp.ID,
			Name:        kp.Name,
			Description: kp.Description,
			Latitude:    kp.Latitude,
			Longitude:   kp.Longitude,
			ImageURL:    kp.ImageURL,
		})
	}

	return &pb.TourResponse{
		Id:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		Difficulty:  t.Difficulty,
		Tags:        t.Tags,
		Price:       t.Price,
		Status:      t.Status,
		AuthorId:    t.AuthorID,
		KeyPoints:   pbKeyPoints,
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

	return mapTourToPb(tour), nil
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
		pbTours = append(pbTours, mapTourToPb(&t))
	}

	return &pb.GetAuthorToursResponse{
		Tours: pbTours,
	}, nil
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
	}

	_, err := h.TourService.AddKeyPointToTour(ctx, req.TourId, keyPoint)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add key point to tour: %v", err)
	}

	return &pb.KeyPointResponse{
		Id:          keyPoint.ID,
		Name:        keyPoint.Name,
		Description: keyPoint.Description,
		Latitude:    keyPoint.Latitude,
		Longitude:   keyPoint.Longitude,
		ImageURL:    keyPoint.ImageURL,
	}, nil
}
