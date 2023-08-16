package producer

import (
	"context"
	"encoding/json"

	"locations/internal/models"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writerConfig kafka.WriterConfig
}

func NewKafkaProducer(kafkaBrokers []string, topic string) *KafkaProducer {
	return &KafkaProducer{
		writerConfig : kafka.WriterConfig{
			Brokers: kafkaBrokers,
			Topic: topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *KafkaProducer) ProduceLocationUpdate(ctx context.Context, location models.LocationUpdate) error {
	writer := kafka.NewWriter(p.writerConfig)
	defer writer.Close()

	locationBytes, err := json.Marshal(location)
	if err != nil {
		return err
	}

	return writer.WriteMessages(ctx, kafka.Message{
		Key:   nil,
		Value: locationBytes,
	})
}
