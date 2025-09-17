package repo

import (
	"context"
	"time"
	"errors"

	"database-example/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PositionRepository struct {
	Collection *mongo.Collection
}

func NewPositionRepository(collection *mongo.Collection) *PositionRepository {
	return &PositionRepository{Collection: collection}
}

//SaveOrUpdate cuva novu poziciju ili update-uje postojecu
func (r *PositionRepository) SaveOrUpdate(ctx context.Context, pos *model.Position) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"touristId": pos.TouristID}
	update := bson.M{
		"$set": bson.M{
			"latitude":  pos.Latitude,
			"longitude": pos.Longitude,
			"updatedAt": pos.UpdatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.Collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetPositionByTouristID vraca posljednju poznatu poziciju turiste
func (r *PositionRepository) GetPositionByTouristId(ctx context.Context, touristID string) (*model.Position, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var pos model.Position
	err := r.Collection.FindOne(ctx, bson.M{"touristId": touristID}).Decode(&pos)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // nema pozicije
		}
		return nil, err
	}
	return &pos, nil
}