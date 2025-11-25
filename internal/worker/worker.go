package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/utils"
)

// FCMSender sends FCM notifications
type FCMSender interface {
	SendNotification(token, title, body string) error
}

// UserRepo provides user operations
type UserRepo interface {
	GetByID(ctx context.Context, userID string) (*models.User, error)
	DisableFCM(ctx context.Context, userID string) error
	SetFCMError(ctx context.Context, userID string, errMsg string) error
}

// ReminderRepo provides reminder operations
type ReminderRepo interface {
	GetDueReminders(ctx context.Context, beforeTime time.Time) ([]*models.Reminder, error)
	Update(ctx context.Context, reminder *models.Reminder) error
	UpdateCRPCount(ctx context.Context, id string, crpCount int) error
	UpdateNextRecurring(ctx context.Context, id string, nextRecurring time.Time) error
	UpdateNextCRP(ctx context.Context, id string, nextCRP time.Time) error
	UpdateNextActionAt(ctx context.Context, id string, nextActionAt time.Time) error
	UpdateLastSent(ctx context.Context, id string, lastSentAt string) error
	UpdateStatus(ctx context.Context, id string, status string) error
}

// ScheduleCalc provides schedule calculation
type ScheduleCalc interface {
	CalculateNextActionAt(reminder *models.Reminder, now time.Time) time.Time
	CalculateNextRecurring(reminder *models.Reminder, now time.Time) (time.Time, error)
	CanSendCRP(reminder *models.Reminder, now time.Time) bool
}

// SystemStatusRepo provides system status operations
type SystemStatusRepo interface {
	IsWorkerEnabled(ctx context.Context) (bool, error)
	UpdateError(ctx context.Context, errMsg string) error
	ClearError(ctx context.Context) error
	DisableWorker(ctx context.Context) error
}

// Worker processes reminders using FRP+CRP logic
type Worker struct {
	sysRepo      SystemStatusRepo
	reminderRepo ReminderRepo
	userRepo     UserRepo
	fcmSender    FCMSender
	schedCalc    ScheduleCalc
	interval     time.Duration
}

// NewWorker creates a new worker
func NewWorker(
	sysRepo SystemStatusRepo,
	reminderRepo ReminderRepo,
	userRepo UserRepo,
	fcmSender FCMSender,
	schedCalc ScheduleCalc,
	interval time.Duration,
) *Worker {
	return &Worker{
		sysRepo:      sysRepo,
		reminderRepo: reminderRepo,
		userRepo:     userRepo,
		fcmSender:    fcmSender,
		schedCalc:    schedCalc,
		interval:     interval,
	}
}

// Start launches the background loop
func (w *Worker) Start(ctx context.Context) {

	if w == nil {
		return
	}

	if w.interval <= 0 {
		w.interval = time.Minute
	}

	ticker := time.NewTicker(w.interval)
	go func() {
		defer ticker.Stop()
		log.Printf("Worker started (interval=%s)", w.interval.String())

		for {
			select {
			case <-ticker.C:
				w.runOnce(ctx)
			case <-ctx.Done():
				log.Println("Worker stopped")
				return
			}
		}
	}()
}

func (w *Worker) RunOnce(ctx context.Context) {
	w.runOnce(ctx)
}
func (w *Worker) DoNoThing(ctx context.Context) {

}

// runOnce processes a single worker cycle
func (w *Worker) runOnce(ctx context.Context) {
	// Check if enabled
	enabled, err := w.sysRepo.IsWorkerEnabled(ctx)
	if err != nil {
		log.Printf("Worker: failed to check system status: %v", err)
		return
	}
	if !enabled {
		return
	}

	now := time.Now().UTC()

	// Get all due reminders
	reminders, err := w.reminderRepo.GetDueReminders(ctx, now)
	if err != nil {
		log.Printf("Worker: failed to get due reminders: %v", err)
		_ = w.sysRepo.UpdateError(ctx, fmt.Sprintf("Failed to get due reminders: %v", err))
		return
	}

	if len(reminders) == 0 {
		return
	}

	log.Printf("Worker: Processing %d due reminders", len(reminders))

	// Track errors
	systemErrorOccurred := false

	for _, reminder := range reminders {
		if err := w.processReminder(ctx, reminder, now); err != nil {
			if isSystemFCMError(err) {
				systemErrorOccurred = true
				log.Printf("Worker: SYSTEM FCM ERROR for reminder %s: %v", reminder.ID, err)
				// Disable worker on system error
				_ = w.sysRepo.DisableWorker(ctx)
				return
			}
			log.Printf("Worker: Error processing reminder %s: %v", reminder.ID, err)
		}
	}

	if !systemErrorOccurred {
		_ = w.sysRepo.ClearError(ctx)
	}
}

