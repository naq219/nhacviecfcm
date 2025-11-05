package pocketbase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"remiaq/internal/models"

	"github.com/pocketbase/dbx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create mock Reminder row
func mockReminderRow(id, userID, title, status string) dbx.NullStringMap {
	pattern := &models.RecurrencePattern{
		Type:            "daily",
		IntervalSeconds: 86400, // 1 day
	}
	patternJSON, _ := json.Marshal(pattern)

	return dbx.NullStringMap{
		"id":                  {String: id, Valid: true},
		"user_id":             {String: userID, Valid: true},
		"title":               {String: title, Valid: true},
		"description":         {String: "Test description", Valid: true},
		"type":                {String: "daily", Valid: true},
		"calendar_type":       {String: "gregorian", Valid: true},
		"next_trigger_at":     {String: time.Now().Format(time.RFC3339), Valid: true},
		"trigger_time_of_day": {String: "09:00", Valid: true},
		"recurrence_pattern":  {String: string(patternJSON), Valid: true},
		"repeat_strategy":     {String: "fixed", Valid: true},
		"retry_interval_sec":  {String: "300", Valid: true},
		"max_retries":         {String: "3", Valid: true},
		"retry_count":         {String: "0", Valid: true},
		"status":              {String: status, Valid: true},
		"snooze_until":        {String: "", Valid: false},
		"last_completed_at":   {String: "", Valid: false},
		"last_sent_at":        {String: "", Valid: false},
		"created":             {String: time.Now().Format(time.RFC3339), Valid: true},
		"updated":             {String: time.Now().Format(time.RFC3339), Valid: true},
	}
}

func TestReminderRepo_Create(t *testing.T) {
	t.Run("should create reminder successfully", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			ExecFn: func(query string, params dbx.Params) error {
				assert.Contains(t, query, "INSERT INTO reminders")
				assert.Equal(t, "test-id", params["id"])
				assert.Equal(t, "user-123", params["user_id"])
				assert.NotContains(t, params, "created")
				assert.NotContains(t, params, "updated")
				return nil
			},
		}

		repo := &ReminderRepo{helper: mockHelper}
		reminder := &models.Reminder{
			ID:     "test-id",
			UserID: "user-123",
			Title:  "Test Reminder",
			Status: "active",
		}

		err := repo.Create(context.Background(), reminder)
		assert.NoError(t, err)
	})
}

func TestReminderRepo_GetByID(t *testing.T) {
	t.Run("should return reminder successfully", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
				assert.Contains(t, query, "SELECT * FROM reminders WHERE id")
				assert.Equal(t, "test-id", params["id"])
				return mockReminderRow("test-id", "user-123", "Test Reminder", "active"), nil
			},
		}

		repo := &ReminderRepo{helper: mockHelper}
		reminder, err := repo.GetByID(context.Background(), "test-id")

		require.NoError(t, err)
		assert.Equal(t, "test-id", reminder.ID)
		assert.Equal(t, "user-123", reminder.UserID)
		assert.Equal(t, "Test Reminder", reminder.Title)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
				return nil, errors.New("database error")
			},
		}

		repo := &ReminderRepo{helper: mockHelper}
		reminder, err := repo.GetByID(context.Background(), "test-id")

		assert.Error(t, err)
		assert.Nil(t, reminder)
	})
}

func TestReminderRepo_GetByUserID(t *testing.T) {
	t.Run("should return user reminders successfully", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			GetAllRowsFn: func(query string, params dbx.Params) ([]dbx.NullStringMap, error) {
				assert.Contains(t, query, "SELECT * FROM reminders WHERE user_id")
				assert.Equal(t, "user-123", params["user_id"])
				return []dbx.NullStringMap{
					mockReminderRow("rem-1", "user-123", "Reminder 1", "active"),
					mockReminderRow("rem-2", "user-123", "Reminder 2", "active"),
				}, nil
			},
		}

		repo := &ReminderRepo{helper: mockHelper}
		reminders, err := repo.GetByUserID(context.Background(), "user-123")

		require.NoError(t, err)
		assert.Len(t, reminders, 2)
		assert.Equal(t, "rem-1", reminders[0].ID)
		assert.Equal(t, "rem-2", reminders[1].ID)
	})

	t.Run("should return empty slice when no reminders found", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			GetAllRowsFn: func(query string, params dbx.Params) ([]dbx.NullStringMap, error) {
				return []dbx.NullStringMap{}, nil
			},
		}

		repo := &ReminderRepo{helper: mockHelper}
		reminders, err := repo.GetByUserID(context.Background(), "user-123")

		require.NoError(t, err)
		assert.Len(t, reminders, 0)
	})
}

func TestReminderRepo_Update(t *testing.T) {
	t.Run("should update reminder successfully", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			ExecFn: func(query string, params dbx.Params) error {
				assert.Contains(t, query, "UPDATE reminders SET")
				assert.Equal(t, "test-id", params["id"])
				assert.Equal(t, "Updated Title", params["title"])
				assert.NotContains(t, params, "updated")
				return nil
			},
		}

		repo := &ReminderRepo{helper: mockHelper}
		reminder := &models.Reminder{
			ID:    "test-id",
			Title: "Updated Title",
		}

		err := repo.Update(context.Background(), reminder)
		assert.NoError(t, err)
	})
}

func TestReminderRepo_Delete(t *testing.T) {
	t.Run("should delete reminder successfully", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			ExecFn: func(query string, params dbx.Params) error {
				assert.Contains(t, query, "DELETE FROM reminders WHERE id")
				assert.Equal(t, "test-id", params["id"])
				return nil
			},
		}

		repo := &ReminderRepo{helper: mockHelper}
		err := repo.Delete(context.Background(), "test-id")
		assert.NoError(t, err)
	})
}

