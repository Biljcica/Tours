package model

import "time"


type Position struct {        
    TouristID string    `gorm:"primaryKey" json:"touristId"`
    Latitude  float64   `json:"latitude"`
    Longitude float64   `json:"longitude"`
    UpdatedAt time.Time `json:"updatedAt"`
}

