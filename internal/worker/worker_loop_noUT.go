package worker

import (
	"context"
	"fmt"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/utils"

	"remiaq/internal/services"

	"github.com/pocketbase/pocketbase"
)

// WorkerLoopNoUT processes recurring reminders without "until complete" strategy
type WorkerLoopNoUT struct {
	repo     *WorkerReminderRepo
	userRepo UserRepo
	sysRepo  SystemStatusRepo
	interval time.Duration
	logger   *utils.Logger
}

// NewWorkerLoopNoUT creates a new worker
func NewWorkerLoopNoUT(
	app *pocketbase.PocketBase,
	sysRepo SystemStatusRepo,
	repo *WorkerReminderRepo,
	userRepo UserRepo,
	interval time.Duration,
) *WorkerLoopNoUT {
	return &WorkerLoopNoUT{
		sysRepo:  sysRepo,
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
	if !IsWorkerSystemEnabled(ctx, w.sysRepo, w.logger.Errorf) {
		return
	}
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

		if err := SendNotification(ctx, r, w.userRepo, w.sysRepo, w.logger.Errorf); err != nil {
			logCase.Errorf("Send failed ID=%s: %v", r.ID, err)
			continue
		}

		// Calculate next recurring time
		nextRecurring, err := w.calculateNextRecurringTH4NoCrpNoUT(*r, now)
		if err != nil {
			logCase.Errorf("64577 Calculate next recurring failed ID=%s: %v", r.ID, err)
			continue
		}
		r.NextRecurring = nextRecurring
		r.LastSentAt = now
		r.NextActionAt = nextRecurring

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

		if err := SendNotification(ctx, r, w.userRepo, w.sysRepo, w.logger.Errorf); err != nil {
			logCase.Errorf("Send failed ID=%s: %v", r.ID, err)
			continue
		}

		// Calculate next recurring time
		nextRecurring, err := w.calculateNextRecurringTH4YesCrpNoUT(*r, now)
		if err != nil {
			logCase.Errorf("75577 Calculate next recurring failed ID=%s: %v", r.ID, err)
			continue
		}
		r.NextRecurring = nextRecurring

		// Reset CRP cycle
		r.CRPCount = 0
		r.NextCRP = now.Add(time.Duration(r.CRPIntervalSec) * time.Second)
		r.LastSentAt = now
		r.NextActionAt = r.NextCRP

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

		if err := SendNotification(ctx, r, w.userRepo, w.sysRepo, w.logger.Errorf); err != nil {
			logCase.Errorf("Send failed ID=%s: %v", r.ID, err)
			continue
		}

		// Increment CRP count and schedule next retry
		r.CRPCount++
		r.NextCRP = now.Add(time.Duration(r.CRPIntervalSec) * time.Second)
		r.LastSentAt = now
		r.NextActionAt = r.NextCRP

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

// calculateNextRecurringTH4NoCrpNoUT calculates next recurring time for Case 4 (No CRP, No UT)
// Based on RecurrencePattern - handles daily, weekly, monthly (solar/lunar), interval_seconds

func (w *WorkerLoopNoUT) calculateNextRecurringTH4NoCrpNoUT(r models.Reminder, now time.Time) (time.Time, error) {
	return services.Tinhtoan_NextRecurringV2(r, now)
}

// func (w *WorkerLoopNoUT) calculateNextRecurringTH4NoCrpNoUT(r *models.Reminder, now time.Time) time.Time {
// 	if r.RecurrencePattern == nil {
// 		w.logger.Errorf("RecurrencePattern is nil for reminder %s", r.ID)
// 		return now.Add(24 * time.Hour) // Fallback
// 	}

// 	pattern := r.RecurrencePattern
// 	current := r.NextRecurring

// 	// If current is zero, initialize to now
// 	if current.IsZero() {
// 		current = now
// 	}

// 	switch pattern.Type {
// 	case models.RecurrenceTypeDaily:
// 		return w.calculateNextDaily(current, pattern, now)
// 	case models.RecurrenceTypeWeekly:
// 		return w.calculateNextWeekly(current, pattern, now)
// 	case models.RecurrenceTypeMonthly:
// 		if r.CalendarType == models.CalendarTypeLunar {
// 			// For lunar monthly, use a simple fallback (proper lunar calc requires lunar calendar service)
// 			return now.Add(30 * 24 * time.Hour)
// 		}
// 		return w.calculateNextSolarMonthly(current, pattern, now)
// 	case models.RecurrenceTypeIntervalSeconds:
// 		return w.calculateNextIntervalSeconds(current, pattern, now)
// 	case models.RecurrenceTypeLunarLastDayOfMonth:
// 		// Fallback for lunar last day
// 		return now.Add(30 * 24 * time.Hour)
// 	default:
// 		w.logger.Errorf("Unsupported recurrence type: %s", pattern.Type)
// 		return now.Add(24 * time.Hour) // Fallback
// 	}
// }

func (w *WorkerLoopNoUT) calculateNextRecurringTH4YesCrpNoUT(r models.Reminder, now time.Time) (time.Time, error) {
	// Same logic as TH4NoCrpNoUT - calculate next FRP based on pattern
	return w.calculateNextRecurringTH4NoCrpNoUT(r, now)
}

// ========================================
// Helper calculation functions
// ========================================

// calculateNextDaily adds interval days and finds first occurrence > now

// calculateNextWeekly finds next target weekday

// calculateNextSolarMonthly adds interval months on day_of_month

// parseTimeOfDay parses HH:MM format
func parseTimeOfDay(timeStr string) (time.Time, error) {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time format, expected HH:MM")
	}
	return t, nil
}
