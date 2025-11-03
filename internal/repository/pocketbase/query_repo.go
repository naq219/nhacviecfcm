package pocketbase

import (
	"context"

	"remiaq/internal/db"
	"remiaq/internal/repository"

	"github.com/pocketbase/pocketbase"
)

// QueryRepo implements QueryRepository for raw SQL queries
type QueryRepo struct {
	helper db.DBHelperInterface
}

// Ensure implementation
var _ repository.QueryRepository = (*QueryRepo)(nil)

// NewQueryRepo creates a new query repository
func NewQueryRepo(app *pocketbase.PocketBase) repository.QueryRepository {
	return &QueryRepo{helper: db.NewDBHelper(app)}
}

// ExecuteSelect executes a SELECT query and returns results
func (r *QueryRepo) ExecuteSelect(ctx context.Context, query string) ([]map[string]interface{}, error) {
	rawResult, err := r.helper.GetAllRows(query, nil)
	if err != nil {
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
func (r *QueryRepo) ExecuteInsert(ctx context.Context, query string) (int64, int64, error) {
	// For raw queries with return values, we need to use the underlying DB
	// This is a limitation of the current DBHelper interface
	// TODO: Consider extending DBHelper interface for this use case
	return 0, 0, r.helper.Exec(query, nil)
}

// ExecuteUpdate executes an UPDATE query
func (r *QueryRepo) ExecuteUpdate(ctx context.Context, query string) (int64, error) {
	// For raw queries with return values, we need to use the underlying DB
	// This is a limitation of the current DBHelper interface
	// TODO: Consider extending DBHelper interface for this use case
	err := r.helper.Exec(query, nil)
	return 0, err
}

// ExecuteDelete executes a DELETE query
func (r *QueryRepo) ExecuteDelete(ctx context.Context, query string) (int64, error) {
	// For raw queries with return values, we need to use the underlying DB
	// This is a limitation of the current DBHelper interface
	// TODO: Consider extending DBHelper interface for this use case
	err := r.helper.Exec(query, nil)
	return 0, err
}
