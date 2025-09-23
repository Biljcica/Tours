package service

import (
	"context"
	"errors"
	"time"

	"database-example/model"
	"database-example/proto/review"
	"database-example/repo"
)

type ReviewService struct {
	review.UnimplementedReviewServiceServer
	ReviewRepo *repo.ReviewRepository
}

func NewReviewService(reviewRepo *repo.ReviewRepository) *ReviewService {
	return &ReviewService{ReviewRepo: reviewRepo}
}

func (s *ReviewService) CreateReview(ctx context.Context, rating int, comment string, touristID string, tourID string, visitDate time.Time, images []string, touristName string, touristImage string) (*model.Review, error) {
	// Validacija
	if rating < 1 || rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}

	if len(comment) < 10 {
		return nil, errors.New("comment must be at least 10 characters long")
	}

	if touristID == "" {
		return nil, errors.New("tourist ID is required")
	}

	if tourID == "" {
		return nil, errors.New("tour ID is required")
	}

	// Ako visitDate nije postavljen, koristimo današnji datum
	if visitDate.IsZero() {
		visitDate = time.Now()
	}

	review := &model.Review{
		ID:           "", // ID će biti generisan u repository-ju
		Rating:       rating,
		Comment:      comment,
		TouristID:    touristID,
		TourID:       tourID,
		VisitDate:    visitDate,
		CommentDate:  time.Now(),
		Images:       images,
		TouristName:  touristName,
		TouristImage: touristImage,
	}

	err := s.ReviewRepo.CreateReview(ctx, review)
	if err != nil {
		return nil, err
	}
	return review, nil
}

func (s *ReviewService) GetReview(ctx context.Context, id string) (*model.Review, error) {
	if id == "" {
		return nil, errors.New("review ID is required")
	}
	return s.ReviewRepo.GetReviewByID(ctx, id)
}

func (s *ReviewService) GetReviewsByTour(ctx context.Context, tourID string) ([]model.Review, error) {
	if tourID == "" {
		return nil, errors.New("tour ID is required")
	}
	return s.ReviewRepo.GetReviewsByTourID(ctx, tourID)
}

func (s *ReviewService) GetReviewsByTourist(ctx context.Context, touristID string) ([]model.Review, error) {
	if touristID == "" {
		return nil, errors.New("tourist ID is required")
	}
	return s.ReviewRepo.GetReviewsByTouristID(ctx, touristID)
}

func (s *ReviewService) UpdateReview(ctx context.Context, id string, rating int, comment string, visitDate time.Time, images []string, touristName string, touristImage string) (*model.Review, error) {
	// Validacija
	if id == "" {
		return nil, errors.New("review ID is required")
	}

	if rating < 1 || rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}

	if len(comment) < 10 {
		return nil, errors.New("comment must be at least 10 characters long")
	}

	// Prvo dobijemo postojeću recenziju
	existingReview, err := s.ReviewRepo.GetReviewByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Ažuriramo samo polja koja se mogu menjati
	updatedReview := &model.Review{
		ID:           id,
		Rating:       rating,
		Comment:      comment,
		TouristID:    existingReview.TouristID, // Ne menja se
		TourID:       existingReview.TourID,    // Ne menja se
		VisitDate:    visitDate,
		CommentDate:  existingReview.CommentDate, // Ne menja se
		Images:       images,
		TouristName:  touristName,
		TouristImage: touristImage,
	}

	err = s.ReviewRepo.UpdateReview(ctx, id, updatedReview)
	if err != nil {
		return nil, err
	}
	return updatedReview, nil
}

func (s *ReviewService) DeleteReview(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("review ID is required")
	}
	return s.ReviewRepo.DeleteReview(ctx, id)
}

func (s *ReviewService) GetAverageRatingForTour(ctx context.Context, tourID string) (float64, error) {
	if tourID == "" {
		return 0, errors.New("tour ID is required")
	}
	return s.ReviewRepo.GetAverageRatingForTour(ctx, tourID)
}
