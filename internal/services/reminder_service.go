package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/repository"
)

// ReminderService handles reminder business logic
type ReminderService struct {
	reminderRepo    repository.ReminderRepository
	userRepo        repository.UserRepository
	fcmService      FCMServiceInterface
	schedCalculator *ScheduleCalculator
}

// NewReminderService creates a new reminder service
func NewReminderService(
	reminderRepo repository.ReminderRepository,
	userRepo repository.UserRepository,
	fcmService FCMServiceInterface,
	schedCalculator *ScheduleCalculator,
) *ReminderService {
	return &ReminderService{
		reminderRepo:    reminderRepo,
		userRepo:        userRepo,
		fcmService:      fcmService,
		schedCalculator: schedCalculator,
	}
}

// CreateReminder creates a new reminder
func (s *ReminderService) CreateReminder(ctx context.Context, reminder *models.Reminder) error {
	// Validate
	if err := reminder.Validate(); err != nil {
		return err
	}

	// PocketBase sẽ tự tạo ID (giới hạn 15 ký tự)
	// Không tạo ID ở đây để tránh xung đột
	reminder.ID = ""

	// Set default values
	if reminder.Status == "" {
		reminder.Status = models.ReminderStatusActive
	}
	if reminder.RepeatStrategy == "" {
		reminder.RepeatStrategy = models.RepeatStrategyNone
	}
	if reminder.CalendarType == "" {
		reminder.CalendarType = models.CalendarTypeSolar
	}

	// Calculate next trigger time if not set
	if reminder.NextTriggerAt == "" {
		nextTrigger, err := s.schedCalculator.CalculateNextTrigger(reminder, time.Now())
		if err != nil {
			return err
		}
		reminder.NextTriggerAt = nextTrigger.Format(time.RFC3339)
	}

	if err := s.reminderRepo.Create(ctx, reminder); err != nil {
		return fmt.Errorf("failed to create reminder: %w", err)
	}

	return nil
}

// GetReminder retrieves a reminder by ID
func (s *ReminderService) GetReminder(ctx context.Context, id string) (*models.Reminder, error) {
	return s.reminderRepo.GetByID(ctx, id)
}

// UpdateReminder updates a reminder
func (s *ReminderService) UpdateReminder(ctx context.Context, reminder *models.Reminder) error {
	// For partial updates, get the existing reminder first to validate only changed fields
	existingReminder, err := s.reminderRepo.GetByID(ctx, reminder.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing reminder: %w", err)
	}

	// Merge the updates into the existing reminder
	if reminder.Title != "" {
		existingReminder.Title = reminder.Title
	}
	if reminder.Description != "" {
		existingReminder.Description = reminder.Description
	}
	if reminder.Type != "" {
		existingReminder.Type = reminder.Type
	}
	if reminder.CalendarType != "" {
		existingReminder.CalendarType = reminder.CalendarType
	}
	if reminder.NextTriggerAt != "" {
		existingReminder.NextTriggerAt = reminder.NextTriggerAt
	}
	if reminder.TriggerTimeOfDay != "" {
		existingReminder.TriggerTimeOfDay = reminder.TriggerTimeOfDay
	}
	if reminder.RecurrencePattern != nil {
		existingReminder.RecurrencePattern = reminder.RecurrencePattern
	}
	if reminder.RepeatStrategy != "" {
		existingReminder.RepeatStrategy = reminder.RepeatStrategy
	}
	if reminder.RetryIntervalSec != 0 {
		existingReminder.RetryIntervalSec = reminder.RetryIntervalSec
	}
	if reminder.MaxRetries != 0 {
		existingReminder.MaxRetries = reminder.MaxRetries
	}
	if reminder.RetryCount != 0 {
		existingReminder.RetryCount = reminder.RetryCount
	}
	if reminder.Status != "" {
		existingReminder.Status = reminder.Status
	}
	if reminder.SnoozeUntil != "" {
		existingReminder.SnoozeUntil = reminder.SnoozeUntil
	}
	if reminder.LastCompletedAt != "" {
		existingReminder.LastCompletedAt = reminder.LastCompletedAt
	}
	if reminder.LastSentAt != "" {
		existingReminder.LastSentAt = reminder.LastSentAt
	}
	
	// Always preserve the user_id from existing reminder to prevent it from being cleared
	if reminder.UserID != "" {
		existingReminder.UserID = reminder.UserID
	}

	// Validate the merged reminder
	if err := existingReminder.Validate(); err != nil {
		return err
	}

	// Recalculate next trigger time if needed for recurring reminders
	if existingReminder.Type == models.ReminderTypeRecurring {
		now := time.Now().UTC()
		nextTriggerTime, err := s.schedCalculator.CalculateNextTrigger(existingReminder, now)
		if err != nil {
			return fmt.Errorf("failed to calculate next trigger time for update: %w", err)
		}
		existingReminder.NextTriggerAt = nextTriggerTime.Format(time.RFC3339)
	}

	return s.reminderRepo.Update(ctx, existingReminder)
}

// DeleteReminder deletes a reminder
func (s *ReminderService) DeleteReminder(ctx context.Context, id string) error {
	return s.reminderRepo.Delete(ctx, id)
}

// GetUserReminders retrieves all reminders for a user
func (s *ReminderService) GetUserReminders(ctx context.Context, userID string) ([]*models.Reminder, error) {
	return s.reminderRepo.GetByUserID(ctx, userID)
}

