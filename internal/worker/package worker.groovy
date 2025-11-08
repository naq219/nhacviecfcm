// ============================================================================
// PROCESSREMINDER: Fix - Check repeat_strategy trước khi trigger FRP
// ============================================================================

func (w *Worker) processReminder(ctx context.Context, reminder *models.Reminder, now time.Time) error {
	log.Printf("📋 Loaded reminder %s: NextCRP=%v, LastSentAt=%v, CRPCount=%d, MaxCRP=%d",
		reminder.ID,
		reminder.NextCRP,
		reminder.LastSentAt,
		reminder.CRPCount,
		reminder.MaxCRP)

	// ========================================
	// STEP 0: Kiểm tra reminder bị snooze không
	// ========================================
	if reminder.IsSnoozeUntilActive(now) {
		log.Printf("😴 Worker: Reminder %s is snoozed until %s", reminder.ID, reminder.SnoozeUntil.Format("15:04:05"))
		nextAction := w.schedCalc.CalculateNextActionAt(reminder, now)
		if !nextAction.Equal(reminder.NextActionAt) {
			_ = w.reminderRepo.UpdateNextActionAt(ctx, reminder.ID, nextAction)
		}
		return nil
	}

	// ========================================
	// STEP 1: Check FRP (Father Recurrence Pattern)
	// ========================================
	// Điều kiện 1: Type phải là recurring
	// Điều kiện 2: NextRecurring phải valid
	// Điều kiện 3: Now phải >= NextRecurring (đến hạn)
	// Điều kiện 4: (NEW) Nếu repeat_strategy = "crp_until_complete"
	//              thì phải check user đã complete lần trước chưa
	if reminder.Type == models.ReminderTypeRecurring && reminder.IsNextRecurringSet() 
    && now.After(reminder.NextRecurring) || now.Equal(reminder.NextRecurring) {
			
        if reminder.RepeatStrategy == models.RepeatStrategyCRPUntilComplete {
            
            if reminder.IsLastCompletedAtSet() {
                
                if reminder.LastCompletedAt.After(reminder.LastSentAt) {
                    // ========================================
                    // User đã complete → OK, trigger FRP
                    // ========================================
                    log.Printf("✅ Worker: repeat_strategy=crp_until_complete, user completed. Proceed to FRP.")
                    log.Printf("   LastCompletedAt=%s > LastSentAt=%s",
                        reminder.LastCompletedAt.Format("15:04:05"),
                        reminder.LastSentAt.Format("15:04:05"))
                    return w.processFRP(ctx, reminder, now)
                } else {
                    // ========================================
                    // User chưa complete → skip FRP
                    // ========================================
                    log.Printf("⏸️  Worker: repeat_strategy=crp_until_complete, user not completed yet. Skip FRP.")
                    log.Printf("   LastCompletedAt=%s <= LastSentAt=%s (waiting for user)",
                        reminder.LastCompletedAt.Format("15:04:05"),
                        reminder.LastSentAt.Format("15:04:05"))
                    // Fall through → Check CRP instead
                }
            } else {
                // ========================================
                // LastCompletedAt trống (lần đầu tiên)
                // ========================================
                // Lần đầu: không có complete event → trigger FRP
                log.Printf("✅ Worker: repeat_strategy=crp_until_complete, first FRP trigger (no previous completion).")
                return w.processFRP(ctx, reminder, now)
            }
        } else if reminder.RepeatStrategy == models.RepeatStrategyNone {
            // ========================================
            // Trường hợp: repeat_strategy = none
            // ========================================
            // "none" = auto-repeat, không chờ user complete
            // → Luôn trigger FRP khi đến hạn
            log.Printf("✅ Worker: repeat_strategy=none. Proceed to FRP.")
            return w.processFRP(ctx, reminder, now)
        } else {
            // ========================================
            // Trường hợp: repeat_strategy khác (future compatibility)
            // ========================================
            log.Printf("✅ Worker: repeat_strategy=%s. Proceed to FRP.", reminder.RepeatStrategy)
            return w.processFRP(ctx, reminder, now)
        }
		
	}

	// ========================================
	// STEP 2: Check CRP (Child Repeat Pattern / Retry)
	// ========================================
	// CRP chỉ trigger nếu FRP chưa trigger
	// Check: CanSendCRP() return true?
	if w.schedCalc.CanSendCRP(reminder, now) {
		return w.processCRP(ctx, reminder, now)
	}

	// ========================================
	// STEP 3: Không có action → chỉ recalc next_action_at
	// ========================================
	nextAction := w.schedCalc.CalculateNextActionAt(reminder, now)
	if !nextAction.Equal(reminder.NextActionAt) {
		_ = w.reminderRepo.UpdateNextActionAt(ctx, reminder.ID, nextAction)
	}

	return nil
}

