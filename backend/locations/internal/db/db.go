package db

import (
	"context"
	"locations/internal/models"
)

type Database interface {
	InsertLocationUpdate(ctx context.Context, update models.LocationUpdate) error
}