// SnoozeReminder postpones a reminder
func (s *ReminderService) SnoozeReminder(ctx context.Context, id string, duration time.Duration) error {
	snoozeUntil := time.Now().Add(duration)
	return s.reminderRepo.UpdateSnooze(ctx, id, snoozeUntil.Format(time.RFC3339))
}

// CompleteReminder marks a reminder as completed
func (s *ReminderService) CompleteReminder(ctx context.Context, id string) error {
	reminder, err := s.reminderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)
	reminder.LastCompletedAt = nowStr
	reminder.Status = models.ReminderStatusCompleted

	// For one-time reminders, we are done.
	if reminder.Type == models.ReminderTypeOneTime {
		return s.reminderRepo.Update(ctx, reminder)
	}

	// For recurring reminders, calculate the next trigger
	if reminder.Type == models.ReminderTypeRecurring {
		nextTrigger, err := s.schedCalculator.CalculateNextTrigger(reminder, now)
		if err != nil {
			// Log the error but don't block completion
			fmt.Printf("WARN: could not calculate next trigger for completed reminder %s: %v\n", reminder.ID, err)
		} else {
			reminder.NextTriggerAt = nextTrigger.Format(time.RFC3339)
			reminder.Status = models.ReminderStatusActive // Reset for the next cycle
		}
	}

	return s.reminderRepo.Update(ctx, reminder)
}

// ProcessDueReminders processes all reminders that are due (called by worker)
func (s *ReminderService) ProcessDueReminders(ctx context.Context) error {
	now := time.Now()

	// Get all due reminders
	reminders, err := s.reminderRepo.GetDueReminders(ctx, now)
	if err != nil {
		return err
	}

	// Track if any system-level errors occurred during processing
	systemErrorOccurred := false

	for _, reminder := range reminders {
		// Process each reminder
		if err := s.processSingleDueReminder(ctx, reminder, now); err != nil {
			// Distinguish device token errors from system-level errors
			if !isTokenInvalidError(err) {
				systemErrorOccurred = true
				// Log detailed system error to terminal
				fmt.Printf("SYSTEM FCM ERROR - Reminder %s: %v\n", reminder.ID, err)
			} else {
				// Log token error details
				fmt.Printf("TOKEN ERROR - Reminder %s: %v\n", reminder.ID, err)
			}
			continue
		}
	}

	if systemErrorOccurred {
		return fmt.Errorf("system_fcm_error: %d reminders failed due to system errors", len(reminders))
	}
	return nil
}

// processSingleDueReminder processes a single reminder
	func (s *ReminderService) processSingleDueReminder(ctx context.Context, reminder *models.Reminder, now time.Time) error {
		// Get user
		user, err := s.userRepo.GetByID(ctx, reminder.UserID)
		if err != nil {
			// Handle user not found error specifically
			if err.Error() == "sql: no rows in result set" || err.Error() == "user not found" {
				fmt.Printf("USER NOT FOUND ERROR - Reminder %s: user %s does not exist\n", reminder.ID, reminder.UserID)
				return fmt.Errorf("user not found: %s", reminder.UserID)
			}
			return err
		}

		// Check if user has active FCM
		if !user.IsFCMActive || user.FCMToken == "" {
			fmt.Printf("USER FCM INACTIVE - Reminder %s: user %s has inactive FCM (is_fcm_active=%t, fcm_token='%s')\n", 
				reminder.ID, reminder.UserID, user.IsFCMActive, user.FCMToken)
			return errors.New("user FCM not active")
		}

	// Send notification (no-op if FCM service is not configured)
	if s.fcmService != nil {
		err = s.fcmService.SendNotification(user.FCMToken, reminder.Title, reminder.Description)
		if err != nil {
			// Handle FCM errors
			if isTokenInvalidError(err) {
				// Disable FCM for this user
				s.userRepo.DisableFCM(ctx, user.ID)
			}
			return err
		}

		// Update last_sent_at only when we actually sent something
		s.reminderRepo.UpdateLastSent(ctx, reminder.ID, now.Format(time.RFC3339))
	}

	// Handle based on type
	if reminder.Type == models.ReminderTypeOneTime {
		return s.handleOneTimeReminder(ctx, reminder)
	} else {
		return s.handleRecurringReminder(ctx, reminder, now)
	}
}

// handleOneTimeReminder handles one-time reminder logic after sending
func (s *ReminderService) handleOneTimeReminder(ctx context.Context, reminder *models.Reminder) error {
	// Mark as completed
	reminder.Status = models.ReminderStatusCompleted
	reminder.LastCompletedAt = time.Now().UTC().Format(time.RFC3339)
	return s.reminderRepo.Update(ctx, reminder)
}

// handleRecurringReminder handles recurring reminder logic after sending
func (s *ReminderService) handleRecurringReminder(ctx context.Context, reminder *models.Reminder, now time.Time) error {
	// For recurring reminders, calculate the next trigger time
	nextTrigger, err := s.schedCalculator.CalculateNextTrigger(reminder, now)
	if err != nil {
		// Log the error but don't block the process
		fmt.Printf("failed to calculate next trigger for reminder %s: %v\n", reminder.ID, err)
		// Decide on a fallback, e.g., retry in 1 hour
		nextTrigger = now.Add(1 * time.Hour)
	}
	reminder.NextTriggerAt = nextTrigger.Format(time.RFC3339)

	// Update the reminder with the new next_trigger_at
	return s.reminderRepo.Update(ctx, reminder)
}

// Helper function to check if FCM error is due to invalid token
func isTokenInvalidError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return errStr == "UNREGISTERED" ||
		errStr == "INVALID_ARGUMENT" ||
		errStr == "NOT_FOUND" ||
		errStr == "user FCM not active"
}
