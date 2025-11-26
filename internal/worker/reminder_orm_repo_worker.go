package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"remiaq/internal/models"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// WorkerReminderRepo handles worker-specific queries
type WorkerReminderRepo struct {
	app *pocketbase.PocketBase
}

// NewWorkerReminderRepo creates a new repo
func NewWorkerReminderRepo(app *pocketbase.PocketBase) *WorkerReminderRepo {
	return &WorkerReminderRepo{app: app}
}

// recordToReminder converts a PocketBase Record to Reminder model
// Duplicated here to avoid dependency on the other repo and keep it self-contained as requested
func (r *WorkerReminderRepo) recordToReminder(record *core.Record) (*models.Reminder, error) {
	reminder := &models.Reminder{
		ID:                 record.Id,
		UserID:             record.GetString("user_id"),
		Title:              record.GetString("title"),
		Description:        record.GetString("description"),
		Tag:                record.GetString("tag"),
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
		IsSendedOneTime:    record.GetBool("is_sended_one_time"),
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

// reminderToRecord updates record from reminder
func (r *WorkerReminderRepo) reminderToRecord(reminder *models.Reminder, record *core.Record) error {
	// Only update fields relevant to worker to be safe, or all?
	// Let's update fields that worker modifies.
	record.Set("status", reminder.Status)
	record.Set("is_sended_one_time", reminder.IsSendedOneTime)
	record.Set("crp_count", reminder.CRPCount)
	record.Set("last_sent_at", reminder.LastSentAt)
	record.Set("last_completed_at", reminder.LastCompletedAt)

	if !reminder.NextCRP.IsZero() {
		record.Set("next_crp", reminder.NextCRP)
	} else {
		record.Set("next_crp", nil)
	}

	if !reminder.NextActionAt.IsZero() {
		record.Set("next_action_at", reminder.NextActionAt)
	} else {
		record.Set("next_action_at", nil)
	}

	if !reminder.NextRecurring.IsZero() {
		record.Set("next_recurring", reminder.NextRecurring)
	} else {
		record.Set("next_recurring", nil)
	}

	return nil
}

// Update updates the reminder
func (r *WorkerReminderRepo) Update(ctx context.Context, reminder *models.Reminder) error {
	record, err := r.app.FindRecordById("reminders", reminder.ID)
	if err != nil {
		return err
	}

	if err := r.reminderToRecord(reminder, record); err != nil {
		return err
	}

	return r.app.Save(record)
}

// 1. Một lần - không CRP
// - type = one_time
// - max_crp = 0
// - status = active
// - next_action_at <= now()
// - isSendedOneTime = false
// - snooze_until <= now() (or null)
func (r *WorkerReminderRepo) GetOneTimeNoCRP(ctx context.Context, now time.Time) ([]*models.Reminder, error) {
	return r.fetchReminders(ctx, dbx.HashExp{
		"type":               models.ReminderTypeOneTime,
		"max_crp":            0,
		"status":             models.ReminderStatusActive,
		"is_sended_one_time": false,
	}, now, "next_action_at")
}

// 2. Một lần - có CRP - isSendedOneTime = false (Lần đầu)
// - type = one_time
// - max_crp > 0
// - status = active
// - next_action_at <= now()
// - isSendedOneTime = false
// - snooze_until <= now()
func (r *WorkerReminderRepo) GetOneTimeFirstSend(ctx context.Context, now time.Time) ([]*models.Reminder, error) {
	return r.fetchReminders(ctx, dbx.HashExp{
		"type":               models.ReminderTypeOneTime,
		"status":             models.ReminderStatusActive,
		"is_sended_one_time": false,
	}, now, "next_action_at", dbx.NewExp("max_crp > 0"))
}

// 3. Một lần - có CRP - isSendedOneTime = true (Retry)
// - type = one_time
// - max_crp > 0
// - crp_count < max_crp
// - status = active
// - next_crp <= now()
// - isSendedOneTime = true
// - snooze_until <= now()
func (r *WorkerReminderRepo) GetOneTimeRetry(ctx context.Context, now time.Time) ([]*models.Reminder, error) {
	return r.fetchReminders(ctx, dbx.HashExp{
		"type":               models.ReminderTypeOneTime,
		"status":             models.ReminderStatusActive,
		"is_sended_one_time": true,
	}, now, "next_crp", dbx.NewExp("max_crp > 0"), dbx.NewExp("crp_count < max_crp"))
}

// 4. Lặp lại - không có CRP - không có crp_until_complete
// - type = recurring
// - status = active
// - max_crp = 0
// - repeat_strategy = none
// - next_recurring <= now()
// - snooze_until <= now()
func (r *WorkerReminderRepo) GetRecurringNoCRP(ctx context.Context, now time.Time) ([]*models.Reminder, error) {
	return r.fetchReminders(ctx, dbx.HashExp{
		"type":            models.ReminderTypeRecurring,
		"status":          models.ReminderStatusActive,
		"max_crp":         0,
		"repeat_strategy": models.RepeatStrategyNone,
	}, now, "next_recurring")
}

// 5.1. Lặp lại - có CRP - không có crp_until_complete - FRP trigger
// - type = recurring
// - status = active
// - max_crp > 0
// - repeat_strategy = none
// - next_recurring <= now()   (Changed from < to <= to match time exactly)
// - snooze_until <= now()
func (r *WorkerReminderRepo) GetRecurringCRPTrigger(ctx context.Context, now time.Time) ([]*models.Reminder, error) {
	return r.fetchReminders(ctx, dbx.HashExp{
		"type":            models.ReminderTypeRecurring,
		"status":          models.ReminderStatusActive,
		"repeat_strategy": models.RepeatStrategyNone,
	}, now, "next_recurring", dbx.NewExp("max_crp > 0"))
}

// 5.2. Lặp lại - có CRP - không có crp_until_complete - CRP retry
// - type = recurring
// - status = active
// - max_crp > 0
// - repeat_strategy = none
// - next_recurring > now()
// - crp_count < max_crp
// - next_crp <= now()
// - snooze_until <= now()
func (r *WorkerReminderRepo) GetRecurringCRPRetry(ctx context.Context, now time.Time) ([]*models.Reminder, error) {
	// This case is special: we check next_crp <= now, but next_recurring > now
	return r.fetchRemindersCustom(ctx, dbx.HashExp{
		"type":            models.ReminderTypeRecurring,
		"status":          models.ReminderStatusActive,
		"repeat_strategy": models.RepeatStrategyNone,
	}, now,
		dbx.NewExp("max_crp > 0"),
		dbx.NewExp("crp_count < max_crp"),
		dbx.NewExp("next_crp <= {:now}", dbx.Params{"now": now}),
		dbx.NewExp("next_recurring > {:now}", dbx.Params{"now": now}),
		dbx.NewExp("(snooze_until IS NULL OR snooze_until = '' OR snooze_until <= {:now})", dbx.Params{"now": now}))
}

// Helper for custom conditions (for case 5.2)
func (r *WorkerReminderRepo) fetchRemindersCustom(ctx context.Context, baseCond dbx.HashExp, now time.Time, extraConds ...dbx.Expression) ([]*models.Reminder, error) {
	var records []struct {
		ID                 string         `db:"id"`
		UserID             string         `db:"user_id"`
		Title              string         `db:"title"`
		Description        string         `db:"description"`
		Tag                string         `db:"tag"`
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
		IsSendedOneTime    bool           `db:"is_sended_one_time"`
	}

	q := r.app.DB().Select("*").From("reminders").Where(baseCond)
	for _, cond := range extraConds {
		q.AndWhere(cond)
	}

	if err := q.All(&records); err != nil {
		return nil, err
	}

	reminders := make([]*models.Reminder, 0, len(records))
	for _, rec := range records {
		reminder := &models.Reminder{
			ID:              rec.ID,
			UserID:          rec.UserID,
			Title:           rec.Title,
			Description:     rec.Description,
			Type:            rec.Type,
			Status:          rec.Status,
			IsSendedOneTime: rec.IsSendedOneTime,
			MaxCRP:          rec.MaxCRP,
			CRPCount:        rec.CRPCount,
			CRPIntervalSec:  rec.CRPIntervalSec,
			RepeatStrategy:  rec.RepeatStrategy,
			NextRecurring:   parseTime(rec.NextRecurring),
			NextActionAt:    parseTime(rec.NextActionAt),
			NextCRP:         parseTime(rec.NextCRP),
			SnoozeUntil:     parseTime(rec.SnoozeUntil),
		}
		reminders = append(reminders, reminder)
	}

	return reminders, nil
}

// Helper to fetch reminders with common conditions
func (r *WorkerReminderRepo) fetchReminders(ctx context.Context, baseCond dbx.HashExp, now time.Time, timeField string, extraConds ...dbx.Expression) ([]*models.Reminder, error) {
	var records []struct {
		ID                 string         `db:"id"`
		UserID             string         `db:"user_id"`
		Title              string         `db:"title"`
		Description        string         `db:"description"`
		Tag                string         `db:"tag"`
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
		IsSendedOneTime    bool           `db:"is_sended_one_time"`
	}

	// Build query
	q := r.app.DB().Select("*").From("reminders").
		Where(baseCond).
		AndWhere(dbx.NewExp(fmt.Sprintf("%s <= {:now}", timeField), dbx.Params{"now": now})).
		AndWhere(dbx.NewExp("(snooze_until IS NULL OR snooze_until = '' OR snooze_until <= {:now})", dbx.Params{"now": now}))

	for _, cond := range extraConds {
		q.AndWhere(cond)
	}

	if err := q.All(&records); err != nil {
		return nil, err
	}

	reminders := make([]*models.Reminder, 0, len(records))
	for _, rec := range records {
		// Manual mapping because we used a struct for scanning
		// We can also use recordToReminder if we fetched *core.Record, but DB().Select().All() maps to struct easier for custom queries
		// Actually, let's use the helper parseTimeDB from the other file? No, I should copy it or implement it.
		// Or just use time.Parse. PocketBase usually returns standard formats.

		reminder := &models.Reminder{
			ID:              rec.ID,
			UserID:          rec.UserID,
			Title:           rec.Title,
			Description:     rec.Description,
			Type:            rec.Type,
			Status:          rec.Status,
			IsSendedOneTime: rec.IsSendedOneTime,
			MaxCRP:          rec.MaxCRP,
			CRPCount:        rec.CRPCount,
			CRPIntervalSec:  rec.CRPIntervalSec,
			NextActionAt:    parseTime(rec.NextActionAt),
			NextCRP:         parseTime(rec.NextCRP),
			SnoozeUntil:     parseTime(rec.SnoozeUntil),
			// Map other fields if needed for notification
		}
		reminders = append(reminders, reminder)
	}

	return reminders, nil
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	// Try standard formats
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02 15:04:05.999Z", s); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02 15:04:05Z", s); err == nil {
		return t
	}
	return time.Time{}
}
