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

type MessageProcessor interface {
	ProcessMessage(ctx context.Context, msg kafka.Message) error
}

type DefaultKafkaMessageProcessor struct {
	database db.Database
}

func NewKafkaMessageProcessor(database db.Database) *DefaultKafkaMessageProcessor {
	return &DefaultKafkaMessageProcessor{
		database: database,
	}
}

func (p *DefaultKafkaMessageProcessor) ProcessMessage(ctx context.Context, msg kafka.Message) error {
	fmt.Printf("Received Kafka message: %s\n", msg.Value)

	var locationUpdate models.LocationUpdate
	err := json.Unmarshal(msg.Value, &locationUpdate)
	if err != nil {
		log.Printf("Error parsing location update: %v\n", err)
		return err
	}

	if err := p.database.InsertLocationUpdate(ctx, locationUpdate); err != nil {
		log.Printf("Error saving location: %v\n", err)
		return err
	}

	return nil
}
