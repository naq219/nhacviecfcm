package pocketbase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/repository"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/dbx"
)

// ReminderORMRepo implements ReminderRepository using PocketBase ORM
type ReminderORMRepo struct {
	app *pocketbase.PocketBase
}

var _ repository.ReminderRepository = (*ReminderORMRepo)(nil)

const reminderCollectionName = "reminders"

func NewReminderORMRepo(app *pocketbase.PocketBase) repository.ReminderRepository {
	return &ReminderORMRepo{app: app}
}

// recordToReminder converts a PocketBase Record to Reminder model
func recordToReminder(record *core.Record) (*models.Reminder, error) {
	reminder := &models.Reminder{
		ID:               record.Id,
		UserID:           record.GetString("user_id"),
		Title:            record.GetString("title"),
		Description:      record.GetString("description"),
		Type:             record.GetString("type"),
		CalendarType:     record.GetString("calendar_type"),
		NextTriggerAt:    record.GetString("next_trigger_at"),
		TriggerTimeOfDay: record.GetString("trigger_time_of_day"),
		RepeatStrategy:   record.GetString("repeat_strategy"),
		RetryIntervalSec: record.GetInt("retry_interval_sec"),
		MaxRetries:       record.GetInt("max_retries"),
		RetryCount:       record.GetInt("retry_count"),
		Status:           record.GetString("status"),
		SnoozeUntil:      record.GetString("snooze_until"),
		LastCompletedAt:  record.GetString("last_completed_at"),
		LastSentAt:       record.GetString("last_sent_at"),
		Created:          record.GetDateTime("created").Time(),
		Updated:          record.GetDateTime("updated").Time(),
	}

	// Parse RecurrencePattern if present
	recurrenceJSON := record.GetString("recurrence_pattern")
	if recurrenceJSON != "" {
		var pattern models.RecurrencePattern
		if err := json.Unmarshal([]byte(recurrenceJSON), &pattern); err != nil {
			return nil, fmt.Errorf("failed to parse recurrence_pattern: %w", err)
		}
		reminder.RecurrencePattern = &pattern
	}

	return reminder, nil
}

// reminderToRecord converts a Reminder model to PocketBase Record
func reminderToRecord(reminder *models.Reminder, record *core.Record) error {
	record.Set("user_id", reminder.UserID)
	record.Set("title", reminder.Title)
	record.Set("description", reminder.Description)
	record.Set("type", reminder.Type)
	record.Set("calendar_type", reminder.CalendarType)
	record.Set("next_trigger_at", reminder.NextTriggerAt)
	record.Set("trigger_time_of_day", reminder.TriggerTimeOfDay)
	record.Set("repeat_strategy", reminder.RepeatStrategy)
	record.Set("retry_interval_sec", reminder.RetryIntervalSec)
	record.Set("max_retries", reminder.MaxRetries)
	record.Set("retry_count", reminder.RetryCount)
	record.Set("status", reminder.Status)
	record.Set("snooze_until", reminder.SnoozeUntil)
	record.Set("last_completed_at", reminder.LastCompletedAt)
	record.Set("last_sent_at", reminder.LastSentAt)

	// Serialize RecurrencePattern if present
	if reminder.RecurrencePattern != nil {
		data, err := json.Marshal(reminder.RecurrencePattern)
		if err != nil {
			return fmt.Errorf("failed to marshal recurrence_pattern: %w", err)
		}
		record.Set("recurrence_pattern", string(data))
	}

	return nil
}



// Create creates a new reminder
func (r *ReminderORMRepo) Create(ctx context.Context, reminder *models.Reminder) error {
	collection, err := r.app.FindCollectionByNameOrId(reminderCollectionName)
	if err != nil {
		return fmt.Errorf("failed to find collection: %w", err)
	}

	record := core.NewRecord(collection)
	if err := reminderToRecord(reminder, record); err != nil {
		return err
	}

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to create reminder: %w", err)
	}

	reminder.ID = record.Id
	reminder.Created = record.GetDateTime("created").Time()
	reminder.Updated = record.GetDateTime("updated").Time()

	return nil
}

// GetByID retrieves a reminder by ID
func (r *ReminderORMRepo) GetByID(ctx context.Context, id string) (*models.Reminder, error) {
	record, err := r.app.FindRecordById(reminderCollectionName, id)
	if err != nil {
		return nil, fmt.Errorf("reminder not found: %w", err)
	}

	return recordToReminder(record)
}

// Update updates an existing reminder
func (r *ReminderORMRepo) Update(ctx context.Context, reminder *models.Reminder) error {
	record, err := r.app.FindRecordById(reminderCollectionName, reminder.ID)
	if err != nil {
		return fmt.Errorf("reminder not found: %w", err)
	}

	if err := reminderToRecord(reminder, record); err != nil {
		return err
	}

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to update reminder: %w", err)
	}

	reminder.Updated = record.GetDateTime("updated").Time()
	return nil
}

