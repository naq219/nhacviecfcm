package services

import (
	"context"
	"errors"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/repository"

	"github.com/google/uuid"
)

// ReminderService handles reminder business logic
type ReminderService struct {
	reminderRepo    repository.ReminderRepository
	userRepo        repository.UserRepository
	fcmService      *FCMService
	schedCalculator *ScheduleCalculator
}

// NewReminderService creates a new reminder service
func NewReminderService(
	reminderRepo repository.ReminderRepository,
	userRepo repository.UserRepository,
	fcmService *FCMService,
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

	// Generate ID if not provided
	if reminder.ID == "" {
		reminder.ID = uuid.New().String()
	}

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
	if reminder.NextTriggerAt.IsZero() {
		nextTrigger, err := s.schedCalculator.CalculateNextTrigger(reminder, time.Now())
		if err != nil {
			return err
		}
		reminder.NextTriggerAt = nextTrigger
	}

	return s.reminderRepo.Create(ctx, reminder)
}

// GetReminder retrieves a reminder by ID
func (s *ReminderService) GetReminder(ctx context.Context, id string) (*models.Reminder, error) {
	return s.reminderRepo.GetByID(ctx, id)
}

// UpdateReminder updates a reminder
func (s *ReminderService) UpdateReminder(ctx context.Context, reminder *models.Reminder) error {
	if err := reminder.Validate(); err != nil {
		return err
	}

	return s.reminderRepo.Update(ctx, reminder)
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

	// For one-time reminders, mark as completed
	if reminder.Type == models.ReminderTypeOneTime {
		return s.reminderRepo.MarkCompleted(ctx, id, now)
	}

	// For recurring reminders with base_on=completion
	if reminder.RecurrencePattern != nil &&
		reminder.RecurrencePattern.BaseOn == models.BaseOnCompletion {
		// Calculate next trigger from completion time
		nextTrigger, err := s.schedCalculator.CalculateNextTrigger(reminder, now)
		if err != nil {
			return err
		}

		// Update last_completed_at and next_trigger_at
		reminder.LastCompletedAt = now.Format(time.RFC3339)
		reminder.NextTriggerAt = nextTrigger
		return s.reminderRepo.Update(ctx, reminder)
	}

	// For other recurring reminders, just update last_completed_at
	reminder.LastCompletedAt = now.Format(time.RFC3339)
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
		if err := s.processReminder(ctx, reminder, now); err != nil {
			// Distinguish device token errors from system-level errors
			if !isTokenInvalidError(err) {
				systemErrorOccurred = true
			}
			// Continue with other reminders regardless
			continue
		}
	}

	if systemErrorOccurred {
		return errors.New("system_fcm_error")
	}
	return nil
}

// processReminder processes a single reminder
func (s *ReminderService) processReminder(ctx context.Context, reminder *models.Reminder, now time.Time) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, reminder.UserID)
	if err != nil {
		return err
	}

	// Check if user has active FCM
	if !user.IsFCMActive || user.FCMToken == "" {
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
		s.reminderRepo.UpdateLastSent(ctx, reminder.ID, now)
	}

	// Handle based on type
	if reminder.Type == models.ReminderTypeOneTime {
		return s.handleOneTimeReminder(ctx, reminder, now)
	} else {
		return s.handleRecurringReminder(ctx, reminder, now)
	}
}

// handleOneTimeReminder handles one-time reminder logic
func (s *ReminderService) handleOneTimeReminder(ctx context.Context, reminder *models.Reminder, now time.Time) error {
	// Check if should retry
	if reminder.RepeatStrategy == models.RepeatStrategyRetryUntilComplete && reminder.IsRetryable() {
		// Increment retry count
		s.reminderRepo.IncrementRetryCount(ctx, reminder.ID)

		// Calculate next retry time
		nextRetry := now.Add(time.Duration(reminder.RetryIntervalSec) * time.Second)
		return s.reminderRepo.UpdateNextTrigger(ctx, reminder.ID, nextRetry)
	}

	// Otherwise, mark as completed
	return s.reminderRepo.MarkCompleted(ctx, reminder.ID, now)
}

// handleRecurringReminder handles recurring reminder logic
func (s *ReminderService) handleRecurringReminder(ctx context.Context, reminder *models.Reminder, now time.Time) error {
	// Calculate next trigger
	nextTrigger, err := s.schedCalculator.CalculateNextTrigger(reminder, now)
	if err != nil {
		return err
	}

	// Update next trigger time
	return s.reminderRepo.UpdateNextTrigger(ctx, reminder.ID, nextTrigger)
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
