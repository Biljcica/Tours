package service

import (
	"context"
	"errors"
	"time"

	"database-example/model"
	"database-example/repo"

	"github.com/google/uuid"
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

	// Dodaj key point u listu
	tour.KeyPoints = append(tour.KeyPoints, *keyPoint)

	// Sačuvaj izmene u bazi
	err = s.TourRepo.UpdateTour(ctx, tourID, tour)
	if err != nil {
		return nil, err
	}

	return tour, nil
}
