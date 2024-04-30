package consumer_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"locations/internal/consumer"
	"locations/internal/db"
	"locations/internal/models"
)

// MockDatabase is a mock implementation of the Database interface for testing purposes.
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) InsertLocationUpdate(ctx context.Context, update models.LocationUpdate) error {
	args := m.Called(ctx, update)
	return args.Error(0)
}

func TestProcessMessage_Success(t *testing.T) {
	// Mock database instance
	mockDB := new(MockDatabase)

	// Create a message processor with the mock database
	processor := consumer.NewKafkaMessageProcessor(mockDB)

	// Sample Kafka message payload
	messagePayload := `{"driver_id":"123","latitude":37.7749,"longitude":-122.4194}`

	// Create a Kafka message
	message := consumer.Message{
		Value: []byte(messagePayload),
	}

	// Expected location update
	expectedUpdate := models.LocationUpdate{
		DriverID:  "123",
		Latitude:  37.7749,
		Longitude: -122.4194,
	}

	// Mock the database insert method
	mockDB.On("InsertLocationUpdate", mock.Anything, expectedUpdate).Return(nil)

	// Process the message
	err := processor.ProcessMessage(context.Background(), message)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Assert that the database insert method was called with the expected parameters
	mockDB.AssertCalled(t, "InsertLocationUpdate", mock.Anything, expectedUpdate)
}

func TestProcessMessage_ParseError(t *testing.T) {
	// Create a message processor with a nil database (to trigger parsing error)
	processor := consumer.NewKafkaMessageProcessor(nil)

	// Sample Kafka message payload with invalid format
	invalidPayload := `invalid_json_format`

	// Create a Kafka message with the invalid payload
	invalidMessage := consumer.Message{
		Value: []byte(invalidPayload),
	}

	// Process the message
	err := processor.ProcessMessage(context.Background(), invalidMessage)

	// Assert that an error occurred during parsing
	assert.Error(t, err)
}

func TestProcessMessage_DatabaseError(t *testing.T) {
	// Mock database instance
	mockDB := new(MockDatabase)

	// Create a message processor with the mock database
	processor := consumer.NewKafkaMessageProcessor(mockDB)

	// Sample Kafka message payload
	messagePayload := `{"driver_id":"123","latitude":37.7749,"longitude":-122.4194}`

	// Create a Kafka message
	message := consumer.Message{
		Value: []byte(messagePayload),
	}

	// Expected location update
	expectedUpdate := models.LocationUpdate{
	
