package pocketbase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/repository"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
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
		ID:                 record.Id,
		UserID:             record.GetString("user_id"),
		Title:              record.GetString("title"),
		Description:        record.GetString("description"),
		Type:               record.GetString("type"),
		CalendarType:       record.GetString("calendar_type"),
		NextRecurring:      record.GetDateTime("next_recurring").Time(),
		NextCRP:            record.GetDateTime("next_crp").Time(),
		NextActionAt:       record.GetDateTime("next_action_at").Time(),
		CRPIntervalSec:     record.GetInt("crp_interval_sec"),
		MaxCRP:             record.GetInt("max_crp"),
		CRPCount:           record.GetInt("crp_count"),
		LastCRPCompletedAt: record.GetDateTime("last_crp_completed_at").Time(),
		RepeatStrategy:     record.GetString("repeat_strategy"),
		Status:             record.GetString("status"),
		SnoozeUntil:        record.GetDateTime("snooze_until").Time(),
		LastSentAt:         record.GetDateTime("last_sent_at").Time(),
		LastCompletedAt:    record.GetDateTime("last_completed_at").Time(),
		Created:            record.GetDateTime("created").Time(),
		Updated:            record.GetDateTime("updated").Time(),
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
	
	// Convert time.Time to RFC3339 string format for PocketBase
	if !reminder.NextRecurring.IsZero() {
		record.Set("next_recurring", reminder.NextRecurring.Format(time.RFC3339Nano))
	} else {
		record.Set("next_recurring", nil)
	}
	
	if !reminder.NextCRP.IsZero() {
		record.Set("next_crp", reminder.NextCRP.Format(time.RFC3339Nano))
	} else {
		record.Set("next_crp", nil)
	}
	
	if !reminder.NextActionAt.IsZero() {
		record.Set("next_action_at", reminder.NextActionAt.Format(time.RFC3339Nano))
	} else {
		record.Set("next_action_at", nil)
	}
	
	record.Set("crp_interval_sec", reminder.CRPIntervalSec)
	record.Set("max_crp", reminder.MaxCRP)
	record.Set("crp_count", reminder.CRPCount)
	
	if !reminder.LastCRPCompletedAt.IsZero() {
		record.Set("last_crp_completed_at", reminder.LastCRPCompletedAt.Format(time.RFC3339Nano))
	} else {
		record.Set("last_crp_completed_at", nil)
	}
	
	record.Set("repeat_strategy", reminder.RepeatStrategy)
	record.Set("status", reminder.Status)
	
	if !reminder.SnoozeUntil.IsZero() {
		record.Set("snooze_until", reminder.SnoozeUntil.Format(time.RFC3339Nano))
	} else {
		record.Set("snooze_until", nil)
	}
	
	if !reminder.LastSentAt.IsZero() {
		record.Set("last_sent_at", reminder.LastSentAt.Format(time.RFC3339Nano))
	} else {
		record.Set("last_sent_at", nil)
	}
	
	if !reminder.LastCompletedAt.IsZero() {
		record.Set("last_completed_at", reminder.LastCompletedAt.Format(time.RFC3339Nano))
	} else {
		record.Set("last_completed_at", nil)
	}

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

