package service

import (
	"context"
	"time"

	"database-example/model"
	"database-example/repo"
)

type PositionService struct {
	repo *repo.PositionRepository
}

func NewPositionService(r *repo.PositionRepository) *PositionService {
	return &PositionService{repo: r}
}

// Cuva novu poziciju ili update postojeceg reda
func (s *PositionService) RecordPosition(ctx context.Context, touristID string, lat, lon float64) (*model.Position, error) {
	pos := &model.Position{
		TouristID: touristID,
		Latitude: lat,
		Longitude: lon,
		UpdatedAt: time.Now(),
	}

	err := s.repo.SaveOrUpdate(ctx, pos)
	if err != nil {
		return nil, err
	}

	return pos, nil
}

// Vraca trenutnu pozciju turiste
func (s *PositionService) GetCurrentPosition(ctx context.Context, touristID string) (*model.Position, error){
	return s.repo.GetPositionByTouristId(ctx, touristID)
}
