package model

import (
	"time"

	"github.com/google/uuid"
)

type TourExecutionStatus string

const (
	StatusActive    TourExecutionStatus = "ACTIVE"
	StatusCompleted TourExecutionStatus = "COMPLETED"
	StatusAbandoned TourExecutionStatus = "ABANDONED"
)

// CompletedKeyPoint beleži kada je turista završio ključnu tačku
type CompletedKeyPoint struct {
	KeyPointID  string    `json:"keyPointId" bson:"keyPointId"`
	CompletedAt time.Time `json:"completedAt" bson:"completedAt"`
}

// TourExecution prati aktivnu sesiju ture
type TourExecution struct {
	ID                 string              `json:"id" bson:"_id,omitempty"`
	TourID             string              `json:"tourId" bson:"tourId"`
	TouristID          string              `json:"touristId" bson:"touristId"`
	Status             TourExecutionStatus `json:"status" bson:"status"`
	StartTime          time.Time           `json:"startTime" bson:"startTime"`
	EndTime            *time.Time          `json:"endTime,omitempty" bson:"endTime,omitempty"`
	LastActivityTime   time.Time           `json:"lastActivityTime" bson:"lastActivityTime"`
	CompletedKeyPoints []CompletedKeyPoint `json:"completedKeyPoints" bson:"completedKeyPoints"`
}

func (t *TourExecution) BeforeCreate() {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
}
