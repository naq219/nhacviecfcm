package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/services/fcmutils"
)

// WorkerOneTimeV2 processes one-time reminders
type WorkerOneTimeV2 struct {
	repo     *WorkerReminderRepo
	userRepo UserRepo // Interface defined in worker.go
	interval time.Duration
}

// NewWorkerOneTimeV2 creates a new worker
func NewWorkerOneTimeV2(
	repo *WorkerReminderRepo,
	userRepo UserRepo,
	interval time.Duration,
) *WorkerOneTimeV2 {
	return &WorkerOneTimeV2{
		repo:     repo,
		userRepo: userRepo,
		interval: interval,
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
		log.Printf("WorkerOneTimeV2 started (interval=%s)", w.interval.String())

		for {
			select {
			case <-ticker.C:
				w.runOnce(ctx)
			case <-ctx.Done():
				log.Println("WorkerOneTimeV2 stopped")
				return
			}
		}
	}()
}

func (w *WorkerOneTimeV2) runOnce(ctx context.Context) {
	now := time.Now().UTC()

	// 1. One Time - No CRP
	if err := w.processNoCRP(ctx, now); err != nil {
		log.Printf("WorkerOneTimeV2: Error processing NoCRP: %v", err)
	}

	// 2. One Time - First Send (CRP)
	if err := w.processFirstSend(ctx, now); err != nil {
		log.Printf("WorkerOneTimeV2: Error processing FirstSend: %v", err)
	}

	// 3. One Time - Retry (CRP)
	if err := w.processRetry(ctx, now); err != nil {
		log.Printf("WorkerOneTimeV2: Error processing Retry: %v", err)
	}
}

func (w *WorkerOneTimeV2) processNoCRP(ctx context.Context, now time.Time) error {
	reminders, err := w.repo.GetOneTimeNoCRP(ctx, now)
	if err != nil {
		return err
	}
	for _, r := range reminders {
		log.Printf("WorkerOneTimeV2: Processing NoCRP for %s", r.ID)
		if err := w.sendNotification(ctx, r); err != nil {
			log.Printf("Failed to send notification for %s: %v", r.ID, err)
			continue
		}

		r.IsSendedOneTime = true
		r.Status = models.ReminderStatusCompleted
		r.LastCompletedAt = now
		r.LastSentAt = now
		r.NextActionAt = time.Time{} // Clear

		if err := w.repo.Update(ctx, r); err != nil {
			log.Printf("Failed to update reminder %s: %v", r.ID, err)
		}
	}
	return nil
}

func (w *WorkerOneTimeV2) processFirstSend(ctx context.Context, now time.Time) error {
	reminders, err := w.repo.GetOneTimeFirstSend(ctx, now)
	if err != nil {
		return err
	}
	for _, r := range reminders {
		log.Printf("WorkerOneTimeV2: Processing FirstSend for %s", r.ID)
		if err := w.sendNotification(ctx, r); err != nil {
			log.Printf("Failed to send notification for %s: %v", r.ID, err)
			continue
		}

		r.IsSendedOneTime = true
		r.LastSentAt = now
		// set next_crp = now() + crp_interval_sec
		if r.CRPIntervalSec > 0 {
			r.NextCRP = now.Add(time.Duration(r.CRPIntervalSec) * time.Second)
		} else {
			r.NextCRP = now.Add(time.Minute)
		}

		if err := w.repo.Update(ctx, r); err != nil {
			log.Printf("Failed to update reminder %s: %v", r.ID, err)
		}
	}
	return nil
}

func (w *WorkerOneTimeV2) processRetry(ctx context.Context, now time.Time) error {
	reminders, err := w.repo.GetOneTimeRetry(ctx, now)
	if err != nil {
		return err
	}
	for _, r := range reminders {
		log.Printf("WorkerOneTimeV2: Processing Retry for %s (%d/%d)", r.ID, r.CRPCount+1, r.MaxCRP)
		if err := w.sendNotification(ctx, r); err != nil {
			log.Printf("Failed to send notification for %s: %v", r.ID, err)
			continue
		}

		r.CRPCount++
		r.LastSentAt = now
		r.NextCRP = now.Add(time.Duration(r.CRPIntervalSec) * time.Second)

		if r.CRPCount >= r.MaxCRP {
			r.Status = models.ReminderStatusCompleted
			r.LastCompletedAt = now
			r.NextActionAt = time.Time{}
		}

		if err := w.repo.Update(ctx, r); err != nil {
			log.Printf("Failed to update reminder %s: %v", r.ID, err)
		}
	}
	return nil
}

func (w *WorkerOneTimeV2) sendNotification(ctx context.Context, reminder *models.Reminder) error {
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