// Delete deletes a reminder by ID
func (r *ReminderORMRepo) Delete(ctx context.Context, id string) error {
	record, err := r.app.FindRecordById(reminderCollectionName, id)
	if err != nil {
		return fmt.Errorf("reminder not found: %w", err)
	}

	if err := r.app.Delete(record); err != nil {
		return fmt.Errorf("failed to delete reminder: %w", err)
	}

	return nil
}

// GetDueReminders retrieves reminders that are due before a given time
func (r *ReminderORMRepo) GetDueReminders(ctx context.Context, beforeTime time.Time) ([]*models.Reminder, error) {
	type ReminderRecord struct {
	ID               string         `db:"id"`
	UserID           string         `db:"user_id"`
	Title            string         `db:"title"`
	Description      string         `db:"description"`
	Type             string         `db:"type"`
	CalendarType     string         `db:"calendar_type"`
	NextTriggerAt    string         `db:"next_trigger_at"`
	TriggerTimeOfDay string         `db:"trigger_time_of_day"`
	RecurrenceJSON   sql.NullString `db:"recurrence_pattern"`
	RepeatStrategy   string         `db:"repeat_strategy"`
	RetryIntervalSec int            `db:"retry_interval_sec"`
	MaxRetries       int            `db:"max_retries"`
	RetryCount       int            `db:"retry_count"`
	Status           string         `db:"status"`
	SnoozeUntil      string         `db:"snooze_until"`
	LastCompletedAt  string         `db:"last_completed_at"`
	LastSentAt       string         `db:"last_sent_at"`
	Created          string         `db:"created"`
	Updated          string         `db:"updated"`
}

	records := []ReminderRecord{}
	err := r.app.DB().
		Select("*").
		From(reminderCollectionName).
		Where(dbx.NewExp("next_trigger_at <= {:beforeTime}", dbx.Params{
			"beforeTime": beforeTime.Format(time.RFC3339Nano),
		})).
		AndWhere(dbx.In("status", "active", "paused")).
		OrderBy("next_trigger_at ASC").
		All(&records)
	if err != nil {
		return nil, fmt.Errorf("failed to query due reminders: %w", err)
	}

	reminders := make([]*models.Reminder, 0, len(records))
	for _, rec := range records {
		reminder := &models.Reminder{
			ID:               rec.ID,
			UserID:           rec.UserID,
			Title:            rec.Title,
			Description:      rec.Description,
			Type:             rec.Type,
			CalendarType:     rec.CalendarType,
			NextTriggerAt:    rec.NextTriggerAt,
			TriggerTimeOfDay: rec.TriggerTimeOfDay,
			RepeatStrategy:   rec.RepeatStrategy,
			RetryIntervalSec: rec.RetryIntervalSec,
			MaxRetries:       rec.MaxRetries,
			RetryCount:       rec.RetryCount,
			Status:           rec.Status,
			SnoozeUntil:      rec.SnoozeUntil,
			LastCompletedAt:  rec.LastCompletedAt,
			LastSentAt:       rec.LastSentAt,
			Created:          parseTime(rec.Created),
			Updated:          parseTime(rec.Updated),
		}

		// Parse RecurrencePattern if present
	if rec.RecurrenceJSON.Valid && rec.RecurrenceJSON.String != "" {
		var pattern models.RecurrencePattern
		if err := json.Unmarshal([]byte(rec.RecurrenceJSON.String), &pattern); err != nil {
			return nil, fmt.Errorf("failed to parse recurrence_pattern: %w", err)
		}
		reminder.RecurrencePattern = &pattern
	}

		reminders = append(reminders, reminder)
	}

	return reminders, nil
}

