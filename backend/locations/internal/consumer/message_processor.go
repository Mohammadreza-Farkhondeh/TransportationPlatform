package consumer

import (
	"context"
	"fmt"
	"log"
	"encoding/json"

	"github.com/segmentio/kafka-go"

	"locations/internal/db"
	"locations/internal/models"
)

// MessageProcessor is an interface that defines a single method, ProcessMessage, which takes a context and a Kafka message and returns an error.
type MessageProcessor interface {
	ProcessMessage(ctx context.Context, msg kafka.Message) error
}

// DefaultKafkaMessageProcessor is a struct that implements the MessageProcessor interface and contains a database instance.
type DefaultKafkaMessageProcessor struct {
	database db.Database
}

// NewKafkaMessageProcessor is a constructor function that creates a new DefaultKafkaMessageProcessor with the provided database.
func NewKafkaMessageProcessor(database db.Database) *DefaultKafkaMessageProcessor {
	return &DefaultKafkaMessageProcessor{
		database: database,
	}
}

// ProcessMessage is a method on DefaultKafkaMessageProcessor that handles the processing of Kafka messages.
func (p *DefaultKafkaMessageProcessor) ProcessMessage(ctx context.Context, msg kafka.Message) error {
	fmt.Printf("Received Kafka message: %s\n", msg.Value)

	// Declare a variable of type LocationUpdate from the models package.
	var locationUpdate models.LocationUpdate
	// Unmarshal the JSON-encoded Kafka message into the locationUpdate variable.
	err := json.Unmarshal(msg.Value, &locationUpdate)
	if err != nil {
		// If there is an error during unmarshaling, log the error and return it.
		log.Printf("Error parsing location update: %v\n", err)
		return err
	}

	// Insert the location update into the database using the InsertLocationUpdate method.
	if err := p.database.InsertLocationUpdate(ctx, locationUpdate); err != nil {
		// If there is an error during insertion, log the error and return it.
		log.Printf("Error saving location: %v\n", err)
		return err
	}

	// If everything went well, return nil indicating no error occurred.
	return nil
}
