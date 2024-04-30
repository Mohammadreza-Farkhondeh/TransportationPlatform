package producer_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"locations/internal/models"
	"locations/internal/producer"
)

// MockKafkaWriter is a mock implementation of the Kafka writer for testing purposes.
type MockKafkaWriter struct {
	mock.Mock
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, messages ...kafka.Message) error {
	args := m.Called(ctx, messages)
	return args.Error(0)
}

func TestProduceLocationUpdate_Success(t *testing.T) {
	mockWriter := new(MockKafkaWriter)

	location := models.LocationUpdate{
		ID:        "1",
		DriverID:  "123",
		Latitude:  37.7749,
		Longitude: -122.4194,
	}

	mockWriter.On("WriteMessages", mock.Anything, mock.Anything).Return(nil)

	kafkaProducer := producer.KafkaProducer{
		writerConfig: kafka.WriterConfig{
			Brokers:  []string{"localhost:9092"},
			Topic:    "test-topic",
			Balancer: &kafka.LeastBytes{},
		},
	}

	err := kafkaProducer.ProduceLocationUpdate(context.Background(), location)

	assert.NoError(t, err)
}

func TestProduceLocationUpdate_Error(t *testing.T) {
	mockWriter := new(MockKafkaWriter)

	location := models.LocationUpdate{
		ID:        "1",
		DriverID:  "123",
		Latitude:  37.7749,
		Longitude: -122.4194,
	}

	mockWriter.On("WriteMessages", mock.Anything, mock.Anything).Return(errors.New("kafka error"))

	kafkaProducer := producer.KafkaProducer{
		writerConfig: kafka.WriterConfig{
			Brokers:  []string{"localhost:9092"},
			Topic:    "test-topic",
			Balancer: &kafka.LeastBytes{},
		},
	}

	err := kafkaProducer.ProduceLocationUpdate(context.Background(), location)

	assert.Error(t, err)
}
