package worker

import (
	"context"
	"fmt"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/services"
	"remiaq/internal/services/fcmutils"
	"remiaq/internal/utils"

	"github.com/pocketbase/pocketbase"
)

// WorkerLoopUT processes recurring reminders with "until complete" strategy
type WorkerLoopUT struct {
	repo     *WorkerReminderRepo
	userRepo UserRepo
	sysRepo  SystemStatusRepo
	interval time.Duration
	logger   *utils.Logger
}

// NewWorkerLoopUT creates a new worker
func NewWorkerLoopUT(
	app *pocketbase.PocketBase,
	sysRepo SystemStatusRepo,
	repo *WorkerReminderRepo,
	userRepo UserRepo,
	interval time.Duration,
) *WorkerLoopUT {
	return &WorkerLoopUT{
		sysRepo:  sysRepo,
		repo:     repo,
		userRepo: userRepo,
		interval: interval,
		logger:   utils.NewLogger(app, "WORKER_LOOP_UT"),
	}
}

// Start launches the background loop
func (w *WorkerLoopUT) Start(ctx context.Context) {
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

func (w *WorkerLoopUT) runOnce(ctx context.Context) {
	if !IsWorkerSystemEnabled(ctx, w.sysRepo, w.logger.Errorf) {
		return
	}
	now := time.Now().UTC()

	// Case 6.1: Recurring - No CRP - Until Complete - First Time
	if err := w.processRecurringUntilCompleteFirstTime(ctx, now); err != nil {
		w.logger.Errorf("Error processing Recurring UT First Time: %v", err)
	}

	// Case 6.2: Recurring - No CRP - Until Complete - Finished
	if err := w.processRecurringUntilCompleteFinished(ctx, now); err != nil {
		w.logger.Errorf("Error processing Recurring UT Finished: %v", err)
	}

	// Case 7.1.1: Recurring - CRP - Until Complete - FRP First Time
	if err := w.processRecurringUntilCompleteCRPFirstTime(ctx, now); err != nil {
		w.logger.Errorf("Error processing Recurring UT CRP First Time: %v", err)
	}

	// Case 7.1.2: Recurring - CRP - Until Complete - FRP After Complete
	if err := w.processRecurringUntilCompleteCRPAfterComplete(ctx, now); err != nil {
		w.logger.Errorf("Error processing Recurring UT CRP After Complete: %v", err)
	}

	// Case 7.2: Recurring - CRP - Until Complete - CRP Retry
	if err := w.processRecurringUntilCompleteCRPRetry(ctx, now); err != nil {
		w.logger.Errorf("Error processing Recurring UT CRP Retry: %v", err)
	}
}

// Case 6.1: Recurring - No CRP - Until Complete - First Time
func (w *WorkerLoopUT) processRecurringUntilCompleteFirstTime(ctx context.Context, now time.Time) error {
	reminders, err := w.repo.GetRecurringUntilCompleteFirstTime(ctx, now)
	if err != nil {
		return err
	}

	if len(reminders) == 0 {
		return nil
	}

	logCase := w.logger.WithTag("Case6.1_RecurringUTFirstTime")
	logCase.Infof("Found %d reminders", len(reminders))

	for _, r := range reminders {
		logCase.Infof("Processing ID=%s, Title=%s", r.ID, r.Title)

		if err := w.sendNotification(ctx, r); err != nil {
			logCase.Errorf("Send failed ID=%s: %v", r.ID, err)
			continue
		}

		// Update state: is_sended_one_time = true
		r.IsSendedOneTime = true
		r.LastSentAt = now
		// next_recurring does NOT change here

		if err := w.repo.Update(ctx, r); err != nil {
			logCase.Errorf("Update failed ID=%s: %v", r.ID, err)
		} else {
			logCase.Infof("✓ Sent First Time ID=%s", r.ID)
		}
	}
	return nil
}

// Case 6.2: Recurring - No CRP - Until Complete - Finished
func (w *WorkerLoopUT) processRecurringUntilCompleteFinished(ctx context.Context, now time.Time) error {
	reminders, err := w.repo.GetRecurringUntilCompleteFinished(ctx, now)
	if err != nil {
		return err
	}

	if len(reminders) == 0 {
		return nil
	}

	logCase := w.logger.WithTag("Case6.2_RecurringUTFinished")
	logCase.Infof("Found %d reminders", len(reminders))

	for _, r := range reminders {
		logCase.Infof("Processing ID=%s, Title=%s", r.ID, r.Title)

		if err := w.sendNotification(ctx, r); err != nil {
			logCase.Errorf("Send failed ID=%s: %v", r.ID, err)
			continue
		}

		// Update state: just last_sent_at
		r.LastSentAt = now
		// next_recurring is updated by API Complete, not here

		if err := w.repo.Update(ctx, r); err != nil {
			logCase.Errorf("Update failed ID=%s: %v", r.ID, err)
		} else {
			logCase.Infof("✓ Sent Next Cycle ID=%s", r.ID)
		}
	}
	return nil
}

func (w *WorkerLoopUT) sendNotification(ctx context.Context, reminder *models.Reminder) error {
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

// calculateNextRecurringTH4NoCrpNoUT calculates next recurring time (reused logic if needed later)
func (w *WorkerLoopUT) calculateNextRecurring(r models.Reminder, now time.Time) (time.Time, error) {
	return services.Tinhtoan_NextRecurringV2(r, now)
}

// Case 7.1.1: Recurring - CRP - Until Complete - FRP Trigger - First Time
func (w *WorkerLoopUT) processRecurringUntilCompleteCRPFirstTime(ctx context.Context, now time.Time) error {
	reminders, err := w.repo.GetRecurringUntilCompleteCRPFirstTime(ctx, now)
	if err != nil {
		return err
	}

	if len(reminders) == 0 {
		return nil
	}

	logCase := w.logger.WithTag("Case7.1.1_RecurringUTCRPFirstTime")
	logCase.Infof("Found %d reminders", len(reminders))

	for _, r := range reminders {
		logCase.Infof("Processing ID=%s, Title=%s, MaxCRP=%d", r.ID, r.Title, r.MaxCRP)

		if err := w.sendNotification(ctx, r); err != nil {
			logCase.Errorf("Send failed ID=%s: %v", r.ID, err)
			continue
		}

		// Update state: is_sended_one_time = true, start CRP cycle
		r.IsSendedOneTime = true
		r.LastSentAt = now
		r.CRPCount = 0
		r.NextCRP = now.Add(time.Duration(r.CRPIntervalSec) * time.Second)
		// next_recurring does NOT change here

		if err := w.repo.Update(ctx, r); err != nil {
			logCase.Errorf("Update failed ID=%s: %v", r.ID, err)
		} else {
			logCase.Infof("✓ Sent FRP (First Time), NextCRP=%s ID=%s", r.NextCRP.Format("15:04:05"), r.ID)
		}
	}
	return nil
}

// Case 7.1.2: Recurring - CRP - Until Complete - FRP Trigger - After Complete
func (w *WorkerLoopUT) processRecurringUntilCompleteCRPAfterComplete(ctx context.Context, now time.Time) error {
	reminders, err := w.repo.GetRecurringUntilCompleteCRPAfterComplete(ctx, now)
	if err != nil {
		return err
	}

	if len(reminders) == 0 {
		return nil
	}

	logCase := w.logger.WithTag("Case7.1.2_RecurringUTCRPAfterComplete")
	logCase.Infof("Found %d reminders", len(reminders))

	for _, r := range reminders {
		logCase.Infof("Processing ID=%s, Title=%s, MaxCRP=%d", r.ID, r.Title, r.MaxCRP)

		if err := w.sendNotification(ctx, r); err != nil {
			logCase.Errorf("Send failed ID=%s: %v", r.ID, err)
			continue
		}

		// Update state: reset CRP cycle for new FRP
		r.LastSentAt = now
		r.CRPCount = 0
		r.NextCRP = now.Add(time.Duration(r.CRPIntervalSec) * time.Second)
		// next_recurring is updated by API Complete, not here

		if err := w.repo.Update(ctx, r); err != nil {
			logCase.Errorf("Update failed ID=%s: %v", r.ID, err)
		} else {
			logCase.Infof("✓ Sent FRP (After Complete), NextCRP=%s ID=%s", r.NextCRP.Format("15:04:05"), r.ID)
		}
	}
	return nil
}

// Case 7.2: Recurring - CRP - Until Complete - CRP Retry
func (w *WorkerLoopUT) processRecurringUntilCompleteCRPRetry(ctx context.Context, now time.Time) error {
	reminders, err := w.repo.GetRecurringUntilCompleteCRPRetry(ctx, now)
	if err != nil {
		return err
	}

	if len(reminders) == 0 {
		return nil
	}

	logCase := w.logger.WithTag("Case7.2_RecurringUTCRPRetry")
	logCase.Infof("Found %d reminders", len(reminders))

	for _, r := range reminders {
		logCase.Infof("Processing ID=%s, CRP=%d/%d", r.ID, r.CRPCount+1, r.MaxCRP)

		if err := w.sendNotification(ctx, r); err != nil {
			logCase.Errorf("Send failed ID=%s: %v", r.ID, err)
			continue
		}

		// Increment CRP count and schedule next retry
		r.CRPCount++
		r.LastSentAt = now
		r.NextCRP = now.Add(time.Duration(r.CRPIntervalSec) * time.Second)

		if err := w.repo.Update(ctx, r); err != nil {
			logCase.Errorf("Update failed ID=%s: %v", r.ID, err)
		} else {
			if r.CRPCount >= r.MaxCRP {
				logCase.Infof("✓ Sent CRP (quota reached), waiting for user complete or FRP ID=%s", r.ID)
			} else {
				logCase.Infof("✓ Sent CRP, NextCRP=%s ID=%s", r.NextCRP.Format("15:04:05"), r.ID)
			}
		}
	}
	return nil
}
