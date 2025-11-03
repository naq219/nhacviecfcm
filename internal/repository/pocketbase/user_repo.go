package pocketbase

import (
	"context"
	"encoding/json"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/repository"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
)

// PocketBaseUserRepo implements UserRepository for PocketBase
type PocketBaseUserRepo struct {
	app *pocketbase.PocketBase
}

// Ensure implementation
var _ repository.UserRepository = (*PocketBaseUserRepo)(nil)

// NewPocketBaseUserRepo creates a new user repository
func NewPocketBaseUserRepo(app *pocketbase.PocketBase) repository.UserRepository {
	return &PocketBaseUserRepo{app: app}
}

// Create inserts a new user
func (r *PocketBaseUserRepo) Create(ctx context.Context, user *models.User) error {
    query := `
        INSERT INTO musers (id, email, fcm_token, is_fcm_active, created, updated)
        VALUES ({:id}, {:email}, {:fcm_token}, {:is_fcm_active}, {:created}, {:updated})
    `

	_, err := r.app.DB().NewQuery(query).Bind(dbx.Params{
		"id":            user.ID,
		"email":         user.Email,
		"fcm_token":     user.FCMToken,
		"is_fcm_active": user.IsFCMActive,
		"created":       time.Now().UTC(),
		"updated":       time.Now().UTC(),
	}).Execute()

	return err
}

// GetByID retrieves a user by ID
func (r *PocketBaseUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
    query := `SELECT * FROM musers WHERE id = {:id}`

	q := r.app.DB().NewQuery(query)
	q.Bind(dbx.Params{
		"id": id,
	})

	var rawResult dbx.NullStringMap
	err := q.One(&rawResult)
	if err != nil {
		return nil, err
	}

	return r.mapToUser(rawResult)
}

// GetByEmail retrieves a user by email
func (r *PocketBaseUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
    query := `SELECT * FROM musers WHERE email = {:email}`

	q := r.app.DB().NewQuery(query)
	q.Bind(dbx.Params{
		"email": email,
	})

	var rawResult dbx.NullStringMap
	err := q.One(&rawResult)
	if err != nil {
		return nil, err
	}

	return r.mapToUser(rawResult)
}

// Update updates user information
func (r *PocketBaseUserRepo) Update(ctx context.Context, user *models.User) error {
    query := `
        UPDATE musers 
        SET email = {:email}, fcm_token = {:fcm_token}, is_fcm_active = {:is_fcm_active}, updated = {:updated}
        WHERE id = {:id}
    `
	q := r.app.DB().NewQuery(query)
	q.Bind(dbx.Params{
		"email": user.Email,
		"fcm_token": user.FCMToken,
		"is_fcm_active": user.IsFCMActive,
		"updated": time.Now().UTC(),
		"id": user.ID,
	})
	_, err := q.Execute()
	return err
}

// UpdateFCMToken updates only the FCM token
func (r *PocketBaseUserRepo) UpdateFCMToken(ctx context.Context, userID, token string) error {
    query := `UPDATE musers SET fcm_token = {:token}, is_fcm_active = TRUE, updated = {:updated} WHERE id = {:id}`
	q := r.app.DB().NewQuery(query)
	q.Bind(dbx.Params{
		"token": token,
		"updated": time.Now().UTC(),
		"id": userID,
	})
	_, err := q.Execute()
	return err
}

// DisableFCM disables FCM for a user (token invalid)
func (r *PocketBaseUserRepo) DisableFCM(ctx context.Context, userID string) error {
    query := `UPDATE musers SET is_fcm_active = FALSE, fcm_token = NULL, updated = {:updated} WHERE id = {:id}`
	q := r.app.DB().NewQuery(query)
	q.Bind(dbx.Params{
		"updated": time.Now().UTC(),
		"id": userID,
	})
	_, err := q.Execute()
	return err
}

// EnableFCM re-enables FCM with a new token
func (r *PocketBaseUserRepo) EnableFCM(ctx context.Context, userID string, token string) error {
    query := `UPDATE musers SET fcm_token = {:token}, is_fcm_active = TRUE, updated = {:updated} WHERE id = {:id}`
	q := r.app.DB().NewQuery(query)
	q.Bind(dbx.Params{
		"token": token,
		"updated": time.Now().UTC(),
		"id": userID,
	})
	_, err := q.Execute()
	return err
}

// GetActiveUsers retrieves all users with active FCM
func (r *PocketBaseUserRepo) GetActiveUsers(ctx context.Context) ([]*models.User, error) {
    query := `
        SELECT * FROM musers 
        WHERE is_fcm_active = TRUE 
          AND fcm_token IS NOT NULL 
          AND fcm_token != ''
    `

	var rawResults []dbx.NullStringMap
	err := r.app.DB().NewQuery(query).All(&rawResults)
	if err != nil {
		return nil, err
	}

	return r.mapToUsers(rawResults)
}

// Helper functions

func (r *PocketBaseUserRepo) mapToUser(raw dbx.NullStringMap) (*models.User, error) {
	user := &models.User{}

	user.ID = raw["id"].String
	user.Email = raw["email"].String
	user.FCMToken = raw["fcm_token"].String

	// Parse boolean
	if raw["is_fcm_active"].Valid {
		var val bool
		json.Unmarshal([]byte(raw["is_fcm_active"].String), &val)
		user.IsFCMActive = val
	}

	// Parse timestamps
	if raw["created"].Valid {
		t, _ := time.Parse(time.RFC3339, raw["created"].String)
		user.Created = t
	}
	if raw["updated"].Valid {
		t, _ := time.Parse(time.RFC3339, raw["updated"].String)
		user.Updated = t
	}

	return user, nil
}

func (r *PocketBaseUserRepo) mapToUsers(rawList []dbx.NullStringMap) ([]*models.User, error) {
	users := make([]*models.User, 0, len(rawList))

	for _, raw := range rawList {
		user, err := r.mapToUser(raw)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
