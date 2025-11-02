package pocketbase

import (
	"context"
	"encoding/json"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/repository"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
)

// PocketBaseReminderRepo implements ReminderRepository for PocketBase
type PocketBaseReminderRepo struct {
	app *pocketbase.PocketBase
}

// Ensure PocketBaseReminderRepo implements ReminderRepository
var _ repository.ReminderRepository = (*PocketBaseReminderRepo)(nil)

// NewPocketBaseReminderRepo creates a new PocketBase reminder repository
func NewPocketBaseReminderRepo(app *pocketbase.PocketBase) repository.ReminderRepository {
	return &PocketBaseReminderRepo{app: app}
}

// Create inserts a new reminder
func (r *PocketBaseReminderRepo) Create(ctx context.Context, reminder *models.Reminder) error {
	patternJSON, err := json.Marshal(reminder.RecurrencePattern)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO reminders (
			id, user_id, title, description, type, calendar_type,
			next_trigger_at, trigger_time_of_day, recurrence_pattern,
			repeat_strategy, retry_interval_sec, max_retries, retry_count,
			status, snooze_until, last_completed_at, last_sent_at,
			created, updated
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.app.DB().NewQuery(query).Execute(
		reminder.ID,
		reminder.UserID,
		reminder.Title,
		reminder.Description,
		reminder.Type,
		reminder.CalendarType,
		reminder.NextTriggerAt,
		reminder.TriggerTimeOfDay,
		string(patternJSON),
		reminder.RepeatStrategy,
		reminder.RetryIntervalSec,
		reminder.MaxRetries,
		reminder.RetryCount,
		reminder.Status,
		reminder.SnoozeUntil,
		reminder.LastCompletedAt,
		reminder.LastSentAt,
		time.Now().UTC(),
		time.Now().UTC(),
	)

	return err
}

// GetByID retrieves a reminder by ID
func (r *PocketBaseReminderRepo) GetByID(ctx context.Context, id string) (*models.Reminder, error) {
	query := `SELECT * FROM reminders WHERE id = ?`

	var rawResult dbx.NullStringMap
	err := r.app.DB().NewQuery(query).One(&rawResult, id)
	if err != nil {
		return nil, errpackage pocketbase

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
	
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"remiaq/internal/models"
	"remiaq/internal/repository"
)

// PocketBaseReminderRepo implements ReminderRepository for PocketBase
type PocketBaseReminderRepo struct {
	app *pocketbase.PocketBase
}

// Ensure PocketBaseReminderRepo implements ReminderRepository
var _ repository.ReminderRepository = (*PocketBaseReminderRepo)(nil)

// NewPocketBaseReminderRepo creates a new PocketBase reminder repository
func NewPocketBaseReminderRepo(app *pocketbase.PocketBase) repository.ReminderRepository {
	return &PocketBaseReminderRepo{app: app}
}

// Create inserts a new reminder
func (r *PocketBaseReminderRepo) Create(ctx context.Context, reminder *models.Reminder) error {
	patternJSON, err := json.Marshal(reminder.RecurrencePattern)
	if err != nil {
		return err
	}
	
	query := `
		INSERT INTO reminders (
			id, user_id, title, description, type, calendar_type,
			next_trigger_at, trigger_time_of_day, recurrence_pattern,
			repeat_strategy, retry_interval_sec, max_retries, retry_count,
			status, snooze_until, last_completed_at, last_sent_at,
			created, updated
		) VALUES (
			{:id}, {:user_id}, {:title}, {:description}, {:type}, {:calendar_type},
			{:next_trigger_at}, {:trigger_time_of_day}, {:recurrence_pattern},
			{:repeat_strategy}, {:retry_interval_sec}, {:max_retries}, {:retry_count},
			{:status}, {:snooze_until}, {:last_completed_at}, {:last_sent_at},
			{:created}, {:updated}
		)
	`
	
	_, err = r.app.DB().NewQuery(query).Bind(dbx.Params{
		"id":                   reminder.ID,
		"user_id":              reminder.UserID,
		"title":                reminder.Title,
		"description":          reminder.Description,
		"type":                 reminder.Type,
		"calendar_type":        reminder.CalendarType,
		"next_trigger_at":      reminder.NextTriggerAt,
		"trigger_time_of_day":  reminder.TriggerTimeOfDay,
		"recurrence_pattern":   string(patternJSON),
		"repeat_strategy":      reminder.RepeatStrategy,
		"retry_interval_sec":   reminder.RetryIntervalSec,
		"max_retries":          reminder.MaxRetries,
		"retry_count":          reminder.RetryCount,
		"status":               reminder.Status,
		"snooze_until":         reminder.SnoozeUntil,
		"last_completed_at":    reminder.LastCompletedAt,
		"last_sent_at":         reminder.LastSentAt,
		"created":              time.Now().UTC(),
		"updated":              time.Now().UTC(),
	}).Execute()
	
	return err
}

// GetByID retrieves a reminder by ID
func (r *PocketBaseReminderRepo) GetByID(ctx context.Context, id string) (*models.Reminder, error) {
	query := `SELECT * FROM reminders WHERE id = {:id}`
	
	var rawResult dbx.NullStringMap
	err := r.app.DB().NewQuery(query).Bind(dbx.Params{"id": id}).One(&rawResult)
	if err != nil {
		return nil, err
	}
	
	return r.mapToReminder(rawResult)
}

// Update updates an existing reminder
func (r *PocketBaseReminderRepo) Update(ctx context.Context, reminder *models.Reminder) error {
	patternJSON, err := json.Marshal(reminder.RecurrencePattern)
	if err != nil {
		return err
	}
	
	query := `
		UPDATE reminders SET
			title = {:title}, 
			description = {:description}, 
			type = {:type}, 
			calendar_type = {:calendar_type},
			next_trigger_at = {:next_trigger_at}, 
			trigger_time_of_day = {:trigger_time_of_day}, 
			recurrence_pattern = {:recurrence_pattern},
			repeat_strategy = {:repeat_strategy}, 
			retry_interval_sec = {:retry_interval_sec}, 
			max_retries = {:max_retries},
			retry_count = {:retry_count}, 
			status = {:status}, 
			snooze_until = {:snooze_until},
			last_completed_at = {:last_completed_at}, 
			last_sent_at = {:last_sent_at}, 
			updated = {:updated}
		WHERE id = {:id}
	`
	
	_, err = r.app.DB().NewQuery(query).Bind(dbx.Params{
		"title":                reminder.Title,
		"description":          reminder.Description,
		"type":                 reminder.Type,
		"calendar_type":        reminder.CalendarType,
		"next_trigger_at":      reminder.NextTriggerAt,
		"trigger_time_of_day":  reminder.TriggerTimeOfDay,
		"recurrence_pattern":   string(patternJSON),
		"repeat_strategy":      reminder.RepeatStrategy,
		"retry_interval_sec":   reminder.RetryIntervalSec,
		"max_retries":          reminder.MaxRetries,
		"retry_count":          reminder.RetryCount,
		"status":               reminder.Status,
		"snooze_until":         reminder.SnoozeUntil,
		"last_completed_at":    reminder.LastCompletedAt,
		"last_sent_at":         reminder.LastSentAt,
		"updated":              time.Now().UTC(),
		"id":                   reminder.ID,
	}).Execute()
	
	return err
}

// Delete removes a reminder
func (r *PocketBaseReminderRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM reminders WHERE id = {:id}`
	_, err := r.app.DB().NewQuery(query).Bind(dbx.Params{"id": id}).Execute()
	return err
}

// GetDueReminders retrieves reminders that should be triggered
func (r *PocketBaseReminderRepo) GetDueReminders(ctx context.Context, beforeTime time.Time) ([]*models.Reminder, error) {
	query := `
		SELECT * FROM reminders
		WHERE next_trigger_at <= {:before_time}
		  AND status = 'active'
		  AND (snooze_until IS NULL OR snooze_until <= {:before_time})
		ORDER BY next_trigger_at ASC
	`
	
	var rawResults []dbx.NullStringMap
	err := r.app.DB().NewQuery(query).Bind(dbx.Params{
		"before_time": beforeTime,
	}).All(&rawResults)
	if err != nil {
		return nil, err
	}
	
	return r.mapToReminders(rawResults)
}

// GetByUserID retrieves all reminders for a user
func (r *PocketBaseReminderRepo) GetByUserID(ctx context.Context, userID string) ([]*models.Reminder, error) {
	query := `
		SELECT * FROM reminders
		WHERE user_id = {:user_id}
		ORDER BY next_trigger_at ASC
	`
	
	var rawResults []dbx.NullStringMap
	err := r.app.DB().NewQuery(query).Bind(dbx.Params{"user_id": userID}).All(&rawResults)
	if err != nil {
		return nil, err
	}
	
	return r.mapToReminders(rawResults)
}

// UpdateNextTrigger updates the next trigger time
func (r *PocketBaseReminderRepo) UpdateNextTrigger(ctx context.Context, id string, nextTrigger time.Time) error {
	query := `UPDATE reminders SET next_trigger_at = {:next_trigger}, updated = {:updated} WHERE id = {:id}`
	_, err := r.app.DB().NewQuery(query).Bind(dbx.Params{
		"next_trigger": nextTrigger,
		"updated":      time.Now().UTC(),
		"id":           id,
	}).Execute()
	return err
}

// UpdateStatus updates the reminder status
func (r *PocketBaseReminderRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE reminders SET status = {:status}, updated = {:updated} WHERE id = {:id}`
	_, err := r.app.DB().NewQuery(query).Bind(dbx.Params{
		"status":  status,
		"updated": time.Now().UTC(),
		"id":      id,
	}).Execute()
	return err
}

// IncrementRetryCount increments the retry counter
func (r *PocketBaseReminderRepo) IncrementRetryCount(ctx context.Context, id string) error {
	query := `UPDATE reminders SET retry_count = retry_count + 1, updated = {:updated} WHERE id = {:id}`
	_, err := r.app.DB().NewQuery(query).Bind(dbx.Params{
		"updated": time.Now().UTC(),
		"id":      id,
	}).Execute()
	return err
}

// UpdateSnooze updates the snooze_until time
func (r *PocketBaseReminderRepo) UpdateSnooze(ctx context.Context, id string, snoozeUntil *time.Time) error {
	query := `UPDATE reminders SET snooze_until = {:snooze_until}, updated = {:updated} WHERE id = {:id}`
	_, err := r.app.DB().NewQuery(query).Bind(dbx.Params{
		"snooze_until": snoozeUntil,
		"updated":      time.Now().UTC(),
		"id":           id,
	}).Execute()
	return err
}

// MarkCompleted marks a reminder as completed
func (r *PocketBaseReminderRepo) MarkCompleted(ctx context.Context, id string, completedAt time.Time) error {
	query := `
		UPDATE reminders 
		SET status = 'completed', last_completed_at = {:completed_at}, updated = {:updated}
		WHERE id = {:id}
	`
	_, err := r.app.DB().NewQuery(query).Bind(dbx.Params{
		"completed_at": completedAt,
		"updated":      time.Now().UTC(),
		"id":           id,
	}).Execute()
	return err
}

// UpdateLastSent updates the last_sent_at timestamp
func (r *PocketBaseReminderRepo) UpdateLastSent(ctx context.Context, id string, sentAt time.Time) error {
	query := `UPDATE reminders SET last_sent_at = {:sent_at}, updated = {:updated} WHERE id = {:id}`
	_, err := r.app.DB().NewQuery(query).Bind(dbx.Params{
		"sent_at": sentAt,
		"updated": time.Now().UTC(),
		"id":      id,
	}).Execute()
	return err
}

// Helper functions

func (r *PocketBaseReminderRepo) mapToReminder(raw dbx.NullStringMap) (*models.Reminder, error) {
	reminder := &models.Reminder{}
	
	reminder.ID = raw["id"].String
	reminder.UserID = raw["user_id"].String
	reminder.Title = raw["title"].String
	reminder.Description = raw["description"].String
	reminder.Type = raw["type"].String
	reminder.CalendarType = raw["calendar_type"].String
	reminder.TriggerTimeOfDay = raw["trigger_time_of_day"].String
	reminder.RepeatStrategy = raw["repeat_strategy"].String
	reminder.Status = raw["status"].String
	
	// Parse integers
	if raw["retry_interval_sec"].Valid && raw["retry_interval_sec"].String != "" {
		var val int
		json.Unmarshal([]byte(raw["retry_interval_sec"].String), &val)
		reminder.RetryIntervalSec = val
	}
	if raw["max_retries"].Valid && raw["max_retries"].String != "" {
		var val int
		json.Unmarshal([]byte(raw["max_retries"].String), &val)
		reminder.MaxRetries = val
	}
	if raw["retry_count"].Valid && raw["retry_count"].String != "" {
		var val int
		json.Unmarshal([]byte(raw["retry_count"].String), &val)
		reminder.RetryCount = val
	}
	
	// Parse timestamps
	if raw["next_trigger_at"].Valid && raw["next_trigger_at"].String != "" {
		t, _ := time.Parse(time.RFC3339, raw["next_trigger_at"].String)
		reminder.NextTriggerAt = t
	}
	if raw["created"].Valid && raw["created"].String != "" {
		t, _ := time.Parse(time.RFC3339, raw["created"].String)
		reminder.Created = t
	}
	if raw["updated"].Valid && raw["updated"].String != "" {
		t, _ := time.Parse(time.RFC3339, raw["updated"].String)
		reminder.Updated = t
	}
	
	// Parse nullable timestamps
	if raw["snooze_until"].Valid && raw["snooze_until"].String != "" {
		t, _ := time.Parse(time.RFC3339, raw["snooze_until"].String)
		reminder.SnoozeUntil = &t
	}
	if raw["last_completed_at"].Valid && raw["last_completed_at"].String != "" {
		t, _ := time.Parse(time.RFC3339, raw["last_completed_at"].String)
		reminder.LastCompletedAt = &t
	}
	if raw["last_sent_at"].Valid && raw["last_sent_at"].String != "" {
		t, _ := time.Parse(time.RFC3339, raw["last_sent_at"].String)
		reminder.LastSentAt = &t
	}
	
	// Parse recurrence pattern
	if raw["recurrence_pattern"].Valid && raw["recurrence_pattern"].String != "" {
		var pattern models.RecurrencePattern
		if err := json.Unmarshal([]byte(raw["recurrence_pattern"].String), &pattern); err == nil {
			reminder.RecurrencePattern = &pattern
		}
	}
	
	return reminder, nil
}

func (r *PocketBaseReminderRepo) mapToReminders(rawList []dbx.NullStringMap) ([]*models.Reminder, error) {
	reminders := make([]*models.Reminder, 0, len(rawList))
	
	for _, raw := range rawList {
		reminder, err := r.mapToReminder(raw)
		if err != nil {
			return nil, err
		}
		reminders = append(reminders, reminder)
	}
	
	return reminders, nil
}
	}

	return r.mapToReminder(rawResult)
}

