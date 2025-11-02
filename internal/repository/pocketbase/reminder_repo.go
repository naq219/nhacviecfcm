// internal/repository/pocketbase/reminder_repo.go
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

type PocketBaseReminderRepo struct {
	app *pocketbase.PocketBase
}

var _ repository.ReminderRepository = (*PocketBaseReminderRepo)(nil)

func NewPocketBaseReminderRepo(app *pocketbase.PocketBase) repository.ReminderRepository {
	return &PocketBaseReminderRepo{app: app}
}

func (r *PocketBaseReminderRepo) Create(ctx context.Context, reminder *models.Reminder) error {
	patternJSON, _ := json.Marshal(reminder.RecurrencePattern)

	query := `
        INSERT INTO reminders (
            id, user_id, title, description, type, calendar_type,
            next_trigger_at, trigger_time_of_day, recurrence_pattern,
            repeat_strategy, retry_interval_sec, max_retries, status,
            snooze_until, last_completed_at, last_sent_at,
            created, updated
        ) VALUES (
            {:id}, {:user_id}, {:title}, {:description}, {:type}, {:calendar_type},
            {:next_trigger_at}, {:trigger_time_of_day}, {:recurrence_pattern},
            {:repeat_strategy}, {:retry_interval_sec}, {:max_retries}, {:status},
            {:snooze_until}, {:last_completed_at}, {:last_sent_at},
            {:created}, {:updated}
        )
    `

	q := r.app.DB().NewQuery(query)
	q.Bind(dbx.Params{
		"id":                reminder.ID,
		"user_id":           reminder.UserID,
		"title":             reminder.Title,
		"description":       reminder.Description,
		"type":              reminder.Type,
		"calendar_type":     reminder.CalendarType,
		"next_trigger_at":   reminder.NextTriggerAt,
		"trigger_time_of_day": reminder.TriggerTimeOfDay,
		"recurrence_pattern": string(patternJSON),
		"repeat_strategy":    reminder.RepeatStrategy,
		"retry_interval_sec": reminder.RetryIntervalSec,
		"max_retries":       reminder.MaxRetries,
		"status":            reminder.Status,
		"snooze_until":      reminder.SnoozeUntil,
		"last_completed_at": reminder.LastCompletedAt,
		"last_sent_at":      reminder.LastSentAt,
		"created":           reminder.Created,
		"updated":           reminder.Updated,
	})
	_, err := q.Execute()
	return err
}

func (r *PocketBaseReminderRepo) GetByID(ctx context.Context, id string) (*models.Reminder, error) {
	query := `SELECT * FROM reminders WHERE id = {:id} LIMIT 1`
	q := r.app.DB().NewQuery(query)
	q.Bind(dbx.Params{
		"id": id,
	})

	var raw map[string]any
	err := q.One(&raw)
	if err != nil {
		return nil, err
	}
	return r.mapToReminder(raw)
}

// GetByUserID retrieves all reminders for a specific user
func (r *PocketBaseReminderRepo) GetByUserID(ctx context.Context, userID string) ([]*models.Reminder, error) {
	query := `SELECT * FROM reminders WHERE user_id = {:user_id} ORDER BY next_trigger_at ASC`
	q := r.app.DB().NewQuery(query)
	q.Bind(dbx.Params{
		"user_id": userID,
	})

	var rawResults []map[string]any
	err := q.All(&rawResults)
	if err != nil {
		return nil, err
	}

	reminders := make([]*models.Reminder, 0, len(rawResults))
	for _, raw := range rawResults {
		rem, err := r.mapToReminder(raw)
		if err != nil {
			continue
		}
		reminders = append(reminders, rem)
	}
	return reminders, nil
}

func (r *PocketBaseReminderRepo) Update(ctx context.Context, reminder *models.Reminder) error {
	patternJSON, _ := json.Marshal(reminder.RecurrencePattern)

	query := `
        UPDATE reminders SET
            user_id = {:user_id}, title = {:title}, description = {:description}, 
            type = {:type}, calendar_type = {:calendar_type},
            next_trigger_at = {:next_trigger_at}, trigger_time_of_day = {:trigger_time_of_day}, 
            recurrence_pattern = {:recurrence_pattern},
            repeat_strategy = {:repeat_strategy}, retry_interval_sec = {:retry_interval_sec}, 
            max_retries = {:max_retries}, status = {:status},
            snooze_until = {:snooze_until}, last_completed_at = {:last_completed_at}, 
            last_sent_at = {:last_sent_at},
            updated = {:updated}
        WHERE id = {:id}
    `

	q := r.app.DB().NewQuery(query)
	q.Bind(dbx.Params{
		"user_id":           reminder.UserID,
		"title":             reminder.Title,
		"description":       reminder.Description,
		"type":              reminder.Type,
		"calendar_type":     reminder.CalendarType,
		"next_trigger_at":   reminder.NextTriggerAt,
		"trigger_time_of_day": reminder.TriggerTimeOfDay,
		"recurrence_pattern": string(patternJSON),
		"repeat_strategy":    reminder.RepeatStrategy,
		"retry_interval_sec": reminder.RetryIntervalSec,
		"max_retries":       reminder.MaxRetries,
		"status":            reminder.Status,
		"snooze_until":      reminder.SnoozeUntil,
		"last_completed_at": reminder.LastCompletedAt,
		"last_sent_at":      reminder.LastSentAt,
		"updated":           time.Now(),
		"id":                reminder.ID,
	})
	_, err := q.Execute()
	return err
}

