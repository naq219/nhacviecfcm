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

func IsTimeValid(t time.Time) bool {
	return !t.IsZero() && t.Year() >= 2000
}

// IsNextCRPSet checks if NextCRP field is properly set
func (r *Reminder) IsNextCRPSet() bool {
	return IsTimeValid(r.NextCRP)
}

// checks if LastSentAt field is properly set
func (r *Reminder) IsLastSentAtSet() bool {
	return IsTimeValid(r.LastSentAt)
}

// one_time, xem đã đến lúc gửi chưa
func (r *Reminder) CanSendFRPOneTime() bool {
	if r.Type == ReminderTypeOneTime && r.CanTriggerNow(r.NextActionAt) {
		return true
	}
	return false
}

// nếu dữ liệu không đúng, cảnh báo
func (r *Reminder) ValidateData() (bool, string) {
	// Check NextActionAt valid
	if !IsTimeValid(r.NextActionAt) {
		return false, r.ID + " NextActionAt không hợp lệ"
	}

	// Check Type valid
	if r.Type != ReminderTypeOneTime && r.Type != ReminderTypeRecurring {
		return false, r.ID + " Type phải là one_time hoặc recurring"
	}

	if r.Type != ReminderTypeRecurring && !IsTimeValid(r.NextRecurring) {
		return false, r.ID + " NextRecurring không hợp lệ"
	}

	// Check Status valid
	if r.Status != ReminderStatusActive &&
		r.Status != ReminderStatusCompleted &&
		r.Status != ReminderStatusPaused {
		return false, r.ID + " Status không hợp lệ"
	}

	// Check Title không trống
	if r.Title == "" {
		return false, r.ID + " Title không được trống"
	}

	// Check MaxCRP >= 0
	if r.MaxCRP < 0 {
		return false, r.ID + " MaxCRP không được âm"
	}

	// Check CRPIntervalSec > 0 (nếu có CRP)
	if r.MaxCRP > 0 && r.CRPIntervalSec <= 0 {
		return false, r.ID + " CRPIntervalSec phải > 0"
	}

	// Check Recurring có NextRecurring
	if r.Type == ReminderTypeRecurring && !IsTimeValid(r.NextRecurring) {
		return false, r.ID + " Recurring phải có NextRecurring"
	}

	// Check UserID không trống
	if r.UserID == "" {
		return false, r.ID + " UserID không được trống"
	}

	// Check LastCompletedAt valid
	if r.RepeatStrategy == RepeatStrategyCRPUntilComplete && !IsTimeValid(r.LastCompletedAt) {
		return false, r.ID + " LastCompletedAt lặp chờ complete nên phải có giá trị"
	}

	return true, ""
}

// IsNextRecurringSet checks if NextRecurring field is properly set
func (r *Reminder) IsNextRecurringSet() bool {
	return IsTimeValid(r.NextRecurring)
}

// IsSnoozeUntilActive checks if reminder is currently snoozed
func (r *Reminder) IsSnoozeUntilActive(now time.Time) bool {
	return IsTimeValid(r.SnoozeUntil) && r.SnoozeUntil.After(now)
}

// đã đến thời điểm so với now chưa
func (r *Reminder) CanTriggerNow(actionTime time.Time) bool {
	now := time.Now()
	return IsTimeValid(actionTime) && (now.After(actionTime) || now.Equal(actionTime))
}
