package model

import (
	"github.com/google/uuid"
)

// KeyPoint predstavlja ključnu tačku ture
type KeyPoint struct {
	ID          string  `json:"id" bson:"id"`                   // UUID
	Name        string  `json:"name" bson:"name"`               // Naziv tačke
	Description string  `json:"description" bson:"description"` // Opis tačke
	Latitude    float64 `json:"latitude" bson:"latitude"`       // Geografska širina
	Longitude   float64 `json:"longitude" bson:"longitude"`     // Geografska dužina
	ImageURL    string  `json:"imageUrl" bson:"imageUrl"`       // URL slike (opciono)
}

// Pre-create hook da se generiše ID ako nije postavljen
func (kp *KeyPoint) BeforeCreate() {
	if kp.ID == "" {
		kp.ID = uuid.New().String()
	}
}
