package handler

import (
	"context"
	"database-example/model"
	pb "database-example/proto/review"
	"database-example/service"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ReviewHandler struct {
	pb.UnimplementedReviewServiceServer
	ReviewService *service.ReviewService
}

func NewReviewHandler(reviewService *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		ReviewService: reviewService,
	}
}

// ------------------------------------------------
// Pomocna funkcija: mapiranje model.Review -> pb.Review
// ------------------------------------------------
func mapReviewToPb(r *model.Review) *pb.Review {
	return &pb.Review{
		Id:          r.ID,
		Rating:      int32(r.Rating),
		Comment:     r.Comment,
		TouristId:   r.TouristID,
		TourId:      r.TourID,
		VisitDate:   timestamppb.New(r.VisitDate),
		CommentDate: timestamppb.New(r.CommentDate),
		Images:      r.Images,
		TouristName: r.TouristName,
	}
}

// ------------------------------------------------
// Pomocna funkcija: mapiranje pb.Review -> model.Review
// ------------------------------------------------
func mapPbToReview(pbReview *pb.Review) *model.Review {
	return &model.Review{
		ID:          pbReview.Id,
		Rating:      int(pbReview.Rating),
		Comment:     pbReview.Comment,
		TouristID:   pbReview.TouristId,
		TourID:      pbReview.TourId,
		VisitDate:   pbReview.VisitDate.AsTime(),
		CommentDate: pbReview.CommentDate.AsTime(),
		Images:      pbReview.Images,
		TouristName: pbReview.TouristName,
	}
}

// ------------------------------------------------
// Kreiranje recenzije
// ------------------------------------------------
func (h *ReviewHandler) CreateReview(ctx context.Context, req *pb.CreateReviewRequest) (*pb.ReviewResponse, error) {
	var visitDate time.Time
	if req.VisitDate != nil {
		visitDate = req.VisitDate.AsTime()
	}

	review, err := h.ReviewService.CreateReview(
		ctx,
		int(req.Rating),
		req.Comment,
		req.TouristId,
		req.TourId,
		visitDate,
		req.Images,
		req.TouristName,
		req.TouristImage,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create review: %v", err)
	}

	return &pb.ReviewResponse{
		Review: mapReviewToPb(review),
	}, nil
}

// ------------------------------------------------
// Dohvatanje recenzije po ID-u
// ------------------------------------------------
func (h *ReviewHandler) GetReview(ctx context.Context, req *pb.GetReviewRequest) (*pb.ReviewResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "review ID is required")
	}

	review, err := h.ReviewService.GetReview(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get review: %v", err)
	}

	return &pb.ReviewResponse{
		Review: mapReviewToPb(review),
	}, nil
}

// ------------------------------------------------
// Dohvatanje recenzija za turu
// ------------------------------------------------
func (h *ReviewHandler) GetReviewsByTour(ctx context.Context, req *pb.GetReviewsByTourRequest) (*pb.GetReviewsResponse, error) {
	if req.TourId == "" {
		return nil, status.Error(codes.InvalidArgument, "tour ID is required")
	}

	reviews, err := h.ReviewService.GetReviewsByTour(ctx, req.TourId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get reviews for tour: %v", err)
	}

	var pbReviews []*pb.Review
	for _, r := range reviews {
		pbReviews = append(pbReviews, mapReviewToPb(&r))
	}

	return &pb.GetReviewsResponse{
		Reviews: pbReviews,
	}, nil
}

// ------------------------------------------------
// Dohvatanje recenzija za turistu
// ------------------------------------------------
func (h *ReviewHandler) GetReviewsByTourist(ctx context.Context, req *pb.GetReviewsByTouristRequest) (*pb.GetReviewsResponse, error) {
	if req.TouristId == "" {
		return nil, status.Error(codes.InvalidArgument, "tourist ID is required")
	}

	reviews, err := h.ReviewService.GetReviewsByTourist(ctx, req.TouristId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get reviews for tourist: %v", err)
	}

	var pbReviews []*pb.Review
	for _, r := range reviews {
		pbReviews = append(pbReviews, mapReviewToPb(&r))
	}

	return &pb.GetReviewsResponse{
		Reviews: pbReviews,
	}, nil
}

// ------------------------------------------------
// Ažuriranje recenzije
// ------------------------------------------------
func (h *ReviewHandler) UpdateReview(ctx context.Context, req *pb.UpdateReviewRequest) (*pb.ReviewResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "review ID is required")
	}

	var visitDate time.Time
	if req.VisitDate != nil {
		visitDate = req.VisitDate.AsTime()
	}

	review, err := h.ReviewService.UpdateReview(
		ctx,
		req.Id,
		int(req.Rating),
		req.Comment,
		visitDate,
		req.Images,
		req.TouristName,
		req.TouristImage,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update review: %v", err)
	}

	return &pb.ReviewResponse{
		Review: mapReviewToPb(review),
	}, nil
}

// ------------------------------------------------
// Brisanje recenzije
// ------------------------------------------------
func (h *ReviewHandler) DeleteReview(ctx context.Context, req *pb.DeleteReviewRequest) (*pb.DeleteReviewResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "review ID is required")
	}

	err := h.ReviewService.DeleteReview(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete review: %v", err)
	}

	return &pb.DeleteReviewResponse{
		Success: true,
	}, nil
}

// ------------------------------------------------
// Dohvatanje prosečne ocene za turu
// ------------------------------------------------
func (h *ReviewHandler) GetAverageRating(ctx context.Context, req *pb.GetAverageRatingRequest) (*pb.GetAverageRatingResponse, error) {
	if req.TourId == "" {
		return nil, status.Error(codes.InvalidArgument, "tour ID is required")
	}

	averageRating, err := h.ReviewService.GetAverageRatingForTour(ctx, req.TourId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get average rating: %v", err)
	}

	return &pb.GetAverageRatingResponse{
		AverageRating: averageRating,
	}, nil
}
