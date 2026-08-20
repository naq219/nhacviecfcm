package services

import (
	"testing"
	"time"

	"remiaq/internal/models"
)

// Helper function to create time easily
func createTime(year, month, day, hour, minute int) time.Time {
	return time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC)
}

// ==================== Test calcNextDailyTime ====================

func TestCalcNextDailyTime_SimpleDaily(t *testing.T) {
	reminder := models.Reminder{
		OriginTime: createTime(2025, 1, 1, 8, 0), // 8:00 AM
		RecurrencePattern: &models.RecurrencePattern{
			Type:     models.RecurrenceTypeDaily,
			Interval: 1, // Every day
		},
	}

	lastTime := createTime(2025, 1, 5, 8, 0)
	now := createTime(2025, 1, 6, 7, 0) // Before 8:00 AM

	next, err := calcNextDailyTime(reminder, lastTime, now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := createTime(2025, 1, 6, 8, 0)
	if !next.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, next)
	}
}

func TestCalcNextDailyTime_EveryTwoDays(t *testing.T) {
	reminder := models.Reminder{
		OriginTime: createTime(2025, 1, 1, 9, 30),
		RecurrencePattern: &models.RecurrencePattern{
			Type:     models.RecurrenceTypeDaily,
			Interval: 2, // Every 2 days
		},
	}

	lastTime := createTime(2025, 1, 5, 9, 30)
	now := createTime(2025, 1, 6, 10, 0) // After the time

	next, err := calcNextDailyTime(reminder, lastTime, now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := createTime(2025, 1, 7, 9, 30) // +2 days
	if !next.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, next)
	}
}

func TestCalcNextDailyTime_CrossMonth(t *testing.T) {
	reminder := models.Reminder{
		OriginTime: createTime(2025, 1, 1, 10, 0),
		RecurrencePattern: &models.RecurrencePattern{
			Type:     models.RecurrenceTypeDaily,
			Interval: 1,
		},
	}

	lastTime := createTime(2025, 1, 31, 10, 0)
	now := createTime(2025, 2, 1, 9, 0)

	next, err := calcNextDailyTime(reminder, lastTime, now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := createTime(2025, 2, 1, 10, 0)
	if !next.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, next)
	}
}

// ==================== Test calcNextWeekly ====================

func TestCalcNextWeekly_SameWeek(t *testing.T) {
	// OriginTime = Friday Jan 3, 2025
	reminder := models.Reminder{
		OriginTime: createTime(2025, 1, 3, 10, 0), // Friday
		RecurrencePattern: &models.RecurrencePattern{
			Type:     models.RecurrenceTypeWeekly,
			Interval: 1,
		},
	}

	// Monday Jan 6, 2025
	lastTime := createTime(2025, 1, 6, 10, 0)
	now := createTime(2025, 1, 6, 11, 0)

	next, err := calcNextWeekly(reminder, lastTime, now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Next Friday (Jan 10)
	expected := createTime(2025, 1, 10, 10, 0)
	if !next.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, next)
	}
}

func TestCalcNextWeekly_EveryTwoWeeks(t *testing.T) {
	// OriginTime = Monday Jan 6, 2025
	reminder := models.Reminder{
		OriginTime: createTime(2025, 1, 6, 14, 0), // Monday
		RecurrencePattern: &models.RecurrencePattern{
			Type:     models.RecurrenceTypeWeekly,
			Interval: 2, // Every 2 weeks
		},
	}

	// Monday Jan 6
	lastTime := createTime(2025, 1, 6, 14, 0)
	now := createTime(2025, 1, 7, 10, 0)

	next, err := calcNextWeekly(reminder, lastTime, now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Monday Jan 20 (2 weeks later)
	expected := createTime(2025, 1, 20, 14, 0)
	if !next.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, next)
	}
}

