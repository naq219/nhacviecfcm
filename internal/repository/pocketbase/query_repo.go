package pocketbase

import (
	"context"

	"remiaq/internal/repository"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
)

// PocketBaseQueryRepo implements QueryRepository for raw SQL queries
type PocketBaseQueryRepo struct {
	app *pocketbase.PocketBase
}

// Ensure implementation
var _ repository.QueryRepository = (*PocketBaseQueryRepo)(nil)

// NewPocketBaseQueryRepo creates a new query repository
func NewPocketBaseQueryRepo(app *pocketbase.PocketBase) repository.QueryRepository {
	return &PocketBaseQueryRepo{app: app}
}

// ExecuteSelect executes a SELECT query and returns results
func (r *PocketBaseQueryRepo) ExecuteSelect(ctx context.Context, query string) ([]map[string]interface{}, error) {
	var rawResult []dbx.NullStringMap
	if err := r.app.DB().NewQuery(query).All(&rawResult); err != nil {
		return nil, err
	}

	// Convert NullStringMap to regular map
	result := make([]map[string]interface{}, 0, len(rawResult))
	for _, row := range rawResult {
		cleaned := map[string]interface{}{}
		for key, val := range row {
			if val.Valid {
				cleaned[key] = val.String
			} else {
				cleaned[key] = nil
			}
		}
		result = append(result, cleaned)
	}

	return result, nil
}

// ExecuteInsert executes an INSERT query
func (r *PocketBaseQueryRepo) ExecuteInsert(ctx context.Context, query string) (int64, int64, error) {
	result, err := r.app.DB().NewQuery(query).Execute()
	if err != nil {
		return 0, 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	lastInsertId, _ := result.LastInsertId()

	return rowsAffected, lastInsertId, nil
}

// ExecuteUpdate executes an UPDATE query
func (r *PocketBaseQueryRepo) ExecuteUpdate(ctx context.Context, query string) (int64, error) {
	result, err := r.app.DB().NewQuery(query).Execute()
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

// ExecuteDelete executes a DELETE query
func (r *PocketBaseQueryRepo) ExecuteDelete(ctx context.Context, query string) (int64, error) {
	result, err := r.app.DB().NewQuery(query).Execute()
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}
