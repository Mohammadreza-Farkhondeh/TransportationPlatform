package consumer

import (
	"context"
	"fmt"
	"locations/internal/db"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	readerConfig kafka.ReaderConfig
}

func NewKafkaConsumer(kafkaBrokers []string, topic string) *KafkaConsumer {
	return &KafkaConsumer{
		readerConfig: kafka.ReaderConfig{
			Brokers:   kafkaBrokers,
			Topic:     topic,
			MinBytes:  10e3,
			MaxBytes:  10e6,
			GroupID:   "location-consumer-group",
		},
	}
}


func (c *KafkaConsumer) ConsumeLocationUpdates(ctx context.Context, messageProcessor MessageProcessor) {
	reader := kafka.NewReader(c.readerConfig)
	defer reader.Close()

	fmt.Println("Location consumer started and listening for updates...")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Location consumer shutting down...")
			return
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				fmt.Println("Error reading Kafka message:", err)
				return
			}

			// Process the Kafka message using the provided message processor
			if err := messageProcessor.ProcessMessage(ctx, msg); err != nil {
				fmt.Println("Error processing Kafka message:", err)
			}
		}
	}
}

func RunKafkaConsumer(ctx context.Context, kafkaBrokers []string, topic string, db db.Database) error {
	kafkaConsumer := NewKafkaConsumer(kafkaBrokers, topic)
	messageProcessor := NewKafkaMessageProcessor(db)

	consumerCtx, cancelConsumer := context.WithCancel(ctx)
	defer cancelConsumer()

	go kafkaConsumer.ConsumeLocationUpdates(consumerCtx, messageProcessor)

	select {
	case <-ctx.Done():
		cancelConsumer()
	}
	return nil
}