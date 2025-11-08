package repository

import (
	"context"
	"time"

	"remiaq/internal/models"
)

// ReminderRepository defines operations for reminder data access
type ReminderRepository interface {
	Create(ctx context.Context, reminder *models.Reminder) error
	GetByID(ctx context.Context, id string) (*models.Reminder, error)
	Update(ctx context.Context, reminder *models.Reminder) error
	Delete(ctx context.Context, id string) error
	GetDueReminders(ctx context.Context, beforeTime time.Time) ([]*models.Reminder, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Reminder, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	UpdateSnooze(ctx context.Context, id string, snoozeUntil string) error
	MarkCompleted(ctx context.Context, id string, completedAt string) error
	UpdateLastSent(ctx context.Context, id string, lastSentAt string) error

	// FIX: Add these methods
	UpdateCRPCount(ctx context.Context, id string, crpCount int) error
	UpdateNextRecurring(ctx context.Context, id string, nextRecurring time.Time) error
	UpdateNextCRP(ctx context.Context, id string, nextCRP time.Time) error
	UpdateNextActionAt(ctx context.Context, id string, nextActionAt time.Time) error
	IncrementRetryCount(ctx context.Context, id string) error                      // Keep for compatibility
	UpdateNextTrigger(ctx context.Context, id string, nextTrigger time.Time) error // Keep for compatibility
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
	SetFCMError(ctx context.Context, userID, errorMsg string) error

	// Query operations
	GetActiveUsers(ctx context.Context) ([]*models.User, error)
}

// SystemStatusRepository defines operations for system status management
type SystemStatusRepository interface {
	GetSystemStatus(ctx context.Context) (*models.SystemStatus, error)
	UpdateSystemStatus(ctx context.Context, status *models.SystemStatus) error
	IsWorkerEnabled(ctx context.Context) (bool, error)
	UpdateError(ctx context.Context, errMsg string) error
	ClearError(ctx context.Context) error
	DisableWorker(ctx context.Context) error // CHANGE: Remove errMsg parameter
}

// QueryRepository defines operations for raw SQL queries (existing functionality)
type QueryRepository interface {
	// Raw query operations
	ExecuteSelect(ctx context.Context, query string) ([]map[string]interface{}, error)
	ExecuteInsert(ctx context.Context, query string) (rowsAffected int64, lastInsertId int64, err error)
	ExecuteUpdate(ctx context.Context, query string) (rowsAffected int64, err error)
	ExecuteDelete(ctx context.Context, query string) (rowsAffected int64, err error)
}
