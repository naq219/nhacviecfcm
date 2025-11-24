package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID          string    `json:"id" db:"id"`
	Email       string    `json:"email" db:"email"`
	Username    string    `json:"username" db:"username"`
	FCMToken    string    `json:"fcm_token" db:"fcm_token"`
	IsFCMActive bool      `json:"is_fcm_active" db:"is_fcm_active"`
	FCMError    string    `json:"fcm_error" db:"fcm_error"`
	Created     time.Time `json:"created" db:"created"`
	Updated     time.Time `json:"updated" db:"updated"`
}

// Validate checks if user data is valid
func (u *User) Validate() error {
	// Add validation logic here if needed
	return nil
}