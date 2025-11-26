package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/services/fcmutils"
	"remiaq/internal/utils"
)

// WorkerV2 processes reminders using new V2 logic
type WorkerV2 struct {
	sysRepo      SystemStatusRepo
	reminderRepo ReminderRepo
	userRepo     UserRepo
	interval     time.Duration
}

// NewWorkerV2 creates a new worker v2
func NewWorkerV2(
	sysRepo SystemStatusRepo,
	reminderRepo ReminderRepo,
	userRepo UserRepo,
	interval time.Duration,
) *WorkerV2 {
	return &WorkerV2{
		sysRepo:      sysRepo,
		reminderRepo: reminderRepo,
		userRepo:     userRepo,
		interval:     interval,
	}
}

// Start launches the background loop
func (w *WorkerV2) Start(ctx context.Context) {
	if w == nil {
		return
	}

	if w.interval <= 0 {
		w.interval = time.Minute
	}

	ticker := time.NewTicker(w.interval)
	go func() {
		defer ticker.Stop()
		log.Printf("Worker V2 started (interval=%s)", w.interval.String())

		for {
			select {
			case <-ticker.C:
				w.runOnce(ctx)
			case <-ctx.Done():
				log.Println("Worker V2 stopped")
				return
			}
		}
	}()
}

// runOnce processes a single worker cycle
func (w *WorkerV2) runOnce(ctx context.Context) {
	// Check if enabled
	enabled, err := w.sysRepo.IsWorkerEnabled(ctx)
	if err != nil {
		log.Printf("Worker V2: failed to check system status: %v", err)
		return
	}
	if !enabled {
		return
	}

	now := time.Now().UTC()

	// Get all due reminders
	// Note: GetDueReminders filters by next_action_at <= now AND snooze_until <= now
	reminders, err := w.reminderRepo.GetDueReminders(ctx, now)
	if err != nil {
		log.Printf("Worker V2: failed to get due reminders: %v", err)
		return
	}

	if len(reminders) == 0 {
		return
	}

	for _, reminder := range reminders {
		if err := w.processReminder(ctx, reminder, now); err != nil {
			log.Printf("Worker V2: Error processing reminder %s: %v", reminder.ID, err)
		}
	}
}

func (w *WorkerV2) processReminder(ctx context.Context, reminder *models.Reminder, now time.Time) error {
	// Validate basic data
	if reminder.Status != models.ReminderStatusActive {
		return nil
	}

	// Logic for One Time Reminders
	if reminder.Type == models.ReminderTypeOneTime {
		return w.processOneTime(ctx, reminder, now)
	}

	// Logic for Recurring Reminders (Placeholder for now as requested)
	// 4. Lặp lại - không có CRP - không có crp_until_complete
	// 5. Lặp lại - không có CRP - có crp_until_complete
	// 6. Lặp lại - có CRP - không có crp_until_complete
	// 7. Lặp lại - có CRP - có crp_until_complete

	return nil
}

