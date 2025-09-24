package service

import (
	"context"
	"database-example/model"
	"database-example/repo"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo" // <-- dodaj ovo

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TourExecutionService struct {
	ExecutionRepo *repo.TourExecutionRepository
	TourRepo      *repo.TourRepository
}

func NewTourExecutionService(repo *repo.TourExecutionRepository) *TourExecutionService {
	return &TourExecutionService{
		ExecutionRepo: repo,
	}
}

// StartTour kreira novu sesiju ture za korisnika
func (s *TourExecutionService) StartTour(ctx context.Context, tourID, userID string) (*model.TourExecution, error) {
	if tourID == "" || userID == "" {
		return nil, status.Error(codes.InvalidArgument, "tourID i userID su obavezni")
	}

	// Kreiranje TourExecution objekta
	exec := &model.TourExecution{
		ID:                 uuid.New().String(),
		TourID:             tourID,
		TouristID:          userID,
		Status:             model.StatusActive,
		StartTime:          time.Now(),
		LastActivityTime:   time.Now(),
		CompletedKeyPoints: []model.CompletedKeyPoint{},
	}

	// Sačuvaj u bazi
	if err := s.ExecutionRepo.CreateTourExecution(ctx, exec); err != nil {
		return nil, status.Errorf(codes.Internal, "greska prilikom kreiranja sesije: %v", err)
	}

	return exec, nil
}

// UpdatePosition ažurira trenutnu poziciju i lastActivityTime
func (s *TourExecutionService) UpdatePosition(ctx context.Context, execID string) error {
	return s.ExecutionRepo.UpdateLastActivityTime(ctx, execID)
}

// CompleteTour završava turu
func (s *TourExecutionService) CompleteTour(ctx context.Context, execID string) error {
	return s.ExecutionRepo.CompleteTour(ctx, execID)
}

// AbandonTour napušta turu
func (s *TourExecutionService) AbandonTour(ctx context.Context, execID string, userID string) error {
	return s.ExecutionRepo.AbandonTour(ctx, execID, userID)
}

// AddCompletedKeyPoint beleži završenu ključnu tačku
func (s *TourExecutionService) AddCompletedKeyPoint(ctx context.Context, execID string, keyPointID string) error {
	ckp := model.CompletedKeyPoint{
		KeyPointID:  keyPointID,
		CompletedAt: time.Now(),
	}
	return s.ExecutionRepo.AddCompletedKeyPoint(ctx, execID, ckp)
}

// NotifyNearKeyPoint beleži da je korisnik blizu ključne tačke
func (s *TourExecutionService) NotifyNearKeyPoint(ctx context.Context, execID, userID, keyPointID string) error {
	// 1️⃣ Dohvati TourExecution iz baze
	exec, err := s.ExecutionRepo.GetTourExecutionByID(ctx, execID)
	if err != nil {
		return err
	}

	if exec.TouristID != userID {
		return fmt.Errorf("user %s not authorized for tour execution %s", userID, execID)
	}

	// 2️⃣ Proveri da li je ključna tačka već završena
	for _, kp := range exec.CompletedKeyPoints {
		if kp.KeyPointID == keyPointID {
			// Već završena, samo update lastActivityTime
			return s.ExecutionRepo.UpdateLastActivityTime(ctx, execID)
		}
	}

	// 3️⃣ Dodaj novu završenu ključnu tačku
	return s.ExecutionRepo.AddCompletedKeyPoint(ctx, execID, model.CompletedKeyPoint{
		KeyPointID:  keyPointID,
		CompletedAt: time.Now(),
	})
}

func (s *TourExecutionService) GetActiveTour(ctx context.Context, touristID string) (*model.TourExecution, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := map[string]interface{}{
		"touristId": touristID,
		"status":    model.StatusActive,
	}

	var exec model.TourExecution
	err := s.ExecutionRepo.Collection.FindOne(ctx, filter).Decode(&exec)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // nema aktivne ture
		}
		return nil, err
	}

	return &exec, nil
}

func (s *TourExecutionService) HasTourExecution(ctx context.Context, userID, tourID string) (bool, error) {
	exec, err := s.ExecutionRepo.GetByUserAndTour(ctx, userID, tourID)
	if err != nil && err != mongo.ErrNoDocuments {
		return false, err
	}
	return exec != nil, nil
}

// GetTourExecutionByID vraća samo TourExecution objekat po execID
func (s *TourExecutionService) GetTourExecutionByID(ctx context.Context, execID string) (*model.TourExecution, error) {
	if execID == "" {
		return nil, status.Error(codes.InvalidArgument, "executionId je obavezan")
	}

	// Dohvati TourExecution iz baze
	exec, err := s.ExecutionRepo.GetTourExecutionByID(ctx, execID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Error(codes.NotFound, "tour execution not found")
		}
		return nil, status.Errorf(codes.Internal, "greska prilikom dohvatanja tour execution: %v", err)
	}

	return exec, nil
}

// AbandonTour napušta turu
func (s *TourExecutionService) UpdateTourExecution(ctx context.Context, execID string) error {
	return s.ExecutionRepo.UpdateTourExecution(ctx, execID)
}
