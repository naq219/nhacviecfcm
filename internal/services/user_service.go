package services

import (
	"context"
	"fmt"

	"remiaq/internal/repository"
)

// UserService handles business logic for users
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// UpdateFCMToken updates the FCM token for a user
func (s *UserService) UpdateFCMToken(ctx context.Context, userID, fcmToken string) error {
	// Logic nghiệp vụ: Cập nhật FCM token cho người dùng
	if err := s.userRepo.UpdateFCMToken(ctx, userID, fcmToken); err != nil {
		return fmt.Errorf("failed to update FCM token in repository: %w", err)
	}
	return nil
}