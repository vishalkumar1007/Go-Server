package main

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// TestMain is the entry point for testing.
// It waits 5 minutes and then fails the test intentionally.
func TestMain(m *testing.M) {
	fmt.Println("🕒 Waiting 5 minutes before failing intentionally...")
	time.Sleep(5 * time.Minute)

	fmt.Println("❌ Test intentionally failed after 5-minute delay.")
	os.Exit(1) // Exit with non-zero status to mark test as failed
}
