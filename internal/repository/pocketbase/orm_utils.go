package pocketbase

import (
	"time"
)

// parseTime converts a time string from PocketBase format to time.Time
// PocketBase stores time in RFC3339 format (e.g., "2024-01-01 12:00:00.000Z")
func parseTime(timeStr string) time.Time {
	if timeStr == "" {
		return time.Time{}
	}
	
	// Try parsing with different formats
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05.000Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t
		}
	}
	
	// If all parsing fails, return zero time
	return time.Time{}
}