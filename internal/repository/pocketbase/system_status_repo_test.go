package pocketbase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDBHelper is a simple mock for db.DBHelperInterface
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

// Helper function to create mock SystemStatus row
func mockSystemStatusRow(enabled bool, lastError string) dbx.NullStringMap {
	hasError := lastError != ""
	return dbx.NullStringMap{
		"mid":            {fmt.Sprint("1"), true},
		"worker_enabled": {fmt.Sprint(enabled), true},
		"last_error":     {lastError, hasError},
		"updated":        {time.Now().UTC().String(), true},
	}
}

func TestSystemStatusRepo_Get(t *testing.T) {
	t.Run("should return system status successfully", func(t *testing.T) {
		repo := &SystemStatusRepo{
			helper: &MockDBHelper{
				GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
					assert.Contains(t, query, "SELECT * FROM system_status WHERE mid = {:mid}")
					assert.Equal(t, 1, params["mid"])
					return mockSystemStatusRow(true, ""), nil
				},
			},
		}

		result, err := repo.Get(context.Background())
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.True(t, result.WorkerEnabled)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		repo := &SystemStatusRepo{
			helper: &MockDBHelper{
				GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
					return nil, errors.New("database connection failed")
				},
			},
		}

		result, err := repo.Get(context.Background())
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestSystemStatusRepo_IsWorkerEnabled(t *testing.T) {
	t.Run("should return true when worker is enabled", func(t *testing.T) {
		repo := &SystemStatusRepo{
			helper: &MockDBHelper{
				GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
					return mockSystemStatusRow(true, ""), nil
				},
			},
		}

		enabled, err := repo.IsWorkerEnabled(context.Background())
		require.NoError(t, err)
		assert.True(t, enabled)
	})

	t.Run("should return false when worker is disabled", func(t *testing.T) {
		repo := &SystemStatusRepo{
			helper: &MockDBHelper{
				GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
					return mockSystemStatusRow(false, "connection timeout"), nil
				},
			},
		}

		enabled, err := repo.IsWorkerEnabled(context.Background())
		require.NoError(t, err)
		assert.False(t, enabled)
	})

	t.Run("should propagate error from Get", func(t *testing.T) {
		repo := &SystemStatusRepo{
			helper: &MockDBHelper{
				GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
					return nil, errors.New("database error")
				},
			},
		}

		_, err := repo.IsWorkerEnabled(context.Background())
		require.Error(t, err)
	})
}

func TestSystemStatusRepo_EnableWorker(t *testing.T) {
	t.Run("should enable worker successfully", func(t *testing.T) {
		execCalled := false
		beforeTime := time.Now().UTC()

		repo := &SystemStatusRepo{
			helper: &MockDBHelper{
				ExecFn: func(query string, params dbx.Params) error {
					execCalled = true
					assert.Contains(t, query, "UPDATE system_status SET worker_enabled = TRUE")
					assert.Contains(t, query, "updated = {:updated}")

					updated := params["updated"].(time.Time)
					assert.True(t, updated.After(beforeTime) || updated.Equal(beforeTime))
					return nil
				},
			},
		}

		err := repo.EnableWorker(context.Background())
		require.NoError(t, err)
		assert.True(t, execCalled)
	})

	t.Run("should return error if exec fails", func(t *testing.T) {
		repo := &SystemStatusRepo{
			helper: &MockDBHelper{
				ExecFn: func(query string, params dbx.Params) error {
					return errors.New("database error")
				},
			},
		}

		err := repo.EnableWorker(context.Background())
		require.Error(t, err)
	})
}

func TestSystemStatusRepo_DisableWorker(t *testing.T) {
	t.Run("should disable worker with error message", func(t *testing.T) {
		errorMsg := "connection timeout"
		execCalled := false

		repo := &SystemStatusRepo{
			helper: &MockDBHelper{
				ExecFn: func(query string, params dbx.Params) error {
					execCalled = true
					assert.Contains(t, query, "UPDATE system_status SET worker_enabled = FALSE")
					assert.Contains(t, query, "last_error = {:error_msg}")
					assert.Equal(t, errorMsg, params["error_msg"])
					assert.NotNil(t, params["updated"])
					return nil
				},
			},
		}

		err := repo.DisableWorker(context.Background(), errorMsg)
		require.NoError(t, err)
		assert.True(t, execCalled)
	})

	t.Run("should handle empty error message", func(t *testing.T) {
		repo := &SystemStatusRepo{
			helper: &MockDBHelper{
				ExecFn: func(query string, params dbx.Params) error {
					assert.Equal(t, "", params["error_msg"])
					return nil
				},
			},
		}

		err := repo.DisableWorker(context.Background(), "")
		require.NoError(t, err)
	})
}

func TestSystemStatusRepo_UpdateError(t *testing.T) {
	t.Run("should update error message", func(t *testing.T) {
		errorMsg := "database error occurred"
		execCalled := false

		repo := &SystemStatusRepo{
			helper: &MockDBHelper{
				ExecFn: func(query string, params dbx.Params) error {
					execCalled = true
					assert.Contains(t, query, "UPDATE system_status SET last_error = {:error_msg}")
					assert.Equal(t, errorMsg, params["error_msg"])
					assert.NotNil(t, params["updated"])
					return nil
				},
			},
		}

		err := repo.UpdateError(context.Background(), errorMsg)
		require.NoError(t, err)
		assert.True(t, execCalled)
	})

	t.Run("should handle long error messages", func(t *testing.T) {
		longError := "this is a very long error message that contains multiple lines and special characters: \n\t!@#$%"
		repo := &SystemStatusRepo{
			helper: &MockDBHelper{
				ExecFn: func(query string, params dbx.Params) error {
					assert.Equal(t, longError, params["error_msg"])
					return nil
				},
			},
		}

		err := repo.UpdateError(context.Background(), longError)
		require.NoError(t, err)
	})
}

func TestSystemStatusRepo_ClearError(t *testing.T) {
	t.Run("should clear error message", func(t *testing.T) {
		execCalled := false

		repo := &SystemStatusRepo{
			helper: &MockDBHelper{
				ExecFn: func(query string, params dbx.Params) error {
					execCalled = true
					assert.Contains(t, query, "UPDATE system_status SET last_error = ''")
					assert.NotNil(t, params["updated"])
					return nil
				},
			},
		}

		err := repo.ClearError(context.Background())
		require.NoError(t, err)
		assert.True(t, execCalled)
	})

	t.Run("should return error if exec fails", func(t *testing.T) {
		repo := &SystemStatusRepo{
			helper: &MockDBHelper{
				ExecFn: func(query string, params dbx.Params) error {
					return errors.New("database error")
				},
			},
		}

		err := repo.ClearError(context.Background())
		require.Error(t, err)
	})
}

// Benchmark tests
func BenchmarkSystemStatusRepo_Get(b *testing.B) {
	repo := &SystemStatusRepo{
		helper: &MockDBHelper{
			GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
				return mockSystemStatusRow(true, ""), nil
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Get(context.Background())
	}
}

func BenchmarkSystemStatusRepo_IsWorkerEnabled(b *testing.B) {
	repo := &SystemStatusRepo{
		helper: &MockDBHelper{
			GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
				return mockSystemStatusRow(true, ""), nil
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.IsWorkerEnabled(context.Background())
	}
}

func BenchmarkSystemStatusRepo_DisableWorker(b *testing.B) {
	repo := &SystemStatusRepo{
		helper: &MockDBHelper{
			ExecFn: func(query string, params dbx.Params) error {
				return nil
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.DisableWorker(context.Background(), "benchmark error")
	}
}
