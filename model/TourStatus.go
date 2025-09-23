package model

type TourStatus string

const (
    Draft     TourStatus = "draft"
    Published TourStatus = "published"
    Archived  TourStatus = "archived"
)