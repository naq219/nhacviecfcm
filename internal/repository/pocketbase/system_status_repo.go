package pocketbase

import (
	"context"
	"encoding/json"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/repository"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
)

// PocketBaseSystemStatusRepo implements SystemStatusRepository
type PocketBaseSystemStatusRepo struct {
	app *pocketbase.PocketBase
}

// Ensure implementation
var _ repository.SystemStatusRepository = (*PocketBaseSystemStatusRepo)(nil)

// NewPocketBaseSystemStatusRepo creates a new system status repository
func NewPocketBaseSystemStatusRepo(app *pocketbase.PocketBase) repository.SystemStatusRepository {
	return &PocketBaseSystemStatusRepo{app: app}
}

// Get retrieves the system status (singleton, id=1)
func (r *PocketBaseSystemStatusRepo) Get(ctx context.Context) (*models.SystemStatus, error) {
	query := `SELECT * FROM system_status WHERE id = 1`

	var rawResult dbx.NullStringMap
	err := r.app.DB().NewQuery(query).One(&rawResult)
	if err != nil {
		return nil, err
	}

	return r.mapToSystemStatus(rawResult)
}

// IsWorkerEnabled checks if worker is enabled
func (r *PocketBaseSystemStatusRepo) IsWorkerEnabled(ctx context.Context) (bool, error) {
	query := `SELECT worker_enabled FROM system_status WHERE id = 1`

	var rawResult dbx.NullStringMap
	err := r.app.DB().NewQuery(query).One(&rawResult)
	if err != nil {
		return false, err
	}

	if rawResult["worker_enabled"].Valid {
		var enabled bool
		json.Unmarshal([]byte(rawResult["worker_enabled"].String), &enabled)
		return enabled, nil
	}

	return false, nil
}

// EnableWorker enables the worker
func (r *PocketBaseSystemStatusRepo) EnableWorker(ctx context.Context) error {
	query := `UPDATE system_status SET worker_enabled = TRUE, updated = ? WHERE id = 1`
	_, err := r.app.DB().NewQuery(query).Execute(time.Now().UTC())
	return err
}

// DisableWorker disables the worker with an error message
func (r *PocketBaseSystemStatusRepo) DisableWorker(ctx context.Context, errorMsg string) error {
	query := `UPDATE system_status SET worker_enabled = FALSE, last_error = ?, updated = ? WHERE id = 1`
	_, err := r.app.DB().NewQuery(query).Execute(errorMsg, time.Now().UTC())
	return err
}

// UpdateError updates the last error message
func (r *PocketBaseSystemStatusRepo) UpdateError(ctx context.Context, errorMsg string) error {
	query := `UPDATE system_status SET last_error = ?, updated = ? WHERE id = 1`
	_, err := r.app.DB().NewQuery(query).Execute(errorMsg, time.Now().UTC())
	return err
}

// ClearError clears the error message
func (r *PocketBaseSystemStatusRepo) ClearError(ctx context.Context) error {
	query := `UPDATE system_status SET last_error = '', updated = ? WHERE id = 1`
	_, err := r.app.DB().NewQuery(query).Execute(time.Now().UTC())
	return err
}

// Helper function

func (r *PocketBaseSystemStatusRepo) mapToSystemStatus(raw dbx.NullStringMap) (*models.SystemStatus, error) {
	status := &models.SystemStatus{}

	// Parse ID
	if raw["id"].Valid {
		var id int
		json.Unmarshal([]byte(raw["id"].String), &id)
		status.ID = id
	}

	// Parse worker_enabled
	if raw["worker_enabled"].Valid {
		var enabled bool
		json.Unmarshal([]byte(raw["worker_enabled"].String), &enabled)
		status.WorkerEnabled = enabled
	}

	// Last error
	status.LastError = raw["last_error"].String

	// Parse timestamp
	if raw["updated"].Valid {
		t, _ := time.Parse(time.RFC3339, raw["updated"].String)
		status.Updated = t
	}

	return status, nil
}
