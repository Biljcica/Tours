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
	TourService          *service.TourService
	TourExecutionService *service.TourExecutionService
}

func NewToursHandler(tourService *service.TourService, tourExecutionService *service.TourExecutionService) *ToursHandler {
	return &ToursHandler{
		TourService:          tourService,
		TourExecutionService: tourExecutionService,
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

func (h *ToursHandler) GetAllTours(ctx context.Context, req *pb.GetAllToursRequest) (*pb.GetAllToursResponse, error) {
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

func (h *ToursHandler) StartTour(ctx context.Context, req *pb.StartTourRequest) (*pb.StartTourResponse, error) {
	if req.TourId == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "tourId i userId su obavezni")
	}

	// Kreiraj novu TourExecution sesiju preko TourExecutionService
	exec, err := h.TourExecutionService.StartTour(ctx, req.TourId, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start tour: %v", err)
	}

	// Vraćamo samo ID sesije i status
	return &pb.StartTourResponse{
		TourExecutionId: exec.ID,
		Status:          string(exec.Status), // ACTIVE
	}, nil
}

func (h *ToursHandler) LeaveTour(ctx context.Context, req *pb.LeaveTourRequest) (*pb.LeaveTourResponse, error) {
	if req.TourExecutionId == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "tourExecutionId i userId su obavezni")
	}

	// ovde pozovi servis koji menja status u bazi
	err := h.TourExecutionService.AbandonTour(ctx, req.TourExecutionId, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "neuspjelo napuštanje ture: %v", err)
	}

	return &pb.LeaveTourResponse{
		Status: "ABANDONED",
	}, nil
}

func (h *ToursHandler) NotifyNearKeyPoint(ctx context.Context, req *pb.NotifyNearKeyPointRequest) (*pb.NotifyNearKeyPointResponse, error) {
	if req.ExecutionId == "" || req.KeyPointId == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "executionId, keyPointId i userId su obavezni")
	}

	err := h.TourExecutionService.NotifyNearKeyPoint(ctx, req.ExecutionId, req.UserId, req.KeyPointId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to notify key point: %v", err)
	}

	return &pb.NotifyNearKeyPointResponse{
		Success: true,
		Message: "Key point completion recorded",
	}, nil
}

func (h *ToursHandler) GetActiveTour(ctx context.Context, req *pb.GetActiveTourRequest) (*pb.GetActiveTourResponse, error) {
	exec, err := h.TourExecutionService.GetActiveTour(ctx, req.TouristId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch active tour: %v", err)
	}

	if exec == nil {
		return &pb.GetActiveTourResponse{}, nil
	}

	return &pb.GetActiveTourResponse{
		ExecutionId: exec.ID,
		TourId:      exec.TourID,
		Status:      string(exec.Status),
	}, nil
}

func (h *ToursHandler) HasTourExecution(ctx context.Context, req *pb.HasTourExecutionRequest) (*pb.HasTourExecutionResponse, error) {
	// Validacija ulaznih parametara
	if req.UserId == "" || req.TourId == "" {
		return nil, status.Error(codes.InvalidArgument, "userId i tourId su obavezni")
	}

	// Pozivanje servisa koji proverava da li korisnik ima aktivnu turu
	hasExec, err := h.TourExecutionService.HasTourExecution(ctx, req.UserId, req.TourId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check tour execution: %v", err)
	}

	// Vraćanje rezultata
	return &pb.HasTourExecutionResponse{
		HasExecution: hasExec,
	}, nil
}

func (h *ToursHandler) CheckTourCompletion(ctx context.Context, req *pb.CheckTourCompletionRequest) (*pb.CheckTourCompletionResponse, error) {
	// 1️⃣ Validacija ulaznih parametara
	if req.ExecutionId == "" {
		return nil, status.Error(codes.InvalidArgument, "executionId je obavezan")
	}

	// 2️⃣ Pribavi izvršenje ture
	exec, err := h.TourExecutionService.ExecutionRepo.GetTourExecutionByID(ctx, req.ExecutionId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch tour execution: %v", err)
	}
	if exec == nil {
		return nil, status.Error(codes.NotFound, "tour execution not found")
	}

	// 3️⃣ Pribavi celu turu sa ključnim tačkama
	tour, err := h.TourService.GetTour(ctx, exec.TourID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch tour: %v", err)
	}
	if tour == nil || len(tour.KeyPoints) == 0 {
		return &pb.CheckTourCompletionResponse{Completed: false}, nil
	}

	// 4️⃣ Provera da li su sve ključne tačke završene
	allVisited := true
	for _, kp := range tour.KeyPoints {
		found := false
		for _, ckp := range exec.CompletedKeyPoints {
			if ckp.KeyPointID == kp.ID {
				found = true
				break
			}
		}
		if !found {
			allVisited = false
			break
		}
	}

	// 5️⃣ Ako jesu, promeni status ture na COMPLETED
	if allVisited && exec.Status != model.StatusCompleted {
		if err := h.TourExecutionService.CompleteTour(ctx, exec.ID); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update tour status: %v", err)
		}
	}

	// 6️⃣ Vraćanje odgovora frontendu
	return &pb.CheckTourCompletionResponse{
		Completed: allVisited,
	}, nil
}
