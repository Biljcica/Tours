package repo

import (
	"context"
	"database-example/model"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TourExecutionRepository upravlja TourExecution dokumentima u Mongo
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

	exec.BeforeCreate()
	exec.LastActivityTime = time.Now()

	_, err := r.Collection.InsertOne(ctx, exec)
	return err
}

// UpdateLastActivityTime ažurira lastActivityTime
func (r *TourExecutionRepository) UpdateLastActivityTime(ctx context.Context, execID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	update := bson.M{
		"lastActivityTime": time.Now(),
	}

	_, err := r.Collection.UpdateOne(ctx, bson.M{"_id": execID}, bson.M{"$set": update})
	return err
}

// CompleteTour završava sesiju ture
func (r *TourExecutionRepository) CompleteTour(ctx context.Context, execID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()
	_, err := r.Collection.UpdateOne(
		ctx,
		bson.M{"_id": execID},
		bson.M{"$set": bson.M{
			"status":  model.StatusCompleted,
			"endTime": now,
		}},
	)
	return err
}

// AbandonTour napušta turu
func (r *TourExecutionRepository) AbandonTour(ctx context.Context, execID string, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()

	filter := bson.M{
		"_id":       execID,
		"touristId": userID,
	}

	update := bson.M{
		"$set": bson.M{
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

	update := bson.M{
		"$push": bson.M{
			"completedKeyPoints": keyPoint,
		},
		"$set": bson.M{
			"lastActivityTime": time.Now(),
		},
	}

	_, err := r.Collection.UpdateOne(ctx, bson.M{"_id": execID}, update)
	return err
}

// GetTourExecutionByID vraća TourExecution po execID
func (r *TourExecutionRepository) GetTourExecutionByID(ctx context.Context, execID string) (*model.TourExecution, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var exec model.TourExecution
	err := r.Collection.FindOne(ctx, bson.M{"_id": execID}).Decode(&exec)
	if err != nil {
		return nil, err
	}
	return &exec, nil
}

// GetByUserAndTour vraća TourExecution po userID i tourID
func (r *TourExecutionRepository) GetByUserAndTour(ctx context.Context, userID, tourID string) (*model.TourExecution, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{
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

// UpdateTourExecution ažurira status ture (npr. COMPLETE)
func (r *TourExecutionRepository) UpdateTourExecution(ctx context.Context, execID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()

	update := bson.M{
		"$set": bson.M{
			"status":  model.StatusCompleted,
			"endTime": now,
		},
	}

	res, err := r.Collection.UpdateOne(ctx, bson.M{"_id": execID}, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return fmt.Errorf("no tour execution found for id=%s", execID)
	}

	return nil
}
