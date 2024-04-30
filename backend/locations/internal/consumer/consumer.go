package consumer

import (
	"context" // Provides functionality to define a deadline or cancellation signal for operations.
	"fmt"     // Implements formatted I/O functions.
	"locations/internal/db" // Internal package for database operations.

	"github.com/segmentio/kafka-go" // Kafka library for Go.
)

// KafkaConsumer struct holds the configuration for a Kafka reader.
type KafkaConsumer struct {
	readerConfig kafka.ReaderConfig
}

// NewKafkaConsumer creates a new KafkaConsumer with the specified brokers and topic.
func NewKafkaConsumer(kafkaBrokers []string, topic string) *KafkaConsumer {
	return &KafkaConsumer{
		readerConfig: kafka.ReaderConfig{
			Brokers:   kafkaBrokers, // List of Kafka broker addresses.
			Topic:     topic,        // Kafka topic to subscribe to.
			MinBytes:  10e3,         // Minimum number of bytes to fetch in a single request.
			MaxBytes:  10e6,         // Maximum number of bytes to fetch in a single request.
			GroupID:   "location-consumer-group", // Consumer group ID.
		},
	}
}

// ConsumeLocationUpdates is responsible for continuously polling the Kafka topic for new messages.
// Once a message is received, it's passed to a message processor which contains the logic for handling the message.
// This function is designed to run indefinitely until it receives a signal to stop via the context's cancellation.
func (c *KafkaConsumer) ConsumeLocationUpdates(ctx context.Context, messageProcessor MessageProcessor) {
	reader := kafka.NewReader(c.readerConfig) // Create a new Kafka reader with the specified configuration.
	defer reader.Close() // Ensure the reader is closed when the function returns.

	fmt.Println("Location consumer started and listening for updates...")

	for {
		select {
		case <-ctx.Done(): // Check if the context has been canceled.
			fmt.Println("Location consumer shutting down...")
			return // Exit the function if the context is canceled.
		default:
			msg, err := reader.ReadMessage(ctx) // Read a message from the Kafka topic.
			if err != nil {
				fmt.Println("Error reading Kafka message:", err)
				return // Exit the function if there's an error reading messages.
			}

			// Process the Kafka message using the provided message processor.
			if err := messageProcessor.ProcessMessage(ctx, msg); err != nil {
				fmt.Println("Error processing Kafka message:", err)
			}
		}
	}
}

// RunKafkaConsumer initializes the necessary components for consuming messages from a Kafka topic.
// It creates a KafkaConsumer instance, sets up a message processor, and starts the message consumption process.
// This function is typically called at the start of the application to begin listening for messages.
func RunKafkaConsumer(ctx context.Context, kafkaBrokers []string, topic string, db db.Database) error {
	kafkaConsumer := NewKafkaConsumer(kafkaBrokers, topic) // Create a new Kafka consumer.
	messageProcessor := NewKafkaMessageProcessor(db) // Create a new message processor with the database instance.

	consumerCtx, cancelConsumer := context.WithCancel(ctx) // Create a cancellable context for the consumer.
	defer cancelConsumer() // Ensure the cancel function is called when the function exits.

	go kafkaConsumer.ConsumeLocationUpdates(consumerCtx, messageProcessor) // Start consuming messages in a new goroutine.

	select {
	case <-ctx.Done(): // Wait for the parent context to be canceled.
		cancelConsumer() // Cancel the consumer context to stop consuming messages.
	}
	return nil // Return nil to indicate no error occurred.
}
