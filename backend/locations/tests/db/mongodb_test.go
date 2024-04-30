package db_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"locations/internal/db"
	"locations/internal/models"
)

// MockMongoDB is a mock implementation of the MongoDB struct for testing purposes.
type MockMongoDB struct {
	mock.Mock
}

func (m *MockMongoDB) InsertLocationUpdate(ctx context.Context, update models.LocationUpdate) error {
	args := m.Called(ctx, update)
	return args.Error(0)
}

func (m *MockMongoDB) GetLocationByID(ctx context.Context, id string) (*models.LocationUpdate, error) {
	args := m.Called(ctx, id)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*models.LocationUpdate), args.Error(1)
}

func (m *MockMongoDB) UpdateLocation(ctx context.Context, id string, update models.LocationUpdate) error {
	args := m.Called(ctx, id, update)
	return args.Error(0)
}

func TestInsertLocationUpdate_Success(t *testing.T) {
	mockClient := new(MockMongoDB)

	update := models.LocationUpdate{
		ID:        "1",
		DriverID:  "123",
		Latitude:  37.7749,
		Longitude: -122.4194,
		Timestamp: time.Now(),
	}

	mockClient.On("InsertLocationUpdate", mock.Anything, update).Return(nil)

	err := mockClient.InsertLocationUpdate(context.Background(), update)

	assert.NoError(t, err)
}

func TestInsertLocationUpdate_Error(t *testing.T) {
	mockClient := new(MockMongoDB)

	update := models.LocationUpdate{
		ID:        "1",
		DriverID:  "123",
		Latitude:  37.7749,
		Longitude: -122.4194,
		Timestamp: time.Now(),
	}

	mockClient.On("InsertLocationUpdate", mock.Anything, update).Return(errors.New("database error"))

	err := mockClient.InsertLocationUpdate(context.Background(), update)

	assert.Error(t, err)
}

func TestGetLocationByID_Success(t *testing.T) {
	mockClient := new(MockMongoDB)

	expectedUpdate := &models.LocationUpdate{
		ID:        "1",
		DriverID:  "123",
		Latitude:  37.7749,
		Longitude: -122.4194,
		Timestamp: time.Now(),
	}

	mockClient.On("GetLocationByID", mock.Anything, "1").Return(expectedUpdate, nil)

	update, err := mockClient.GetLocationByID(context.Background(), "1")

	assert.NoError(t, err)
	assert.Equal(t, expectedUpdate, update)
}

func TestGetLocationByID_NotFound(t *testing.T) {
	mockClient := new(MockMongoDB)

	mockClient.On("GetLocationByID", mock.Anything, "1").Return(nil, mongo.ErrNoDocuments)

	update, err := mockClient.GetLocationByID(context.Background(), "1")

	assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	assert.Nil(t, update)
}

func TestGetLocationByID_Error(t *testing.T) {
	mockClient := new(MockMongoDB)

	mockClient.On("GetLocationByID", mock.Anything, "1").Return(nil, errors.New("database error"))

	update, err := mockClient.GetLocationByID(context.Background(), "1")

	assert.Error(t, err)
	assert.Nil(t, update)
}

func TestUpdateLocation_Success(t *testing.T) {
	mockClient := new(MockMongoDB)

	update := models.LocationUpdate{
		ID:        "1",
		DriverID:  "123",
		Latitude:  37.7749,
		Longitude: -122.4194,
		Timestamp: time.Now(),
	}

	mockClient.On("UpdateLocation", mock.Anything, "1", update).Return(nil)

	err := mockClient.UpdateLocation(context.Background(), "1", update)

	assert.NoError(t, err)
}

func TestUpdateLocation_Error(t *testing.T) {
	mockClient := new(MockMongoDB)

	update := models.LocationUpdate{
		ID:        "1",
		DriverID:  "123",
		Latitude:  37.7749,
		Longitude: -122.4194,
		Timestamp: time.Now(),
	}

	mockClient.On("UpdateLocation", mock.Anything, "1", update).Return(errors.New("database error"))

	err := mockClient.UpdateLocation(context.Background(), "1", update)

	assert.Error(t, err)
}
