package producer

import (
	"context" // Provides functionality to define a deadline or cancellation signal for operations.
	"encoding/json" // Implements encoding and decoding of JSON.

	"locations/internal/models" // Internal package for data models.

	"github.com/segmentio/kafka-go" // Kafka library for Go.
)

// KafkaProducer encapsulates the Kafka writer configuration needed to send messages to a Kafka topic.
type KafkaProducer struct {
	writerConfig kafka.WriterConfig
}

// NewKafkaProducer initializes a new KafkaProducer with the specified broker addresses and topic.
// It configures the Kafka writer with a LeastBytes balancer to evenly distribute messages across partitions.
func NewKafkaProducer(kafkaBrokers []string, topic string) *KafkaProducer {
	return &KafkaProducer{
		writerConfig: kafka.WriterConfig{
			Brokers:  kafkaBrokers, // Kafka broker addresses.
			Topic:    topic,        // Kafka topic for publishing messages.
			Balancer: &kafka.LeastBytes{}, // Balancer for distributing messages across brokers.
		},
	}
}

// ProduceLocationUpdate takes a LocationUpdate model and sends it to the configured Kafka topic.
// It first converts the LocationUpdate into a JSON byte slice, then creates a Kafka message,
// and finally writes the message to the Kafka topic using the Kafka writer.
func (p *KafkaProducer) ProduceLocationUpdate(ctx context.Context, location models.LocationUpdate) error {
	writer := kafka.NewWriter(p.writerConfig) // Instantiate a new Kafka writer with the provided configuration.
	defer writer.Close() // Ensure the writer is closed properly after message production.

	locationBytes, err := json.Marshal(location) // Convert the LocationUpdate to JSON format.
	if err != nil {
		return err // If JSON marshaling fails, return the error.
	}

	// Construct a Kafka message with the JSON-encoded location update as the value.
	message := kafka.Message{
		Key:   nil, // No key is specified, as the message key is not used in this context.
		Value: locationBytes, // The JSON-encoded location update.
	}

	// Write the constructed message to the Kafka topic.
	// The context allows for timeout or cancellation of the message production.
	return writer.WriteMessages(ctx, message)
}
