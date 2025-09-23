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
	KeyPointID string    `json:"keyPointId"`
	CompletedAt time.Time `json:"completedAt"`
}

// TourExecution prati aktivnu sesiju ture
type TourExecution struct {
	ID                string              `json:"id"`
	TourID            string              `json:"tourId"`
	TouristID         string              `json:"touristId"`
	Status            TourExecutionStatus `json:"status"`
	StartTime         time.Time           `json:"startTime"`
	EndTime           *time.Time          `json:"endTime"` // nil dok nije završena
	LastActivityTime  time.Time           `json:"lastActivityTime"`
	CurrentPosition   Position            `json:"currentPosition"`
	CompletedKeyPoints []CompletedKeyPoint `json:"completedKeyPoints"`
}

func (t *TourExecution) BeforeCreate() {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
}