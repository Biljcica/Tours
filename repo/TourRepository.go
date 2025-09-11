package repo

import (
	"context"
	"time"

	"database-example/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TourRepository struct {
	Collection *mongo.Collection
}

// Novi repozitorijum
func NewTourRepository(collection *mongo.Collection) *TourRepository {
	return &TourRepository{Collection: collection}
}

func (r *TourRepository) CreateTour(ctx context.Context, tour *model.Tour) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.Collection.InsertOne(ctx, tour)
	return err
}

func (r *TourRepository) GetTourByID(ctx context.Context, id string) (*model.Tour, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var tour model.Tour
	err := r.Collection.FindOne(ctx, bson.M{"id": id}).Decode(&tour)
	if err != nil {
		return nil, err
	}
	return &tour, nil
}

func (r *TourRepository) GetToursByAuthor(ctx context.Context, authorID string) ([]model.Tour, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cursor, err := r.Collection.Find(ctx, bson.M{"authorId": authorID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tours []model.Tour
	if err := cursor.All(ctx, &tours); err != nil {
		return nil, err
	}
	return tours, nil
}

func (r *TourRepository) UpdateTourStatus(ctx context.Context, id string, status string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.Collection.UpdateOne(ctx, bson.M{"id": id}, bson.M{"$set": bson.M{"status": status}})
	return err
}

func (r *TourRepository) GetAllTours(ctx context.Context) ([]model.Tour, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tours []model.Tour
	for cursor.Next(ctx) {
		var t model.Tour
		if err := cursor.Decode(&t); err != nil {
			return nil, err
		}
		tours = append(tours, t)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return tours, nil
}
