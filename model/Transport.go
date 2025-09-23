package model

type TransportType string

const (
    Walking TransportType = "walking"
    Bike TransportType = "bike"
    Car  TransportType = "car"
)

type Duration struct {
    Mode    TransportType `json:"mode" bson:"mode"`
    Minutes int           `json:"minutes" bson:"minutes"`
}