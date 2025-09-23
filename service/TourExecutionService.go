package service

import (
	"context"
	"time"

	"database-example/model"
	"database-example/repo"

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
func (s *TourExecutionService) UpdatePosition(ctx context.Context, execID string, pos model.Position) error {
	return s.ExecutionRepo.UpdateLastActivityTime(ctx, execID, &pos)
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
