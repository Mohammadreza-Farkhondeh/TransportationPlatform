package main

import (
	"context" // Provides functionality to define a deadline or cancellation signal for operations.
	"fmt"     // Implements formatted I/O functions.
	"log"     // Implements a simple logging package.
	"os"      // Provides a platform-independent interface to operating system functionality.
	"os/signal" // Allows the program to receive notifications from the operating system about incoming signals.
	"syscall" // Contains an interface to the low-level operating system primitives.

	"github.com/joho/godotenv" // package for loading environment variables from .env files.

	// Internal packages for the location service application.
	"locations/internal/consumer"
	"locations/internal/db"
	"locations/internal/http"
	"locations/internal/producer"
)

func main() {
	// Load environment variables from .env file.
	// If the .env file is not found, log a fatal error and exit the application.
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// Read environment variables required for the application to connect to MongoDB and Kafka.
	mongoURI := os.Getenv("MONGODB_URI")       // MongoDB connection string.
	kafkaBrokers := os.Getenv("KAFKA_BROKERS") // Kafka broker list.
	kafkaTopic := os.Getenv("KAFKA_TOPIC")     // Kafka topic name.
	httpAddr := ":8080"                        // HTTP server address.

	// Create a context with cancellation capabilities to manage the lifecycle of the application.
	// 'context.Background()' returns a non-nil, empty context. It is never canceled, has no values, and has no deadline.
	// It is typically used by the main function, initialization, and tests, and as the top-level Context for incoming requests.
	ctx, cancel := context.WithCancel(context.Background())

	// 'defer' is used to ensure that a function call is performed later in a program’s execution, usually for purposes of cleanup.
	// 'cancel' is a function that, when called, will signal to the context that it should be canceled.
	// 'defer cancel()' ensures that the 'cancel' function is called when the current function returns, which is 'main' in this case.
	// This is important for resource cleanup and to signal any goroutines running in parallel to stop execution.
	defer cancel()

	// This pattern is common in Go for handling graceful shutdowns.
	// By calling 'cancel', all operations or goroutines that are using 'ctx' will receive a cancellation signal.
	// They can listen for this signal by checking the 'Done' channel on the 'ctx' object and perform any necessary cleanup operations before terminating.
	// This allows the application to shut down gracefully and avoid leaving resources in an inconsistent state.


	// Initialize a new MongoDB instance with the provided URI.
	// Log a fatal error and exit if the instance cannot be created.
	mongoDB, err := db.NewMongoDB(mongoURI)
	if err != nil {
		log.Fatal("Error creating MongoDB instance:", err)
	}
	// Ensure the MongoDB connection is closed properly when the application exits.
	defer func() {
		if err := mongoDB.Close(); err != nil {
			log.Println("Error closing MongoDB connection:", err)
		}
	}()

	// Attempt to create a new MongoDB instance using the provided URI.
	// 'db.NewMongoDB' is a function that takes a MongoDB URI and returns a new MongoDB instance and an error value.
	mongoDB, err := db.NewMongoDB(mongoURI)

	// Check if there was an error while creating the MongoDB instance.
	// If there is an error ('err' is not nil), log a fatal error and exit the application.
	// 'log.Fatal' function logs the error message and then calls 'os.Exit(1)' to terminate the program.
	if err != nil {
		log.Fatal("Error creating MongoDB instance:", err)
	}

	// Defer a function call to ensure the MongoDB connection is closed properly when the application exits.
	defer func() {
		// Attempt to close the MongoDB connection.
		// 'mongoDB.Close' is a method that cleans up the resources associated with the 'mongoDB' instance.
		// It returns an error if the closing process encounters any issues.
		if err := mongoDB.Close(); err != nil {
			// If there is an error while closing the connection, log the error message.
			// 'log.Println' function logs the error message but unlike 'log.Fatal', it does not terminate the program.
			log.Println("Error closing MongoDB connection:", err)
		}
	}() // The '()' at the end of the 'defer' statement is used to immediately invoke the function literal.

	// Create a new Kafka producer instance with the specified brokers and topic.
	// 'producer.NewKafkaProducer' is a function that takes a slice of broker addresses and a topic name,
	// and returns a new Kafka producer instance that can send messages to the Kafka topic.
	kafkaProducer := producer.NewKafkaProducer([]string{kafkaBrokers}, kafkaTopic)

	// Start the Kafka consumer in a separate goroutine to process incoming messages.
	// 'context.WithCancel' creates a new context that is a copy of the parent context (ctx) but with a new done channel.
	// The done channel is closed when the 'cancelConsumer' function is called, signaling that the context should be canceled.
	consumerCtx, cancelConsumer := context.WithCancel(ctx)
	// 'defer cancelConsumer()' ensures that the cancel function for the consumer context is called when the function exits,
	// which helps to stop the consumer gracefully when the application is shutting down.
	defer cancelConsumer()

	// 'go' keyword starts a new goroutine, which is a lightweight thread managed by the Go runtime.
	// Goroutines run concurrently with other functions or goroutines.
	go func() {
		// 'consumer.RunKafkaConsumer' is a function that takes a context, a slice of broker addresses, a topic name,
		// and a MongoDB instance. It listens for messages on the Kafka topic and processes them.
		// If an error occurs while running the consumer, it will be logged.
		if err := consumer.RunKafkaConsumer(consumerCtx, []string{kafkaBrokers}, kafkaTopic, mongoDB); err != nil {
			// 'log.Println' logs the error message but does not terminate the program.
			// This allows the application to continue running even if the consumer encounters an issue.
			log.Println("Error running Kafka consumer:", err)
		}
	}()

	// Start the HTTP server in a separate goroutine to handle incoming HTTP requests.
	// 'context.WithCancel' creates a new context that can be canceled. This context is derived from 'ctx', which is the main application context.
	httpCtx, cancelHTTP := context.WithCancel(ctx)
	// 'defer cancelHTTP()' schedules the 'cancelHTTP' function to be called when the current function exits.
	// This is used to signal the HTTP server to shutdown gracefully when the application is terminating.
	defer cancelHTTP()

	go func() {
		// 'http.RunHTTPServer' is a function that starts an HTTP server listening on the address specified by 'httpAddr'.
		// The server uses the Kafka producer 'kafkaProducer' for certain operations, such as publishing messages.
		// The 'httpCtx' is passed to manage the lifecycle of the HTTP server and allow for a graceful shutdown.
		if err := http.RunHTTPServer(httpCtx, httpAddr, *kafkaProducer); err != nil {
			// If there is an error while running the HTTP server, it will be logged using 'log.Println'.
			// This allows the application to continue running and log the error for troubleshooting.
			log.Println("Error running HTTP server:", err)
		}
	}()


	// Prepare to handle termination signals (e.g., SIGINT, SIGTERM) for graceful shutdown.
	// 'make(chan os.Signal, 1)' creates a new channel for receiving 'os.Signal' values with a buffer size of one.
	signalCh := make(chan os.Signal, 1)
	// 'signal.Notify' registers the given channel to receive notifications of the specified signals.
	// In this case, it's set up to receive SIGINT (interrupt from keyboard, Ctrl+C) and SIGTERM (termination signal from the operating system).
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// 'select' statement waits for multiple communication operations and proceeds with the first one that's ready.
	// Here, it's used to wait for a signal to be received on the 'signalCh' channel.
	select {
	case sig := <-signalCh:
		// When a signal is received, it's assigned to the variable 'sig' and the following code block is executed.
		fmt.Printf("Received signal %s. Shutting down...\n", sig)
		// 'cancel()' is called to trigger the cancellation of all contexts created with 'context.WithCancel'.
		// This includes 'ctx', 'consumerCtx', 'httpCtx', and any other derived contexts.
		// It signals all goroutines using these contexts to stop their work and helps in the graceful shutdown of the application.
		cancel()
	}
}
