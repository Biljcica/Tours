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

// CreateTour dodaje novu turu
func (r *TourRepository) CreateTour(tour *model.Tour) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.Collection.InsertOne(ctx, tour)
	return err
}

// GetTourByID vraća turu po ID-u
func (r *TourRepository) GetTourByID(id string) (*model.Tour, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var tour model.Tour
	err := r.Collection.FindOne(ctx, bson.M{"id": id}).Decode(&tour)
	if err != nil {
		return nil, err
	}
	return &tour, nil
}

// GetToursByAuthor vraća sve ture za datog autora
func (r *TourRepository) GetToursByAuthor(authorID string) ([]model.Tour, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

// UpdateTourStatus menja status ture (npr. draft -> published)
func (r *TourRepository) UpdateTourStatus(id string, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.Collection.UpdateOne(
		ctx,
		bson.M{"id": id},
		bson.M{"$set": bson.M{"status": status}},
	)
	return err
}

func (r *TourRepository) GetAllTours() ([]model.Tour, error) {
	var tours []model.Tour
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var tour model.Tour
		if err := cursor.Decode(&tour); err != nil {
			return nil, err
		}
		tours = append(tours, tour)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return tours, nil
}

