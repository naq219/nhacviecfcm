package worker

import (
	"context"
	"fmt"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/services/fcmutils"
	"remiaq/internal/utils"

	"github.com/pocketbase/pocketbase"
)

// WorkerLoopNoUT processes recurring reminders without "until complete" strategy
type WorkerLoopNoUT struct {
	repo     *WorkerReminderRepo
	userRepo UserRepo
	interval time.Duration
	logger   *utils.Logger
}

// NewWorkerLoopNoUT creates a new worker
func NewWorkerLoopNoUT(
	app *pocketbase.PocketBase,
	repo *WorkerReminderRepo,
	userRepo UserRepo,
	interval time.Duration,
) *WorkerLoopNoUT {
	return &WorkerLoopNoUT{
		repo:     repo,
		userRepo: userRepo,
		interval: interval,
		logger:   utils.NewLogger(app, "WORKER_LOOP_NO_UT"),
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
		w.logger.Infof("Started (interval=%s)", w.interval.String())

		for {
			select {
			case <-ticker.C:
				w.runOnce(ctx)
			case <-ctx.Done():
				w.logger.Info("Stopped")
				return
			}
		}
	}()
}

func (w *WorkerLoopNoUT) runOnce(ctx context.Context) {
	now := time.Now().UTC()

	// Case 4: Recurring - No CRP
	if err := w.processRecurringNoCRP(ctx, now); err != nil {

		w.logger.Errorf("Error processing Recurring No CRP: %v", err)
	}

	// Case 5.1: Recurring - CRP Trigger
	if err := w.processRecurringCRPTrigger(ctx, now); err != nil {
		w.logger.Errorf("Error processing Recurring CRP Trigger: %v", err)
	}

	// Case 5.2: Recurring - CRP Retry
	if err := w.processRecurringCRPRetry(ctx, now); err != nil {
		w.logger.Errorf("Error processing Recurring CRP Retry: %v", err)
	}
}

// Case 4: Recurring - No CRP - No until_complete
func (w *WorkerLoopNoUT) processRecurringNoCRP(ctx context.Context, now time.Time) error {

	reminders, err := w.repo.GetRecurringNoCRP(ctx, now)
	if err != nil {
		return err
	}

	if len(reminders) == 0 {
		return nil
	}

	logCase := w.logger.WithTag("##4---RecurringNoCRP")

	logCase.Infof("Found %d reminders", len(reminders))

	for _, r := range reminders {
		logCase.Infof("Processing ID=%s, Title=%s", r.ID, r.Title)

		if err := w.sendNotification(ctx, r); err != nil {
			logCase.Errorf("Send failed ID=%s: %v", r.ID, err)
			continue
		}

		// Calculate next recurring time
		nextRecurring := w.calculateNextRecurringTH4NoCrpNoUT(r, now)
		r.NextRecurring = nextRecurring
		r.LastSentAt = now

		if err := w.repo.Update(ctx, r); err != nil {
			logCase.Errorf("Update failed ID=%s: %v", r.ID, err)
		} else {
			logCase.Infof("✓ Sent, NextRecurring=%s ID=%s", r.NextRecurring.Format("15:04:05"), r.ID)
		}
	}
	return nil
}

// Case 5.1: Recurring - CRP Trigger (FRP)
func (w *WorkerLoopNoUT) processRecurringCRPTrigger(ctx context.Context, now time.Time) error {
	logCase := w.logger.WithTag("Case5.1_RecurringCRPTrigger")

	reminders, err := w.repo.GetRecurringCRPTrigger(ctx, now)
	if err != nil {
		return err
	}

	if len(reminders) == 0 {
		return nil
	}

	logCase.Infof("Found %d reminders", len(reminders))

	for _, r := range reminders {
		logCase.Infof("Processing ID=%s, Title=%s, MaxCRP=%d", r.ID, r.Title, r.MaxCRP)

		if err := w.sendNotification(ctx, r); err != nil {
			logCase.Errorf("Send failed ID=%s: %v", r.ID, err)
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
			logCase.Errorf("Update failed ID=%s: %v", r.ID, err)
		} else {
			logCase.Infof("✓ Sent FRP, NextRecurring=%s, NextCRP=%s ID=%s",
				r.NextRecurring.Format("15:04:05"), r.NextCRP.Format("15:04:05"), r.ID)
		}
	}
	return nil
}

// Case 5.2: Recurring - CRP Retry
func (w *WorkerLoopNoUT) processRecurringCRPRetry(ctx context.Context, now time.Time) error {
	logCase := w.logger.WithTag("Case5.2_RecurringCRPRetry")

	reminders, err := w.repo.GetRecurringCRPRetry(ctx, now)
	if err != nil {
		return err
	}

	if len(reminders) == 0 {
		return nil
	}

	logCase.Infof("Found %d reminders", len(reminders))

	for _, r := range reminders {
		logCase.Infof("Processing ID=%s, CRP=%d/%d", r.ID, r.CRPCount+1, r.MaxCRP)

		if err := w.sendNotification(ctx, r); err != nil {
			logCase.Errorf("Send failed ID=%s: %v", r.ID, err)
			continue
		}

		// Increment CRP count and schedule next retry
		r.CRPCount++
		r.NextCRP = now.Add(time.Duration(r.CRPIntervalSec) * time.Second)
		r.LastSentAt = now

		if err := w.repo.Update(ctx, r); err != nil {
			logCase.Errorf("Update failed ID=%s: %v", r.ID, err)
		} else {
			if r.CRPCount >= r.MaxCRP {
				logCase.Infof("✓ Sent CRP (quota reached), waiting for FRP ID=%s", r.ID)
			} else {
				logCase.Infof("✓ Sent CRP, NextCRP=%s ID=%s", r.NextCRP.Format("15:04:05"), r.ID)
			}
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
