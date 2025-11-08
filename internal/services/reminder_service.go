package services

import (
	"context"
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
	if err := reminder.Validate(); err != nil {
		return err
	}

	reminder.ID = ""

	// Set defaults
	if reminder.Status == "" {
		reminder.Status = models.ReminderStatusActive
	}
	if reminder.RepeatStrategy == "" {
		reminder.RepeatStrategy = models.RepeatStrategyNone
	}
	if reminder.CalendarType == "" {
		reminder.CalendarType = models.CalendarTypeSolar
	}
	if reminder.MaxCRP == 0 {
		reminder.MaxCRP = 0 // Default: send once
	}
	if reminder.CRPIntervalSec == 0 {
		reminder.CRPIntervalSec = 0 // Will be set per reminder
	}

	now := time.Now().UTC()

	// For one_time: set next_crp = now (send immediately)
	if reminder.Type == models.ReminderTypeOneTime {
		reminder.NextCRP = now
		reminder.CRPCount = 0
	} else {
		// For recurring: calculate next_recurring from pattern
		nextRecurring, err := s.schedCalculator.CalculateNextRecurring(reminder, now)
		if err != nil {
			return fmt.Errorf("failed to calculate next_recurring: %w", err)
		}
		reminder.NextRecurring = nextRecurring
		reminder.NextCRP = nextRecurring // CRP starts at FRP time
		reminder.CRPCount = 0
	}

	// Calculate next_action_at
	reminder.NextActionAt = s.schedCalculator.CalculateNextActionAt(reminder, now)

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
	existingReminder, err := s.reminderRepo.GetByID(ctx, reminder.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing reminder: %w", err)
	}

	// Merge updates
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
	if reminder.RecurrencePattern != nil {
		existingReminder.RecurrencePattern = reminder.RecurrencePattern
	}
	if reminder.RepeatStrategy != "" {
		existingReminder.RepeatStrategy = reminder.RepeatStrategy
	}
	if reminder.CRPIntervalSec != 0 {
		existingReminder.CRPIntervalSec = reminder.CRPIntervalSec
	}
	if reminder.MaxCRP != 0 {
		existingReminder.MaxCRP = reminder.MaxCRP
	}
	if reminder.Status != "" {
		existingReminder.Status = reminder.Status
	}

	if err := existingReminder.Validate(); err != nil {
		return err
	}

	// If recurring and pattern changed, recalc next_recurring
	if existingReminder.Type == models.ReminderTypeRecurring {
		now := time.Now().UTC()
		nextRecurring, err := s.schedCalculator.CalculateNextRecurring(existingReminder, now)
		if err != nil {
			return fmt.Errorf("failed to calculate next_recurring: %w", err)
		}
		existingReminder.NextRecurring = nextRecurring
		existingReminder.NextActionAt = s.schedCalculator.CalculateNextActionAt(existingReminder, now)
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

// SnoozeReminder snoozes a reminder for specified duration
func (s *ReminderService) SnoozeReminder(ctx context.Context, id string, duration time.Duration) error {
	reminder, err := s.reminderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	reminder.SnoozeUntil = now.Add(duration)

	// Recalc next_action_at
	reminder.NextActionAt = s.schedCalculator.CalculateNextActionAt(reminder, now)

	return s.reminderRepo.Update(ctx, reminder)
}

// OnUserComplete handles user completing a reminder
func (s *ReminderService) OnUserComplete(ctx context.Context, id string) error {
	reminder, err := s.reminderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	reminder.LastCRPCompletedAt = now

	if reminder.Type == models.ReminderTypeOneTime {
		// Mark as completed
		reminder.Status = models.ReminderStatusCompleted
		reminder.LastCompletedAt = now
	} else if reminder.Type == models.ReminderTypeRecurring {
		// Reset CRP for current cycle
		reminder.CRPCount = 0

		if reminder.RepeatStrategy == models.RepeatStrategyCRPUntilComplete {
			// Recalc next_recurring from now
			nextRecurring, err := s.schedCalculator.CalculateNextRecurring(reminder, now)
			if err != nil {
				// Log but don't fail
				fmt.Printf("WARN: Failed to calc next_recurring on complete: %v\n", err)
				nextRecurring = now.Add(24 * time.Hour)
			}
			reminder.NextRecurring = nextRecurring
			reminder.NextCRP = nextRecurring
		}

		reminder.LastCompletedAt = now
	}

	// Recalc next_action_at
	reminder.NextActionAt = s.schedCalculator.CalculateNextActionAt(reminder, now)

	return s.reminderRepo.Update(ctx, reminder)
}

// CompleteReminder marks a reminder as completed (legacy, delegates to OnUserComplete)
func (s *ReminderService) CompleteReminder(ctx context.Context, id string) error {
	return s.OnUserComplete(ctx, id)
}

// ProcessDueReminders processes all active reminders that are due
// Called by worker - NOT worker logic itself, just pre-processing
func (s *ReminderService) ProcessDueReminders(ctx context.Context) error {
	now := time.Now().UTC()
	reminders, err := s.reminderRepo.GetDueReminders(ctx, now)
	if err != nil {
		return err
	}

	// This just returns reminders, actual processing is in worker
	// Note: Keep for compatibility if needed by old code
	_ = reminders
	return nil
}