func (r *PocketBaseReminderRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM reminders WHERE id = {:id}`
	q := r.app.DB().NewQuery(query)
	q.Bind(dbx.Params{
		"id": id,
	})
	_, err := q.Execute()
	return err
}

func (r *PocketBaseReminderRepo) GetDueReminders(ctx context.Context, beforeTime time.Time) ([]*models.Reminder, error) {
	query := `
        SELECT * FROM reminders
        WHERE next_trigger_at <= {:before_time}
          AND status = 'active'
          AND (snooze_until IS NULL OR snooze_until <= {:before_time})
    `
	q := r.app.DB().NewQuery(query)
	q.Bind(dbx.Params{
		"before_time": beforeTime,
	})

	var rawResults []map[string]any
	err := q.All(&rawResults)
	if err != nil {
		return nil, err
	}

	reminders := make([]*models.Reminder, 0, len(rawResults))
	for _, raw := range rawResults {
		rem, err := r.mapToReminder(raw)
		if err != nil {
			continue
		}
		reminders = append(reminders, rem)
	}
	return reminders, nil
}

func (r *PocketBaseReminderRepo) UpdateNextTrigger(ctx context.Context, id string, nextTrigger time.Time) error {
	query := `UPDATE reminders SET next_trigger_at = {:next_trigger}, updated = {:updated} WHERE id = {:id}`
	q := r.app.DB().NewQuery(query)
	q.Bind(dbx.Params{
		"next_trigger": nextTrigger,
		"updated":      time.Now(),
		"id":           id,
	})
	_, err := q.Execute()
	return err
}

func (r *PocketBaseReminderRepo) IncrementRetryCount(ctx context.Context, id string) error {
	query := `UPDATE reminders SET retry_count = retry_count + 1, updated = {:updated} WHERE id = {:id}`
	q := r.app.DB().NewQuery(query)
	q.Bind(dbx.Params{
		"updated": time.Now(),
		"id":      id,
	})
	_, err := q.Execute()
	return err
}

func (r *PocketBaseReminderRepo) MarkCompleted(ctx context.Context, id string, completedAt time.Time) error {
	query := `UPDATE reminders SET status = {:status}, last_completed_at = {:completed_at}, updated = {:updated} WHERE id = {:id}`
	q := r.app.DB().NewQuery(query)
	q.Bind(dbx.Params{
		"status": "completed",
		"completed_at": completedAt,
		"updated": time.Now(),
		"id": id,
	})
	_, err := q.Execute()
	return err
}

func (r *PocketBaseReminderRepo) UpdateSnooze(ctx context.Context, id string, snoozeUntil *time.Time) error {
	query := `UPDATE reminders SET snooze_until = {:snooze_until}, updated = {:updated} WHERE id = {:id}`
	q := r.app.DB().NewQuery(query)
	q.Bind(dbx.Params{
		"snooze_until": snoozeUntil,
		"updated": time.Now(),
		"id": id,
	})
	_, err := q.Execute()
	return err
}

func (r *PocketBaseReminderRepo) UpdateLastSent(ctx context.Context, id string, sentAt time.Time) error {
	query := `UPDATE reminders SET last_sent_at = {:sent_at}, updated = {:updated} WHERE id = {:id}`
	q := r.app.DB().NewQuery(query)
	q.Bind(dbx.Params{
		"sent_at": sentAt,
		"updated": time.Now(),
		"id": id,
	})
	_, err := q.Execute()
	return err
}

func (r *PocketBaseReminderRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE reminders SET status = {:status}, updated = {:updated} WHERE id = {:id}`
	q := r.app.DB().NewQuery(query)
	q.Bind(dbx.Params{
		"status": status,
		"updated": time.Now(),
		"id": id,
	})
	_, err := q.Execute()
	return err
}

// Helper: map raw DB row → Reminder struct
func (r *PocketBaseReminderRepo) mapToReminder(raw map[string]any) (*models.Reminder, error) {
	var pattern models.RecurrencePattern
	if p, ok := raw["recurrence_pattern"].(string); ok && p != "" {
		json.Unmarshal([]byte(p), &pattern)
	}

	rem := &models.Reminder{
		ID:                getString(raw, "id"),
		UserID:            getString(raw, "user_id"),
		Title:             getString(raw, "title"),
		Description:       getString(raw, "description"),
		Type:              getString(raw, "type"),
		CalendarType:      getString(raw, "calendar_type"),
		NextTriggerAt:     getTime(raw, "next_trigger_at"),
		TriggerTimeOfDay:  getString(raw, "trigger_time_of_day"),
		RecurrencePattern: &pattern,
		RepeatStrategy:    getString(raw, "repeat_strategy"),
		RetryIntervalSec:  getInt(raw, "retry_interval_sec"),
		MaxRetries:        getInt(raw, "max_retries"),
		RetryCount:        getInt(raw, "retry_count"),
		Status:            getString(raw, "status"),
		Created:           getTime(raw, "created"),
		Updated:           getTime(raw, "updated"),
	}

	if v := raw["snooze_until"]; v != nil {
		t := getTime(raw, "snooze_until")
		rem.SnoozeUntil = &t
	}
	if v := raw["last_completed_at"]; v != nil {
		t := getTime(raw, "last_completed_at")
		rem.LastCompletedAt = &t
	}
	if v := raw["last_sent_at"]; v != nil {
		t := getTime(raw, "last_sent_at")
		rem.LastSentAt = &t
	}

	return rem, nil
}

// Helper functions
func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]any, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getTime(m map[string]any, key string) time.Time {
	if v, ok := m[key].(time.Time); ok {
		return v
	}
	return time.Time{}
}
