package pocketbase

import (
	"context"

	"remiaq/internal/db"
	"remiaq/internal/models"
	"remiaq/internal/repository"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/tools/types"
)

// SystemStatusRepo implements repository.SystemStatusRepository
type SystemStatusRepo struct {
	helper db.DBHelperInterface // Use interface from db package
}

// Ensure implementation
var _ repository.SystemStatusRepository = (*SystemStatusRepo)(nil)

// NewSystemStatusRepo creates a new system status repository
func NewSystemStatusRepo(app *pocketbase.PocketBase) repository.SystemStatusRepository {
	return &SystemStatusRepo{helper: db.NewDBHelper(app)}
}

// Get retrieves the system status (singleton, id=1)
func (r *SystemStatusRepo) Get(ctx context.Context) (*models.SystemStatus, error) {
	return db.GetOne[models.SystemStatus](
		r.helper,
		"SELECT * FROM system_status WHERE mid = {:mid}",
		dbx.Params{"mid": 1},
	)
}

// IsWorkerEnabled checks if worker is enabled
func (r *SystemStatusRepo) IsWorkerEnabled(ctx context.Context) (bool, error) {
	status, err := r.Get(ctx)
	if err != nil {
		return false, err
	}
	return status.WorkerEnabled, nil
}

// EnableWorker enables the worker
func (r *SystemStatusRepo) EnableWorker(ctx context.Context) error {
	return r.helper.Exec(
		"UPDATE system_status SET worker_enabled = TRUE, updated = {:updated} WHERE mid = 1",
		dbx.Params{"updated": types.NowDateTime()},
	)
}

// DisableWorker disables the worker with an error message
func (r *SystemStatusRepo) DisableWorker(ctx context.Context, errorMsg string) error {
	return r.helper.Exec(
		"UPDATE system_status SET worker_enabled = FALSE, last_error = {:error_msg}, updated = {:updated} WHERE mid = 1",
		dbx.Params{
			"error_msg": errorMsg,
			"updated":   types.NowDateTime(),
		},
	)
}

// UpdateError updates the last error message
func (r *SystemStatusRepo) UpdateError(ctx context.Context, errorMsg string) error {
	return r.helper.Exec(
		"UPDATE system_status SET last_error = {:error_msg}, updated = {:updated} WHERE mid = 1",
		dbx.Params{
			"error_msg": errorMsg,
			"updated":   types.NowDateTime(),
		},
	)
}

// ClearError clears the error message
func (r *SystemStatusRepo) ClearError(ctx context.Context) error {
	return r.helper.Exec(
		"UPDATE system_status SET last_error = '', updated = {:updated} WHERE mid = 1",
		dbx.Params{"updated": types.NowDateTime()},
	)
}
