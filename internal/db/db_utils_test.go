package db

import (
	"fmt"
	"testing"
	"time"

	"remiaq/internal/models"

	"github.com/pocketbase/dbx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetOne_SuccessMapping tests successful mapping from SQL row to struct
func TestGetOne_SuccessMapping(t *testing.T) {
	t.Run("should map valid SQL row to User struct", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
				return dbx.NullStringMap{
					"id":            {"user123", true},
					"email":         {"test@example.com", true},
					"fcm_token":     {"token_abc", true},
					"is_fcm_active": {"true", true},
					"created":       {time.Now().UTC().Format(time.RFC3339), true},
					"updated":       {time.Now().UTC().Format(time.RFC3339), true},
				}, nil
			},
		}

		result, err := GetOne[models.User](mockHelper, "SELECT * FROM musers WHERE id = ?", dbx.Params{})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "user123", result.ID)
		assert.Equal(t, "test@example.com", result.Email)
	})

	t.Run("should map valid SQL row to SystemStatus struct", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
				return dbx.NullStringMap{
					"mid":            {"1", true},
					"worker_enabled": {"true", true},
					"last_error":     {"", false},
					"updated":        {time.Now().UTC().Format(time.RFC3339), true},
				}, nil
			},
		}

		result, err := GetOne[models.SystemStatus](mockHelper, "SELECT * FROM system_status", dbx.Params{})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.True(t, result.WorkerEnabled)
	})
}

// TestGetOne_MappingErrors tests mapping error cases
func TestGetOne_MappingErrors(t *testing.T) {
	t.Run("should allow missing id field (defaults to empty string)", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
				// Missing "id" field - mapper allows this (returns zero value)
				return dbx.NullStringMap{
					"email":         {"test@example.com", true},
					"fcm_token":     {"token", true},
					"is_fcm_active": {"true", true},
					"created":       {time.Now().UTC().Format(time.RFC3339), true},
					"updated":       {time.Now().UTC().Format(time.RFC3339), true},
				}, nil
			},
		}

		result, err := GetOne[models.User](mockHelper, "SELECT * FROM musers", dbx.Params{})
		require.NoError(t, err, "mapper allows missing fields - uses zero values")
		require.NotNil(t, result)
		assert.Equal(t, "", result.ID) // Empty string for missing string field
		assert.Equal(t, "test@example.com", result.Email)
	})

	t.Run("should return error on invalid type conversion (string to int)", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
				return dbx.NullStringMap{
					"mid":            {"invalid_id_not_a_number", true}, // ❌ Should be number
					"worker_enabled": {"true", true},
					"last_error":     {"", false},
					"updated":        {time.Now().UTC().Format(time.RFC3339), true},
				}, nil
			},
		}

		result, err := GetOne[models.SystemStatus](mockHelper, "SELECT * FROM system_status", dbx.Params{})
		assert.Error(t, err, "should error when 'mid' cannot convert to int")
		assert.Nil(t, result)
		// Error message uses field name not db tag
		assert.Contains(t, err.Error(), "ID")
	})

	t.Run("should return error on invalid boolean conversion", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
				return dbx.NullStringMap{
					"mid":            {"1", true},
					"worker_enabled": {"not_a_boolean", true}, // ❌ Invalid bool
					"last_error":     {"", false},
					"updated":        {time.Now().UTC().Format(time.RFC3339), true},
				}, nil
			},
		}

		result, err := GetOne[models.SystemStatus](mockHelper, "SELECT * FROM system_status", dbx.Params{})
		assert.Error(t, err, "should error on invalid boolean value")
		assert.Nil(t, result)
	})

	t.Run("should return error on invalid time format", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
				return dbx.NullStringMap{
					"id":            {"user123", true},
					"email":         {"test@example.com", true},
					"fcm_token":     {"token", true},
					"is_fcm_active": {"true", true},
					"created":       {"invalid-date-format", true}, // ❌ Invalid time
					"updated":       {time.Now().UTC().Format(time.RFC3339), true},
				}, nil
			},
		}

		result, err := GetOne[models.User](mockHelper, "SELECT * FROM musers", dbx.Params{})
		assert.Error(t, err, "should error on invalid time format")
		assert.Nil(t, result)
		// Error message uses field name not db tag
		assert.Contains(t, err.Error(), "Created")
	})

	t.Run("should allow NULL field (defaults to zero value)", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
				return dbx.NullStringMap{
					"mid":            {"", false}, // NULL value - mapper allows this
					"worker_enabled": {"true", true},
					"last_error":     {"", false},
					"updated":        {time.Now().UTC().Format(time.RFC3339), true},
				}, nil
			},
		}

		result, err := GetOne[models.SystemStatus](mockHelper, "SELECT * FROM system_status", dbx.Params{})
		require.NoError(t, err, "mapper allows NULL fields - uses zero values")
		require.NotNil(t, result)
		assert.Equal(t, 0, result.ID) // Zero value for missing int field
		assert.True(t, result.WorkerEnabled)
	})
}

