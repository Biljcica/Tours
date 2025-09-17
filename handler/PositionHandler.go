package handler

import (
	"context"

	pb "database-example/proto/position"
	"database-example/service"
)

type PositionHandler struct {
	pb.UnimplementedPositionServiceServer
	service *service.PositionService
}

func NewPositionHandler(s *service.PositionService) *PositionHandler {
	return &PositionHandler{service: s}
}

// RecordPosition RPC
func (h *PositionHandler) RecordPosition(ctx context.Context, req *pb.RecordPositionRequest) (*pb.PositionResponse, error) {
	pos, err := h.service.RecordPosition(ctx, req.TouristId, req.Latitude, req.Longitude)
	if err != nil {
		return nil, err
	}

	return &pb.PositionResponse{
		TouristId: pos.TouristID,
		Latitude:  pos.Latitude,
		Longitude: pos.Longitude,
		UpdatedAt: pos.UpdatedAt.Unix(),
	}, nil
}

// GetCurrentPosition RPC
func (h *PositionHandler) GetCurrentPosition(ctx context.Context, req *pb.GetCurrentPositionRequest) (*pb.PositionResponse, error) {
	pos, err := h.service.GetCurrentPosition(ctx, req.TouristId)
	if err != nil {
		return nil, err
	}

	if pos == nil {
		// turista jos nema poziciju u bazi
		return &pb.PositionResponse{
			TouristId: req.TouristId,
			Latitude:  0,
			Longitude: 0,
			UpdatedAt: 0,
		}, nil
	}

	return &pb.PositionResponse{
		TouristId: pos.TouristID,
		Latitude:  pos.Latitude,
		Longitude: pos.Longitude,
		UpdatedAt: pos.UpdatedAt.Unix(),
	}, nil
}