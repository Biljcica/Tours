package handlers

import (
	"context"
	pb "database-example/proto/tours"
	"database-example/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ToursHandler struct {
	pb.UnimplementedToursServiceServer
	TourService *service.TourService
}

func NewToursHandler(toursService *service.TourService) *ToursHandler {
	return &ToursHandler{TourService: toursService}
}

// Kreiranje ture
func (h *ToursHandler) CreateTour(ctx context.Context, req *pb.CreateTourRequest) (*pb.TourResponse, error) {
	tour, err := h.TourService.CreateTour(ctx, req.AuthorId, req.Name, req.Description, req.Difficulty, req.Tags)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create tour: %v", err)
	}

	return &pb.TourResponse{
		Id:          tour.ID,
		Name:        tour.Name,
		Description: tour.Description,
		Difficulty:  tour.Difficulty,
		Tags:        tour.Tags,
		Price:       tour.Price,
		Status:      tour.Status,
	}, nil
}

// Vraća sve ture
func (h *ToursHandler) GetAuthorTours(ctx context.Context, req *pb.GetAuthorToursRequest) (*pb.GetAuthorToursResponse, error) {
	tours, err := h.TourService.GetToursByAuthor(ctx, req.AuthorId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get tours: %v", err)
	}

	var pbTours []*pb.TourResponse
	for _, t := range tours {
		pbTours = append(pbTours, &pb.TourResponse{
			Id:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			Difficulty:  t.Difficulty,
			Tags:        t.Tags,
			Price:       t.Price,
			Status:      t.Status,
		})
	}

	return &pb.GetAuthorToursResponse{
		Tours: pbTours,
	}, nil
}

// Vraća turu po ID-u
/*func (h *ToursHandler) GetTour(ctx context.Context, req *pb.GetTourRequest) (*pb.TourResponse, error) {
	tour, err := h.TourService.GetTour(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "tour not found: %v", err)
	}

	return &pb.TourResponse{
		Id:          tour.ID,
		Name:        tour.Name,
		Description: tour.Description,
		Difficulty:  tour.Difficulty,
		Tags:        tour.Tags,
		Price:       tour.Price,
		Status:      tour.Status,
	}, nil
}

// Vraća sve ture određenog autora
func (h *ToursHandler) GetToursByAuthor(ctx context.Context, req *pb.GetToursByAuthorRequest) (*pb.GetToursResponse, error) {
	tours, err := h.TourService.GetToursByAuthor(ctx, req.AuthorId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get tours by author: %v", err)
	}

	var pbTours []*pb.TourResponse
	for _, t := range tours {
		pbTours = append(pbTours, &pb.TourResponse{
			Id:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			Difficulty:  t.Difficulty,
			Tags:        t.Tags,
			Price:       t.Price,
			Status:      t.Status,
		})
	}

	return &pb.GetToursResponse{
		Tours: pbTours,
	}, nil
}

// Publikuje turu
func (h *ToursHandler) PublishTour(ctx context.Context, req *pb.PublishTourRequest) (*pb.PublishTourResponse, error) {
	err := h.TourService.PublishTour(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish tour: %v", err)
	}

	return &pb.PublishTourResponse{
		Message: "Tour published successfully",
	}, nil
}
*/