func TestCalcNextWeekly_Sunday(t *testing.T) {
	// OriginTime = Sunday Jan 5, 2025
	reminder := models.Reminder{
		OriginTime: createTime(2025, 1, 5, 9, 0), // Sunday
		RecurrencePattern: &models.RecurrencePattern{
			Type:     models.RecurrenceTypeWeekly,
			Interval: 1,
		},
	}

	// Saturday Jan 11
	lastTime := createTime(2025, 1, 11, 9, 0)
	now := createTime(2025, 1, 11, 10, 0)

	next, err := calcNextWeekly(reminder, lastTime, now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Sunday Jan 12
	expected := createTime(2025, 1, 12, 9, 0)
	if !next.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, next)
	}
}

// ==================== Test calcNextSolarMonthly ====================

func TestCalcNextSolarMonthly_SimpleMonthly(t *testing.T) {
	reminder := models.Reminder{
		OriginTime: createTime(2025, 1, 15, 10, 30), // 15th day
		RecurrencePattern: &models.RecurrencePattern{
			Type:     models.RecurrenceTypeMonthly,
			Interval: 1,
		},
	}

	lastTime := createTime(2025, 1, 15, 10, 30)
	now := createTime(2025, 2, 1, 12, 0)

	next, err := calcNextSolarMonthly(reminder, lastTime, now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := createTime(2025, 2, 15, 10, 30)
	if !next.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, next)
	}
}

func TestCalcNextSolarMonthly_Day31ToFebruary(t *testing.T) {
	reminder := models.Reminder{
		OriginTime: createTime(2025, 1, 31, 8, 0), // 31st day
		RecurrencePattern: &models.RecurrencePattern{
			Type:     models.RecurrenceTypeMonthly,
			Interval: 1,
		},
	}

	lastTime := createTime(2025, 1, 31, 8, 0)
	now := createTime(2025, 2, 1, 9, 0)

	next, err := calcNextSolarMonthly(reminder, lastTime, now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// February only has 28 days in 2025, should use last day
	expected := createTime(2025, 2, 28, 8, 0)
	if !next.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, next)
	}
}

func TestCalcNextSolarMonthly_EveryTwoMonths(t *testing.T) {
	reminder := models.Reminder{
		OriginTime: createTime(2025, 1, 10, 15, 0),
		RecurrencePattern: &models.RecurrencePattern{
			Type:     models.RecurrenceTypeMonthly,
			Interval: 2, // Every 2 months
		},
	}

	lastTime := createTime(2025, 1, 10, 15, 0)
	now := createTime(2025, 2, 15, 10, 0)

	next, err := calcNextSolarMonthly(reminder, lastTime, now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// March 10 (2 months after Jan)
	expected := createTime(2025, 3, 10, 15, 0)
	if !next.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, next)
	}
}

func TestCalcNextSolarMonthly_LeapYear(t *testing.T) {
	reminder := models.Reminder{
		OriginTime: createTime(2024, 1, 29, 12, 0), // 29th
		RecurrencePattern: &models.RecurrencePattern{
			Type:     models.RecurrenceTypeMonthly,
			Interval: 1,
		},
	}

	lastTime := createTime(2024, 1, 29, 12, 0)
	now := createTime(2024, 2, 15, 10, 0)

	next, err := calcNextSolarMonthly(reminder, lastTime, now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// February 2024 has 29 days (leap year)
	expected := createTime(2024, 2, 29, 12, 0)
	if !next.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, next)
	}
}

// ==================== Test calcNextIntervalSeconds ====================

func TestCalcNextIntervalSeconds_OneHour(t *testing.T) {
	reminder := models.Reminder{
		RecurrencePattern: &models.RecurrencePattern{
			Type:            models.RecurrenceTypeIntervalSeconds,
			IntervalSeconds: 3600, // 1 hour
		},
	}

	lastTime := createTime(2025, 1, 1, 10, 0)
	now := createTime(2025, 1, 1, 10, 30)

	next, err := calcNextIntervalSeconds(reminder, lastTime, now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := createTime(2025, 1, 1, 11, 0) // +1 hour
	if !next.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, next)
	}
}

