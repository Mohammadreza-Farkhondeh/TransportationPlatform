package models

import "time"

type LocationUpdate struct {
	ID        string    `json:"id,omitempty"`
	DriverID  string    `json:"driver_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp"`
}
