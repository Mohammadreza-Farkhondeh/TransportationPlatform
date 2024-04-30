package db

import (
	"context" // Provides functionality to define a deadline or cancellation signal for operations.
	"fmt"     // Implements formatted I/O functions.
	"time"    // Provides functionality for measuring and displaying time.

	"go.mongodb.org/mongo-driver/mongo"         // Official MongoDB driver for Go.
	"go.mongodb.org/mongo-driver/mongo/options" // Provides options to configure the MongoDB driver.
	"locations/internal/models"                // Internal package for data models.
)

// MongoDB wraps the official MongoDB client.
type MongoDB struct {
	client *mongo.Client // The client field holds the connection to the MongoDB instance.
}

// NewMongoDB creates a new MongoDB client and establishes a connection to the database.
// It takes a MongoDB URI and returns a connected MongoDB instance or an error if the connection fails.
func NewMongoDB(uri string) (*MongoDB, error) {
	// Set client options using the provided URI.
	clientOptions := options.Client().ApplyURI(uri)
	// Attempt to connect to MongoDB using the specified client options.
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	// Create a context with a 10-second timeout for the ping operation.
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	// Ping the MongoDB server to verify the connection is active.
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to MongoDB")

	// Return a new MongoDB instance with the established client.
	return &MongoDB{client: client}, nil
}

// InsertLocationUpdate inserts a location update into the MongoDB database.
// It validates the update, connects to the 'locations' collection, and inserts the update.
func (db *MongoDB) InsertLocationUpdate(ctx context.Context, update models.LocationUpdate) error {
	// Connect to the 'locations' collection in the 'database'.
	collection := db.client.Database("database").Collection("locations")

	// Insert the location update into the collection.
	_, err := collection.InsertOne(ctx, update)
	if err != nil {
		return fmt.Errorf("failed to insert location update: %w", err)
	}

	return nil
}

// GetLocationByID retrieves a location update by its ID from the MongoDB database.
// It connects to the 'locations' collection and searches for the update by ID.
func (db *MongoDB) GetLocationByID(ctx context.Context, id string) (*models.LocationUpdate, error) {
	// Connect to the 'locations' collection in the 'database'.
	collection := db.client.Database("database").Collection("locations")

	var location models.LocationUpdate
	// Find the document with the matching ID and decode it into the location variable.
	err := collection.FindOne(ctx, bson.M{"id": id}).Decode(&location)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve location: %w", err)
	}

	return &location, nil
}

// UpdateLocation updates a location update in the MongoDB database.
// It connects to the 'locations' collection and replaces the existing document with the new update.
func (db *MongoDB) UpdateLocation(ctx context.Context, id string, update models.LocationUpdate) error {
	// Connect to the 'locations' collection in the 'database'.
	collection := db.client.Database("database").Collection("locations")

	// Replace the existing document with the new update.
	_, err := collection.ReplaceOne(ctx, bson.M{"id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to update location: %w", err)
	}

	return nil
}
