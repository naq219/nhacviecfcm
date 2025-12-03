package worker

import (
	"context"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/utils"

	"github.com/pocketbase/pocketbase"
)

// WorkerOneTimeV2 processes one-time reminders
type WorkerOneTimeV2 struct {
	repo     *WorkerReminderRepo
	userRepo UserRepo
	sysRepo  SystemStatusRepo
	interval time.Duration
	logger   *utils.Logger
}

// NewWorkerOneTimeV2 creates a new worker
func NewWorkerOneTimeV2(
	app *pocketbase.PocketBase,
	sysRepo SystemStatusRepo,
	repo *WorkerReminderRepo,
	userRepo UserRepo,
	interval time.Duration,
) *WorkerOneTimeV2 {
	return &WorkerOneTimeV2{
		sysRepo:  sysRepo,
		repo:     repo,
		userRepo: userRepo,
		interval: interval,
		logger:   utils.NewLogger(app, "WORKER_ONE_TIME_V2"),
	}
}

// Start launches the background loop
func (w *WorkerOneTimeV2) Start(ctx context.Context) {
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

func (w *WorkerOneTimeV2) runOnce(ctx context.Context) {
	if !IsWorkerSystemEnabled(ctx, w.sysRepo, w.logger.Errorf) {
		return
	}
	now := time.Now().UTC()

	// 1. One Time - No CRP
	if err := w.processNoCRP(ctx, now); err != nil {
		w.logger.Errorf("Error processing NoCRP: %v", err)
	}

	// 2. One Time - First Send (CRP)
	if err := w.processFirstSend(ctx, now); err != nil {
		w.logger.Errorf("Error processing FirstSend: %v", err)
	}

	// 3. One Time - Retry (CRP)
	if err := w.processRetry(ctx, now); err != nil {
		w.logger.Errorf("Error processing Retry: %v", err)
	}
}

func (w *WorkerOneTimeV2) processNoCRP(ctx context.Context, now time.Time) error {
	logCase := w.logger.WithTag("Case1_NoCRP")

	reminders, err := w.repo.GetOneTimeNoCRP(ctx, now)
	if err != nil {
		return err
	}

	if len(reminders) == 0 {
		return nil
	}

	logCase.Infof("Found %d reminders", len(reminders))

	for _, r := range reminders {
		logCase.Infof("Processing ID=%s, Title=%s", r.ID, r.Title)

		if err := SendNotification(ctx, r, w.userRepo, w.sysRepo, w.logger.Errorf); err != nil {
			logCase.Errorf("Send failed ID=%s: %v", r.ID, err)
			continue
		}

		r.IsSendedOneTime = true
		r.Status = models.ReminderStatusCompleted
		r.LastCompletedAt = now
		r.LastSentAt = now
		r.NextActionAt = time.Time{}

		if err := w.repo.Update(ctx, r); err != nil {
			logCase.Errorf("Update failed ID=%s: %v", r.ID, err)
		} else {
			logCase.Infof("✓ Completed ID=%s", r.ID)
		}
	}
	return nil
}

func (w *WorkerOneTimeV2) processFirstSend(ctx context.Context, now time.Time) error {
	logCase := w.logger.WithTag("Case2_FirstSend")

	reminders, err := w.repo.GetOneTimeFirstSend(ctx, now)
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

		r.IsSendedOneTime = true
		r.LastSentAt = now

		if r.CRPIntervalSec > 0 {
			r.NextCRP = now.Add(time.Duration(r.CRPIntervalSec) * time.Second)
		} else {
			r.NextCRP = now.Add(time.Minute)
		}

		if err := w.repo.Update(ctx, r); err != nil {
			logCase.Errorf("Update failed ID=%s: %v", r.ID, err)
		} else {
			logCase.Infof("✓ Sent, NextCRP=%s ID=%s", r.NextCRP.Format("15:04:05"), r.ID)
		}
	}
	return nil
}

func (w *WorkerOneTimeV2) processRetry(ctx context.Context, now time.Time) error {
	logCase := w.logger.WithTag("Case3_Retry")

	reminders, err := w.repo.GetOneTimeRetry(ctx, now)
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

		r.CRPCount++
		r.LastSentAt = now
		r.NextCRP = now.Add(time.Duration(r.CRPIntervalSec) * time.Second)

		if r.CRPCount >= r.MaxCRP {
			r.Status = models.ReminderStatusCompleted
			r.LastCompletedAt = now
			r.NextActionAt = time.Time{}
			logCase.Infof("✓ Completed (quota reached) ID=%s", r.ID)
		} else {
			logCase.Infof("✓ Sent retry, NextCRP=%s ID=%s", r.NextCRP.Format("15:04:05"), r.ID)
		}

		if err := w.repo.Update(ctx, r); err != nil {
			logCase.Errorf("Update failed ID=%s: %v", r.ID, err)
		}
	}
	return nil
}