// processReminder processes a single reminder (FRP + CRP logic)
// processReminder processes a single reminder (FRP + CRP logic)
// processReminder processes a single reminder (FRP + CRP logic)
func (w *Worker) processReminder(ctx context.Context, reminder *models.Reminder, now time.Time) error {
	// ========================================
	// DEBUG: Log reminder state on load
	// ========================================

	if valid, reason := reminder.ValidateData(); !valid {
		log.Printf("❌❌❌❌❌ Validation failed: %s", reason)
		// Log error to notifi_error.txt
		utils.LogErrorToFile(reminder.ID, fmt.Sprintf("Validation failed: %s", reason))
		return nil
	}
	// tạm dừng, dù là 1 lần hay lặp đều nghỉ
	if reminder.Status == models.ReminderStatusPaused {
		return nil
	}

	log.Printf("📋 Loaded reminder %s: NextCRP=%v, LastSentAt=%v, CRPCount=%d, MaxCRP=%d",
		reminder.ID,
		reminder.NextCRP,
		reminder.LastSentAt,
		reminder.CRPCount,
		reminder.MaxCRP)

	// ========================================
	// STEP 0: Check if snoozed
	// ========================================
	if reminder.IsSnoozeUntilActive(now) {
		log.Printf("😴 Worker: Reminder %s is snoozed until %s", reminder.ID, reminder.SnoozeUntil.Format("15:04:05"))
		// Recalc next_action_at and skip
		nextAction := w.schedCalc.CalculateNextActionAt(reminder, now)
		if !nextAction.Equal(reminder.NextActionAt) {
			_ = w.reminderRepo.UpdateNextActionAt(ctx, reminder.ID, nextAction)
		}
		return nil
	}

	// ========================================
	// STEP 1: Check FRP (has priority)
	// ========================================

	if reminder.Type == models.ReminderTypeOneTime {
		if reminder.CanSendFRPOneTime() {
			// Chưa gửi → gửi ngay
			return w.processCRPForOneTime(ctx, reminder, now)
		}
		if w.schedCalc.CanSendCRP(reminder, now) {
			// Gửi retry
			return w.processCRPForOneTime(ctx, reminder, now)
		}

		return nil // về thôi làm gì nữa
	}
	// còn lại phần của lặp FRP
	// không cần reminder.Type == models.ReminderTypeRecurring vì đã check return  ở trên

	// if now.After(reminder.NextRecurring) || now.Equal(reminder.NextRecurring) {
	// 	return w.processFRP(ctx, reminder, now)
	// }

	if reminder.CanTriggerNow(reminder.NextRecurring) {

		if reminder.RepeatStrategy == models.RepeatStrategyCRPUntilComplete {
			if !utils.IsTimeValid(reminder.LastSentAt) || reminder.LastCompletedAt.After(reminder.LastSentAt) {
				// User đã complete → OK, trigger FRP
				return w.processFRP(ctx, reminder, now)

			}
		} else {
			// vòng lặp cứ đến thời điểm là được notifi tiếp
			return w.processFRP(ctx, reminder, now)
		}
	}

	// ========================================
	// STEP 2: Check CRP
	// ========================================
	if w.schedCalc.CanSendCRP(reminder, now) {
		return w.processCRP(ctx, reminder, now)
	}

	// ========================================
	// STEP 3: Just recalc next_action_at
	// ========================================
	nextAction := w.schedCalc.CalculateNextActionAt(reminder, now)
	if !nextAction.Equal(reminder.NextActionAt) {
		_ = w.reminderRepo.UpdateNextActionAt(ctx, reminder.ID, nextAction)
	}

	return nil
}