func (w *WorkerV2) processOneTime(ctx context.Context, reminder *models.Reminder, now time.Time) error {
	// Check snooze (already handled by GetDueReminders query, but double check logic if needed)
	// The user requirements say: snooze_until <= now(). GetDueReminders does this.

	//validate
	if !reminder.NextActionAt.IsZero() && (reminder.NextActionAt.Before(now) || reminder.NextActionAt.Equal(now)) {
		return fmt.Errorf("next_action_at must be in the future")
	}
	if !utils.IsTimeValid(reminder.NextCRP) || !isInFuture(reminder.NextCRP) && reminder.IsSendedOneTime == 1 {
		return fmt.Errorf("next_crp phải được set và phải ở trong tương lai khi isSendedOneTime = true")
	}

	// Case 1: Một lần - không CRP
	// - type = one_time
	// - max_crp = 0
	// - status = active
	// - next_action_at <= now()
	// - isSendedOneTime = false
	// - snooze_until <= now()
	if reminder.MaxCRP == 0 && reminder.IsSendedOneTime == 0 {
		log.Printf("Worker V2: Processing One Time - No CRP for %s", reminder.ID)

		// Send notification
		if err := w.sendNotification(ctx, reminder); err != nil {
			return err
		}

		// Update state
		reminder.IsSendedOneTime = 1
		reminder.Status = models.ReminderStatusCompleted
		reminder.LastCompletedAt = now
		reminder.LastSentAt = now

		// Clear next_action_at so it's not picked up again
		reminder.NextActionAt = time.Time{}

		return w.reminderRepo.Update(ctx, reminder)
	}

	// Case 2: Một lần - có CRP - isSendedOneTime = false
	// - type = one_time
	// - max_crp > 0
	// - status = active
	// - next_action_at <= now()
	// - isSendedOneTime = false
	// - snooze_until <= now()
	if reminder.MaxCRP > 0 && reminder.IsSendedOneTime == 0 {
		log.Printf("***#2 Một lần - có CRP - isSendedOneTime = false for %s", reminder.ID)

		// Send notification
		if err := w.sendNotification(ctx, reminder); err != nil {
			return err
		}

		// Update state
		reminder.IsSendedOneTime = 1
		reminder.LastSentAt = now

		// Set next_crp
		if reminder.CRPIntervalSec > 0 {
			reminder.NextCRP = now.Add(time.Duration(reminder.CRPIntervalSec) * time.Second)
			// Also update NextActionAt to NextCRP so it's picked up then
			reminder.NextActionAt = reminder.NextCRP
		} else {
			// Fallback if interval is 0, though validation should prevent this
			reminder.NextActionAt = time.Time{}
		}

		return w.reminderRepo.Update(ctx, reminder)
	}

	// Case 3: Một lần - có CRP - isSendedOneTime = true
	// - type = one_time
	// - max_crp > 0
	// - crp_count < max_crp
	// - status = active
	// - next_crp <= now() (This is implied because NextActionAt is set to NextCRP in previous step)
	// - isSendedOneTime = true
	// - snooze_until <= now()
	if reminder.MaxCRP > 0 && reminder.IsSendedOneTime == 1 {

		if utils.IsTimeValid(reminder.NextCRP) && !isInFuture(reminder.NextCRP) {
			if reminder.CRPCount < reminder.MaxCRP {
				log.Printf("***#3 Một lần - có CRP - isSendedOneTime = true for %s", reminder.ID)
				// Send notification
				if err := w.sendNotification(ctx, reminder); err != nil {
					return err
				}

				// Update state
				reminder.CRPCount++
				reminder.LastSentAt = now
				reminder.NextCRP = now.Add(time.Duration(reminder.CRPIntervalSec) * time.Second)

				// Check if completed
				if reminder.CRPCount >= reminder.MaxCRP {
					reminder.Status = models.ReminderStatusCompleted
					reminder.LastCompletedAt = now
					reminder.NextActionAt = time.Time{} // Stop
				} else {
					reminder.NextActionAt = reminder.NextCRP // Continue
				}

				return w.reminderRepo.Update(ctx, reminder)
			}
		}
	}

	return nil
}

func isInFuture(time1 time.Time) bool {
	// Check if time is in the future
	return !time1.IsZero() && (time1.After(time.Now()) || time1.Equal(time.Now()))
}
func isNotInFuture(time1 time.Time) bool {
	// Check if time is not in the future
	return !isInFuture(time1)
}

func (w *WorkerV2) sendNotification(ctx context.Context, reminder *models.Reminder) error {
	user, err := w.userRepo.GetByID(ctx, reminder.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if !user.IsFCMActive || user.FCMToken == "" {
		return fmt.Errorf("user FCM not active or token empty")
	}

	// Use fcmutils to send
	_, err = fcmutils.SendFCMNotification(ctx, reminder.Title, reminder.Description, user.FCMToken, user.Email)
	if err != nil {
		return fmt.Errorf("fcm send error: %w", err)
	}

	return nil
}
