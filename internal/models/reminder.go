package models

import (
	"time"
)

// Reminder represents a notification reminder
type Reminder struct {
	ID                string             `json:"id" db:"id"`
	UserID            string             `json:"user_id" db:"user_id"`
	Title             string             `json:"title" db:"title"`
	Description       string             `json:"description" db:"description"`
	Type              string             `json:"type" db:"type"`                   // one_time, recurring
	CalendarType      string             `json:"calendar_type" db:"calendar_type"` // solar, lunar
	NextTriggerAt     time.Time          `json:"next_trigger_at" db:"next_trigger_at"`
	TriggerTimeOfDay  string             `json:"trigger_time_of_day" db:"trigger_time_of_day"` // HH:MM format
	RecurrencePattern *RecurrencePattern `json:"recurrence_pattern" db:"recurrence_pattern"`   // JSON field
	RepeatStrategy    string             `json:"repeat_strategy" db:"repeat_strategy"`         // none, retry_until_complete
	RetryIntervalSec  int                `json:"retry_interval_sec" db:"retry_interval_sec"`
	MaxRetries        int                `json:"max_retries" db:"max_retries"`
	RetryCount        int                `json:"retry_count" db:"retry_count"`
	Status            string             `json:"status" db:"status"` // active, completed, paused
	SnoozeUntil       string             `json:"snooze_until" db:"snooze_until"`
	LastCompletedAt   string             `json:"last_completed_at" db:"last_completed_at"`
	LastSentAt        string             `json:"last_sent_at" db:"last_sent_at"`
	Created           time.Time          `json:"created" db:"created"`
	Updated           time.Time          `json:"updated" db:"updated"`
}

// RecurrencePattern defines how a reminder repeats
type RecurrencePattern struct {
	Type            string `json:"type"`                       // daily, weekly, monthly, lunar_last_day_of_month
	IntervalSeconds int    `json:"interval_seconds,omitempty"` // For interval-based recurrence
	DayOfMonth      int    `json:"day_of_month,omitempty"`     // For monthly recurrence
	DayOfWeek       int    `json:"day_of_week,omitempty"`      // For weekly recurrence (0=Sunday)
	BaseOn          string `json:"base_on,omitempty"`          // creation, completion
}

// User represents a user with FCM token
type User struct {
	ID          string    `json:"id" db:"id"`
	Email       string    `json:"email" db:"email"`
	FCMToken    string    `json:"fcm_token" db:"fcm_token"`
	IsFCMActive bool      `json:"is_fcm_active" db:"is_fcm_active"`
	Created     time.Time `json:"created" db:"created"`
	Updated     time.Time `json:"updated" db:"updated"`
}

// SystemStatus represents system configuration (singleton)
type SystemStatus struct {
	ID            int       `json:"mid" db:"mid"` // Always 1
	WorkerEnabled bool      `json:"worker_enabled" db:"worker_enabled"`
	LastError     string    `json:"last_error" db:"last_error"`
	Updated       time.Time `json:"updated" db:"updated"`
}

// Constants for reminder types
const (
	ReminderTypeOneTime   = "one_time"
	ReminderTypeRecurring = "recurring"
)

// Constants for calendar types
const (
	CalendarTypeSolar = "solar"
	CalendarTypeLunar = "lunar"
)

// Constants for repeat strategies
const (
	RepeatStrategyNone               = "none"
	RepeatStrategyRetryUntilComplete = "retry_until_complete"
)

// Constants for reminder status
const (
	ReminderStatusActive    = "active"
	ReminderStatusCompleted = "completed"
	ReminderStatusPaused    = "paused"
)

// Constants for recurrence pattern types
const (
	RecurrenceTypeDaily               = "daily"
	RecurrenceTypeWeekly              = "weekly"
	RecurrenceTypeMonthly             = "monthly"
	RecurrenceTypeLunarLastDayOfMonth = "lunar_last_day_of_month"
)

// Constants for base_on
const (
	BaseOnCreation   = "creation"
	BaseOnCompletion = "completion"
)

// Validate checks if reminder data is valid
func (r *Reminder) Validate() error {
	if r.Title == "" {
		return &ValidationError{Field: "title", Message: "Title is required"}
	}
	if r.Type != ReminderTypeOneTime && r.Type != ReminderTypeRecurring {
		return &ValidationError{Field: "type", Message: "Type must be one_time or recurring"}
	}
	if r.CalendarType != CalendarTypeSolar && r.CalendarType != CalendarTypeLunar {
		return &ValidationError{Field: "calendar_type", Message: "Calendar type must be solar or lunar"}
	}
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// IsRetryable checks if reminder can be retried
func (r *Reminder) IsRetryable() bool {
	return r.RepeatStrategy == RepeatStrategyRetryUntilComplete &&
		r.RetryCount < r.MaxRetries
}

// ShouldSend checks if reminder should be sent now
func (r *Reminder) ShouldSend(now time.Time) bool {
	if r.Status != ReminderStatusActive {
		return false
	}

	// 🔧 FIX: Check snooze - parse string to time.Time
	if r.SnoozeUntil != "" {
		snoozeTime, err := time.Parse(time.RFC3339, r.SnoozeUntil)
		if err == nil && now.Before(snoozeTime) {
			return false
		}
	}

	// Check trigger time
	return !now.Before(r.NextTriggerAt)
}
