package pocketbase

import (
	"context"
	"errors"
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryRepo_ExecuteSelect(t *testing.T) {
	t.Run("should execute select query successfully", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			GetAllRowsFn: func(query string, params dbx.Params) ([]dbx.NullStringMap, error) {
				assert.Equal(t, "SELECT * FROM users", query)
				assert.Nil(t, params)
				
				return []dbx.NullStringMap{
					{
						"id":    {String: "user1", Valid: true},
						"email": {String: "user1@example.com", Valid: true},
						"name":  {String: "", Valid: false}, // NULL value
					},
					{
						"id":    {String: "user2", Valid: true},
						"email": {String: "user2@example.com", Valid: true},
						"name":  {String: "John Doe", Valid: true},
					},
				}, nil
			},
		}

		repo := &QueryRepo{helper: mockHelper}
		result, err := repo.ExecuteSelect(context.Background(), "SELECT * FROM users")

		require.NoError(t, err)
		require.Len(t, result, 2)
		
		// Check first row
		assert.Equal(t, "user1", result[0]["id"])
		assert.Equal(t, "user1@example.com", result[0]["email"])
		assert.Nil(t, result[0]["name"]) // NULL value should be nil
		
		// Check second row
		assert.Equal(t, "user2", result[1]["id"])
		assert.Equal(t, "user2@example.com", result[1]["email"])
		assert.Equal(t, "John Doe", result[1]["name"])
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			GetAllRowsFn: func(query string, params dbx.Params) ([]dbx.NullStringMap, error) {
				return nil, errors.New("database error")
			},
		}

		repo := &QueryRepo{helper: mockHelper}
		result, err := repo.ExecuteSelect(context.Background(), "SELECT * FROM invalid_table")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "database error")
	})

	t.Run("should return empty slice for no results", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			GetAllRowsFn: func(query string, params dbx.Params) ([]dbx.NullStringMap, error) {
				return []dbx.NullStringMap{}, nil
			},
		}

		repo := &QueryRepo{helper: mockHelper}
		result, err := repo.ExecuteSelect(context.Background(), "SELECT * FROM empty_table")

		require.NoError(t, err)
		assert.Len(t, result, 0)
	})
}

func TestQueryRepo_ExecuteInsert(t *testing.T) {
	t.Run("should execute insert query successfully", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			ExecFn: func(query string, params dbx.Params) error {
				assert.Equal(t, "INSERT INTO users (id, email) VALUES ('user1', 'test@example.com')", query)
				assert.Nil(t, params)
				return nil
			},
		}

		repo := &QueryRepo{helper: mockHelper}
		rowsAffected, lastInsertId, err := repo.ExecuteInsert(context.Background(), "INSERT INTO users (id, email) VALUES ('user1', 'test@example.com')")

		assert.NoError(t, err)
		// Note: Current implementation returns 0, 0 due to DBHelper limitation
		assert.Equal(t, int64(0), rowsAffected)
		assert.Equal(t, int64(0), lastInsertId)
	})

	t.Run("should return error when insert fails", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			ExecFn: func(query string, params dbx.Params) error {
				return errors.New("insert failed")
			},
		}

		repo := &QueryRepo{helper: mockHelper}
		rowsAffected, lastInsertId, err := repo.ExecuteInsert(context.Background(), "INSERT INTO users (id, email) VALUES ('user1', 'test@example.com')")

		assert.Error(t, err)
		assert.Equal(t, int64(0), rowsAffected)
		assert.Equal(t, int64(0), lastInsertId)
		assert.Contains(t, err.Error(), "insert failed")
	})
}

func TestQueryRepo_ExecuteUpdate(t *testing.T) {
	t.Run("should execute update query successfully", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			ExecFn: func(query string, params dbx.Params) error {
				assert.Equal(t, "UPDATE users SET email = 'new@example.com' WHERE id = 'user1'", query)
				assert.Nil(t, params)
				return nil
			},
		}

		repo := &QueryRepo{helper: mockHelper}
		rowsAffected, err := repo.ExecuteUpdate(context.Background(), "UPDATE users SET email = 'new@example.com' WHERE id = 'user1'")

		assert.NoError(t, err)
		// Note: Current implementation returns 0 due to DBHelper limitation
		assert.Equal(t, int64(0), rowsAffected)
	})

	t.Run("should return error when update fails", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			ExecFn: func(query string, params dbx.Params) error {
				return errors.New("update failed")
			},
		}

		repo := &QueryRepo{helper: mockHelper}
		rowsAffected, err := repo.ExecuteUpdate(context.Background(), "UPDATE users SET email = 'new@example.com' WHERE id = 'user1'")

		assert.Error(t, err)
		assert.Equal(t, int64(0), rowsAffected)
		assert.Contains(t, err.Error(), "update failed")
	})
}

func TestQueryRepo_ExecuteDelete(t *testing.T) {
	t.Run("should execute delete query successfully", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			ExecFn: func(query string, params dbx.Params) error {
				assert.Equal(t, "DELETE FROM users WHERE id = 'user1'", query)
				assert.Nil(t, params)
				return nil
			},
		}

		repo := &QueryRepo{helper: mockHelper}
		rowsAffected, err := repo.ExecuteDelete(context.Background(), "DELETE FROM users WHERE id = 'user1'")

		assert.NoError(t, err)
		// Note: Current implementation returns 0 due to DBHelper limitation
		assert.Equal(t, int64(0), rowsAffected)
	})

	t.Run("should return error when delete fails", func(t *testing.T) {
		mockHelper := &MockDBHelper{
			ExecFn: func(query string, params dbx.Params) error {
				return errors.New("delete failed")
			},
		}

		repo := &QueryRepo{helper: mockHelper}
		rowsAffected, err := repo.ExecuteDelete(context.Background(), "DELETE FROM users WHERE id = 'user1'")

		assert.Error(t, err)
		assert.Equal(t, int64(0), rowsAffected)
		assert.Contains(t, err.Error(), "delete failed")
	})
}

// Benchmark tests
func BenchmarkQueryRepo_ExecuteSelect(b *testing.B) {
	mockHelper := &MockDBHelper{
		GetAllRowsFn: func(query string, params dbx.Params) ([]dbx.NullStringMap, error) {
			return []dbx.NullStringMap{
				{
					"id":    {String: "user1", Valid: true},
					"email": {String: "user1@example.com", Valid: true},
				},
			}, nil
		},
	}

	repo := &QueryRepo{helper: mockHelper}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.ExecuteSelect(ctx, "SELECT * FROM users")
	}
}