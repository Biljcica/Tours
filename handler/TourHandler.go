package handler

import (
	"context"
	"database-example/model"
	pb "database-example/proto/tours"
	"database-example/service"
	"fmt"
	"strings"

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
			Order:       kp.Order,
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
	resp := mapTourToPb(tour)

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

	// 🟢 Debug print: šta vraća GetAllTours
	fmt.Println("=== DEBUG: Sve ture iz baze ===")
	for i, t := range tours {
		fmt.Printf("%d) Tour ID: %s, Name: %s, Status: '%s', KeyPoints len: %d\n",
			i+1, t.ID, t.Name, t.Status, len(t.KeyPoints))
	}
	fmt.Println("=== KRAJ DEBUG ===")

	var pbTours []*pb.PublishedTour
	for _, t := range tours {
		if strings.ToLower(strings.TrimSpace(t.Status)) != "publish" {
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
			Length:      10.0,
			StartTime:   "2025-01-01T09:00:00Z",
			KeyPoints:   pbKeyPoints,
		})
	}

	fmt.Printf("=== DEBUG: Filtered published tours count: %d ===\n", len(pbTours))

	return &pb.GetPublishedToursResponse{
		Tours: pbTours,
	}, nil
}
