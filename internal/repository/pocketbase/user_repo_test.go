package pocketbase

import (
	"context"
	"errors"
	"testing"
	"time"

	"remiaq/internal/models"

	"github.com/pocketbase/dbx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create mock User row
func mockUserRow(id, email, fcmToken string, isFCMActive bool) dbx.NullStringMap {
	return dbx.NullStringMap{
		"id":            {id, true},
		"email":         {email, true},
		"fcm_token":     {fcmToken, fcmToken != ""},
		"is_fcm_active": {boolToString(isFCMActive), true},
		"created":       {time.Now().UTC().Format(time.RFC3339), true},
		"updated":       {time.Now().UTC().Format(time.RFC3339), true},
	}
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func TestUserRepo_Create(t *testing.T) {
	t.Run("should create user successfully", func(t *testing.T) {
		execCalled := false
		user := &models.User{
			ID:          "user123",
			Email:       "test@example.com",
			FCMToken:    "fcm_token_123",
			IsFCMActive: true,
		}

		repo := &UserRepo{
			helper: &MockDBHelper{
				ExecFn: func(query string, params dbx.Params) error {
					execCalled = true
					assert.Contains(t, query, "INSERT INTO musers")
					assert.Equal(t, user.ID, params["id"])
					assert.Equal(t, user.Email, params["email"])
					assert.Equal(t, user.FCMToken, params["fcm_token"])
					assert.Equal(t, user.IsFCMActive, params["is_fcm_active"])
					assert.NotNil(t, params["created"])
					assert.NotNil(t, params["updated"])
					return nil
				},
			},
		}

		err := repo.Create(context.Background(), user)
		require.NoError(t, err)
		assert.True(t, execCalled)
	})

	t.Run("should return error when exec fails", func(t *testing.T) {
		user := &models.User{ID: "user123", Email: "test@example.com"}
		repo := &UserRepo{
			helper: &MockDBHelper{
				ExecFn: func(query string, params dbx.Params) error {
					return errors.New("database error")
				},
			},
		}

		err := repo.Create(context.Background(), user)
		require.Error(t, err)
	})
}

func TestUserRepo_GetByID(t *testing.T) {
	t.Run("should return user successfully", func(t *testing.T) {
		userID := "user123"
		repo := &UserRepo{
			helper: &MockDBHelper{
				GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
					assert.Contains(t, query, "SELECT * FROM musers WHERE id = {:id}")
					assert.Equal(t, userID, params["id"])
					return mockUserRow(userID, "test@example.com", "fcm_token", true), nil
				},
			},
		}

		result, err := repo.GetByID(context.Background(), userID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, userID, result.ID)
		assert.Equal(t, "test@example.com", result.Email)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		repo := &UserRepo{
			helper: &MockDBHelper{
				GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
					return nil, errors.New("user not found")
				},
			},
		}

		result, err := repo.GetByID(context.Background(), "nonexistent")
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestUserRepo_GetByEmail(t *testing.T) {
	t.Run("should return user by email successfully", func(t *testing.T) {
		email := "test@example.com"
		repo := &UserRepo{
			helper: &MockDBHelper{
				GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
					assert.Contains(t, query, "SELECT * FROM musers WHERE email = {:email}")
					assert.Equal(t, email, params["email"])
					return mockUserRow("user123", email, "fcm_token", true), nil
				},
			},
		}

		result, err := repo.GetByEmail(context.Background(), email)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, email, result.Email)
		assert.Equal(t, "user123", result.ID)
	})
}

func TestUserRepo_Update(t *testing.T) {
	t.Run("should update user successfully", func(t *testing.T) {
		execCalled := false
		user := &models.User{
			ID:          "user123",
			Email:       "updated@example.com",
			FCMToken:    "new_fcm_token",
			IsFCMActive: false,
		}

		repo := &UserRepo{
			helper: &MockDBHelper{
				ExecFn: func(query string, params dbx.Params) error {
					execCalled = true
					assert.Contains(t, query, "UPDATE musers")
					assert.Contains(t, query, "WHERE id = {:id}")
					assert.Equal(t, user.Email, params["email"])
					assert.Equal(t, user.FCMToken, params["fcm_token"])
					assert.Equal(t, user.IsFCMActive, params["is_fcm_active"])
					assert.Equal(t, user.ID, params["id"])
					return nil
				},
			},
		}

		err := repo.Update(context.Background(), user)
		require.NoError(t, err)
		assert.True(t, execCalled)
	})
}

func TestUserRepo_UpdateFCMToken(t *testing.T) {
	t.Run("should update FCM token successfully", func(t *testing.T) {
		userID := "user123"
		token := "new_fcm_token"
		execCalled := false

		repo := &UserRepo{
			helper: &MockDBHelper{
				ExecFn: func(query string, params dbx.Params) error {
					execCalled = true
					assert.Contains(t, query, "UPDATE musers SET fcm_token = {:token}, is_fcm_active = TRUE")
					assert.Equal(t, token, params["token"])
					assert.Equal(t, userID, params["id"])
					return nil
				},
			},
		}

		err := repo.UpdateFCMToken(context.Background(), userID, token)
		require.NoError(t, err)
		assert.True(t, execCalled)
	})
}

func TestUserRepo_DisableFCM(t *testing.T) {
	t.Run("should disable FCM successfully", func(t *testing.T) {
		userID := "user123"
		execCalled := false

		repo := &UserRepo{
			helper: &MockDBHelper{
				ExecFn: func(query string, params dbx.Params) error {
					execCalled = true
					assert.Contains(t, query, "UPDATE musers SET is_fcm_active = FALSE, fcm_token = NULL")
					assert.Equal(t, userID, params["id"])
					return nil
				},
			},
		}

		err := repo.DisableFCM(context.Background(), userID)
		require.NoError(t, err)
		assert.True(t, execCalled)
	})
}

func TestUserRepo_EnableFCM(t *testing.T) {
	t.Run("should enable FCM with new token", func(t *testing.T) {
		userID := "user123"
		token := "new_token"
		execCalled := false

		repo := &UserRepo{
			helper: &MockDBHelper{
				ExecFn: func(query string, params dbx.Params) error {
					execCalled = true
					assert.Contains(t, query, "UPDATE musers SET fcm_token = {:token}, is_fcm_active = TRUE")
					assert.Equal(t, token, params["token"])
					assert.Equal(t, userID, params["id"])
					return nil
				},
			},
		}

		err := repo.EnableFCM(context.Background(), userID, token)
		require.NoError(t, err)
		assert.True(t, execCalled)
	})
}

func TestUserRepo_GetActiveUsers(t *testing.T) {
	t.Run("should return active users successfully", func(t *testing.T) {
		repo := &UserRepo{
			helper: &MockDBHelper{
				GetAllRowsFn: func(query string, params dbx.Params) ([]dbx.NullStringMap, error) {
					assert.Contains(t, query, "SELECT * FROM musers")
					assert.Contains(t, query, "WHERE is_fcm_active = TRUE")
					assert.Contains(t, query, "AND fcm_token IS NOT NULL")
					return []dbx.NullStringMap{
						mockUserRow("user1", "user1@example.com", "token1", true),
						mockUserRow("user2", "user2@example.com", "token2", true),
					}, nil
				},
			},
		}

		result, err := repo.GetActiveUsers(context.Background())
		require.NoError(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, "user1", result[0].ID)
		assert.Equal(t, "user2", result[1].ID)
	})

	t.Run("should return empty slice when no active users", func(t *testing.T) {
		repo := &UserRepo{
			helper: &MockDBHelper{
				GetAllRowsFn: func(query string, params dbx.Params) ([]dbx.NullStringMap, error) {
					return []dbx.NullStringMap{}, nil
				},
			},
		}

		result, err := repo.GetActiveUsers(context.Background())
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

// Benchmark tests
func BenchmarkUserRepo_GetByID(b *testing.B) {
	repo := &UserRepo{
		helper: &MockDBHelper{
			GetOneRowFn: func(query string, params dbx.Params) (dbx.NullStringMap, error) {
				return mockUserRow("user123", "test@example.com", "token", true), nil
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByID(context.Background(), "user123")
	}
}

func BenchmarkUserRepo_GetActiveUsers(b *testing.B) {
	repo := &UserRepo{
		helper: &MockDBHelper{
			GetAllRowsFn: func(query string, params dbx.Params) ([]dbx.NullStringMap, error) {
				return []dbx.NullStringMap{
					mockUserRow("user1", "user1@example.com", "token1", true),
					mockUserRow("user2", "user2@example.com", "token2", true),
				}, nil
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetActiveUsers(context.Background())
	}
}