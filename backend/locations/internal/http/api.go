package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"locations/internal/models"
	"locations/internal/producer"
)

func LocationUpdateHandler(w http.ResponseWriter, r *http.Request, kafkaProducer producer.KafkaProducer) {
	var location models.LocationUpdate
	err := json.NewDecoder(r.Body).Decode(&location)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = kafkaProducer.ProduceLocationUpdate(r.Context(), location)
	if err != nil {
		http.Error(w, "Failed to produce Kafka message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func RunHTTPServer(ctx context.Context, addr string, kafkaProducer producer.KafkaProducer) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/location", func(w http.ResponseWriter, r *http.Request) {
		LocationUpdateHandler(w, r, kafkaProducer)
	})

	server := http.Server{
		Addr: addr,
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