// Update updates an existing reminder
func (r *PocketBaseReminderRepo) Update(ctx context.Context, reminder *models.Reminder) error {
	patternJSON, err := json.Marshal(reminder.RecurrencePattern)
	if err != nil {
		return err
	}

	query := `
		UPDATE reminders SET
			title = ?, description = ?, type = ?, calendar_type = ?,
			next_trigger_at = ?, trigger_time_of_day = ?, recurrence_pattern = ?,
			repeat_strategy = ?, retry_interval_sec = ?, max_retries = ?,
			retry_count = ?, status = ?, snooze_until = ?,
			last_completed_at = ?, last_sent_at = ?, updated = ?
		WHERE id = ?
	`

	_, err = r.app.DB().NewQuery(query).Execute(
		reminder.Title,
		reminder.Description,
		reminder.Type,
		reminder.CalendarType,
		reminder.NextTriggerAt,
		reminder.TriggerTimeOfDay,
		string(patternJSON),
		reminder.RepeatStrategy,
		reminder.RetryIntervalSec,
		reminder.MaxRetries,
		reminder.RetryCount,
		reminder.Status,
		reminder.SnoozeUntil,
		reminder.LastCompletedAt,
		reminder.LastSentAt,
		time.Now().UTC(),
		reminder.ID,
	)

	return err
}

