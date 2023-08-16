package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"log"

    "github.com/joho/godotenv"

	"locations/internal/consumer"
	"locations/internal/producer"
	"locations/internal/http"
	"locations/internal/db"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	mongoURI := os.Getenv("MONGODB_URI")
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	httpAddr := ":8080"

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create MongoDB instance
	mongoDB, err := db.NewMongoDB(mongoURI)
	if err != nil {
		panic(err)
	}
	defer mongoDB.Close()

	// Create kafka producer instance
	kafkaProducer := producer.NewKafkaProducer([]string{kafkaBrokers}, kafkaTopic)

	// Run Kafka consumer
	go func() {
		if err := consumer.RunKafkaConsumer(ctx, []string{kafkaBrokers}, kafkaTopic, mongoDB); err != nil {
			fmt.Println("Kafka consumer error:", err)
		}
	}()

	// Run HTTP server
	go func() {
		if err := http.RunHTTPServer(ctx, httpAddr, *kafkaProducer); err != nil {
			fmt.Println("HTTP server error:", err)
		}
	}()

	// Wait for termination signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	// Cancel the context to trigger graceful shutdown
	cancel()
}