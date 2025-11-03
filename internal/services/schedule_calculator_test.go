package services

import (
	"testing"
	"time"

	"remiaq/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestNewScheduleCalculator(t *testing.T) {
	lunarCalendar := NewLunarCalendar()
	calculator := NewScheduleCalculator(lunarCalendar)

	assert.NotNil(t, calculator)
	assert.Equal(t, lunarCalendar, calculator.lunarCalendar)
}

func TestScheduleCalculator_CalculateNextTrigger(t *testing.T) {
	lunarCalendar := NewLunarCalendar()
	calculator := NewScheduleCalculator(lunarCalendar)

	t.Run("should handle one-time reminder", func(t *testing.T) {
		now := time.Now()
		reminder := &models.Reminder{
			Type:          models.ReminderTypeOneTime,
			NextTriggerAt: now.Add(time.Hour),
		}

		result, err := calculator.CalculateNextTrigger(reminder, now)

		assert.NoError(t, err)
		assert.Equal(t, reminder.NextTriggerAt, result)
	})

	t.Run("should handle recurring reminder", func(t *testing.T) {
		now := time.Now()
		reminder := &models.Reminder{
			Type: models.ReminderTypeRecurring,
			RecurrencePattern: &models.RecurrencePattern{
				Type:            models.RecurrenceTypeDaily,
				IntervalSeconds: 0,
			},
			TriggerTimeOfDay: "09:00",
		}

		result, err := calculator.CalculateNextTrigger(reminder, now)

		assert.NoError(t, err)
		assert.True(t, result.After(now))
	})

	t.Run("should return error for recurring reminder without pattern", func(t *testing.T) {
		now := time.Now()
		reminder := &models.Reminder{
			Type:              models.ReminderTypeRecurring,
			RecurrencePattern: nil,
		}

		_, err := calculator.CalculateNextTrigger(reminder, now)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "recurrence pattern is required")
	})
}

func TestScheduleCalculator_calculateOneTime(t *testing.T) {
	calculator := NewScheduleCalculator(NewLunarCalendar())

	t.Run("should return existing NextTriggerAt", func(t *testing.T) {
		now := time.Now()
		triggerTime := now.Add(2 * time.Hour)
		reminder := &models.Reminder{
			NextTriggerAt: triggerTime,
		}

		result, err := calculator.calculateOneTime(reminder, now)

		assert.NoError(t, err)
		assert.Equal(t, triggerTime, result)
	})

	t.Run("should return fromTime when NextTriggerAt is zero", func(t *testing.T) {
		now := time.Now()
		reminder := &models.Reminder{
			NextTriggerAt: time.Time{},
		}

		result, err := calculator.calculateOneTime(reminder, now)

		assert.NoError(t, err)
		assert.Equal(t, now, result)
	})
}

func TestScheduleCalculator_calculateIntervalBased(t *testing.T) {
	calculator := NewScheduleCalculator(NewLunarCalendar())

	t.Run("should calculate from completion time when base_on=completion", func(t *testing.T) {
		now := time.Now()
		completedAt := now.Add(-time.Hour)
		reminder := &models.Reminder{
			RecurrencePattern: &models.RecurrencePattern{
				IntervalSeconds: 3600, // 1 hour
				BaseOn:          models.BaseOnCompletion,
			},
			LastCompletedAt: &completedAt,
		}

		result, err := calculator.calculateIntervalBased(reminder, now)

		assert.NoError(t, err)
		expected := completedAt.Add(time.Hour)
		assert.Equal(t, expected, result)
	})

	t.Run("should calculate from creation time when never completed", func(t *testing.T) {
		now := time.Now()
		created := now.Add(-2 * time.Hour)
		reminder := &models.Reminder{
			RecurrencePattern: &models.RecurrencePattern{
				IntervalSeconds: 3600, // 1 hour
				BaseOn:          models.BaseOnCompletion,
			},
			Created:         created,
			LastCompletedAt: nil,
		}

		result, err := calculator.calculateIntervalBased(reminder, now)

		assert.NoError(t, err)
		expected := created.Add(time.Hour)
		assert.Equal(t, expected, result)
	})

	t.Run("should calculate from fromTime for creation-based", func(t *testing.T) {
		now := time.Now()
		reminder := &models.Reminder{
			RecurrencePattern: &models.RecurrencePattern{
				IntervalSeconds: 3600, // 1 hour
				BaseOn:          models.BaseOnCreation,
			},
		}

		result, err := calculator.calculateIntervalBased(reminder, now)

		assert.NoError(t, err)
		expected := now.Add(time.Hour)
		assert.Equal(t, expected, result)
	})
}

