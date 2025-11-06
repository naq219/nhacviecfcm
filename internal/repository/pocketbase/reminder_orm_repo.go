package pocketbase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/repository"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// ReminderORMRepo implements ReminderRepository using PocketBase ORM
type ReminderORMRepo struct {
	app *pocketbase.PocketBase
}

var _ repository.ReminderRepository = (*ReminderORMRepo)(nil)

func NewReminderORMRepo(app *pocketbase.PocketBase) repository.ReminderRepository {
	return &ReminderORMRepo{app: app}
}

func (r *ReminderORMRepo) Create(ctx context.Context, reminder *models.Reminder) error {
	// Lấy collection reminders
	collection, err := r.app.FindCollectionByNameOrId("reminders")
	if err != nil {
		return fmt.Errorf("failed to find reminders collection: %w", err)
	}

	// Tạo record mới
	record := core.NewRecord(collection)
	
	// Thiết lập các giá trị từ reminder model
	record.Set("id", reminder.ID)
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

	// Xử lý recurrence_pattern dạng JSON
	if reminder.RecurrencePattern != nil {
		patternJSON, err := json.Marshal(reminder.RecurrencePattern)
		if err != nil {
			return fmt.Errorf("failed to marshal recurrence pattern: %w", err)
		}
		record.Set("recurrence_pattern", string(patternJSON))
	} else {
		record.Set("recurrence_pattern", "")
	}

	// PocketBase tự động xử lý created và updated
	// Không cần set created và updated vì PocketBase tự động xử lý

	// Lưu record
	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to save reminder record: %w", err)
	}

	return nil
}

// Các phương thức khác vẫn sử dụng SQL thông qua DBHelper cũ
// Để đơn giản, chúng ta sẽ giữ nguyên các phương thức này và chỉ thay đổi Create
// Trong thực tế, có thể chuyển dần các phương thức khác sang ORM

func (r *ReminderORMRepo) GetByID(ctx context.Context, id string) (*models.Reminder, error) {
	return nil, fmt.Errorf("not implemented: use SQL repository for this method")
}

func (r *ReminderORMRepo) GetByUserID(ctx context.Context, userID string) ([]*models.Reminder, error) {
	return nil, fmt.Errorf("not implemented: use SQL repository for this method")
}

func (r *ReminderORMRepo) Update(ctx context.Context, reminder *models.Reminder) error {
	return fmt.Errorf("not implemented: use SQL repository for this method")
}

func (r *ReminderORMRepo) Delete(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented: use SQL repository for this method")
}

func (r *ReminderORMRepo) GetDueReminders(ctx context.Context, beforeTime time.Time) ([]*models.Reminder, error) {
	return nil, fmt.Errorf("not implemented: use SQL repository for this method")
}

func (r *ReminderORMRepo) UpdateNextTrigger(ctx context.Context, id string, nextTrigger time.Time) error {
	return fmt.Errorf("not implemented: use SQL repository for this method")
}

func (r *ReminderORMRepo) IncrementRetryCount(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented: use SQL repository for this method")
}

func (r *ReminderORMRepo) MarkCompleted(ctx context.Context, id string, completedAt string) error {
	return fmt.Errorf("not implemented: use SQL repository for this method")
}

func (r *ReminderORMRepo) UpdateSnooze(ctx context.Context, id string, snoozeUntil string) error {
	return fmt.Errorf("not implemented: use SQL repository for this method")
}

func (r *ReminderORMRepo) UpdateLastSent(ctx context.Context, id string, lastSentAt string) error {
	return fmt.Errorf("not implemented: use SQL repository for this method")
}

func (r *ReminderORMRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	return fmt.Errorf("not implemented: use SQL repository for this method")
}