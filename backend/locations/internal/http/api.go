package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"locations/internal/models"
	"locations/internal/producer"
)

// LocationUpdateHandler handles POST requests to update location data.
// It decodes the JSON request body into a LocationUpdate model, validates the data,
// and uses a KafkaProducer to send the location update to a Kafka topic.
func LocationUpdateHandler(w http.ResponseWriter, r *http.Request, kafkaProducer producer.KafkaProducer) {
	var location models.LocationUpdate
	err := json.NewDecoder(r.Body).Decode(&location)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate location data
	if err := validateLocationData(location); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = kafkaProducer.ProduceLocationUpdate(r.Context(), location)
	if err != nil {
		http.Error(w, "Failed to produce Kafka message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetLocationHandler handles GET requests to retrieve location data by ID.
// It extracts the location ID from the query parameters, retrieves the location data from the database,
// and encodes the location data into a JSON response.
func GetLocationHandler(w http.ResponseWriter, r *http.Request, db Database) {
	locationID := r.URL.Query().Get("id")
	if locationID == "" {
		http.Error(w, "Location ID is required", http.StatusBadRequest)
		return
	}

	location, err := db.GetLocationByID(r.Context(), locationID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get location: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(location)
}

// UpdateLocationHandler handles PUT requests to modify existing location data.
// It decodes the JSON request body into a LocationUpdate model, validates the data,
// and updates the location data in the database using the provided ID.
func UpdateLocationHandler(w http.ResponseWriter, r *http.Request, db Database) {
	var location models.LocationUpdate
	err := json.NewDecoder(r.Body).Decode(&location)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	locationID := r.URL.Query().Get("id")
	if locationID == "" {
		http.Error(w, "Location ID is required", http.StatusBadRequest)
		return
	}

	// Validate location data
	if err := validateLocationData(location); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.UpdateLocation(r.Context(), locationID, location)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update location: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// NearbyDriversHandler handles the request to retrieve nearby drivers based on longitude and latitude.
func NearbyDriversHandler(w http.ResponseWriter, r *http.Request, db db.Database) {
	// Parse the request parameters (latitude and longitude).
	latitude := r.URL.Query().Get("latitude")
	longitude := r.URL.Query().Get("longitude")

	// Validate the request parameters.
	if latitude == "" || longitude == "" {
		http.Error(w, "Latitude and longitude are required parameters", http.StatusBadRequest)
		return
	}

	// Convert latitude and longitude to appropriate data types (float64 or other).
	// Perform any necessary validation or sanitization.

	// Query the database for nearby drivers based on the provided latitude and longitude.
	nearbyDrivers, err := db.GetNearbyDrivers(r.Context(), latitude, longitude)
	if err != nil {
		http.Error(w, "Failed to retrieve nearby drivers", http.StatusInternalServerError)
		return
	}

	// Marshal the retrieved nearby drivers into JSON format.
	response, err := json.Marshal(nearbyDrivers)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to indicate JSON format.
	w.Header().Set("Content-Type", "application/json")
	// Write the JSON response to the HTTP response writer.
	w.Write(response)
}

// RunHTTPServer starts an HTTP server with handlers for location data.
// It listens for incoming HTTP requests and routes them to the appropriate handlers based on the request method.
// The server runs in a separate goroutine and can be shut down gracefully when the context is canceled.
func RunHTTPServer(ctx context.Context, addr string, kafkaProducer producer.KafkaProducer, db Database) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/location", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			LocationUpdateHandler(w, r, kafkaProducer)
		case http.MethodGet:
			GetLocationHandler(w, r, db)
		case http.MethodPut:
			UpdateLocationHandler(w, r, db)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/nearby", func(w http.ResponseWriter, r *http.Request) {
		NearbyDriversHandler(w, r, db)
	})

	server := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	fmt.Printf("HTTP server listening on %s\n", addr)
	return server.ListenAndServe()
}

func validateLocationData(location models.LocationUpdate) error {
	if location.DriverID == "" {
		return fmt.Errorf("driver ID is required")
	}
	if location.Latitude < -90 || location.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if location.Longitude < -180 || location.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	if location.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	return nil
}
