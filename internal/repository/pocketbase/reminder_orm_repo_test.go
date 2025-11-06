package pocketbase

import (
	"context"
	"errors"
	"testing"
	"time"

	"remiaq/internal/models"

	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReminderORMRepo_Create(t *testing.T) {
	t.Run("should create reminder successfully using ORM", func(t *testing.T) {
		mockApp := &MockApp{
			FindCollectionByNameOrIdFn: func(nameOrId string) (*core.Collection, error) {
				assert.Equal(t, "reminders", nameOrId)
				collection := core.NewBaseCollection("reminders")
				return collection, nil
			},
			SaveFn: func(record *core.Record) error {
				// Kiểm tra các giá trị được set trong record
				assert.Equal(t, "test-id", record.GetString("id"))
				assert.Equal(t, "user-123", record.GetString("user_id"))
				assert.Equal(t, "Test Reminder", record.GetString("title"))
				assert.Equal(t, "Test description", record.GetString("description"))
				assert.Equal(t, "daily", record.GetString("type"))
				assert.Equal(t, "gregorian", record.GetString("calendar_type"))
				assert.Equal(t, "active", record.GetString("status"))
				
				// Kiểm tra recurrence pattern
				recurrencePattern := record.GetString("recurrence_pattern")
				assert.Contains(t, recurrencePattern, `"type":"daily"`)
				assert.Contains(t, recurrencePattern, `"intervalSeconds":86400`)
				
				return nil
			},
		}

		repo := &ReminderORMRepo{app: mockApp}
		
		nextTrigger := time.Now().Add(24 * time.Hour)
		reminder := &models.Reminder{
			ID:               "test-id",
			UserID:           "user-123",
			Title:            "Test Reminder",
			Description:      "Test description",
			Type:             "daily",
			CalendarType:     "gregorian",
			NextTriggerAt:    nextTrigger,
			TriggerTimeOfDay: "09:00",
			RecurrencePattern: &models.RecurrencePattern{
				Type:            "daily",
				IntervalSeconds: 86400, // 1 day
			},
			RepeatStrategy:    "fixed",
			RetryIntervalSec: 300,
			MaxRetries:       3,
			RetryCount:       0,
			Status:           "active",
		}

		err := repo.Create(context.Background(), reminder)
		assert.NoError(t, err)
	})

	t.Run("should return error when collection not found", func(t *testing.T) {
		mockApp := &MockApp{
			FindCollectionByNameOrIdFn: func(nameOrId string) (*core.Collection, error) {
				return nil, errors.New("collection not found")
			},
		}

		repo := &ReminderORMRepo{app: mockApp}
		reminder := &models.Reminder{
			ID:     "test-id",
			UserID: "user-123",
			Title:  "Test Reminder",
		}

		err := repo.Create(context.Background(), reminder)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find reminders collection")
	})

	t.Run("should return error when save fails", func(t *testing.T) {
		mockApp := &MockApp{
			FindCollectionByNameOrIdFn: func(nameOrId string) (*core.Collection, error) {
				collection := core.NewBaseCollection("reminders")
				return collection, nil
			},
			SaveFn: func(record *core.Record) error {
				return errors.New("save failed")
			},
		}

		repo := &ReminderORMRepo{app: mockApp}
		reminder := &models.Reminder{
			ID:     "test-id",
			UserID: "user-123",
			Title:  "Test Reminder",
		}

		err := repo.Create(context.Background(), reminder)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save reminder record")
	})

	t.Run("should handle nil recurrence pattern", func(t *testing.T) {
		mockApp := &MockApp{
			FindCollectionByNameOrIdFn: func(nameOrId string) (*core.Collection, error) {
				collection := core.NewBaseCollection("reminders")
				return collection, nil
			},
			SaveFn: func(record *core.Record) error {
				assert.Equal(t, "", record.GetString("recurrence_pattern"))
				return nil
			},
		}

		repo := &ReminderORMRepo{app: mockApp}
		reminder := &models.Reminder{
			ID:                "test-id",
			UserID:            "user-123",
			Title:             "Test Reminder",
			RecurrencePattern: nil,
		}

		err := repo.Create(context.Background(), reminder)
		assert.NoError(t, err)
	})
}

func TestReminderORMRepo_OtherMethods(t *testing.T) {
	t.Run("should return not implemented error for other methods", func(t *testing.T) {
		mockApp := &MockApp{}
		repo := &ReminderORMRepo{app: mockApp}

		// Test GetByID
		_, err := repo.GetByID(context.Background(), "test-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")

		// Test GetByUserID
		_, err = repo.GetByUserID(context.Background(), "user-123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")

		// Test Update
		err = repo.Update(context.Background(), &models.Reminder{ID: "test-id"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")

		// Test Delete
		err = repo.Delete(context.Background(), "test-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
	})
}