func (w *Worker) processCRPForOneTime(ctx context.Context, reminder *models.Reminder, now time.Time) error {
	log.Printf("🔔 Worker: CRP triggered for ONE-TIME %s (count: %d/%d)",
		reminder.ID, reminder.CRPCount+1, reminder.MaxCRP)

	if err := w.sendNotification(ctx, reminder); err != nil {
		return err // Lỗi gửi → stop, retry lần sau
	}

	reminder.LastSentAt = now
	reminder.CRPCount++

	reminder.NextCRP = now.Add(time.Duration(reminder.CRPIntervalSec) * time.Second)

	if reminder.MaxCRP == 0 {
		// ========================================
		// Case: MaxCRP = 0 (gửi 1 lần)
		// ========================================
		// Chỉ gửi 1 lần → xong
		log.Printf("🏁 Worker: One-time MaxCRP=0, mark completed")
		reminder.Status = models.ReminderStatusCompleted
		reminder.LastCompletedAt = now
		reminder.NextActionAt = time.Time{}
		// Clear next_action_at (không còn action nào)
		// Worker sẽ không query reminder này nữa (status = completed)

	} else if reminder.CRPCount >= reminder.MaxCRP {

		log.Printf("🏁 Worker: One-time quota reached (%d/%d), mark completed",
			reminder.CRPCount, reminder.MaxCRP)
		reminder.Status = models.ReminderStatusCompleted
		reminder.LastCompletedAt = now
		reminder.NextActionAt = time.Time{}

	} else {

		log.Printf("⏳ Worker: One-time still has quota (%d/%d), wait for next CRP",
			reminder.CRPCount, reminder.MaxCRP)

		reminder.NextActionAt = w.schedCalc.CalculateNextActionAt(reminder, now)
	}

	// ========================================
	// 5. Update database với tất cả changes
	// ========================================
	if err := w.reminderRepo.Update(ctx, reminder); err != nil {
		return fmt.Errorf("failed to update reminder after CRP: %w", err)
	}

	log.Printf("✅ Worker: One-time CRP processed (count=%d/%d). Status=%s, NextActionAt=%s",
		reminder.CRPCount, reminder.MaxCRP,
		reminder.Status,
		reminder.NextActionAt.Format("15:04:05"))
	return nil
}

// processFRP handles Father Recurrence Pattern trigger

func (w *Worker) processFRP(ctx context.Context, reminder *models.Reminder, now time.Time) error {
	log.Printf("Worker: FRP triggered for reminder %s", reminder.ID)
	reminder.SnoozeUntil = time.Time{} // ✅ Clear snooze

	// Get user and send notification
	if err := w.sendNotification(ctx, reminder); err != nil {
		log.Printf("❌ FRP sendNotification failed for %s, snoozing 30s: %v", reminder.ID, err)

		reminder.SnoozeUntil = now.Add(60 * time.Second)
		reminder.NextActionAt = reminder.SnoozeUntil

		_ = w.reminderRepo.Update(ctx, reminder)

		return err
	}

	// Update tracking
	reminder.LastSentAt = now
	reminder.CRPCount = 0

	// ========================================
	// CRITICAL FIX: Calculate next FRP
	// ========================================
	// For repeat_strategy = "crp_until_complete":
	// NextRecurring should NOT be recalculated here.
	// It should only be advanced when the user completes the reminder.
	// For other repeat strategies, it should be recalculated.

	if reminder.RepeatStrategy != models.RepeatStrategyCRPUntilComplete {
		nextRecurring, err := w.schedCalc.CalculateNextRecurring(reminder, now)
		if err != nil {
			log.Printf("Worker: Warning - failed to calc next FRP for %s: %v", reminder.ID, err)
			nextRecurring = now.Add(24 * time.Hour)
		}
		reminder.NextRecurring = nextRecurring
		log.Printf("📅 Calculated NextRecurring: %s (from now: %s)",
			nextRecurring.Format("15:04:05"), now.Format("15:04:05"))
	} else {
		// For CRPUntilComplete, NextRecurring is only updated on user completion.
		// We just log that it's not being advanced here.
		log.Printf("📅 NextRecurring for CRPUntilComplete reminder %s not advanced by FRP trigger.", reminder.ID)
	}

	// Recalc next_action_at
	reminder.NextActionAt = w.schedCalc.CalculateNextActionAt(reminder, now)
	reminder.NextCRP = now.Add(time.Duration(reminder.CRPIntervalSec) * time.Second)
	log.Printf("naq CRPIntervalSec: %d", reminder.CRPIntervalSec)
	log.Printf("NextRecurring: %s", reminder.NextRecurring)
	log.Printf("NextCRP: %s", reminder.NextCRP)
	// Update DB
	if err := w.reminderRepo.Update(ctx, reminder); err != nil {
		return fmt.Errorf("failed to update reminder after FRP: %w", err)
	}

	log.Printf("Worker: FRP processed. Next FRP: %s", reminder.NextRecurring.Format("15:04:05"))
	return nil
}

