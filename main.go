package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

// DefaultResponse represents the default route response
type DefaultResponse struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Service:   "go-server",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding health response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	response := DefaultResponse{
		Message:   "Welcome to the Go Server!",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding default response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func main() {
	// Create a new HTTP multiplexer
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/", defaultHandler)

	// Server configuration
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the server
	log.Printf("Starting server on port 8080...")
	log.Printf("Health check endpoint: http://localhost:8080/health")
	log.Printf("Default endpoint: http://localhost:8080/")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// this is command to create a new SSH key pair
// ssh-keygen -t rsa -b 4096 -C "vishalkumarnke93@gmail.com"
