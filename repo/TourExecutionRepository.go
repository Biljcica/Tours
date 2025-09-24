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
func (r *TourExecutionRepository) UpdateLastActivityTime(ctx context.Context, execID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	update := map[string]interface{}{
		"lastActivityTime": time.Now(),
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

	filter := map[string]interface{}{
		"id":        execID,
		"touristId": userID, // ✔ koristi tačan naziv iz Mongo dokumenta
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

	if res.MatchedCount == 0 {
		return fmt.Errorf("no tour execution found for id=%s and touristId=%s", execID, userID)
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

func (r *TourExecutionRepository) GetTourExecutionByID(ctx context.Context, execID string) (*model.TourExecution, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var exec model.TourExecution
	err := r.Collection.FindOne(ctx, map[string]interface{}{"id": execID}).Decode(&exec)
	if err != nil {
		return nil, err
	}
	return &exec, nil
}

func (r *TourExecutionRepository) GetByUserAndTour(ctx context.Context, userID, tourID string) (*model.TourExecution, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := map[string]interface{}{
		"touristId": userID,
		"tourId":    tourID,
	}

	var exec model.TourExecution
	err := r.Collection.FindOne(ctx, filter).Decode(&exec)
	if err != nil {
		return nil, err
	}
	return &exec, nil
}

func (r *TourExecutionRepository) UpdateTourExecution(ctx context.Context, execID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()

	filter := map[string]interface{}{
		"id": execID,
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"status":  model.StatusCompleted,
			"endTime": now,
		},
	}

	res, err := r.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return fmt.Errorf("no tour execution found for id=%s and touristId=%s", execID)
	}

	return nil
}
