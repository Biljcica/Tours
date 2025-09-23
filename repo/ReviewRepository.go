package repo

import (
	"context"
	"database-example/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReviewRepository struct {
	Collection *mongo.Collection
}

func NewReviewRepository(collection *mongo.Collection) *ReviewRepository {
	return &ReviewRepository{Collection: collection}
}

func (r *ReviewRepository) CreateReview(ctx context.Context, review *model.Review) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Ako ID nije postavljen, generišemo ObjectID
	if review.ID == "" {
		objectID := primitive.NewObjectID()
		review.ID = objectID.Hex()
	}

	_, err := r.Collection.InsertOne(ctx, review)
	return err
}

func (r *ReviewRepository) GetReviewByID(ctx context.Context, id string) (*model.Review, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var review model.Review
	err := r.Collection.FindOne(ctx, bson.M{"id": id}).Decode(&review)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *ReviewRepository) GetReviewsByTourID(ctx context.Context, tourID string) ([]model.Review, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cursor, err := r.Collection.Find(ctx, bson.M{"tourid": tourID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reviews []model.Review
	if err := cursor.All(ctx, &reviews); err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewRepository) GetReviewsByTouristID(ctx context.Context, touristID string) ([]model.Review, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cursor, err := r.Collection.Find(ctx, bson.M{"touristId": touristID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reviews []model.Review
	if err := cursor.All(ctx, &reviews); err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewRepository) UpdateReview(ctx context.Context, id string, updatedReview *model.Review) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"rating":       updatedReview.Rating,
			"comment":      updatedReview.Comment,
			"visitDate":    updatedReview.VisitDate,
			"commentDate":  updatedReview.CommentDate,
			"images":       updatedReview.Images,
			"touristName":  updatedReview.TouristName,
			"touristImage": updatedReview.TouristImage,
		},
	}

	_, err := r.Collection.UpdateOne(ctx, bson.M{"id": id}, update)
	return err
}

func (r *ReviewRepository) DeleteReview(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.Collection.DeleteOne(ctx, bson.M{"id": id})
	return err
}

func (r *ReviewRepository) GetAverageRatingForTour(ctx context.Context, tourID string) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		bson.D{{"$match", bson.D{{"tourid", tourID}}}},
		bson.D{{"$group", bson.D{
			{"_id", "$tourid"},
			{"averageRating", bson.D{{"$avg", "$rating"}}},
		}}},
	}

	cursor, err := r.Collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result struct {
		AverageRating float64 `bson:"averageRating"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
		return result.AverageRating, nil
	}

	return 0, nil // Nema recenzija za ovu turu
}
