package worker

import (
	"context"
	"remiaq/internal/models"
)

// UserRepo provides user operations
type UserRepo interface {
	GetByID(ctx context.Context, userID string) (*models.User, error)
	DisableFCM(ctx context.Context, userID string) error
	SetFCMError(ctx context.Context, userID string, errMsg string) error
}

type SystemStatusRepo interface {
	IsWorkerEnabled(ctx context.Context) (bool, error)
	UpdateError(ctx context.Context, errMsg string) error
	ClearError(ctx context.Context) error
	DisableWorker(ctx context.Context) error
}
