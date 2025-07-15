// main_test.go - THIS VERSION WILL CAUSE TEST FAILURES
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalculator_Add(t *testing.T) {
	calc := &Calculator{}

	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"positive numbers", 2, 3, 5},
		{"negative numbers", -2, -3, -5},
		{"zero", 0, 5, 5},
		{"mixed", -2, 3, 1},
		{"THIS WILL FAIL", 2, 2, 5}, // INTENTIONAL FAILURE: 2+2 != 5
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.Add(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestCalculator_Divide(t *testing.T) {
	calc := &Calculator{}

	// Test successful division
	result, err := calc.Divide(10, 2)
	if err != nil {
		t.Errorf("Divide(10, 2) returned error: %v", err)
	}
	if result != 5 {
		t.Errorf("Divide(10, 2) = %d, want 5", result)
	}

	// INTENTIONAL FAILURE: This test expects no error but will get one
	_, err = calc.Divide(10, 0)
	if err != nil {
		t.Error("THIS WILL FAIL: Divide(10, 0) should NOT return error (intentional wrong test)")
	}
}

func TestCalculator_Multiply(t *testing.T) {
	calc := &Calculator{}

	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"positive numbers", 3, 4, 12},
		{"negative numbers", -3, -4, 12},
		{"zero", 0, 5, 0},
		{"mixed", -3, 4, -12},
		{"INTENTIONAL FAILURE", 3, 3, 10}, // WRONG: 3*3 != 10
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.Multiply(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Multiply(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestIsEven(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want bool
	}{
		{"even positive", 4, true},
		{"odd positive", 5, false},
		{"even negative", -4, true},
		{"odd negative", -5, false},
		{"zero", 0, true},
		{"WRONG TEST", 3, true}, // INTENTIONAL FAILURE: 3 is not even
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEven(tt.n)
			if got != tt.want {
				t.Errorf("IsEven(%d) = %v, want %v", tt.n, got, tt.want)
			}
		})
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	// INTENTIONAL FAILURE: Expecting wrong status code
	if w.Code != http.StatusCreated {
		t.Errorf("healthHandler returned wrong status code: got %v want %v", w.Code, http.StatusCreated)
	}

	expected := "OK"
	if w.Body.String() != expected {
		t.Errorf("healthHandler returned wrong body: got %v want %v", w.Body.String(), expected)
	}
}

func TestCalculateHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/calculate", nil)
	w := httptest.NewRecorder()

	calculateHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("calculateHandler returned wrong status code: got %v want %v", w.Code, http.StatusOK)
	}

	// INTENTIONAL FAILURE: Expecting wrong content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("calculateHandler returned wrong content type: got %v want %v", contentType, "text/plain")
	}
}
