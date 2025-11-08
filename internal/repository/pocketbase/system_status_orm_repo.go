package pocketbase

import (
	"context"
	"fmt"

	"remiaq/internal/models"
	"remiaq/internal/repository"

	"github.com/pocketbase/pocketbase"
)

// SystemStatusORMRepo implements SystemStatusRepository using PocketBase ORM
type SystemStatusORMRepo struct {
	app *pocketbase.PocketBase
}

var _ repository.SystemStatusRepository = (*SystemStatusORMRepo)(nil)

const systemStatusCollectionName = "system_status"

func NewSystemStatusORMRepo(app *pocketbase.PocketBase) repository.SystemStatusRepository {
	return &SystemStatusORMRepo{app: app}
}

// GetSystemStatus retrieves the system status record (mid=1)
func (r *SystemStatusORMRepo) GetSystemStatus(ctx context.Context) (*models.SystemStatus, error) {
	records, err := r.app.FindRecordsByFilter(
		systemStatusCollectionName,
		"mid = 1",
		"",
		1,
		0,
	)
	if err != nil || len(records) == 0 {
		return nil, fmt.Errorf("system_status record not found: %w", err)
	}

	record := records[0]
	status := &models.SystemStatus{
		ID:            record.GetInt("mid"),
		WorkerEnabled: record.GetBool("worker_enabled"),
		LastError:     record.GetString("last_error"),
		Updated:       record.GetDateTime("updated").Time(),
	}

	return status, nil
}

// UpdateSystemStatus updates the system status record
func (r *SystemStatusORMRepo) UpdateSystemStatus(ctx context.Context, status *models.SystemStatus) error {
	records, err := r.app.FindRecordsByFilter(
		systemStatusCollectionName,
		"mid = 1",
		"",
		1,
		0,
	)
	if err != nil || len(records) == 0 {
		return fmt.Errorf("system_status record not found: %w", err)
	}

	record := records[0]
	record.Set("worker_enabled", status.WorkerEnabled)
	record.Set("last_error", status.LastError)

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to update system_status: %w", err)
	}

	return nil
}

// IsWorkerEnabled checks if worker is enabled
func (r *SystemStatusORMRepo) IsWorkerEnabled(ctx context.Context) (bool, error) {
	records, err := r.app.FindRecordsByFilter(
		systemStatusCollectionName,
		"mid = 1",
		"",
		1,
		0,
	)
	if err != nil || len(records) == 0 {
		return false, fmt.Errorf("system_status record not found: %w", err)
	}

	record := records[0]
	return record.GetBool("worker_enabled"), nil
}

// UpdateError records an error
func (r *SystemStatusORMRepo) UpdateError(ctx context.Context, errMsg string) error {
	records, err := r.app.FindRecordsByFilter(
		systemStatusCollectionName,
		"mid = 1",
		"",
		1,
		0,
	)
	if err != nil || len(records) == 0 {
		return fmt.Errorf("system_status record not found: %w", err)
	}

	record := records[0]
	record.Set("last_error", errMsg)

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to update error: %w", err)
	}

	return nil
}

// ClearError clears the error
func (r *SystemStatusORMRepo) ClearError(ctx context.Context) error {
	records, err := r.app.FindRecordsByFilter(
		systemStatusCollectionName,
		"mid = 1",
		"",
		1,
		0,
	)
	if err != nil || len(records) == 0 {
		return fmt.Errorf("system_status record not found: %w", err)
	}

	record := records[0]
	record.Set("last_error", "")

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to clear error: %w", err)
	}

	return nil
}

// DisableWorker disables the worker
func (r *SystemStatusORMRepo) DisableWorker(ctx context.Context) error {
	records, err := r.app.FindRecordsByFilter(
		systemStatusCollectionName,
		"mid = 1",
		"",
		1,
		0,
	)
	if err != nil || len(records) == 0 {
		return fmt.Errorf("system_status record not found: %w", err)
	}

	record := records[0]
	record.Set("worker_enabled", false)
	record.Set("last_error", "Worker disabled")

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to disable worker: %w", err)
	}

	return nil
}
