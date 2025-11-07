package pocketbase

import (
	"context"
	"fmt"

	"remiaq/internal/models"
	"remiaq/internal/repository"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// UserORMRepo implements UserRepository using PocketBase ORM
type UserORMRepo struct {
	app *pocketbase.PocketBase
}

var _ repository.UserRepository = (*UserORMRepo)(nil)

const userCollectionName = "musers"

func NewUserORMRepo(app *pocketbase.PocketBase) repository.UserRepository {
	return &UserORMRepo{app: app}
}

// recordToUser converts a PocketBase Record to User model
func recordToUser(record *core.Record) (*models.User, error) {
	user := &models.User{
		ID:          record.Id,
		Email:       record.GetString("email"),
		FCMToken:    record.GetString("fcm_token"),
		IsFCMActive: record.GetBool("is_fcm_active"),
		FCMError:    record.GetString("fcm_error"),
		Created:     record.GetDateTime("created").Time(),
		Updated:     record.GetDateTime("updated").Time(),
	}
	return user, nil
}

// userToRecord converts a User model to PocketBase Record
func userToRecord(user *models.User, record *core.Record) error {
	record.Set("email", user.Email)
	record.Set("fcm_token", user.FCMToken)
	record.Set("is_fcm_active", user.IsFCMActive)
	record.Set("fcm_error", user.FCMError)
	return nil
}

// Create creates a new user
func (r *UserORMRepo) Create(ctx context.Context, user *models.User) error {
	collection, err := r.app.FindCollectionByNameOrId(userCollectionName)
	if err != nil {
		return fmt.Errorf("failed to find collection: %w", err)
	}

	record := core.NewRecord(collection)
	if err := userToRecord(user, record); err != nil {
		return err
	}

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.ID = record.Id
	user.Created = record.GetDateTime("created").Time()
	user.Updated = record.GetDateTime("updated").Time()

	return nil
}

// GetByID retrieves a user by ID
func (r *UserORMRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	record, err := r.app.FindRecordById(userCollectionName, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return recordToUser(record)
}

// GetByEmail retrieves a user by email
func (r *UserORMRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	type UserRecord struct {
		ID          string `db:"id"`
		Email       string `db:"email"`
		FCMToken    string `db:"fcm_token"`
		IsFCMActive bool   `db:"is_fcm_active"`
		Created     string `db:"created"`
		Updated     string `db:"updated"`
	}

	userRec := UserRecord{}
	err := r.app.DB().
		Select("*").
		From(userCollectionName).
		Where(dbx.HashExp{"email": email}).
		Limit(1).
		One(&userRec)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user := &models.User{
		ID:          userRec.ID,
		Email:       userRec.Email,
		FCMToken:    userRec.FCMToken,
		IsFCMActive: userRec.IsFCMActive,
		Created:     parseTime(userRec.Created),
		Updated:     parseTime(userRec.Updated),
	}

	return user, nil
}

// Update updates an existing user
func (r *UserORMRepo) Update(ctx context.Context, user *models.User) error {
	record, err := r.app.FindRecordById(userCollectionName, user.ID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if err := userToRecord(user, record); err != nil {
		return err
	}

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	user.Updated = record.GetDateTime("updated").Time()
	return nil
}

// UpdateFCMToken updates the FCM token for a user
func (r *UserORMRepo) UpdateFCMToken(ctx context.Context, userID, token string) error {
	record, err := r.app.FindRecordById(userCollectionName, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	record.Set("fcm_token", token)
	record.Set("is_fcm_active", true)

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to update fcm_token: %w", err)
	}

	return nil
}

// DisableFCM disables FCM for a user
func (r *UserORMRepo) DisableFCM(ctx context.Context, userID string) error {
	record, err := r.app.FindRecordById(userCollectionName, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	record.Set("is_fcm_active", false)

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to disable FCM: %w", err)
	}

	return nil
}

// EnableFCM enables FCM for a user with a new token
func (r *UserORMRepo) EnableFCM(ctx context.Context, userID string, token string) error {
	record, err := r.app.FindRecordById(userCollectionName, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	record.Set("fcm_token", token)
	record.Set("is_fcm_active", true)

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to enable FCM: %w", err)
	}

	return nil
}

// GetActiveUsers retrieves all users with FCM enabled
func (r *UserORMRepo) GetActiveUsers(ctx context.Context) ([]*models.User, error) {
	type UserRecord struct {
		ID          string `db:"id"`
		Email       string `db:"email"`
		FCMToken    string `db:"fcm_token"`
		IsFCMActive bool   `db:"is_fcm_active"`
		Created     string `db:"created"`
		Updated     string `db:"updated"`
	}

	records := []UserRecord{}
	err := r.app.DB().
		Select("*").
		From(userCollectionName).
		Where(dbx.NewExp("is_fcm_active = {:is_fcm_active} AND (fcm_error IS NULL OR fcm_error = '')", dbx.Params{"is_fcm_active": true})).
		OrderBy("created DESC").
		All(&records)
	if err != nil {
		return nil, fmt.Errorf("failed to query active users: %w", err)
	}

	users := make([]*models.User, 0, len(records))
	for _, rec := range records {
		user := &models.User{
			ID:          rec.ID,
			Email:       rec.Email,
			FCMToken:    rec.FCMToken,
			IsFCMActive: rec.IsFCMActive,
			Created:     parseTime(rec.Created),
			Updated:     parseTime(rec.Updated),
		}
		users = append(users, user)
	}

	return users, nil
}

// SetFCMError sets FCM error message for a user
func (r *UserORMRepo) SetFCMError(ctx context.Context, userID, errorMsg string) error {
	record, err := r.app.FindRecordById(userCollectionName, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	record.Set("fcm_error", errorMsg)

	if err := r.app.Save(record); err != nil {
		return fmt.Errorf("failed to set FCM error: %w", err)
	}

	return nil
}