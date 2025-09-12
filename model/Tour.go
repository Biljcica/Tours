package model

import (
	"time"

	"github.com/google/uuid"
)

type Tour struct {
	ID          string     `json:"id" bson:"id"`
	AuthorID    string     `json:"authorId" bson:"authorId"`
	Name        string     `json:"name" bson:"name"`
	Description string     `json:"description" bson:"description"`
	Difficulty  string     `json:"difficulty" bson:"difficulty"`
	Tags        []string   `json:"tags" bson:"tags"`
	Status      string     `json:"status" bson:"status"`
	Price       float64    `json:"price" bson:"price"`
	KeyPoints   []KeyPoint `json:"keypoints" bson:"keypoints"`
	CreatedAt   time.Time  `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt" bson:"updatedAt"`
}

func (t *Tour) BeforeCreate() {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
}
