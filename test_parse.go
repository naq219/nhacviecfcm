package main

import (
	"fmt"
	"time"
)

func main() {
	// Test parse format từ database
	testValue := "2025-11-05 18:44:49.659Z"
	
	// Thử các format khác nhau
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05.999Z", // PocketBase DateTime format với milliseconds
		"2006-01-02 15:04:05.000Z", // PocketBase DateTime format
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.9999999 -0700 MST", // Go time.String() format
	}
	
	fmt.Printf("Testing parse for: %s\n", testValue)
	for i, format := range formats {
		t, err := time.Parse(format, testValue)
		if err == nil {
			fmt.Printf("✓ Format %d (%s): %v\n", i, format, t)
		} else {
			fmt.Printf("✗ Format %d (%s): %v\n", i, format, err)
		}
	}
}