func TestCalcNextIntervalSeconds_SystemDowntime(t *testing.T) {
	reminder := models.Reminder{
		RecurrencePattern: &models.RecurrencePattern{
			Type:            models.RecurrenceTypeIntervalSeconds,
			IntervalSeconds: 600, // 10 minutes
		},
	}

	lastTime := createTime(2025, 1, 1, 10, 0)
	now := createTime(2025, 1, 1, 12, 0) // 2 hours later

	next, err := calcNextIntervalSeconds(reminder, lastTime, now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should catch up: 10:00 + 10min repeated until > 12:00
	// Expected: 12:00 or 12:10 (next interval after now)
	if !next.After(now) {
		t.Errorf("Next time should be after now. Got %v", next)
	}

	// Check it's a multiple of interval from lastTime
	diff := next.Sub(lastTime).Seconds()
	if int(diff)%600 != 0 {
		t.Errorf("Next time should be a multiple of 600 seconds from lastTime. Diff: %v", diff)
	}
}

// ==================== Test calcNextSolarLastDayOfMonth ====================

func TestCalcNextSolarLastDayOfMonth_SameMonth(t *testing.T) {
	// January 15, should return January 31
	now := createTime(2025, 1, 15, 14, 30)

	next, err := calcNextSolarLastDayOfMonth(now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := createTime(2025, 1, 31, 14, 30)
	if !next.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, next)
	}
}

func TestCalcNextSolarLastDayOfMonth_AlreadyLastDay(t *testing.T) {
	// January 31, should return February 28
	now := createTime(2025, 1, 31, 10, 0)

	next, err := calcNextSolarLastDayOfMonth(now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := createTime(2025, 2, 28, 10, 0)
	if !next.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, next)
	}
}

func TestCalcNextSolarLastDayOfMonth_February(t *testing.T) {
	// February 15, should return February 28 (2025 not leap year)
	now := createTime(2025, 2, 15, 9, 0)

	next, err := calcNextSolarLastDayOfMonth(now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := createTime(2025, 2, 28, 9, 0)
	if !next.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, next)
	}
}

func TestCalcNextSolarLastDayOfMonth_LeapYearFebruary(t *testing.T) {
	// February 15, 2024 (leap year), should return February 29
	now := createTime(2024, 2, 15, 11, 0)

	next, err := calcNextSolarLastDayOfMonth(now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := createTime(2024, 2, 29, 11, 0)
	if !next.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, next)
	}
}

// ==================== Test Tinhtoan_NextRecurringV2 Integration ====================

func TestTinhtoan_NextRecurringV2_Daily(t *testing.T) {
	reminder := models.Reminder{
		OriginTime: createTime(2025, 1, 1, 8, 0),
		RecurrencePattern: &models.RecurrencePattern{
			Type:     models.RecurrenceTypeDaily,
			Interval: 1,
		},
	}

	now := createTime(2025, 1, 5, 9, 0)

	next, err := Tinhtoan_NextRecurringV2(reminder, now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should return tomorrow at 8:00
	if !next.After(now) {
		t.Errorf("Next time should be after now. Got %v", next)
	}
}

func TestTinhtoan_NextRecurringV2_Weekly(t *testing.T) {
	// OriginTime = Wednesday Jan 1, 2025
	reminder := models.Reminder{
		OriginTime: createTime(2025, 1, 1, 10, 0), // Wednesday
		RecurrencePattern: &models.RecurrencePattern{
			Type:     models.RecurrenceTypeWeekly,
			Interval: 1,
		},
	}

	now := createTime(2025, 1, 5, 12, 0) // Sunday

	next, err := Tinhtoan_NextRecurringV2(reminder, now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should return next Wednesday (Jan 8) at 10:00
	expected := createTime(2025, 1, 8, 10, 0)
	if !next.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, next)
	}
}