// processCRP handles Child Repeat Pattern (retry) trigger
// processCRP handles Child Repeat Pattern (retry) trigger
func (w *Worker) processCRP(ctx context.Context, reminder *models.Reminder, now time.Time) error {
	log.Printf("🔔 Worker: CRP triggered for reminder %s (count: %d/%d)", reminder.ID, reminder.CRPCount+1, reminder.MaxCRP)

	// Send notification
	if err := w.sendNotification(ctx, reminder); err != nil {
		return err
	}

	// Update tracking - IMPORTANT: Update BOTH LastSentAt and NextCRP
	reminder.LastSentAt = now
	reminder.CRPCount++

	// Calculate next CRP trigger time (CRITICAL FIX!)
	reminder.NextCRP = now.Add(time.Duration(reminder.CRPIntervalSec) * time.Second)

	// Check if one_time and reached quota
	if reminder.Type == models.ReminderTypeOneTime {
		if reminder.MaxCRP == 0 || reminder.CRPCount >= reminder.MaxCRP {
			log.Printf("🏁 Worker: One-time reminder reached quota, marking completed")
			reminder.Status = models.ReminderStatusCompleted
			reminder.LastCompletedAt = now
			reminder.NextActionAt = time.Time{} // Clear next_action_at
		} else {
			// Still has quota, recalc next_action_at
			reminder.NextActionAt = w.schedCalc.CalculateNextActionAt(reminder, now)
		}
	} else if reminder.Type == models.ReminderTypeRecurring { // chờ user bấm complete FRP mới được post tiếp
		// Recurring: Check if reached CRP quota
		if reminder.MaxCRP > 0 && reminder.CRPCount >= reminder.MaxCRP {
			log.Printf("⏸️  Worker: Recurring reminder CRP quota reached (%d/%d), waiting for FRP", reminder.CRPCount, reminder.MaxCRP)
			// Quota reached: Don't recalc next_action_at yet
			// Only FRP can advance, so next_action_at = next_recurring (if exists)
			if !reminder.NextRecurring.IsZero() {
				reminder.NextActionAt = reminder.NextRecurring
			} else {
				// No FRP set (shouldn't happen for recurring), clear it
				reminder.NextActionAt = time.Time{}
			}
		} else {
			// Still has CRP quota
			reminder.NextActionAt = w.schedCalc.CalculateNextActionAt(reminder, now)
		}
	}

	// Update DB
	if err := w.reminderRepo.Update(ctx, reminder); err != nil {
		return fmt.Errorf("failed to update reminder after CRP: %w", err)
	}

	log.Printf("✅ Worker: CRP processed (count=%d/%d). Next CRP: %s, NextActionAt: %s",
		reminder.CRPCount, reminder.MaxCRP,
		reminder.NextCRP.Format("15:04:05"),
		reminder.NextActionAt.Format("15:04:05"))
	return nil
}

// sendNotification sends FCM notification to user
func (w *Worker) sendNotification(ctx context.Context, reminder *models.Reminder) error {
	user, err := w.userRepo.GetByID(ctx, reminder.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if !user.IsFCMActive || user.FCMToken == "" {
		return fmt.Errorf("user FCM not active")
	}

	// Send FCM (no-op if not configured)
	if w.fcmSender != nil {
		err := w.fcmSender.SendNotification(user.FCMToken, reminder.Title, reminder.Description)
		if err != nil {
			// Check if token error
			if isTokenError(err.Error()) {
				_ = w.userRepo.DisableFCM(ctx, user.ID)
				return fmt.Errorf("token disabled: %w", err)
			}
			// System error
			_ = w.userRepo.SetFCMError(ctx, user.ID, err.Error())
			return fmt.Errorf("fcm system error: %w", err)
		}
	}

	return nil
}

// ========================================
// HELPERS
// ========================================

func isTokenError(errStr string) bool {
	return errStr == "UNREGISTERED" ||
		errStr == "INVALID_ARGUMENT" ||
		errStr == "NOT_FOUND" ||
		errStr == "user FCM not active"
}

func isSystemFCMError(err error) bool {
	if err == nil {
		return false
	}

	// System errors: auth, permission, timeout
	return errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, context.Canceled)
}