// Delete removes a reminder
func (r *PocketBaseReminderRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM reminders WHERE id = ?`
	_, err := r.app.DB().NewQuery(query).Execute(id)
	return err
}

// GetDueReminders retrieves reminders that should be triggered
func (r *PocketBaseReminderRepo) GetDueReminders(ctx context.Context, beforeTime time.Time) ([]*models.Reminder, error) {
	query := `
		SELECT * FROM reminders
		WHERE next_trigger_at <= ?
		  AND status = 'active'
		  AND (snooze_until IS NULL OR snooze_until <= ?)
		ORDER BY next_trigger_at ASC
	`

	var rawResults []dbx.NullStringMap
	err := r.app.DB().NewQuery(query).All(&rawResults, beforeTime, beforeTime)
	if err != nil {
		return nil, err
	}

	return r.mapToReminders(rawResults)
}

// GetByUserID retrieves all reminders for a user
func (r *PocketBaseReminderRepo) GetByUserID(ctx context.Context, userID string) ([]*models.Reminder, error) {
	query := `
		SELECT * FROM reminders
		WHERE user_id = ?
		ORDER BY next_trigger_at ASC
	`

	var rawResults []dbx.NullStringMap
	err := r.app.DB().NewQuery(query).All(&rawResults, userID)
	if err != nil {
		return nil, err
	}

	return r.mapToReminders(rawResults)
}

// UpdateNextTrigger updates the next trigger time
func (r *PocketBaseReminderRepo) UpdateNextTrigger(ctx context.Context, id string, nextTrigger time.Time) error {
	query := `UPDATE reminders SET next_trigger_at = ?, updated = ? WHERE id = ?`
	_, err := r.app.DB().NewQuery(query).Execute(nextTrigger, time.Now().UTC(), id)
	return err
}

// UpdateStatus updates the reminder status
func (r *PocketBaseReminderRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE reminders SET status = ?, updated = ? WHERE id = ?`
	_, err := r.app.DB().NewQuery(query).Execute(status, time.Now().UTC(), id)
	return err
}

