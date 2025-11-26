package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/services/fcmutils"
)

// WorkerLoopNoUT processes recurring reminders without "until complete" strategy
type WorkerLoopNoUT struct {
	repo     *WorkerReminderRepo
	userRepo UserRepo
	interval time.Duration
}

// NewWorkerLoopNoUT creates a new worker
func NewWorkerLoopNoUT(
	repo *WorkerReminderRepo,
	userRepo UserRepo,
	interval time.Duration,
) *WorkerLoopNoUT {
	return &WorkerLoopNoUT{
		repo:     repo,
		userRepo: userRepo,
		interval: interval,
	}
}

// Start launches the background loop
func (w *WorkerLoopNoUT) Start(ctx context.Context) {
	if w == nil {
		return
	}
	if w.interval <= 0 {
		w.interval = time.Minute
	}

	ticker := time.NewTicker(w.interval)
	go func() {
		defer ticker.Stop()
		log.Printf("WorkerLoopNoUT started (interval=%s)", w.interval.String())

		for {
			select {
			case <-ticker.C:
				w.runOnce(ctx)
			case <-ctx.Done():
				log.Println("WorkerLoopNoUT stopped")
				return
			}
		}
	}()
}

func (w *WorkerLoopNoUT) runOnce(ctx context.Context) {
	now := time.Now().UTC()

	// Case 4: Recurring - No CRP
	if err := w.processRecurringNoCRP(ctx, now); err != nil {
		log.Printf("WorkerLoopNoUT: Error processing Recurring No CRP: %v", err)
	}

	// Case 5.1: Recurring - With CRP - FRP trigger
	if err := w.processRecurringCRPTrigger(ctx, now); err != nil {
		log.Printf("WorkerLoopNoUT: Error processing Recurring CRP Trigger: %v", err)
	}

	// Case 5.2: Recurring - With CRP - Retry
	if err := w.processRecurringCRPRetry(ctx, now); err != nil {
		log.Printf("WorkerLoopNoUT: Error processing Recurring CRP Retry: %v", err)
	}
}

// Case 4: Recurring - No CRP - No until_complete
func (w *WorkerLoopNoUT) processRecurringNoCRP(ctx context.Context, now time.Time) error {
	reminders, err := w.repo.GetRecurringNoCRP(ctx, now)
	if err != nil {
		return err
	}

	for _, r := range reminders {
		log.Printf("WorkerLoopNoUT: Processing Recurring No CRP for %s", r.ID)

		if err := w.sendNotification(ctx, r); err != nil {
			log.Printf("Failed to send notification for %s: %v", r.ID, err)
			continue
		}

		// Calculate next recurring time
		nextRecurring := w.calculateNextRecurringTH4NoCrpNoUT(r, now)
		r.NextRecurring = nextRecurring
		r.LastSentAt = now

		if err := w.repo.Update(ctx, r); err != nil {
			log.Printf("Failed to update reminder %s: %v", r.ID, err)
		}
	}
	return nil
}

// Case 5.1: Recurring - With CRP - FRP trigger
func (w *WorkerLoopNoUT) processRecurringCRPTrigger(ctx context.Context, now time.Time) error {
	reminders, err := w.repo.GetRecurringCRPTrigger(ctx, now)
	if err != nil {
		return err
	}

	for _, r := range reminders {
		log.Printf("WorkerLoopNoUT: Processing Recurring CRP Trigger for %s", r.ID)

		if err := w.sendNotification(ctx, r); err != nil {
			log.Printf("Failed to send notification for %s: %v", r.ID, err)
			continue
		}

		// Calculate next recurring time
		nextRecurring := w.calculateNextRecurringTH4YesCrpNoUT(r, now)
		r.NextRecurring = nextRecurring

		// Reset CRP cycle
		r.CRPCount = 0
		r.NextCRP = now.Add(time.Duration(r.CRPIntervalSec) * time.Second)
		r.LastSentAt = now

		if err := w.repo.Update(ctx, r); err != nil {
			log.Printf("Failed to update reminder %s: %v", r.ID, err)
		}
	}
	return nil
}

// Case 5.2: Recurring - With CRP - Retry
func (w *WorkerLoopNoUT) processRecurringCRPRetry(ctx context.Context, now time.Time) error {
	reminders, err := w.repo.GetRecurringCRPRetry(ctx, now)
	if err != nil {
		return err
	}

	for _, r := range reminders {
		log.Printf("WorkerLoopNoUT: Processing Recurring CRP Retry for %s (%d/%d)", r.ID, r.CRPCount+1, r.MaxCRP)

		if err := w.sendNotification(ctx, r); err != nil {
			log.Printf("Failed to send notification for %s: %v", r.ID, err)
			continue
		}

		// Increment CRP count and schedule next retry
		r.CRPCount++
		r.NextCRP = now.Add(time.Duration(r.CRPIntervalSec) * time.Second)
		r.LastSentAt = now

		if err := w.repo.Update(ctx, r); err != nil {
			log.Printf("Failed to update reminder %s: %v", r.ID, err)
		}
	}
	return nil
}

func (w *WorkerLoopNoUT) sendNotification(ctx context.Context, reminder *models.Reminder) error {
	user, err := w.userRepo.GetByID(ctx, reminder.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if !user.IsFCMActive || user.FCMToken == "" {
		return fmt.Errorf("user FCM not active")
	}
	_, err = fcmutils.SendFCMNotification(ctx, reminder.Title, reminder.Description, user.FCMToken, user.Email)
	return err
}

// Placeholder calculation functions - implement based on your recurrence pattern logic
func (w *WorkerLoopNoUT) calculateNextRecurringTH4NoCrpNoUT(r *models.Reminder, now time.Time) time.Time {
	// TODO: Implement based on RecurrencePattern
	// For now, return a simple daily increment
	return now.Add(24 * time.Hour)
}

func (w *WorkerLoopNoUT) calculateNextRecurringTH4YesCrpNoUT(r *models.Reminder, now time.Time) time.Time {
	// TODO: Implement based on RecurrencePattern
	// For now, return a simple daily increment
	return now.Add(24 * time.Hour)
}
