package model

import (
	"time"
)

type Review struct {
	ID           string    `json:"id"`
	Rating       int       `json:"rating"`       // Ocena (1-5)
	Comment      string    `json:"comment"`      // Komentar
	TouristID    string    `json:"touristId"`    // ID turiste
	TourID       string    `json:"tourId"`       // ID turne
	VisitDate    time.Time `json:"visitDate"`    // Datum posete
	CommentDate  time.Time `json:"commentDate"`  // Datum komentara
	Images       []string  `json:"images"`       // Slike (Base64 ili putanje)
	TouristName  string    `json:"touristName"`  // Ime turiste
	TouristImage string    `json:"touristImage"` // Slika turiste (opciono)
}