func TestTinhtoan_NextRecurringV2_NilPattern(t *testing.T) {
	reminder := models.Reminder{
		OriginTime:        createTime(2025, 1, 1, 8, 0),
		RecurrencePattern: nil,
	}

	now := createTime(2025, 1, 5, 9, 0)

	_, err := Tinhtoan_NextRecurringV2(reminder, now)
	if err == nil {
		t.Error("Expected error for nil RecurrencePattern")
	}
}

func TestTinhtoan_NextRecurringV2_FromAPIComplete_ShouldError(t *testing.T) {
	reminder := models.Reminder{
		From:       "api_complete",
		OriginTime: createTime(2025, 1, 1, 8, 0),
		RecurrencePattern: &models.RecurrencePattern{
			Type:     models.RecurrenceTypeDaily,
			Interval: 1,
		},
	}

	now := createTime(2025, 1, 5, 9, 0)

	_, err := Tinhtoan_NextRecurringV2(reminder, now)
	if err == nil {
		t.Error("Expected error when From='api_complete'")
	}
}

// ==================== Test Edge Cases ====================

func TestCalcNextDaily_ZeroInterval(t *testing.T) {
	reminder := models.Reminder{
		OriginTime: createTime(2025, 1, 1, 8, 0),
		RecurrencePattern: &models.RecurrencePattern{
			Type:     models.RecurrenceTypeDaily,
			Interval: 0, // Invalid
		},
	}

	lastTime := createTime(2025, 1, 5, 8, 0)
	now := createTime(2025, 1, 6, 7, 0)

	// Should handle gracefully (treat as 1 or error)
	_, err := calcNextDailyTime(reminder, lastTime, now)
	// Depends on implementation, might error or default to 1
	// For now just checking it doesn't panic
	_ = err
}

func TestCalcNextWeekly_NoOriginTime(t *testing.T) {
	reminder := models.Reminder{
		OriginTime: time.Time{}, // Zero time
		RecurrencePattern: &models.RecurrencePattern{
			Type:      models.RecurrenceTypeWeekly,
			Interval:  1,
			DayOfWeek: 1,
		},
	}

	lastTime := createTime(2025, 1, 6, 10, 0)
	now := createTime(2025, 1, 7, 10, 0)

	_, err := calcNextWeekly(reminder, lastTime, now)
	if err == nil {
		t.Error("Expected error for zero OriginTime")
	}
}

// ==================== Benchmark Tests ====================

func BenchmarkCalcNextDailyTime(b *testing.B) {
	reminder := models.Reminder{
		OriginTime: createTime(2025, 1, 1, 8, 0),
		RecurrencePattern: &models.RecurrencePattern{
			Type:     models.RecurrenceTypeDaily,
			Interval: 1,
		},
	}

	lastTime := createTime(2025, 1, 5, 8, 0)
	now := createTime(2025, 1, 6, 7, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = calcNextDailyTime(reminder, lastTime, now)
	}
}

func BenchmarkCalcNextWeekly(b *testing.B) {
	reminder := models.Reminder{
		OriginTime: createTime(2025, 1, 1, 10, 0),
		RecurrencePattern: &models.RecurrencePattern{
			Type:      models.RecurrenceTypeWeekly,
			Interval:  1,
			DayOfWeek: 1,
		},
	}

	lastTime := createTime(2025, 1, 6, 10, 0)
	now := createTime(2025, 1, 7, 10, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = calcNextWeekly(reminder, lastTime, now)
	}
}

func BenchmarkCalcNextSolarMonthly(b *testing.B) {
	reminder := models.Reminder{
		OriginTime: createTime(2025, 1, 15, 10, 30),
		RecurrencePattern: &models.RecurrencePattern{
			Type:     models.RecurrenceTypeMonthly,
			Interval: 1,
		},
	}

	lastTime := createTime(2025, 1, 15, 10, 30)
	now := createTime(2025, 2, 1, 12, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = calcNextSolarMonthly(reminder, lastTime, now)
	}
}
