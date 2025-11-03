package test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"remiaq/config"
	"remiaq/internal/models"
	"remiaq/internal/services"
)

// MockFCMService implements basic FCM interface for testing
type MockFCMService struct{}

func (m *MockFCMService) SendNotification(token, title, body string) error {
	return nil
}

func TestIntegration_ConfigAndServices(t *testing.T) {
	// Set test environment
	os.Setenv("ENVIRONMENT", "testing")
	os.Setenv("WORKER_INTERVAL", "5")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("WORKER_INTERVAL")
	}()

	t.Run("should load and validate config", func(t *testing.T) {
		cfg, err := config.Load()
		require.NoError(t, err)
		assert.True(t, cfg.IsTesting())
		assert.Equal(t, 5, cfg.WorkerInterval)
	})

	t.Run("should initialize lunar calendar service", func(t *testing.T) {
		lunarCalendar := services.NewLunarCalendar()
		assert.NotNil(t, lunarCalendar)

		// Test basic lunar calendar functionality
		testTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		lunarDate := lunarCalendar.SolarToLunar(testTime)
		assert.NotNil(t, lunarDate)
		assert.True(t, lunarDate.Year > 0)
		assert.True(t, lunarDate.Month >= 1 && lunarDate.Month <= 12)
		assert.True(t, lunarDate.Day >= 1 && lunarDate.Day <= 30)
	})

	t.Run("should initialize schedule calculator", func(t *testing.T) {
		lunarCalendar := services.NewLunarCalendar()
		scheduleCalc := services.NewScheduleCalculator(lunarCalendar)
		assert.NotNil(t, scheduleCalc)

		// Create test reminder for daily recurrence
		testTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
		reminder := &models.Reminder{
			Type:             models.ReminderTypeRecurring,
			TriggerTimeOfDay: "10:00",
			CalendarType:     models.CalendarTypeSolar,
			RecurrencePattern: &models.RecurrencePattern{
				Type: models.RecurrenceTypeDaily,
			},
		}

		// Test solar schedule calculation
		nextTime, err := scheduleCalc.CalculateNextTrigger(reminder, testTime)
		assert.NoError(t, err)
		assert.True(t, nextTime.After(testTime))

		// Test lunar schedule calculation
		reminder.CalendarType = models.CalendarTypeLunar
		nextLunar, err := scheduleCalc.CalculateNextTrigger(reminder, testTime)
		assert.NoError(t, err)
		assert.True(t, nextLunar.After(testTime))
	})
}

func TestIntegration_ConfigValidation(t *testing.T) {
	t.Run("should reject invalid environment", func(t *testing.T) {
		os.Setenv("ENVIRONMENT", "invalid_env")
		defer os.Unsetenv("ENVIRONMENT")

		cfg, err := config.Load()
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "Environment")
	})

	t.Run("should reject invalid worker interval", func(t *testing.T) {
		os.Setenv("WORKER_INTERVAL", "-1")
		defer os.Unsetenv("WORKER_INTERVAL")

		cfg, err := config.Load()
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "WorkerInterval")
	})

	t.Run("should reject invalid server address", func(t *testing.T) {
		os.Setenv("SERVER_ADDR", "invalid_address")
		defer os.Unsetenv("SERVER_ADDR")

		cfg, err := config.Load()
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "ServerAddr")
	})
}

func TestIntegration_ServiceInteraction(t *testing.T) {
	t.Run("should integrate lunar calendar with schedule calculator", func(t *testing.T) {
		lunarCalendar := services.NewLunarCalendar()
		scheduleCalc := services.NewScheduleCalculator(lunarCalendar)

		// Test different schedule types
		testCases := []struct {
			name         string
			recurrenceType string
			calendarType   string
		}{
			{"daily solar", models.RecurrenceTypeDaily, models.CalendarTypeSolar},
			{"weekly solar", models.RecurrenceTypeWeekly, models.CalendarTypeSolar},
			{"monthly solar", models.RecurrenceTypeMonthly, models.CalendarTypeSolar},
			{"daily lunar", models.RecurrenceTypeDaily, models.CalendarTypeLunar},
			{"weekly lunar", models.RecurrenceTypeWeekly, models.CalendarTypeLunar},
			{"monthly lunar", models.RecurrenceTypeMonthly, models.CalendarTypeLunar},
		}

		baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				reminder := &models.Reminder{
					Type:             models.ReminderTypeRecurring,
					TriggerTimeOfDay: "10:00",
					CalendarType:     tc.calendarType,
					RecurrencePattern: &models.RecurrencePattern{
						Type: tc.recurrenceType,
					},
				}

				// Add required fields for monthly
				if tc.recurrenceType == models.RecurrenceTypeMonthly {
					reminder.RecurrencePattern.DayOfMonth = 15
				}

				nextTime, err := scheduleCalc.CalculateNextTrigger(reminder, baseTime)
				assert.NoError(t, err)
				assert.True(t, nextTime.After(baseTime), "Next execution should be after base time")
			})
		}
	})

	t.Run("should handle one-time reminders", func(t *testing.T) {
		lunarCalendar := services.NewLunarCalendar()
		scheduleCalc := services.NewScheduleCalculator(lunarCalendar)

		baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
		triggerTime := time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC)

		reminder := &models.Reminder{
			Type:          models.ReminderTypeOneTime,
			NextTriggerAt: triggerTime,
		}

		nextTime, err := scheduleCalc.CalculateNextTrigger(reminder, baseTime)
		assert.NoError(t, err)
		assert.Equal(t, triggerTime, nextTime)
	})
}

// BenchmarkIntegration_LunarCalculation benchmarks lunar calendar calculations
func BenchmarkIntegration_LunarCalculation(b *testing.B) {
	lunarCalendar := services.NewLunarCalendar()
	testTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lunarCalendar.SolarToLunar(testTime)
	}
}

// BenchmarkIntegration_ScheduleCalculation benchmarks schedule calculations
func BenchmarkIntegration_ScheduleCalculation(b *testing.B) {
	lunarCalendar := services.NewLunarCalendar()
	scheduleCalc := services.NewScheduleCalculator(lunarCalendar)
	testTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)

	reminder := &models.Reminder{
		Type:             models.ReminderTypeRecurring,
		TriggerTimeOfDay: "10:00",
		CalendarType:     models.CalendarTypeSolar,
		RecurrencePattern: &models.RecurrencePattern{
			Type: models.RecurrenceTypeDaily,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := scheduleCalc.CalculateNextTrigger(reminder, testTime)
		if err != nil {
			b.Fatal(err)
		}
	}
}