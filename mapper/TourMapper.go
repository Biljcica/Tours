package mapper

import (
	"database-example/model"
    pb "database-example/proto/tours"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ------------------------------------------------
// Pomocna funkcija: mapiranje KeyPoints
// ------------------------------------------------
func MapKeyPointsToPb(kps []model.KeyPoint) []*pb.KeyPoint {
	var pbKeyPoints []*pb.KeyPoint
	for _, kp := range kps {
		pbKeyPoints = append(pbKeyPoints, &pb.KeyPoint{
			Id:          kp.ID,
			Name:        kp.Name,
			Description: kp.Description,
			Latitude:    kp.Latitude,
			Longitude:   kp.Longitude,
			ImageURL:    kp.ImageURL,
			Order:       kp.Order,
		})
	}
	return pbKeyPoints
}

// ------------------------------------------------
// Pomocna funkcija: mapiranje Durations
// ------------------------------------------------
func mapTransportType(mode model.TransportType) pb.TransportType {
    switch mode {
    case model.Walking:
        return pb.TransportType_WALKING
    case model.Bike:
        return pb.TransportType_BIKE
    case model.Car:
        return pb.TransportType_CAR
    default:
        return pb.TransportType_TRANSPORT_TYPE_UNSPECIFIED
    }
}

func MapDurationsToPb(durations []model.Duration) []*pb.Duration {
	var pbDurations []*pb.Duration
	for _, d := range durations {
		pbDurations = append(pbDurations, &pb.Duration{
			Mode:    mapTransportType(d.Mode),
			Minutes: int32(d.Minutes),
		})
	}
	return pbDurations
}

// ------------------------------------------------
// Pomocna funkcija: mapiranje Status
// ------------------------------------------------
func MapStatusToPb(status model.TourStatus) pb.TourStatus {
	switch status {
	case model.Draft:
		return pb.TourStatus_DRAFT
	case model.Published:
		return pb.TourStatus_PUBLISHED
	case model.Archived:
		return pb.TourStatus_ARCHIVED
	default:
		return pb.TourStatus_TOUR_STATUS_UNSPECIFIED
	}
}

// ------------------------------------------------
// Glavna funkcija: mapiranje model.Tour -> pb.TourResponse
// ------------------------------------------------
func MapTourToPb(t *model.Tour) *pb.TourResponse {
	var publishedAt *timestamppb.Timestamp
	var archivedAt *timestamppb.Timestamp
	
	if t.PublishedAt != nil {
		publishedAt = timestamppb.New(*t.PublishedAt)
	}
	if t.ArchivedAt != nil {
		archivedAt = timestamppb.New(*t.ArchivedAt)
	}

	return &pb.TourResponse{
		Id:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		Difficulty:  t.Difficulty,
		Tags:        t.Tags,
		Status:      MapStatusToPb(t.Status),
		Price:       t.Price,
		AuthorId:    t.AuthorID,
		KeyPoints:   MapKeyPointsToPb(t.KeyPoints),
		Durations:   MapDurationsToPb(t.Durations),
		Distance:    t.Distance,
		PublishedAt: publishedAt,
		ArchivedAt:  archivedAt,
	}
}