func TestScheduleCalculator_calculateDaily(t *testing.T) {
	calculator := NewScheduleCalculator(NewLunarCalendar())

	t.Run("should return error when trigger_time_of_day is missing", func(t *testing.T) {
		now := time.Now()
		reminder := &models.Reminder{
			TriggerTimeOfDay: "",
		}

		_, err := calculator.calculateDaily(reminder, now)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "trigger_time_of_day is required")
	})

	t.Run("should calculate next day when time has passed", func(t *testing.T) {
		// Set current time to 10:00 AM
		now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		reminder := &models.Reminder{
			TriggerTimeOfDay: "09:00", // 9:00 AM (already passed)
		}

		result, err := calculator.calculateDaily(reminder, now)

		assert.NoError(t, err)
		// Should be tomorrow at 9:00 AM
		expected := time.Date(2024, 1, 16, 9, 0, 0, 0, time.UTC)
		assert.Equal(t, expected, result)
	})

	t.Run("should calculate today when time hasn't passed", func(t *testing.T) {
		// Set current time to 8:00 AM
		now := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
		reminder := &models.Reminder{
			TriggerTimeOfDay: "09:00", // 9:00 AM (not yet passed)
		}

		result, err := calculator.calculateDaily(reminder, now)

		assert.NoError(t, err)
		// Should be today at 9:00 AM
		expected := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
		assert.Equal(t, expected, result)
	})
}

func TestScheduleCalculator_calculateWeekly(t *testing.T) {
	calculator := NewScheduleCalculator(NewLunarCalendar())

	t.Run("should return error when trigger_time_of_day is missing", func(t *testing.T) {
		now := time.Now()
		reminder := &models.Reminder{
			TriggerTimeOfDay: "",
			RecurrencePattern: &models.RecurrencePattern{
				DayOfWeek: int(time.Monday),
			},
		}

		_, err := calculator.calculateWeekly(reminder, now)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "trigger_time_of_day is required")
	})

	t.Run("should calculate next Monday", func(t *testing.T) {
		// Set current time to Wednesday
		now := time.Date(2024, 1, 17, 10, 0, 0, 0, time.UTC) // Wednesday
		reminder := &models.Reminder{
			TriggerTimeOfDay: "09:00",
			RecurrencePattern: &models.RecurrencePattern{
				DayOfWeek: int(time.Monday),
			},
		}

		result, err := calculator.calculateWeekly(reminder, now)

		assert.NoError(t, err)
		// Should be next Monday at 9:00 AM
		expected := time.Date(2024, 1, 22, 9, 0, 0, 0, time.UTC)
		assert.Equal(t, expected, result)
	})

	t.Run("should handle same day when time hasn't passed", func(t *testing.T) {
		// Set current time to Monday 8:00 AM
		now := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC) // Monday
		reminder := &models.Reminder{
			TriggerTimeOfDay: "09:00",
			RecurrencePattern: &models.RecurrencePattern{
				DayOfWeek: int(time.Monday),
			},
		}

		result, err := calculator.calculateWeekly(reminder, now)

		assert.NoError(t, err)
		// Should be today at 9:00 AM
		expected := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
		assert.Equal(t, expected, result)
	})
}

func TestScheduleCalculator_calculateMonthly(t *testing.T) {
	calculator := NewScheduleCalculator(NewLunarCalendar())

	t.Run("should return error when trigger_time_of_day is missing for solar calendar", func(t *testing.T) {
		now := time.Now()
		reminder := &models.Reminder{
			TriggerTimeOfDay: "",
			CalendarType:     models.CalendarTypeSolar,
			RecurrencePattern: &models.RecurrencePattern{
				DayOfMonth: 15,
			},
		}

		_, err := calculator.calculateMonthly(reminder, now)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "trigger_time_of_day is required")
	})

	t.Run("should calculate next month when day has passed", func(t *testing.T) {
		// Set current time to January 20th
		now := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC)
		reminder := &models.Reminder{
			TriggerTimeOfDay: "09:00",
			CalendarType:     models.CalendarTypeSolar,
			RecurrencePattern: &models.RecurrencePattern{
				DayOfMonth: 15, // 15th (already passed)
			},
		}

		result, err := calculator.calculateMonthly(reminder, now)

		assert.NoError(t, err)
		// Should be February 15th at 9:00 AM
		expected := time.Date(2024, 2, 15, 9, 0, 0, 0, time.UTC)
		assert.Equal(t, expected, result)
	})

	t.Run("should calculate current month when day hasn't passed", func(t *testing.T) {
		// Set current time to January 10th
		now := time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)
		reminder := &models.Reminder{
			TriggerTimeOfDay: "09:00",
			CalendarType:     models.CalendarTypeSolar,
			RecurrencePattern: &models.RecurrencePattern{
				DayOfMonth: 15, // 15th (not yet passed)
			},
		}

		result, err := calculator.calculateMonthly(reminder, now)

		assert.NoError(t, err)
		// Should be January 15th at 9:00 AM
		expected := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
		assert.Equal(t, expected, result)
	})

	t.Run("should handle lunar calendar", func(t *testing.T) {
		now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		reminder := &models.Reminder{
			TriggerTimeOfDay: "09:00",
			CalendarType:     models.CalendarTypeLunar,
			RecurrencePattern: &models.RecurrencePattern{
				DayOfMonth: 15,
			},
		}

		result, err := calculator.calculateMonthly(reminder, now)

		// Should not error (lunar calculation is complex, just check it doesn't crash)
		assert.NoError(t, err)
		assert.True(t, result.After(now) || result.Equal(now))
	})
}

