package worker

import (
	"context"
	"fmt"

	"remiaq/internal/models"
	"remiaq/internal/services/fcmutils"
)

// IsWorkerSystemEnabled checks if the worker is enabled in the system settings.
// It logs an error if the check fails.
// Returns true if enabled, false otherwise.
func IsWorkerSystemEnabled(ctx context.Context, sysRepo SystemStatusRepo, logError func(format string, v ...any)) bool {
	enabled, err := sysRepo.IsWorkerEnabled(ctx)
	if err != nil {
		logError("Worker: failed to check system status: %v", err)
		return false
	}
	return enabled
}

// SendNotification sends FCM notification to user
// If sending fails, it updates system_status with error and disables worker
func SendNotification(
	ctx context.Context,
	reminder *models.Reminder,
	userRepo UserRepo,
	sysRepo SystemStatusRepo,
	logError func(format string, v ...any),
) error {
	// Get user
	user, err := userRepo.GetByID(ctx, reminder.UserID)
	if err != nil {
		errMsg := fmt.Sprintf("User not found for reminder %s: %v", reminder.ID, err)
		logError(errMsg)
		// Don't disable worker for user not found (might be user-specific issue)
		return fmt.Errorf("user not found: %w", err)
	}

	// Check FCM active
	if !user.IsFCMActive || user.FCMToken == "" {
		logError("User FCM not active for reminder %s (user: %s)", reminder.ID, user.Email)
		// Don't disable worker (user-specific issue)
		return fmt.Errorf("user FCM not active")
	}

	// Send FCM notification
	_, err = fcmutils.SendFCMNotification(ctx, reminder.Title, reminder.Description, user.FCMToken, user.Email)
	if err != nil {
		// FCM sending failed - this is a system error
		errMsg := fmt.Sprintf("FCM send failed for reminder %s: %v", reminder.ID, err)
		logError(errMsg)

		// Update system_status with error
		if updateErr := sysRepo.UpdateError(ctx, errMsg); updateErr != nil {
			logError("Failed to update system error: %v", updateErr)
		}

		// Disable worker to prevent further errors
		if disableErr := sysRepo.DisableWorker(ctx); disableErr != nil {
			logError("Failed to disable worker: %v", disableErr)
		}

		return fmt.Errorf("fcm send failed: %w", err)
	}

	return nil
}
