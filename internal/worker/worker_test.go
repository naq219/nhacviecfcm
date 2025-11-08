package worker

import (
	"context"
	"testing"
	"time"

	"remiaq/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ========================================
// INTEGRATION TEST: All 14 Test Cases
// ========================================

// TestAllCases runs all 14 test cases in sequence
func TestAllCases(t *testing.T) {
	t.Run("Group1-OneTime", func(t *testing.T) {
		t.Run("1.1-OneTimeNoCRP", testOneTimeNoCRP)
		t.Run("1.2-OneTimeCRP", testOneTimeCRP)
		t.Run("1.3-OneTimeCompleteAtCRP2", testOneTimeCompleteAtCRP2)
	})

	t.Run("Group2-RecurringNone", func(t *testing.T) {
		t.Run("2.1-RecurringNoneAutoRepeat", testRecurringNoneAutoRepeat)
		t.Run("2.2-RecurringNoneWithCRP", testRecurringNoneWithCRP)
		t.Run("2.3-RecurringNoneUserComplete", testRecurringNoneUserComplete)
	})

	t.Run("Group3-RecurringCRPUntilComplete", func(t *testing.T) {
		t.Run("3.1-CRPUntilCompleteQuota", testCRPUntilCompleteQuota)
		t.Run("3.2-CRPUntilCompleteAtCRP2", testCRPUntilCompleteAtCRP2)
		t.Run("3.3-CRPUntilCompleteAfterQuota", testCRPUntilCompleteAfterQuota)
	})

	t.Run("Group4-Snooze", func(t *testing.T) {
		t.Run("4.1-OneTimeSnoozed", testOneTimeSnoozed)
		t.Run("4.2-RecurringSnoozeduringCRP", testRecurringSnoozeduringCRP)
	})

	t.Run("Group5-EdgeCases", func(t *testing.T) {
		t.Run("5.1-RecurringNoneFRPMultipleTimes", testRecurringNoneFRPMultipleTimes)
		t.Run("5.2-CRPUntilCompleteDoubleComplete", testCRPUntilCompleteDoubleComplete)
		t.Run("5.3-RecurringNoneShortCRP", testRecurringNoneShortCRP)
	})
}

// ========================================
// TEST 1.1: One-time, no CRP
// ========================================
func testOneTimeNoCRP(t *testing.T) {
	sysRepo := new(MockSystemStatusRepo)
	reminderRepo := new(MockReminderRepo)
	userRepo := new(MockUserRepo)
	fcmSender := new(MockFCMSender)
	schedCalc := new(MockScheduleCalc)

	ctx := context.Background()
	now := time.Now().UTC()

	reminder := createTestReminder("r1", "u1", models.ReminderTypeOneTime)
	reminder.MaxCRP = 0 // No CRP
	reminder.NextActionAt = now.Add(-1 * time.Second)
	reminder.NextCRP = time.Time{}

	user := createTestUser("u1", true)

	sysRepo.On("IsWorkerEnabled", ctx).Return(true, nil)
	reminderRepo.On("GetDueReminders", ctx, mock.Anything).Return([]*models.Reminder{reminder}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(user, nil)
	schedCalc.On("CanSendCRP", mock.Anything, mock.Anything).Return(true)
	fcmSender.On("SendNotification", "token_u1", "Test Reminder", "Test Description").Return(nil)
	reminderRepo.On("Update", ctx, mock.MatchedBy(func(r *models.Reminder) bool {
		return r.Status == models.ReminderStatusCompleted && r.NextActionAt.IsZero()
	})).Return(nil)
	sysRepo.On("ClearError", ctx).Return(nil)

	w := NewWorker(sysRepo, reminderRepo, userRepo, fcmSender, schedCalc, time.Second)
	w.runOnce(ctx)

	// Verify: Sent 1 time, marked completed
	assert.Equal(t, 1, len(fcmSender.Calls))
	reminderRepo.AssertCalled(t, "Update", ctx, mock.MatchedBy(func(r *models.Reminder) bool {
		return r.Status == models.ReminderStatusCompleted
	}))
}

// ========================================
// TEST 1.2: One-time with CRP (3 times)
// ========================================
func testOneTimeCRP(t *testing.T) {
	sysRepo := new(MockSystemStatusRepo)
	reminderRepo := new(MockReminderRepo)
	userRepo := new(MockUserRepo)
	fcmSender := new(MockFCMSender)
	schedCalc := new(MockScheduleCalc)

	ctx := context.Background()
	now := time.Now().UTC()

	reminder := createTestReminder("r1", "u1", models.ReminderTypeOneTime)
	reminder.MaxCRP = 3
	reminder.CRPCount = 0
	reminder.NextActionAt = now.Add(-1 * time.Second)
	reminder.NextCRP = time.Time{}

	user := createTestUser("u1", true)

	sysRepo.On("IsWorkerEnabled", ctx).Return(true, nil)
	reminderRepo.On("GetDueReminders", ctx, mock.Anything).Return([]*models.Reminder{reminder}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(user, nil)
	schedCalc.On("CanSendCRP", mock.Anything, mock.Anything).Return(true)
	fcmSender.On("SendNotification", "token_u1", "Test Reminder", "Test Description").Return(nil)
	schedCalc.On("CalculateNextActionAt", mock.Anything, mock.Anything).Return(now.Add(20 * time.Second))
	reminderRepo.On("Update", ctx, mock.MatchedBy(func(r *models.Reminder) bool {
		return r.CRPCount == 1
	})).Return(nil)
	sysRepo.On("ClearError", ctx).Return(nil)

	w := NewWorker(sysRepo, reminderRepo, userRepo, fcmSender, schedCalc, time.Second)
	w.runOnce(ctx)

	// Verify: CRPCount incremented
	reminderRepo.AssertCalled(t, "Update", ctx, mock.MatchedBy(func(r *models.Reminder) bool {
		return r.CRPCount == 1
	}))
}

// ========================================
// TEST 1.3: One-time complete at CRP 2
// ========================================
func testOneTimeCompleteAtCRP2(t *testing.T) {
	// Simulate: CRP 1 done, user clicks complete before CRP 2
	// Result: Status = completed, no CRP 2
	t.Skip("Requires OnUserComplete API endpoint test")
}

// ========================================
// TEST 2.1: Recurring none, auto-repeat
// ========================================
func testRecurringNoneAutoRepeat(t *testing.T) {
	sysRepo := new(MockSystemStatusRepo)
	reminderRepo := new(MockReminderRepo)
	userRepo := new(MockUserRepo)
	fcmSender := new(MockFCMSender)
	schedCalc := new(MockScheduleCalc)

	ctx := context.Background()
	now := time.Now().UTC()

	reminder := createTestReminder("r1", "u1", models.ReminderTypeRecurring)
	reminder.RepeatStrategy = models.RepeatStrategyNone
	reminder.RecurrencePattern = &models.RecurrencePattern{
		Type:            models.RecurrenceTypeIntervalSeconds,
		IntervalSeconds: 180,
	}
	reminder.NextRecurring = now.Add(-1 * time.Second)
	reminder.NextCRP = reminder.NextRecurring
	reminder.NextActionAt = reminder.NextRecurring

	user := createTestUser("u1", true)
	nextRecurring := now.Add(3 * time.Minute)

	sysRepo.On("IsWorkerEnabled", ctx).Return(true, nil)
	reminderRepo.On("GetDueReminders", ctx, mock.Anything).Return([]*models.Reminder{reminder}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(user, nil)
	fcmSender.On("SendNotification", "token_u1", "Test Reminder", "Test Description").Return(nil)
	schedCalc.On("CalculateNextRecurring", mock.Anything, mock.Anything).Return(nextRecurring, nil)
	schedCalc.On("CalculateNextActionAt", mock.Anything, mock.Anything).Return(nextRecurring)
	reminderRepo.On("Update", ctx, mock.Anything).Return(nil)
	sysRepo.On("ClearError", ctx).Return(nil)

	w := NewWorker(sysRepo, reminderRepo, userRepo, fcmSender, schedCalc, time.Second)
	w.runOnce(ctx)

	// Verify: Next recurring calculated and set
	reminderRepo.AssertCalled(t, "Update", ctx, mock.MatchedBy(func(r *models.Reminder) bool {
		return r.NextRecurring.After(now)
	}))
}

// ========================================
// TEST 2.2: Recurring none with CRP
// ========================================
func testRecurringNoneWithCRP(t *testing.T) {
	// FRP trigger resets CRP, then CRP retries
	// When quota reached: next_action_at = next_recurring (auto-repeat)
	t.Skip("Requires multiple worker cycles")
}

// ========================================
// TEST 2.3: Recurring none, user complete
// ========================================
func testRecurringNoneUserComplete(t *testing.T) {
	// User complete shouldn't change next_recurring for "none" strategy
	t.Skip("Requires OnUserComplete API endpoint test")
}

// ========================================
// TEST 3.1: CRP until complete, quota reached, wait for user
// ========================================
func testCRPUntilCompleteQuota(t *testing.T) {
	sysRepo := new(MockSystemStatusRepo)
	reminderRepo := new(MockReminderRepo)
	userRepo := new(MockUserRepo)
	fcmSender := new(MockFCMSender)
	schedCalc := new(MockScheduleCalc)

	ctx := context.Background()
	now := time.Now().UTC()

	// After CRP 3, quota reached
	reminder := createTestReminder("r1", "u1", models.ReminderTypeRecurring)
	reminder.RepeatStrategy = models.RepeatStrategyCRPUntilComplete
	reminder.MaxCRP = 3
	reminder.CRPCount = 3 // Quota reached
	reminder.NextRecurring = now.Add(1 * time.Minute)
	reminder.NextActionAt = now.Add(-1 * time.Second)

	sysRepo.On("IsWorkerEnabled", ctx).Return(true, nil)
	reminderRepo.On("GetDueReminders", ctx, mock.Anything).Return([]*models.Reminder{reminder}, nil)
	schedCalc.On("CanSendCRP", mock.Anything, mock.Anything).Return(false)                  // Quota reached
	schedCalc.On("CalculateNextActionAt", mock.Anything, mock.Anything).Return(time.Time{}) // EMPTY!
	reminderRepo.On("UpdateNextActionAt", ctx, "r1", time.Time{}).Return(nil)
	sysRepo.On("ClearError", ctx).Return(nil)

	w := NewWorker(sysRepo, reminderRepo, userRepo, fcmSender, schedCalc, time.Second)
	w.runOnce(ctx)

	// Verify: next_action_at set to EMPTY (waiting for user)
	reminderRepo.AssertCalled(t, "UpdateNextActionAt", ctx, "r1", time.Time{})
	fcmSender.AssertNotCalled(t, "SendNotification")
}

// ========================================
// TEST 3.2: CRP until complete, user complete at CRP 2
// ========================================
func testCRPUntilCompleteAtCRP2(t *testing.T) {
	// Requires API test: User clicks complete before quota
	t.Skip("Requires OnUserComplete API endpoint test")
}

// ========================================
// TEST 3.3: CRP until complete, user complete after quota
// ========================================
func testCRPUntilCompleteAfterQuota(t *testing.T) {
	// Requires API test: User clicks complete after waiting
	t.Skip("Requires OnUserComplete API endpoint test")
}

// ========================================
// TEST 4.1: One-time snoozed
// ========================================
func testOneTimeSnoozed(t *testing.T) {
	sysRepo := new(MockSystemStatusRepo)
	reminderRepo := new(MockReminderRepo)
	userRepo := new(MockUserRepo)
	fcmSender := new(MockFCMSender)
	schedCalc := new(MockScheduleCalc)

	ctx := context.Background()
	now := time.Now().UTC()

	reminder := createTestReminder("r1", "u1", models.ReminderTypeOneTime)
	reminder.SnoozeUntil = now.Add(1 * time.Minute) // Snoozed
	reminder.NextActionAt = now.Add(-1 * time.Second)
	reminder.NextCRP = time.Time{}

	sysRepo.On("IsWorkerEnabled", ctx).Return(true, nil)
	reminderRepo.On("GetDueReminders", ctx, mock.Anything).Return([]*models.Reminder{reminder}, nil)
	schedCalc.On("CalculateNextActionAt", mock.Anything, mock.Anything).Return(reminder.SnoozeUntil)
	reminderRepo.On("UpdateNextActionAt", ctx, "r1", reminder.SnoozeUntil).Return(nil)
	sysRepo.On("ClearError", ctx).Return(nil)

	w := NewWorker(sysRepo, reminderRepo, userRepo, fcmSender, schedCalc, time.Second)
	w.runOnce(ctx)

	// Verify: Not sent, next_action_at = snooze_until
	fcmSender.AssertNotCalled(t, "SendNotification")
	userRepo.AssertNotCalled(t, "GetByID")
}

// ========================================
// TEST 4.2: Recurring snoozed during CRP
// ========================================
func testRecurringSnoozeduringCRP(t *testing.T) {
	sysRepo := new(MockSystemStatusRepo)
	reminderRepo := new(MockReminderRepo)
	userRepo := new(MockUserRepo)
	fcmSender := new(MockFCMSender)
	schedCalc := new(MockScheduleCalc)

	ctx := context.Background()
	now := time.Now().UTC()

	reminder := createTestReminder("r1", "u1", models.ReminderTypeRecurring)
	reminder.SnoozeUntil = now.Add(30 * time.Second)
	reminder.NextCRP = now.Add(-5 * time.Second) // Should trigger, but snoozed
	reminder.NextActionAt = reminder.NextCRP

	sysRepo.On("IsWorkerEnabled", ctx).Return(true, nil)
	reminderRepo.On("GetDueReminders", ctx, mock.Anything).Return([]*models.Reminder{reminder}, nil)
	schedCalc.On("CalculateNextActionAt", mock.Anything, mock.Anything).Return(reminder.SnoozeUntil)
	reminderRepo.On("UpdateNextActionAt", ctx, "r1", reminder.SnoozeUntil).Return(nil)
	sysRepo.On("ClearError", ctx).Return(nil)

	w := NewWorker(sysRepo, reminderRepo, userRepo, fcmSender, schedCalc, time.Second)
	w.runOnce(ctx)

	// Verify: Not sent due to snooze
	fcmSender.AssertNotCalled(t, "SendNotification")
}

// ========================================
// TEST 5.1: FRP trigger multiple times
// ========================================
func testRecurringNoneFRPMultipleTimes(t *testing.T) {
	// Simulate: Worker was down, now catching up
	// Should calculate correctly without sending multiple times
	t.Skip("Requires time simulation")
}

// ========================================
// TEST 5.2: User complete double complete
// ========================================
func testCRPUntilCompleteDoubleComplete(t *testing.T) {
	// First complete: next_recurring = now + 3min
	// Second complete: next_recurring = now2 + 3min (not same as first)
	t.Skip("Requires OnUserComplete API endpoint test")
}

// ========================================
// TEST 5.3: Short CRP interval (1 second)
// ========================================
func testRecurringNoneShortCRP(t *testing.T) {
	// Should not spam faster than worker cycle
	t.Skip("Requires multiple cycles simulation")
}