func TestReminderRepo_GetDueReminders(t *testing.T) {
	t.Run("should return due reminders successfully", func(t *testing.T) {
		beforeTime := time.Now()
		mockHelper := &MockDBHelper{
			GetAllRowsFn: func(query string, params dbx.Params) ([]dbx.NullStringMap, error) {
				assert.Contains(t, query, "next_trigger_at <= {:before_time}")
				assert.Contains(t, query, "status = 'active'")
				assert.Equal(t, beforeTime.Format(time.RFC3339), params["before_time"])
				return []dbx.NullStringMap{
					mockReminderRow("rem-1", "user-123", "Due Reminder", "active"),
				}, nil
			},
		}

		repo := &ReminderRepo{helper: mockHelper}
		reminders, err := repo.GetDueReminders(context.Background(), beforeTime)

		require.NoError(t, err)
		assert.Len(t, reminders, 1)
		assert.Equal(t, "rem-1", reminders[0].ID)
	})
}

func TestReminderRepo_UpdateNextTrigger(t *testing.T) {
	t.Run("should update next trigger successfully", func(t *testing.T) {
		nextTrigger := time.Now().Add(24 * time.Hour)
		mockHelper := &MockDBHelper{
			ExecFn: func(query string, params dbx.Params) error {
				assert.Contains(t, query, "UPDATE reminders SET next_trigger_at")
				assert.Equal(t, "test-id", params["id"])
				assert.Equal(t, nextTrigger.Format(time.RFC3339), params["next_trigger"])
				return nil
			},
		}

		repo := &ReminderRepo{helper: mockHelper}
		err := repo.UpdateNextTrigger(context.Background(), "test-id", nextTrigger)
		assert.NoError(t, err)
	})
}

func TestReminderRepo_IncrementRetryCount(t *testing.T) {
	t.Run("should increment retry count successfully", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			ExecFn: func(query string, params dbx.Params) error {
				assert.Contains(t, query, "retry_count = retry_count + 1")
				assert.Equal(t, "test-id", params["id"])
				return nil
			},
		}

		repo := &ReminderRepo{helper: mockHelper}
		err := repo.IncrementRetryCount(context.Background(), "test-id")
		assert.NoError(t, err)
	})
}

func TestReminderRepo_MarkCompleted(t *testing.T) {
	t.Run("should mark reminder as completed", func(t *testing.T) {
		completedAt := time.Now()
		mockHelper := &MockDBHelper{
			ExecFn: func(query string, params dbx.Params) error {
				assert.Contains(t, query, "status = {:status}")
				assert.Equal(t, "completed", params["status"])
				assert.Equal(t, completedAt.Format(time.RFC3339), params["completed_at"])
				return nil
			},
		}

		repo := &ReminderRepo{helper: mockHelper}
		err := repo.MarkCompleted(context.Background(), "test-id", completedAt.Format(time.RFC3339))
		assert.NoError(t, err)
	})
}

func TestReminderRepo_UpdateSnooze(t *testing.T) {
	t.Run("should update snooze time successfully", func(t *testing.T) {
		snoozeUntil := time.Now().Add(1 * time.Hour)
		mockHelper := &MockDBHelper{
			ExecFn: func(query string, params dbx.Params) error {
				assert.Contains(t, query, "snooze_until = {:snooze_until}")
				assert.Equal(t, snoozeUntil.Format(time.RFC3339), params["snooze_until"])
				return nil
			},
		}

		repo := &ReminderRepo{helper: mockHelper}
		err := repo.UpdateSnooze(context.Background(), "test-id", snoozeUntil.Format(time.RFC3339))
		assert.NoError(t, err)
	})
}

func TestReminderRepo_UpdateLastSent(t *testing.T) {
	t.Run("should update last sent time successfully", func(t *testing.T) {
		sentAt := time.Now()
		mockHelper := &MockDBHelper{
			ExecFn: func(query string, params dbx.Params) error {
				assert.Contains(t, query, "last_sent_at = {:sent_at}")
				assert.Equal(t, sentAt.Format(time.RFC3339), params["sent_at"])
				return nil
			},
		}

		repo := &ReminderRepo{helper: mockHelper}
		err := repo.UpdateLastSent(context.Background(), "test-id", sentAt.Format(time.RFC3339))
		assert.NoError(t, err)
	})
}

func TestReminderRepo_UpdateStatus(t *testing.T) {
	t.Run("should update status successfully", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			ExecFn: func(query string, params dbx.Params) error {
				assert.Contains(t, query, "status = {:status}")
				assert.Equal(t, "paused", params["status"])
				return nil
			},
		}

		repo := &ReminderRepo{helper: mockHelper}
		err := repo.UpdateStatus(context.Background(), "test-id", "paused")
		assert.NoError(t, err)
	})
}

// Benchmark tests
func BenchmarkReminderRepo_GetByID(b *testing.B) {
	mockHelper := &MockDBHelper{
		GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
			return mockReminderRow("test-id", "user-123", "Test Reminder", "active"), nil
		},
	}

	repo := &ReminderRepo{helper: mockHelper}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByID(ctx, "test-id")
	}
}

func BenchmarkReminderRepo_GetByUserID(b *testing.B) {
	mockHelper := &MockDBHelper{
		GetAllRowsFn: func(query string, params dbx.Params) ([]dbx.NullStringMap, error) {
			return []dbx.NullStringMap{
				mockReminderRow("rem-1", "user-123", "Reminder 1", "active"),
				mockReminderRow("rem-2", "user-123", "Reminder 2", "active"),
			}, nil
		},
	}

	repo := &ReminderRepo{helper: mockHelper}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByUserID(ctx, "user-123")
	}
}