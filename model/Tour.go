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
	Status      TourStatus `json:"status" bson:"status"`
	Price       float64    `json:"price" bson:"price"`
	KeyPoints   []KeyPoint `json:"keypoints" bson:"keypoints"`
	Distance 	float64	   `json:"distance" bson:"distance"`
	Durations 	[]Duration `json:"durations" bson:"durations"`
	CreatedAt   time.Time  `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt" bson:"updatedAt"`
	PublishedAt  *time.Time  `json:"publishedAt,omitempty" bson:"publishedAt,omitempty"`
	ArchivedAt  *time.Time  `json:"archivedAt,omitempty" bson:"archivedAt,omitempty"`
}

func (t *Tour) BeforeCreate() {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	if t.Status == "" {
		t.Status = Draft
	}
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
}
