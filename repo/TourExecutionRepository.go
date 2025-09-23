package repo

import (
	"context"
	"database-example/model"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// Novi repozitorijum za TourExecution
type TourExecutionRepository struct {
	Collection *mongo.Collection
}

func NewTourExecutionRepository(collection *mongo.Collection) *TourExecutionRepository {
	return &TourExecutionRepository{Collection: collection}
}

// CreateTourExecution kreira novu sesiju ture
func (r *TourExecutionRepository) CreateTourExecution(ctx context.Context, exec *model.TourExecution) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Generiši ID ako nije postavljen
	exec.BeforeCreate()
	exec.LastActivityTime = time.Now()

	_, err := r.Collection.InsertOne(ctx, exec)
	return err
}

// UpdateLastActivityTime ažurira lastActivityTime i eventualno CurrentPosition
func (r *TourExecutionRepository) UpdateLastActivityTime(ctx context.Context, execID string, position *model.Position) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	update := map[string]interface{}{
		"lastActivityTime": time.Now(),
	}

	if position != nil {
		update["currentPosition"] = position
	}

	_, err := r.Collection.UpdateOne(ctx, map[string]interface{}{"id": execID}, map[string]interface{}{
		"$set": update,
	})
	return err
}

// CompleteTour završava sesiju ture
func (r *TourExecutionRepository) CompleteTour(ctx context.Context, execID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()
	_, err := r.Collection.UpdateOne(ctx, map[string]interface{}{"id": execID}, map[string]interface{}{
		"$set": map[string]interface{}{
			"status":  model.StatusCompleted,
			"endTime": now,
		},
	})
	return err
}

func (r *TourExecutionRepository) AbandonTour(ctx context.Context, execID string, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()

	// filter uključuje i executionId i userId
	filter := map[string]interface{}{
		"id":     execID,
		"userId": userID,
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"status":  model.StatusAbandoned,
			"endTime": now,
		},
	}

	res, err := r.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// Ako ništa nije ažurirano, znači da nije pronađena sesija
	if res.MatchedCount == 0 {
		return fmt.Errorf("no tour execution found for id=%s and userId=%s", execID, userID)
	}

	return nil
}

// AddCompletedKeyPoint beleži završenu ključnu tačku
func (r *TourExecutionRepository) AddCompletedKeyPoint(ctx context.Context, execID string, keyPoint model.CompletedKeyPoint) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.Collection.UpdateOne(ctx, map[string]interface{}{"id": execID}, map[string]interface{}{
		"$push": map[string]interface{}{
			"completedKeyPoints": keyPoint,
		},
		"$set": map[string]interface{}{
			"lastActivityTime": time.Now(),
		},
	})
	return err
}