// GetByUserID retrieves all reminders for a specific user
func (r *ReminderORMRepo) GetByUserID(ctx context.Context, userID string) ([]*models.Reminder, error) {
	type ReminderRecord struct {
		ID               string         `db:"id"`
		UserID           string         `db:"user_id"`
		Title            string         `db:"title"`
		Description      string         `db:"description"`
		Type             string         `db:"type"`
		CalendarType     string         `db:"calendar_type"`
		NextTriggerAt    string         `db:"next_trigger_at"`
		TriggerTimeOfDay string         `db:"trigger_time_of_day"`
		RecurrenceJSON   sql.NullString `db:"recurrence_pattern"`
		RepeatStrategy   string         `db:"repeat_strategy"`
		RetryIntervalSec int            `db:"retry_interval_sec"`
		MaxRetries       int            `db:"max_retries"`
		RetryCount       int            `db:"retry_count"`
		Status           string         `db:"status"`
		SnoozeUntil      string         `db:"snooze_until"`
		LastCompletedAt  string         `db:"last_completed_at"`
		LastSentAt       string         `db:"last_sent_at"`
		Created          string         `db:"created"`
		Updated          string         `db:"updated"`
	}

	records := []ReminderRecord{}
	err := r.app.DB().
		Select("*").
		From(reminderCollectionName).
		Where(dbx.HashExp{"user_id": userID}).
		OrderBy("created DESC").
		All(&records)
	if err != nil {
		return nil, fmt.Errorf("failed to query reminders by user: %w", err)
	}

	reminders := make([]*models.Reminder, 0, len(records))
	for _, rec := range records {
		reminder := &models.Reminder{
			ID:               rec.ID,
			UserID:           rec.UserID,
			Title:            rec.Title,
			Description:      rec.Description,
			Type:             rec.Type,
			CalendarType:     rec.CalendarType,
			NextTriggerAt:    rec.NextTriggerAt,
			TriggerTimeOfDay: rec.TriggerTimeOfDay,
			RepeatStrategy:   rec.RepeatStrategy,
			RetryIntervalSec: rec.RetryIntervalSec,
			MaxRetries:       rec.MaxRetries,
			RetryCount:       rec.RetryCount,
			Status:           rec.Status,
			SnoozeUntil:      rec.SnoozeUntil,
			LastCompletedAt:  rec.LastCompletedAt,
			LastSentAt:       rec.LastSentAt,
			Created:          parseTime(rec.Created),
			Updated:          parseTime(rec.Updated),
		}

		// Parse RecurrencePattern if present
		if rec.RecurrenceJSON.Valid && rec.RecurrenceJSON.String != "" {
			var pattern models.RecurrencePattern
			if err := json.Unmarshal([]byte(rec.RecurrenceJSON.String), &pattern); err != nil {
				return nil, fmt.Errorf("failed to parse recurrence_pattern: %w", err)
			}
			reminder.RecurrencePattern = &pattern
		}

		reminders = append(reminders, reminder)
	}

	return reminders, nil
}

// UpdateNextTrigger updates the next trigger time for a reminder
func (r *ReminderORMRepo) UpdateNextTrigger(ctx context.Context, id string, nextTrigger time.Time) error {
	record, err := r.app.FindRecordById(reminderCollectionName, id)
	if err != nil {
		return fmt.Errorf("reminder not found: %w", err)
	}

	record.Set("next_trigger_at", nextTrigger.Format(time.RFC3339Nano))

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to update next_trigger_at: %w", err)
	}

	return nil
}

// UpdateStatus updates the status of a reminder
func (r *ReminderORMRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	record, err := r.app.FindRecordById(reminderCollectionName, id)
	if err != nil {
		return fmt.Errorf("reminder not found: %w", err)
	}

	record.Set("status", status)

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

// IncrementRetryCount increments the retry count of a reminder
func (r *ReminderORMRepo) IncrementRetryCount(ctx context.Context, id string) error {
	record, err := r.app.FindRecordById(reminderCollectionName, id)
	if err != nil {
		return fmt.Errorf("reminder not found: %w", err)
	}

	currentCount := record.GetInt("retry_count")
	record.Set("retry_count", currentCount+1)

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to increment retry_count: %w", err)
	}

	return nil
}

// UpdateSnooze updates the snooze time for a reminder
func (r *ReminderORMRepo) UpdateSnooze(ctx context.Context, id string, snoozeUntil string) error {
	record, err := r.app.FindRecordById(reminderCollectionName, id)
	if err != nil {
		return fmt.Errorf("reminder not found: %w", err)
	}

	record.Set("snooze_until", snoozeUntil)

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to update snooze_until: %w", err)
	}

	return nil
}

// MarkCompleted marks a reminder as completed
func (r *ReminderORMRepo) MarkCompleted(ctx context.Context, id string, completedAt string) error {
	record, err := r.app.FindRecordById(reminderCollectionName, id)
	if err != nil {
		return fmt.Errorf("reminder not found: %w", err)
	}

	record.Set("last_completed_at", completedAt)
	record.Set("status", "completed")

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to mark as completed: %w", err)
	}

	return nil
}

// UpdateLastSent updates the last sent time for a reminder
func (r *ReminderORMRepo) UpdateLastSent(ctx context.Context, id string, lastSentAt string) error {
	record, err := r.app.FindRecordById(reminderCollectionName, id)
	if err != nil {
		return fmt.Errorf("reminder not found: %w", err)
	}

	record.Set("last_sent_at", lastSentAt)

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to update last_sent_at: %w", err)
	}

	return nil
}