package db

import (
	"context"
	"fmt"
	"log"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
)

// DBHelperInterface defines the interface for database operations
// This allows for easy mocking in tests
type DBHelperInterface interface {
	GetOneRow(query string, params dbx.Params) (dbx.NullStringMap, error)
	GetAllRows(query string, params dbx.Params) ([]dbx.NullStringMap, error)
	Exec(query string, params dbx.Params) error
	Count(query string, params dbx.Params) (int, error)
	Exists(query string, params dbx.Params) (bool, error)
	App() *pocketbase.PocketBase
}

type DBHelper struct {
	app *pocketbase.PocketBase
}

// App returns the PocketBase app instance
func (h *DBHelper) App() *pocketbase.PocketBase {
	return h.app
}

// NewDBHelper returns a helper bound to the current PocketBase app
func NewDBHelper(app *pocketbase.PocketBase) *DBHelper {
	return &DBHelper{app: app}
}

// GetOneRow runs a query and returns a single row as raw map.
// Returns error if no row found or query fails.
func (h *DBHelper) GetOneRow(query string, params dbx.Params) (dbx.NullStringMap, error) {
	var result dbx.NullStringMap
	q := h.app.DB().NewQuery(query).Bind(params)
	err := q.One(&result)
	if err != nil {
		log.Printf("[DBHelper] GetOneRow failed (query=%s): %v", query, err)
		return nil, err
	}
	return result, nil
}

// GetOne is a generic function that runs a query and returns a single row mapped to struct T.
// Go doesn't support generic methods, so this is implemented as a function.
// Accepts both *DBHelper and DBHelperInterface for flexibility.
// Usage: user, err := db.GetOne[User](helper, "SELECT * FROM users WHERE id = {:id}", dbx.Params{"id": 1})
func GetOne[T any](h DBHelperInterface, query string, params dbx.Params) (*T, error) {
	raw, err := h.GetOneRow(query, params)
	if err != nil {
		return nil, err
	}
	return MapNullStringMapToStruct[T](raw)
}

// GetOneWithConfig is a generic function that runs a query and returns a single row mapped to struct T with config.
// Supports custom mappers and required field validation.
// Usage: user, err := db.GetOneWithConfig[User](helper, query, params, &db.MapperConfig{RequiredFields: []string{"ID"}})
func GetOneWithConfig[T any](h DBHelperInterface, query string, params dbx.Params, cfg *MapperConfig) (*T, error) {
	raw, err := h.GetOneRow(query, params)
	if err != nil {
		return nil, err
	}
	return MapNullStringMapToStructWithConfig[T](raw, cfg)
}

// GetAllRows runs a query and returns all rows as raw maps.
// Returns error if query fails.
func (h *DBHelper) GetAllRows(query string, params dbx.Params) ([]dbx.NullStringMap, error) {
	var results []dbx.NullStringMap
	q := h.app.DB().NewQuery(query).Bind(params)
	err := q.All(&results)
	if err != nil {
		log.Printf("[DBHelper] GetAllRows failed (query=%s): %v", query, err)
		return nil, err
	}
	return results, nil
}

// GetAll is a generic function that runs a query and returns all rows mapped to slice of struct T.
// Usage: users, err := db.GetAll[User](helper, "SELECT * FROM users", dbx.Params{})
func GetAll[T any](h DBHelperInterface, query string, params dbx.Params) ([]T, error) {
	rows, err := h.GetAllRows(query, params)
	if err != nil {
		return nil, err
	}

	results := make([]T, len(rows))
	for i, row := range rows {
		mapped, err := MapNullStringMapToStruct[T](row)
		if err != nil {
			return nil, fmt.Errorf("failed to map row %d: %w", i, err)
		}
		results[i] = *mapped
	}

	return results, nil
}

// GetAllWithConfig is a generic function that runs a query and returns all rows mapped to slice of struct T with config.
// Usage: users, err := db.GetAllWithConfig[User](helper, query, params, &db.MapperConfig{RequiredFields: []string{"ID"}})
func GetAllWithConfig[T any](h DBHelperInterface, query string, params dbx.Params, cfg *MapperConfig) ([]T, error) {
	rows, err := h.GetAllRows(query, params)
	if err != nil {
		return nil, err
	}

	results := make([]T, len(rows))
	for i, row := range rows {
		mapped, err := MapNullStringMapToStructWithConfig[T](row, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to map row %d: %w", i, err)
		}
		results[i] = *mapped
	}

	return results, nil
}

// Exec runs INSERT, UPDATE, DELETE queries.
// Returns error if execution fails.
func (h *DBHelper) Exec(query string, params dbx.Params) error {
	q := h.app.DB().NewQuery(query).Bind(params)
	// Lấy thông tin câu lệnh và tham số (phiên bản dbx của PocketBase)
	sqlStr, args := q.SQL(), q.Params()
	log.Printf("[DBHelper] Exec SQL: %s | args=%v", sqlStr, args)

	_, err := q.Execute()
	if err != nil {
		log.Printf("[DBHelper] Exec failed (query=%s): %v", query, err)
		return err
	}
	return nil
}

// ExecWithContext runs INSERT, UPDATE, DELETE with context support.
// Returns error if context is cancelled or execution fails.
func (h *DBHelper) ExecWithContext(ctx context.Context, query string, params dbx.Params) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return h.Exec(query, params)
	}
}

// Count runs a COUNT query and returns the row count.
// Query should use "SELECT COUNT(*) AS count" pattern for predictable parsing.
// Example: "SELECT COUNT(*) AS count FROM users WHERE status = {:status}"
func (h *DBHelper) Count(query string, params dbx.Params) (int, error) {
	var result struct {
		Count int `db:"count"`
	}
	q := h.app.DB().NewQuery(query).Bind(params)
	err := q.One(&result)
	if err != nil {
		log.Printf("[DBHelper] Count failed (query=%s): %v", query, err)
		return 0, err
	}
	return result.Count, nil
}

// Exists checks if a query returns any rows using SQL EXISTS for better performance.
// Much faster than Count for large datasets since it stops after finding the first match.
// Example: "SELECT * FROM users WHERE email = {:email}"
func (h *DBHelper) Exists(query string, params dbx.Params) (bool, error) {
	// Wrap query with EXISTS to get true/false
	existsQuery := fmt.Sprintf("SELECT EXISTS(%s) AS ok", query)
	var result struct {
		Ok bool `db:"ok"`
	}
	q := h.app.DB().NewQuery(existsQuery).Bind(params)
	err := q.One(&result)
	if err != nil {
		log.Printf("[DBHelper] Exists failed (query=%s): %v", query, err)
		return false, err
	}
	return result.Ok, nil
}
