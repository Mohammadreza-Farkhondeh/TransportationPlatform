package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"locations/internal/http"
	"locations/internal/models"
	"locations/internal/producer"
)

// MockKafkaProducer is a mock implementation of the Kafka producer for testing purposes.
type MockKafkaProducer struct{}

func (m *MockKafkaProducer) ProduceLocationUpdate(ctx context.Context, location models.LocationUpdate) error {
	// Simulate successful message production.
	return nil
}

func TestLocationUpdateHandler_Success(t *testing.T) {
	mockProducer := &MockKafkaProducer{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.LocationUpdateHandler(w, r, mockProducer)
	}))
	defer server.Close()

	location := models.LocationUpdate{
		ID:        "1",
		DriverID:  "123",
		Latitude:  37.7749,
		Longitude: -122.4194,
		Timestamp: time.Now(),
	}

	locationJSON, err := json.Marshal(location)
	assert.NoError(t, err)

	resp, err := http.Post(server.URL+"/location", "application/json", bytes.NewBuffer(locationJSON))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLocationUpdateHandler_BadRequest(t *testing.T) {
	mockProducer := &MockKafkaProducer{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.LocationUpdateHandler(w, r, mockProducer)
	}))
	defer server.Close()

	// Sending an empty request body to simulate a bad request.
	resp, err := http.Post(server.URL+"/location", "application/json", nil)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRunHTTPServer(t *testing.T) {
	mockProducer := &MockKafkaProducer{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error)

	go func() {
		errCh <- http.RunHTTPServer(ctx, ":8080", mockProducer)
	}()

	// Wait for the server to start listening.
	time.Sleep(100 * time.Millisecond)

	// Send a test request to the server.
	location := models.LocationUpdate{
		ID:        "1",
		DriverID:  "123",
		Latitude:  37.7749,
		Longitude: -122.4194,
		Timestamp: time.Now(),
	}

	locationJSON, err := json.Marshal(location)
	assert.NoError(t, err)

	resp, err := http.Post("http://localhost:8080/location", "application/json", bytes.NewBuffer(locationJSON))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Stop the server by canceling the context.
	cancel()

	// Check if the server shutdown gracefully.
	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		assert.Fail(t, "server did not shut down gracefully")
	}
}
