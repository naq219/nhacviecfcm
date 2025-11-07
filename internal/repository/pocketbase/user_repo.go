package pocketbase

import (
	"context"
	"time"

	"remiaq/internal/db"
	"remiaq/internal/models"
	"remiaq/internal/repository"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
)

// UserRepo implements repository.UserRepository
type UserRepo struct {
	helper db.DBHelperInterface // Use interface from db package
}

// Ensure implementation
var _ repository.UserRepository = (*UserRepo)(nil)

// NewUserRepo creates a new user repository
func NewUserRepo(app *pocketbase.PocketBase) repository.UserRepository {
	return &UserRepo{helper: db.NewDBHelper(app)}
}

// Create inserts a new user
func (r *UserRepo) Create(ctx context.Context, user *models.User) error {
	return r.helper.Exec(
		`INSERT INTO musers (id, email, fcm_token, is_fcm_active, created, updated)
		 VALUES ({:id}, {:email}, {:fcm_token}, {:is_fcm_active}, {:created}, {:updated})`,
		dbx.Params{
			"id":            user.ID,
			"email":         user.Email,
			"fcm_token":     user.FCMToken,
			"is_fcm_active": user.IsFCMActive,
			"created":       time.Now().UTC(),
			"updated":       time.Now().UTC(),
		},
	)
}

// GetByID retrieves a user by ID
func (r *UserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	return db.GetOne[models.User](
		r.helper,
		"SELECT * FROM musers WHERE id = {:id}",
		dbx.Params{"id": id},
	)
}

// GetByEmail retrieves a user by email
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return db.GetOne[models.User](
		r.helper,
		"SELECT * FROM musers WHERE email = {:email}",
		dbx.Params{"email": email},
	)
}

// Update updates user information
func (r *UserRepo) Update(ctx context.Context, user *models.User) error {
	return r.helper.Exec(
		`UPDATE musers 
		 SET email = {:email}, fcm_token = {:fcm_token}, is_fcm_active = {:is_fcm_active}, updated = {:updated}
		 WHERE id = {:id}`,
		dbx.Params{
			"email":         user.Email,
			"fcm_token":     user.FCMToken,
			"is_fcm_active": user.IsFCMActive,
			"updated":       time.Now().UTC(),
			"id":            user.ID,
		},
	)
}

// UpdateFCMToken updates only the FCM token
func (r *UserRepo) UpdateFCMToken(ctx context.Context, userID, token string) error {
	return r.helper.Exec(
		"UPDATE musers SET fcm_token = {:token}, is_fcm_active = TRUE, updated = {:updated} WHERE id = {:id}",
		dbx.Params{
			"token":   token,
			"updated": time.Now().UTC(),
			"id":      userID,
		},
	)
}

// DisableFCM disables FCM for a user (token invalid)
func (r *UserRepo) DisableFCM(ctx context.Context, userID string) error {
	return r.helper.Exec(
		"UPDATE musers SET is_fcm_active = FALSE, fcm_token = NULL, updated = {:updated} WHERE id = {:id}",
		dbx.Params{
			"updated": time.Now().UTC(),
			"id":      userID,
		},
	)
}

// EnableFCM re-enables FCM with a new token
func (r *UserRepo) EnableFCM(ctx context.Context, userID string, token string) error {
	return r.helper.Exec(
		"UPDATE musers SET fcm_token = {:token}, is_fcm_active = TRUE, updated = {:updated} WHERE id = {:id}",
		dbx.Params{
			"token":   token,
			"updated": time.Now().UTC(),
			"id":      userID,
		},
	)
}

// SetFCMError sets FCM error message for a user
func (r *UserRepo) SetFCMError(ctx context.Context, userID, errorMsg string) error {
	return r.helper.Exec(
		"UPDATE musers SET fcm_error = {:error_msg}, updated = {:updated} WHERE id = {:id}",
		dbx.Params{
			"error_msg": errorMsg,
			"updated":   time.Now().UTC(),
			"id":        userID,
		},
	)
}

// GetActiveUsers retrieves all users with active FCM and no FCM errors
func (r *UserRepo) GetActiveUsers(ctx context.Context) ([]*models.User, error) {
	users, err := db.GetAll[models.User](
		r.helper,
		`SELECT * FROM musers 
		 WHERE is_fcm_active = TRUE 
		   AND fcm_token IS NOT NULL 
		   AND fcm_token != ''
		   AND (fcm_error IS NULL OR fcm_error = '')`,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Convert []models.User to []*models.User
	result := make([]*models.User, len(users))
	for i := range users {
		result[i] = &users[i]
	}
	return result, nil
}