func TestParseTimeOfDay(t *testing.T) {
	testCases := []struct {
		name        string
		timeStr     string
		expectError bool
		expectedH   int
		expectedM   int
	}{
		{"valid time", "09:30", false, 9, 30},
		{"midnight", "00:00", false, 0, 0},
		{"noon", "12:00", false, 12, 0},
		{"evening", "23:59", false, 23, 59},
		{"single digit hour", "9:30", false, 9, 30},
		{"invalid hour", "25:00", true, 0, 0},
		{"invalid minute", "12:60", true, 0, 0},
		{"empty string", "", true, 0, 0},
		{"wrong format", "12-30", true, 0, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseTimeOfDay(tc.timeStr)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedH, result.Hour())
				assert.Equal(t, tc.expectedM, result.Minute())
			}
		})
	}
}

func TestScheduleCalculator_calculateLunarLastDay(t *testing.T) {
	calculator := NewScheduleCalculator(NewLunarCalendar())

	t.Run("should calculate last day of lunar month", func(t *testing.T) {
		now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		reminder := &models.Reminder{
			TriggerTimeOfDay: "09:00",
		}

		result, err := calculator.calculateLunarLastDay(reminder, now)

		assert.NoError(t, err)
		assert.True(t, result.After(now))
		// Should have the correct time of day
		assert.Equal(t, 9, result.Hour())
		assert.Equal(t, 0, result.Minute())
	})

	t.Run("should handle missing trigger_time_of_day", func(t *testing.T) {
		now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		reminder := &models.Reminder{
			TriggerTimeOfDay: "",
		}

		result, err := calculator.calculateLunarLastDay(reminder, now)

		assert.NoError(t, err)
		assert.True(t, result.After(now))
	})
}

// Integration tests
func TestScheduleCalculator_Integration(t *testing.T) {
	calculator := NewScheduleCalculator(NewLunarCalendar())

	t.Run("should handle complex recurring pattern", func(t *testing.T) {
		now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		reminder := &models.Reminder{
			Type:             models.ReminderTypeRecurring,
			TriggerTimeOfDay: "09:00",
			CalendarType:     models.CalendarTypeSolar,
			RecurrencePattern: &models.RecurrencePattern{
				Type:       models.RecurrenceTypeWeekly,
				DayOfWeek:  int(time.Friday),
			},
		}

		result, err := calculator.CalculateNextTrigger(reminder, now)

		assert.NoError(t, err)
		assert.True(t, result.After(now))
		assert.Equal(t, time.Friday, result.Weekday())
		assert.Equal(t, 9, result.Hour())
		assert.Equal(t, 0, result.Minute())
	})

	t.Run("should handle unsupported recurrence type", func(t *testing.T) {
		now := time.Now()
		reminder := &models.Reminder{
			Type: models.ReminderTypeRecurring,
			RecurrencePattern: &models.RecurrencePattern{
				Type: "unsupported_type",
			},
		}

		_, err := calculator.CalculateNextTrigger(reminder, now)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported recurrence type")
	})
}

// Benchmark tests
func BenchmarkScheduleCalculator_CalculateDaily(b *testing.B) {
	calculator := NewScheduleCalculator(NewLunarCalendar())
	now := time.Now()
	reminder := &models.Reminder{
		TriggerTimeOfDay: "09:00",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = calculator.calculateDaily(reminder, now)
	}
}

func BenchmarkScheduleCalculator_CalculateWeekly(b *testing.B) {
	calculator := NewScheduleCalculator(NewLunarCalendar())
	now := time.Now()
	reminder := &models.Reminder{
		TriggerTimeOfDay: "09:00",
		RecurrencePattern: &models.RecurrencePattern{
			DayOfWeek: int(time.Monday),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = calculator.calculateWeekly(reminder, now)
	}
}

func BenchmarkParseTimeOfDay(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseTimeOfDay("09:30")
	}
}