// GetDueReminders retrieves reminders that need to be processed (next_action_at <= now AND not snoozed)
func (r *ReminderORMRepo) GetDueReminders(ctx context.Context, beforeTime time.Time) ([]*models.Reminder, error) {
	type ReminderRecord struct {
		ID                 string         `db:"id"`
		UserID             string         `db:"user_id"`
		Title              string         `db:"title"`
		Description        string         `db:"description"`
		Type               string         `db:"type"`
		CalendarType       string         `db:"calendar_type"`
		NextRecurring      string         `db:"next_recurring"`
		NextCRP            string         `db:"next_crp"`
		NextActionAt       string         `db:"next_action_at"`
		RecurrenceJSON     sql.NullString `db:"recurrence_pattern"`
		CRPIntervalSec     int            `db:"crp_interval_sec"`
		MaxCRP             int            `db:"max_crp"`
		CRPCount           int            `db:"crp_count"`
		LastCRPCompletedAt string         `db:"last_crp_completed_at"`
		RepeatStrategy     string         `db:"repeat_strategy"`
		Status             string         `db:"status"`
		SnoozeUntil        string         `db:"snooze_until"`
		LastSentAt         string         `db:"last_sent_at"`
		LastCompletedAt    string         `db:"last_completed_at"`
		Created            string         `db:"created"`
		Updated            string         `db:"updated"`
	}

	records := []ReminderRecord{}

	// Query: next_action_at <= now AND (snooze_until IS NULL OR snooze_until <= now) AND status = 'active'
	err := r.app.DB().
		Select("*").
		From(reminderCollectionName).
		Where(dbx.NewExp("next_action_at <= {:beforeTime}", dbx.Params{
			"beforeTime": beforeTime.Format(time.RFC3339Nano),
		})).
		AndWhere(dbx.NewExp("(snooze_until IS NULL OR snooze_until <= {:beforeTime})", dbx.Params{
			"beforeTime": beforeTime.Format(time.RFC3339Nano),
		})).
		AndWhere(dbx.In("status", "active")).
		OrderBy("next_action_at ASC").
		All(&records)
	if err != nil {
		return nil, fmt.Errorf("failed to query due reminders: %w", err)
	}

	reminders := make([]*models.Reminder, 0, len(records))
	for _, rec := range records {
		reminder := &models.Reminder{
			ID:                 rec.ID,
			UserID:             rec.UserID,
			Title:              rec.Title,
			Description:        rec.Description,
			Type:               rec.Type,
			CalendarType:       rec.CalendarType,
			NextRecurring:      parseTimeDB(rec.NextRecurring),
			NextCRP:            parseTimeDB(rec.NextCRP),
			NextActionAt:       parseTimeDB(rec.NextActionAt),
			CRPIntervalSec:     rec.CRPIntervalSec,
			MaxCRP:             rec.MaxCRP,
			CRPCount:           rec.CRPCount,
			LastCRPCompletedAt: parseTimeDB(rec.LastCRPCompletedAt),
			RepeatStrategy:     rec.RepeatStrategy,
			Status:             rec.Status,
			SnoozeUntil:        parseTimeDB(rec.SnoozeUntil),
			LastSentAt:         parseTimeDB(rec.LastSentAt),
			LastCompletedAt:    parseTimeDB(rec.LastCompletedAt),
			Created:            parseTimeDB(rec.Created),
			Updated:            parseTimeDB(rec.Updated),
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
		ID                 string         `db:"id"`
		UserID             string         `db:"user_id"`
		Title              string         `db:"title"`
		Description        string         `db:"description"`
		Type               string         `db:"type"`
		CalendarType       string         `db:"calendar_type"`
		NextRecurring      string         `db:"next_recurring"`
		NextCRP            string         `db:"next_crp"`
		NextActionAt       string         `db:"next_action_at"`
		RecurrenceJSON     sql.NullString `db:"recurrence_pattern"`
		CRPIntervalSec     int            `db:"crp_interval_sec"`
		MaxCRP             int            `db:"max_crp"`
		CRPCount           int            `db:"crp_count"`
		LastCRPCompletedAt string         `db:"last_crp_completed_at"`
		RepeatStrategy     string         `db:"repeat_strategy"`
		Status             string         `db:"status"`
		SnoozeUntil        string         `db:"snooze_until"`
		LastSentAt         string         `db:"last_sent_at"`
		LastCompletedAt    string         `db:"last_completed_at"`
		Created            string         `db:"created"`
		Updated            string         `db:"updated"`
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
			ID:                 rec.ID,
			UserID:             rec.UserID,
			Title:              rec.Title,
			Description:        rec.Description,
			Type:               rec.Type,
			CalendarType:       rec.CalendarType,
			NextRecurring:      parseTimeDB(rec.NextRecurring),
			NextCRP:            parseTimeDB(rec.NextCRP),
			NextActionAt:       parseTimeDB(rec.NextActionAt),
			CRPIntervalSec:     rec.CRPIntervalSec,
			MaxCRP:             rec.MaxCRP,
			CRPCount:           rec.CRPCount,
			LastCRPCompletedAt: parseTimeDB(rec.LastCRPCompletedAt),
			RepeatStrategy:     rec.RepeatStrategy,
			Status:             rec.Status,
			SnoozeUntil:        parseTimeDB(rec.SnoozeUntil),
			LastSentAt:         parseTimeDB(rec.LastSentAt),
			LastCompletedAt:    parseTimeDB(rec.LastCompletedAt),
			Created:            parseTimeDB(rec.Created),
			Updated:            parseTimeDB(rec.Updated),
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

// UpdateCRPCount updates CRP count
func (r *ReminderORMRepo) UpdateCRPCount(ctx context.Context, id string, crpCount int) error {
	record, err := r.app.FindRecordById(reminderCollectionName, id)
	if err != nil {
		return fmt.Errorf("reminder not found: %w", err)
	}

	record.Set("crp_count", crpCount)

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to update crp_count: %w", err)
	}

	return nil
}

// UpdateNextRecurring updates next_recurring
func (r *ReminderORMRepo) UpdateNextRecurring(ctx context.Context, id string, nextRecurring time.Time) error {
	record, err := r.app.FindRecordById(reminderCollectionName, id)
	if err != nil {
		return fmt.Errorf("reminder not found: %w", err)
	}

	record.Set("next_recurring", nextRecurring)

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to update next_recurring: %w", err)
	}

	return nil
}

// UpdateNextCRP updates next_crp
func (r *ReminderORMRepo) UpdateNextCRP(ctx context.Context, id string, nextCRP time.Time) error {
	record, err := r.app.FindRecordById(reminderCollectionName, id)
	if err != nil {
		return fmt.Errorf("reminder not found: %w", err)
	}

	record.Set("next_crp", nextCRP)

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to update next_crp: %w", err)
	}

	return nil
}

// UpdateNextActionAt updates next_action_at
func (r *ReminderORMRepo) UpdateNextActionAt(ctx context.Context, id string, nextActionAt time.Time) error {
	record, err := r.app.FindRecordById(reminderCollectionName, id)
	if err != nil {
		return fmt.Errorf("reminder not found: %w", err)
	}

	record.Set("next_action_at", nextActionAt)

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to update next_action_at: %w", err)
	}

	return nil
}

// parseTimeDB parses time string from DB
func parseTimeDB(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339Nano, s)
	return t
}

// ADD these methods for backward compatibility:

// IncrementRetryCount increments CRP count (for backward compatibility)
func (r *ReminderORMRepo) IncrementRetryCount(ctx context.Context, id string) error {
	record, err := r.app.FindRecordById(reminderCollectionName, id)
	if err != nil {
		return fmt.Errorf("reminder not found: %w", err)
	}

	currentCount := record.GetInt("crp_count")
	record.Set("crp_count", currentCount+1)

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to increment crp_count: %w", err)
	}

	return nil
}

// UpdateNextTrigger updates next_crp (for backward compatibility)
func (r *ReminderORMRepo) UpdateNextTrigger(ctx context.Context, id string, nextTrigger time.Time) error {
	return r.UpdateNextCRP(ctx, id, nextTrigger)
}
