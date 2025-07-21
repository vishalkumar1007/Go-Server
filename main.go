// This is qa branch code

// main.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// Calculator struct for basic math operations
type Calculator struct{}

// Add performs addition
func (c *Calculator) Add(a, b int) int {
	return a + b
}

// Divide performs division
func (c *Calculator) Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}

// Multiply performs multiplication
func (c *Calculator) Multiply(a, b int) int {
	return a * b
}

// IsEven checks if a number is even
func IsEven(n int) bool {
	return n%2 == 0
}

// HTTP handler for health check
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Server is live"))
}

// HTTP handler for calculator endpoint
func calculateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Calculator API is running", "status": "success"}`))
}

func main() {
	// Set up HTTP routes
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/calculate", calculateHandler)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