// IncrementRetryCount increments the retry counter
func (r *PocketBaseReminderRepo) IncrementRetryCount(ctx context.Context, id string) error {
	query := `UPDATE reminders SET retry_count = retry_count + 1, updated = ? WHERE id = ?`
	_, err := r.app.DB().NewQuery(query).Execute(time.Now().UTC(), id)
	return err
}

// UpdateSnooze updates the snooze_until time
func (r *PocketBaseReminderRepo) UpdateSnooze(ctx context.Context, id string, snoozeUntil *time.Time) error {
	query := `UPDATE reminders SET snooze_until = ?, updated = ? WHERE id = ?`
	_, err := r.app.DB().NewQuery(query).Execute(snoozeUntil, time.Now().UTC(), id)
	return err
}

// MarkCompleted marks a reminder as completed
func (r *PocketBaseReminderRepo) MarkCompleted(ctx context.Context, id string, completedAt time.Time) error {
	query := `
		UPDATE reminders 
		SET status = 'completed', last_completed_at = ?, updated = ?
		WHERE id = ?
	`
	_, err := r.app.DB().NewQuery(query).Execute(completedAt, time.Now().UTC(), id)
	return err
}

// UpdateLastSent updates the last_sent_at timestamp
func (r *PocketBaseReminderRepo) UpdateLastSent(ctx context.Context, id string, sentAt time.Time) error {
	query := `UPDATE reminders SET last_sent_at = ?, updated = ? WHERE id = ?`
	_, err := r.app.DB().NewQuery(query).Execute(sentAt, time.Now().UTC(), id)
	return err
}

