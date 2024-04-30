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

// GetNearbyDrivers retrieves nearby drivers based on the provided latitude and longitude from the MongoDB database.
func (db *MongoDB) GetNearbyDrivers(ctx context.Context, latitude, longitude string) ([]models.Driver, error) {
	// Connect to the 'drivers' collection in the 'database'.
	// This establishes a connection to the specific collection where driver data is stored.
	collection := db.client.Database("database").Collection("drivers")

	// Define a filter to find nearby drivers using the provided latitude and longitude.
	// The filter uses MongoDB's geospatial query operator '$near' to find documents (drivers)
	// that are within a certain distance from the given point.
	filter := bson.M{
		"location": bson.M{
			"$near": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{parseLongitude(longitude), parseLatitude(latitude)},
				},
				"$maxDistance": 1000, // Maximum distance in meters.
			},
		},
	}

	// Execute the query and retrieve the results.
	// A cursor is obtained which can be used to iterate over the results of the query.
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx) // Ensure that the cursor is closed after the operation is complete.

	// Iterate through the cursor and decode each document into a Driver struct.
	// This loop goes through each result returned by the query and decodes the data
	// into the 'Driver' struct which can be used by the application.
	var nearbyDrivers []models.Driver
	for cursor.Next(ctx) {
		var driver models.Driver
		if err := cursor.Decode(&driver); err != nil {
			return nil, err
		}
		nearbyDrivers = append(nearbyDrivers, driver) // Add the decoded driver to the slice of nearby drivers.
	}

	// Check if there were any errors during iteration.
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	// Return the slice of nearby drivers.
	return nearbyDrivers, nil
}


// Helper function to parse latitude string to float64.
func parseLatitude(latitude string) float64 {
	// Parse latitude string to float64.
	lat, err := strconv.ParseFloat(latitude, 64)
	if err != nil {
		// Handle conversion error appropriately, such as returning a default value or logging the error.
		return 0.0
	}
	return lat
}

// Helper function to parse longitude string to float64.
func parseLongitude(longitude string) float64 {
	// Parse longitude string to float64.
	lng, err := strconv.ParseFloat(longitude, 64)
	if err != nil {
		// Handle conversion error appropriately, such as returning a default value or logging the error.
		return 0.0
	}
	return lng
}
