package main

import (
	"testing"
)

// TestCalculatorAdd tests the Add function
func TestCalculatorAdd(t *testing.T) {
	calc := &Calculator{}

	result := calc.Add(2, 3)
	expected := 5

	if result != expected {
		t.Errorf("Add(2, 3) = %d; want %d", result, expected)
	}
}

// TestCalculatorMultiply tests the Multiply function
func TestCalculatorMultiply(t *testing.T) {
	calc := &Calculator{}

	result := calc.Multiply(4, 5)
	expected := 20

	if result != expected {
		t.Errorf("Multiply(4, 5) = %d; want %d", result, expected)
	}
}

// TestCalculatorDivide tests the Divide function
func TestCalculatorDivide(t *testing.T) {
	calc := &Calculator{}

	// Test normal division
	result, err := calc.Divide(10, 2)
	if err != nil {
		t.Errorf("Divide(10, 2) returned error: %v", err)
	}
	if result != 5 {
		t.Errorf("Divide(10, 2) = %d; want 5", result)
	}

	// Test division by zero
	_, err = calc.Divide(10, 0)
	if err == nil {
		t.Error("Divide(10, 0) should return error for division by zero")
	}
}

// TestIsEven tests the IsEven function
func TestIsEven(t *testing.T) {
	tests := []struct {
		input    int
		expected bool
	}{
		{2, true},
		{3, false},
		{0, true},
		{-2, true},
		{-3, false},
	}

	for _, test := range tests {
		result := IsEven(test.input)
		if result != test.expected {
			t.Errorf("IsEven(%d) = %t; want %t", test.input, result, test.expected)
		}
	}
}