// Helper functions

func (r *PocketBaseReminderRepo) mapToReminder(raw dbx.NullStringMap) (*models.Reminder, error) {
	reminder := &models.Reminder{}

	reminder.ID = raw["id"].String
	reminder.UserID = raw["user_id"].String
	reminder.Title = raw["title"].String
	reminder.Description = raw["description"].String
	reminder.Type = raw["type"].String
	reminder.CalendarType = raw["calendar_type"].String
	reminder.TriggerTimeOfDay = raw["trigger_time_of_day"].String
	reminder.RepeatStrategy = raw["repeat_strategy"].String
	reminder.Status = raw["status"].String

	// Parse integers
	if raw["retry_interval_sec"].Valid {
		var val int
		json.Unmarshal([]byte(raw["retry_interval_sec"].String), &val)
		reminder.RetryIntervalSec = val
	}
	if raw["max_retries"].Valid {
		var val int
		json.Unmarshal([]byte(raw["max_retries"].String), &val)
		reminder.MaxRetries = val
	}
	if raw["retry_count"].Valid {
		var val int
		json.Unmarshal([]byte(raw["retry_count"].String), &val)
		reminder.RetryCount = val
	}

	// Parse timestamps
	if raw["next_trigger_at"].Valid {
		t, _ := time.Parse(time.RFC3339, raw["next_trigger_at"].String)
		reminder.NextTriggerAt = t
	}
	if raw["created"].Valid {
		t, _ := time.Parse(time.RFC3339, raw["created"].String)
		reminder.Created = t
	}
	if raw["updated"].Valid {
		t, _ := time.Parse(time.RFC3339, raw["updated"].String)
		reminder.Updated = t
	}

	// Parse nullable timestamps
	if raw["snooze_until"].Valid && raw["snooze_until"].String != "" {
		t, _ := time.Parse(time.RFC3339, raw["snooze_until"].String)
		reminder.SnoozeUntil = &t
	}
	if raw["last_completed_at"].Valid && raw["last_completed_at"].String != "" {
		t, _ := time.Parse(time.RFC3339, raw["last_completed_at"].String)
		reminder.LastCompletedAt = &t
	}
	if raw["last_sent_at"].Valid && raw["last_sent_at"].String != "" {
		t, _ := time.Parse(time.RFC3339, raw["last_sent_at"].String)
		reminder.LastSentAt = &t
	}

	// Parse recurrence pattern
	if raw["recurrence_pattern"].Valid && raw["recurrence_pattern"].String != "" {
		var pattern models.RecurrencePattern
		if err := json.Unmarshal([]byte(raw["recurrence_pattern"].String), &pattern); err == nil {
			reminder.RecurrencePattern = &pattern
		}
	}

	return reminder, nil
}

func (r *PocketBaseReminderRepo) mapToReminders(rawList []dbx.NullStringMap) ([]*models.Reminder, error) {
	reminders := make([]*models.Reminder, 0, len(rawList))

	for _, raw := range rawList {
		reminder, err := r.mapToReminder(raw)
		if err != nil {
			return nil, err
		}
		reminders = append(reminders, reminder)
	}

	return reminders, nil
}
