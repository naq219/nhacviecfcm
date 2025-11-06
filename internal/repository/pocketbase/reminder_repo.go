// internal/repository/pocketbase/reminder_repo.go
package pocketbase

import (
	"context"
	"encoding/json"
	"time"

	"remiaq/internal/db"
	"remiaq/internal/models"
	"remiaq/internal/repository"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
)

type ReminderRepo struct {
	helper db.DBHelperInterface
}

var _ repository.ReminderRepository = (*ReminderRepo)(nil)

func NewReminderRepo(app *pocketbase.PocketBase) repository.ReminderRepository {
	return &ReminderRepo{helper: db.NewDBHelper(app)}
}

func (r *ReminderRepo) Create(ctx context.Context, reminder *models.Reminder) error {
	patternJSON, _ := json.Marshal(reminder.RecurrencePattern)
	now := time.Now().UTC()

	query := `
        INSERT INTO reminders (
            id, user_id, title, description, type, calendar_type,
            next_trigger_at, trigger_time_of_day, recurrence_pattern,
            repeat_strategy, retry_interval_sec, max_retries, status,
            snooze_until, last_completed_at, last_sent_at, created, updated
        ) VALUES (
            {:id}, {:user_id}, {:title}, {:description}, {:type}, {:calendar_type},
            {:next_trigger_at}, {:trigger_time_of_day}, {:recurrence_pattern},
            {:repeat_strategy}, {:retry_interval_sec}, {:max_retries}, {:status},
            {:snooze_until}, {:last_completed_at}, {:last_sent_at}, {:created}, {:updated}
        )
    `

	return r.helper.Exec(query, dbx.Params{
		"id":                  reminder.ID,
		"user_id":             reminder.UserID,
		"title":               reminder.Title,
		"description":         reminder.Description,
		"type":                reminder.Type,
		"calendar_type":       reminder.CalendarType,
		"next_trigger_at":     reminder.NextTriggerAt,
		"trigger_time_of_day": reminder.TriggerTimeOfDay,
		"recurrence_pattern":  string(patternJSON),
		"repeat_strategy":     reminder.RepeatStrategy,
		"retry_interval_sec":  reminder.RetryIntervalSec,
		"max_retries":         reminder.MaxRetries,
		"status":              reminder.Status,
		"snooze_until":        reminder.SnoozeUntil,
		"last_completed_at":   reminder.LastCompletedAt,
		"last_sent_at":        reminder.LastSentAt,
		"created":             now,
		"updated":             now,
	})
}

func (r *ReminderRepo) GetByID(ctx context.Context, id string) (*models.Reminder, error) {
	return db.GetOne[models.Reminder](r.helper,
		"SELECT * FROM reminders WHERE id = {:id} LIMIT 1",
		dbx.Params{"id": id})
}

// GetByUserID retrieves all reminders for a specific user
func (r *ReminderRepo) GetByUserID(ctx context.Context, userID string) ([]*models.Reminder, error) {
	reminders, err := db.GetAll[models.Reminder](r.helper,
		"SELECT * FROM reminders WHERE user_id = {:user_id} ORDER BY next_trigger_at ASC",
		dbx.Params{"user_id": userID})
	if err != nil {
		return nil, err
	}

	// Convert []models.Reminder to []*models.Reminder
	result := make([]*models.Reminder, len(reminders))
	for i := range reminders {
		result[i] = &reminders[i]
	}
	return result, nil
}

func (r *ReminderRepo) Update(ctx context.Context, reminder *models.Reminder) error {
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
            last_sent_at = {:last_sent_at}
        WHERE id = {:id}
    `

	return r.helper.Exec(query, dbx.Params{
		"user_id":             reminder.UserID,
		"title":               reminder.Title,
		"description":         reminder.Description,
		"type":                reminder.Type,
		"calendar_type":       reminder.CalendarType,
		"next_trigger_at":     reminder.NextTriggerAt,
		"trigger_time_of_day": reminder.TriggerTimeOfDay,
		"recurrence_pattern":  string(patternJSON),
		"repeat_strategy":     reminder.RepeatStrategy,
		"retry_interval_sec":  reminder.RetryIntervalSec,
		"max_retries":         reminder.MaxRetries,
		"status":              reminder.Status,
		"snooze_until":        reminder.SnoozeUntil,
		"last_completed_at":   reminder.LastCompletedAt,
		"last_sent_at":        reminder.LastSentAt,
		"id":                  reminder.ID,
	})
}

func (r *ReminderRepo) Delete(ctx context.Context, id string) error {
	return r.helper.Exec("DELETE FROM reminders WHERE id = {:id}",
		dbx.Params{"id": id})
}

func (r *ReminderRepo) GetDueReminders(ctx context.Context, beforeTime time.Time) ([]*models.Reminder, error) {
	query := `
        SELECT * FROM reminders
        WHERE next_trigger_at <= {:before_time}
          AND status = 'active'
          AND (snooze_until IS NULL OR snooze_until <= {:before_time})
    `

	reminders, err := db.GetAll[models.Reminder](r.helper, query,
		dbx.Params{"before_time": beforeTime})
	if err != nil {
		return nil, err
	}

	// Convert []models.Reminder to []*models.Reminder
	result := make([]*models.Reminder, len(reminders))
	for i := range reminders {
		result[i] = &reminders[i]
	}
	return result, nil
}

func (r *ReminderRepo) UpdateNextTrigger(ctx context.Context, id string, nextTrigger time.Time) error {
	return r.helper.Exec(
		"UPDATE reminders SET next_trigger_at = {:next_trigger} WHERE id = {:id}",
		dbx.Params{
			"next_trigger": nextTrigger,
			"id":           id,
		})
}

func (r *ReminderRepo) IncrementRetryCount(ctx context.Context, id string) error {
	return r.helper.Exec(
		"UPDATE reminders SET retry_count = retry_count + 1 WHERE id = {:id}",
		dbx.Params{
			"id": id,
		})
}

func (r *ReminderRepo) MarkCompleted(ctx context.Context, id string, completedAt time.Time) error {
	return r.helper.Exec(
		"UPDATE reminders SET status = {:status}, last_completed_at = {:completed_at} WHERE id = {:id}",
		dbx.Params{
			"status":       "completed",
			"completed_at": completedAt,
			"id":           id,
		})
}

// UpdateSnooze sets the snooze_until time for a reminder.
func (r *ReminderRepo) UpdateSnooze(ctx context.Context, id string, snoozeUntil *time.Time) error {
	return r.helper.Exec(
		"UPDATE reminders SET snooze_until = {:snooze_until} WHERE id = {:id}",
		dbx.Params{
			"snooze_until": snoozeUntil,
			"id":           id,
		},
	)
}

func (r *ReminderRepo) UpdateLastSent(ctx context.Context, id string, lastSentAt time.Time) error {
	return r.helper.Exec(
		"UPDATE reminders SET last_sent_at = {:sent_at} WHERE id = {:id}",
		dbx.Params{
			"sent_at": lastSentAt,
			"id":      id,
		})
}

func (r *ReminderRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	return r.helper.Exec(
		"UPDATE reminders SET status = {:status} WHERE id = {:id}",
		dbx.Params{
			"status": status,
			"id":     id,
		})
}
