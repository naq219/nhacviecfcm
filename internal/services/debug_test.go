package services

import (
	"fmt"
	"testing"
	"time"

	"remiaq/internal/models"
)

func TestDebug_Day31ToFeb(t *testing.T) {
	reminder := models.Reminder{
		OriginTime: time.Date(2025, 1, 31, 8, 0, 0, 0, time.UTC),
		RecurrencePattern: &models.RecurrencePattern{
			Type:     models.RecurrenceTypeMonthly,
			Interval: 1,
		},
	}

	lastTime := time.Date(2025, 1, 31, 8, 0, 0, 0, time.UTC)
	now := time.Date(2025, 2, 1, 9, 0, 0, 0, time.UTC)

	next, err := calcNextSolarMonthly(reminder, lastTime, now)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	fmt.Printf("LastTime: %v\n", lastTime)
	fmt.Printf("Now:      %v\n", now)
	fmt.Printf("Next:     %v\n", next)
	fmt.Printf("Expected: Feb 28, 2025 08:00:00 UTC\n")

	expected := time.Date(2025, 2, 28, 8, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("\n❌ FAIL:\nExpected: %v\nGot:      %v\nDiff: %v", expected, next, next.Sub(expected))
	} else {
		t.Logf("✅ PASS")
	}
}
