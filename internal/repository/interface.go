package repository

import (
	"context"
	"time"

	"remiaq/internal/models"
)

// ReminderRepository defines operations for reminder data access
type ReminderRepository interface {
	// CRUD operations
	Create(ctx context.Context, reminder *models.Reminder) error
	GetByID(ctx context.Context, id string) (*models.Reminder, error)
	Update(ctx context.Context, reminder *models.Reminder) error
	Delete(ctx context.Context, id string) error

	// Query operations
	GetDueReminders(ctx context.Context, beforeTime time.Time) ([]*models.Reminder, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Reminder, error)

	// Specific updates
	UpdateNextTrigger(ctx context.Context, id string, nextTrigger time.Time) error
	UpdateStatus(ctx context.Context, id string, status string) error
	IncrementRetryCount(ctx context.Context, id string) error
	UpdateSnooze(ctx context.Context, id string, snoozeUntil string) error
	MarkCompleted(ctx context.Context, id string, completedAt string) error
	UpdateLastSent(ctx context.Context, id string, lastSentAt string) error
}

// UserRepository defines operations for user data access
type UserRepository interface {
	// CRUD operations
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error

	// FCM token management
	UpdateFCMToken(ctx context.Context, userID, token string) error
	DisableFCM(ctx context.Context, userID string) error
	EnableFCM(ctx context.Context, userID string, token string) error

	// Query operations
	GetActiveUsers(ctx context.Context) ([]*models.User, error)
}

// SystemStatusRepository defines operations for system status management
type SystemStatusRepository interface {
	// Get singleton instance
	Get(ctx context.Context) (*models.SystemStatus, error)

	// Worker control
	IsWorkerEnabled(ctx context.Context) (bool, error)
	EnableWorker(ctx context.Context) error
	DisableWorker(ctx context.Context, errorMsg string) error

	// Error tracking
	UpdateError(ctx context.Context, errorMsg string) error
	ClearError(ctx context.Context) error
}

// QueryRepository defines operations for raw SQL queries (existing functionality)
type QueryRepository interface {
	// Raw query operations
	ExecuteSelect(ctx context.Context, query string) ([]map[string]interface{}, error)
	ExecuteInsert(ctx context.Context, query string) (rowsAffected int64, lastInsertId int64, err error)
	ExecuteUpdate(ctx context.Context, query string) (rowsAffected int64, err error)
	ExecuteDelete(ctx context.Context, query string) (rowsAffected int64, err error)
}
