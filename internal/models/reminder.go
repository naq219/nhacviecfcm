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
	RecurrencePattern *RecurrencePattern `json:"recurrence_pattern" db:"recurrence_pattern"`

	// FRP (Father Recurrence Pattern) - only for recurring
	NextRecurring time.Time `json:"next_recurring" db:"next_recurring"`

	// CRP (Child Repeat Pattern) - for both one_time and recurring
	NextCRP            time.Time `json:"next_crp" db:"next_crp"`
	CRPIntervalSec     int       `json:"crp_interval_sec" db:"crp_interval_sec"`
	MaxCRP             int       `json:"max_crp" db:"max_crp"`
	CRPCount           int       `json:"crp_count" db:"crp_count"`
	LastCRPCompletedAt time.Time `json:"last_crp_completed_at" db:"last_crp_completed_at"`

	// Repeat strategy
	RepeatStrategy string `json:"repeat_strategy" db:"repeat_strategy"` // none, crp_until_complete

	// Tracking
	NextActionAt    time.Time `json:"next_action_at" db:"next_action_at"`
	LastSentAt      time.Time `json:"last_sent_at" db:"last_sent_at"`
	LastCompletedAt time.Time `json:"last_completed_at" db:"last_completed_at"`

	// Snooze
	SnoozeUntil time.Time `json:"snooze_until" db:"snooze_until"`

	// Status
	Status string `json:"status" db:"status"` // active, completed, paused

	// Timestamps
	Created time.Time `json:"created" db:"created"`
	Updated time.Time `json:"updated" db:"updated"`
}

// RecurrencePattern defines how a reminder repeats
type RecurrencePattern struct {
	Type             string `json:"type"` // daily, weekly, monthly, lunar_last_day_of_month
	Interval         int    `json:"interval,omitempty"`
	DayOfMonth       int    `json:"day_of_month,omitempty"`
	DayOfWeek        int    `json:"day_of_week,omitempty"`         // 0=Sunday, 1=Monday, etc.
	CalendarType     string `json:"calendar_type,omitempty"`       // solar, lunar (for monthly/yearly)
	TriggerTimeOfDay string `json:"trigger_time_of_day,omitempty"` // HH:MM format (UTC)
	IntervalSeconds  int    `json:"interval_seconds,omitempty"`    // ⭐ ADD THIS FIELD

}

// User represents a user with FCM token
type User struct {
	ID          string    `json:"id" db:"id"`
	Email       string    `json:"email" db:"email"`
	FCMToken    string    `json:"fcm_token" db:"fcm_token"`
	IsFCMActive bool      `json:"is_fcm_active" db:"is_fcm_active"`
	FCMError    string    `json:"fcm_error" db:"fcm_error"`
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
	RepeatStrategyNone             = "none"
	RepeatStrategyCRPUntilComplete = "crp_until_complete"
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
	RecurrenceTypeIntervalSeconds     = "interval_seconds" // ⭐ ADD THIS

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