// TestGetAll_MappingErrors tests mapping errors in batch operations
func TestGetAll_MappingErrors(t *testing.T) {
	t.Run("should return error when any row fails mapping", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			GetAllRowsFn: func(query string, params dbx.Params) ([]dbx.NullStringMap, error) {
				return []dbx.NullStringMap{
					{
						"id":            {"user1", true},
						"email":         {"user1@example.com", true},
						"fcm_token":     {"token1", true},
						"is_fcm_active": {"true", true},
						"created":       {time.Now().UTC().Format(time.RFC3339), true},
						"updated":       {time.Now().UTC().Format(time.RFC3339), true},
					},
					{
						"id":            {"user2", true},
						"email":         {"user2@example.com", true},
						"fcm_token":     {"token2", true},
						"is_fcm_active": {"invalid_boolean", true}, // ❌ Row 2 has error
						"created":       {time.Now().UTC().Format(time.RFC3339), true},
						"updated":       {time.Now().UTC().Format(time.RFC3339), true},
					},
				}, nil
			},
		}

		result, err := GetAll[models.User](mockHelper, "SELECT * FROM musers", dbx.Params{})
		assert.Error(t, err, "should error when any row fails mapping")
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "row 1") // Error on row 2 (0-indexed = row 1)
	})

	t.Run("should successfully map all rows when all valid", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			GetAllRowsFn: func(query string, params dbx.Params) ([]dbx.NullStringMap, error) {
				return []dbx.NullStringMap{
					{
						"id":            {"user1", true},
						"email":         {"user1@example.com", true},
						"fcm_token":     {"token1", true},
						"is_fcm_active": {"true", true},
						"created":       {time.Now().UTC().Format(time.RFC3339), true},
						"updated":       {time.Now().UTC().Format(time.RFC3339), true},
					},
					{
						"id":            {"user2", true},
						"email":         {"user2@example.com", true},
						"fcm_token":     {"token2", true},
						"is_fcm_active": {"false", true},
						"created":       {time.Now().UTC().Format(time.RFC3339), true},
						"updated":       {time.Now().UTC().Format(time.RFC3339), true},
					},
				}, nil
			},
		}

		result, err := GetAll[models.User](mockHelper, "SELECT * FROM musers", dbx.Params{})
		require.NoError(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, "user1", result[0].ID)
		assert.Equal(t, "user2", result[1].ID)
	})
}

// TestGetOne_DatabaseErrors tests database-level errors
func TestGetOne_DatabaseErrors(t *testing.T) {
	t.Run("should propagate database errors", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
				return nil, fmt.Errorf("database connection lost")
			},
		}

		result, err := GetOne[models.User](mockHelper, "SELECT * FROM musers", dbx.Params{})
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "database connection lost")
	})

	t.Run("should handle empty result set", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
				return nil, fmt.Errorf("no rows found")
			},
		}

		result, err := GetOne[models.User](mockHelper, "SELECT * FROM musers WHERE id = ?", dbx.Params{})
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

// MockDBHelper for testing
type MockDBHelper struct {
	GetOneRowFn  func(query string, params dbx.Params) (dbx.NullStringMap, error)
	GetAllRowsFn func(query string, params dbx.Params) ([]dbx.NullStringMap, error)
	ExecFn       func(query string, params dbx.Params) error
	CountFn      func(query string, params dbx.Params) (int, error)
	ExistsFn     func(query string, params dbx.Params) (bool, error)
}

func (m *MockDBHelper) GetOneRow(query string, params dbx.Params) (dbx.NullStringMap, error) {
	if m.GetOneRowFn != nil {
		return m.GetOneRowFn(query, params)
	}
	return nil, nil
}

func (m *MockDBHelper) GetAllRows(query string, params dbx.Params) ([]dbx.NullStringMap, error) {
	if m.GetAllRowsFn != nil {
		return m.GetAllRowsFn(query, params)
	}
	return nil, nil
}

func (m *MockDBHelper) Exec(query string, params dbx.Params) error {
	if m.ExecFn != nil {
		return m.ExecFn(query, params)
	}
	return nil
}

func (m *MockDBHelper) Count(query string, params dbx.Params) (int, error) {
	if m.CountFn != nil {
		return m.CountFn(query, params)
	}
	return 0, nil
}

func (m *MockDBHelper) Exists(query string, params dbx.Params) (bool, error) {
	if m.ExistsFn != nil {
		return m.ExistsFn(query, params)
	}
	return false, nil
}