// ============================================================================
// HELPER: Check LastCompletedAt
// ============================================================================
// (Thêm vào models/reminder.go nếu chưa có)

// IsLastCompletedAtSet checks if LastCompletedAt is properly set
func (r *Reminder) IsLastCompletedAtSet() bool {
	return IsTimeValid(r.LastCompletedAt)
}

// ============================================================================
// TIMELINE EXAMPLE: repeat_strategy = crp_until_complete
// ============================================================================
/*

Scenario: Recurring reminder mỗi 3 phút, CRP 3x 20s, repeat_strategy=crp_until_complete

=== CYCLE 1 ===

12:00:00 - FRP TRIGGER
  LastSentAt = 12:00:00
  LastCompletedAt = EMPTY (lần đầu)
  NextRecurring = 12:03:00 (3 phút sau)
  CRPCount = 0
  NextCRP = 12:00:00

12:00:20 - CRP 1
  LastSentAt = 12:00:20
  CRPCount = 1
  NextCRP = 12:00:40

12:00:40 - CRP 2
  LastSentAt = 12:00:40
  CRPCount = 2
  NextCRP = 12:01:00

12:01:00 - CRP 3
  LastSentAt = 12:01:00
  CRPCount = 3
  NextCRP = EMPTY (quota reached)
  NextActionAt = EMPTY (chờ user complete)

12:02:00 - USER CLICK "COMPLETE" ✅
  LastCompletedAt = 12:02:00
  CRPCount = 0 (reset)
  NextRecurring = 12:05:00 (tính từ 12:02:00 + 3 phút)
  NextCRP = 12:05:00
  NextActionAt = 12:05:00

=== CYCLE 2 ===

12:05:00 - FRP CHECK
  now (12:05:00) >= NextRecurring (12:05:00)? YES
  repeat_strategy = crp_until_complete
  Check: LastCompletedAt (12:02:00) > LastSentAt (12:01:00)? YES ✅
  → TRIGGER FRP!

12:05:00 - FRP TRIGGER
  LastSentAt = 12:05:00
  CRPCount = 0
  NextRecurring = 12:08:00 (3 phút sau)
  NextCRP = 12:05:00

12:05:20 - CRP 1
  ...

=== FAILURE CASE ===

12:01:00 - CRP 3 (không user complete)
  LastSentAt = 12:01:00
  CRPCount = 3
  NextActionAt = EMPTY (chờ)

12:03:00 - FRP CHECK (time đã tới)
  now (12:03:00) >= NextRecurring (12:03:00)? YES
  repeat_strategy = crp_until_complete
  Check: LastCompletedAt (EMPTY) > LastSentAt (12:01:00)? NO ❌
  → SKIP FRP! (chờ user complete)

12:05:00 - Vẫn chờ user, FRP không trigger

12:05:30 - USER CLICK "COMPLETE"
  LastCompletedAt = 12:05:30
  NextRecurring = 12:08:30 (tính từ complete time)
  NextActionAt = 12:08:30

12:08:30 - FRP TRIGGER (sau user complete)
  Check: LastCompletedAt (12:05:30) > LastSentAt (12:01:00)? YES ✅
  → TRIGGER FRP!

*/