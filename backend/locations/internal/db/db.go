package db

import (
	"context"
	"locations/internal/models"
)

// Database is an interface that defines the methods for interacting with the database.
type Database interface {
	// InsertLocationUpdate inserts a location update into the database.
	InsertLocationUpdate(ctx context.Context, update models.LocationUpdate) error

	// GetLocationByID retrieves a location update by its ID from the database.
	GetLocationByID(ctx context.Context, id string) (*models.LocationUpdate, error)

	// UpdateLocation updates a location update in the database.
	UpdateLocation(ctx context.Context, id string, update models.LocationUpdate) error
}
