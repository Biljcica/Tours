package service

import (
	"errors"
	"time"

	"database-example/model"
	"database-example/repo"
	"github.com/google/uuid"
)

type TourService struct {
	TourRepo *repo.TourRepository
}

func (s *TourService) CreateTour(authorID string, name string, description string, difficulty string, tags []string) (*model.Tour, error) {
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
		Status:      "draft",   // podrazumevano
		Price:       0,         // podrazumevano
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.TourRepo.CreateTour(tour)
	if err != nil {
		return nil, err
	}
	return tour, nil
}

// GetTour vraća turu po ID
func (s *TourService) GetTour(id string) (*model.Tour, error) {
	return s.TourRepo.GetTourByID(id)
}

// GetToursByAuthor vraća sve ture određenog autora
func (s *TourService) GetToursByAuthor(authorID string) ([]model.Tour, error) {
	return s.TourRepo.GetToursByAuthor(authorID)
}

// PublishTour objavljuje turu (menja status iz draft u published)
func (s *TourService) PublishTour(id string) error {
	return s.TourRepo.UpdateTourStatus(id, "published")
}

// GetAllTours vraća sve ture iz baze
func (s *TourService) GetAllTours() ([]model.Tour, error) {
	return s.TourRepo.GetAllTours()
}
