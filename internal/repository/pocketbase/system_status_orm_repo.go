package pocketbase

import (
	"context"
	"fmt"

	"remiaq/internal/models"
	"remiaq/internal/repository"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
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

// recordToSystemStatus converts a PocketBase Record to SystemStatus model
func recordToSystemStatus(record *core.Record) (*models.SystemStatus, error) {
	status := &models.SystemStatus{
		ID:            record.GetInt("mid"),
		WorkerEnabled: record.GetBool("worker_enabled"),
		LastError:     record.GetString("last_error"),
		Updated:       record.GetDateTime("updated").Time(),
	}
	return status, nil
}

// systemStatusToRecord converts a SystemStatus model to PocketBase Record
func systemStatusToRecord(status *models.SystemStatus, record *core.Record) error {
	record.Set("mid", status.ID)
	record.Set("worker_enabled", status.WorkerEnabled)
	record.Set("last_error", status.LastError)
	return nil
}



// Get retrieves the singleton system status record (mid = 1)
func (r *SystemStatusORMRepo) Get(ctx context.Context) (*models.SystemStatus, error) {
	type SystemStatusRecord struct {
		ID            int    `db:"mid"`
		WorkerEnabled bool   `db:"worker_enabled"`
		LastError     string `db:"last_error"`
		Updated       string `db:"updated"`
	}

	statusRec := SystemStatusRecord{}
	err := r.app.DB().
		Select("*").
		From(systemStatusCollectionName).
		Where(dbx.HashExp{"mid": 1}).
		Limit(1).
		One(&statusRec)
	if err != nil {
		return nil, fmt.Errorf("failed to get system status: %w", err)
	}

	status := &models.SystemStatus{
		ID:            statusRec.ID,
		WorkerEnabled: statusRec.WorkerEnabled,
		LastError:     statusRec.LastError,
		Updated:       parseTime(statusRec.Updated),
	}

	return status, nil
}

// IsWorkerEnabled checks if the worker is enabled
func (r *SystemStatusORMRepo) IsWorkerEnabled(ctx context.Context) (bool, error) {
	type SystemStatusRecord struct {
		WorkerEnabled bool `db:"worker_enabled"`
	}

	statusRec := SystemStatusRecord{}
	err := r.app.DB().
		Select("worker_enabled").
		From(systemStatusCollectionName).
		Where(dbx.HashExp{"mid": 1}).
		Limit(1).
		One(&statusRec)
	if err != nil {
		return false, fmt.Errorf("failed to check worker status: %w", err)
	}

	return statusRec.WorkerEnabled, nil
}

// EnableWorker enables the worker
func (r *SystemStatusORMRepo) EnableWorker(ctx context.Context) error {
	records, err := r.app.FindRecordsByFilter(
		systemStatusCollectionName,
		"mid = 1",
		"",
		1,
		0,
	)
	if err != nil || len(records) == 0 {
		return fmt.Errorf("system status record not found: %w", err)
	}

	record := records[0]
	record.Set("worker_enabled", true)
	record.Set("last_error", "")

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to enable worker: %w", err)
	}

	return nil
}

// DisableWorker disables the worker and sets error message
func (r *SystemStatusORMRepo) DisableWorker(ctx context.Context, errorMsg string) error {
	records, err := r.app.FindRecordsByFilter(
		systemStatusCollectionName,
		"mid = 1",
		"",
		1,
		0,
	)
	if err != nil || len(records) == 0 {
		return fmt.Errorf("system status record not found: %w", err)
	}

	record := records[0]
	record.Set("worker_enabled", false)
	record.Set("last_error", errorMsg)

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to disable worker: %w", err)
	}

	return nil
}

// UpdateError updates the last error message
func (r *SystemStatusORMRepo) UpdateError(ctx context.Context, errorMsg string) error {
	records, err := r.app.FindRecordsByFilter(
		systemStatusCollectionName,
		"mid = 1",
		"",
		1,
		0,
	)
	if err != nil || len(records) == 0 {
		return fmt.Errorf("system status record not found: %w", err)
	}

	record := records[0]
	record.Set("last_error", errorMsg)

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to update error: %w", err)
	}

	return nil
}

// ClearError clears the last error message
func (r *SystemStatusORMRepo) ClearError(ctx context.Context) error {
	records, err := r.app.FindRecordsByFilter(
		systemStatusCollectionName,
		"mid = 1",
		"",
		1,
		0,
	)
	if err != nil || len(records) == 0 {
		return fmt.Errorf("system status record not found: %w", err)
	}

	record := records[0]
	record.Set("last_error", "")

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to clear error: %w", err)
	}

	return nil
}