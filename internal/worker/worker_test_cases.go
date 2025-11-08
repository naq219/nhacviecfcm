package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"remiaq/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ========================================
// MOCK IMPLEMENTATIONS
// ========================================

type MockReminderRepo struct {
	mock.Mock
}

func (m *MockReminderRepo) GetDueReminders(ctx context.Context, beforeTime time.Time) ([]*models.Reminder, error) {
	args := m.Called(ctx, beforeTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Reminder), args.Error(1)
}

func (m *MockReminderRepo) Update(ctx context.Context, reminder *models.Reminder) error {
	args := m.Called(ctx, reminder)
	return args.Error(0)
}

func (m *MockReminderRepo) UpdateCRPCount(ctx context.Context, id string, crpCount int) error {
	args := m.Called(ctx, id, crpCount)
	return args.Error(0)
}

func (m *MockReminderRepo) UpdateNextRecurring(ctx context.Context, id string, nextRecurring time.Time) error {
	args := m.Called(ctx, id, nextRecurring)
	return args.Error(0)
}

func (m *MockReminderRepo) UpdateNextCRP(ctx context.Context, id string, nextCRP time.Time) error {
	args := m.Called(ctx, id, nextCRP)
	return args.Error(0)
}

func (m *MockReminderRepo) UpdateNextActionAt(ctx context.Context, id string, nextActionAt time.Time) error {
	args := m.Called(ctx, id, nextActionAt)
	return args.Error(0)
}

func (m *MockReminderRepo) UpdateLastSent(ctx context.Context, id string, lastSentAt string) error {
	args := m.Called(ctx, id, lastSentAt)
	return args.Error(0)
}

func (m *MockReminderRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) GetByID(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) DisableFCM(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepo) SetFCMError(ctx context.Context, userID string, errMsg string) error {
	args := m.Called(ctx, userID, errMsg)
	return args.Error(0)
}

type MockFCMSender struct {
	mock.Mock
}

func (m *MockFCMSender) SendNotification(token, title, body string) error {
	args := m.Called(token, title, body)
	return args.Error(0)
}

type MockScheduleCalc struct {
	mock.Mock
}

func (m *MockScheduleCalc) CalculateNextActionAt(reminder *models.Reminder, now time.Time) time.Time {
	args := m.Called(reminder, now)
	if args.Get(0) == nil {
		return time.Time{}
	}
	return args.Get(0).(time.Time)
}

func (m *MockScheduleCalc) CalculateNextRecurring(reminder *models.Reminder, now time.Time) (time.Time, error) {
	args := m.Called(reminder, now)
	if args.Get(0) == nil {
		return time.Time{}, args.Error(1)
	}
	return args.Get(0).(time.Time), args.Error(1)
}

func (m *MockScheduleCalc) CanSendCRP(reminder *models.Reminder, now time.Time) bool {
	args := m.Called(reminder, now)
	return args.Bool(0)
}

type MockSystemStatusRepo struct {
	mock.Mock
}

func (m *MockSystemStatusRepo) IsWorkerEnabled(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *MockSystemStatusRepo) UpdateError(ctx context.Context, errMsg string) error {
	args := m.Called(ctx, errMsg)
	return args.Error(0)
}

func (m *MockSystemStatusRepo) ClearError(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSystemStatusRepo) DisableWorker(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// ========================================
// TEST HELPERS
// ========================================

func createTestReminder(id, userID string, reminderType string) *models.Reminder {
	return &models.Reminder{
		ID:             id,
		UserID:         userID,
		Title:          "Test Reminder",
		Description:    "Test Description",
		Type:           reminderType,
		CalendarType:   models.CalendarTypeSolar,
		Status:         models.ReminderStatusActive,
		RepeatStrategy: models.RepeatStrategyNone,
		MaxCRP:         3,
		CRPIntervalSec: 20,
		CRPCount:       0,
	}
}

func createTestUser(id string, fcmActive bool) *models.User {
	return &models.User{
		ID:          id,
		Email:       "test@example.com",
		FCMToken:    "token_" + id,
		IsFCMActive: fcmActive,
	}
}

// ========================================
// TEST CASES
// ========================================

// Test 1: Worker skipped when disabled
func TestWorkerSkippedWhenDisabled(t *testing.T) {
	sysRepo := new(MockSystemStatusRepo)
	reminderRepo := new(MockReminderRepo)
	userRepo := new(MockUserRepo)
	fcmSender := new(MockFCMSender)
	schedCalc := new(MockScheduleCalc)

	sysRepo.On("IsWorkerEnabled", mock.Anything).Return(false, nil)

	w := NewWorker(sysRepo, reminderRepo, userRepo, fcmSender, schedCalc, time.Second)
	ctx := context.Background()
	// now := time.Now().UTC()

	w.runOnce(ctx)

	// Should only call IsWorkerEnabled, nothing else
	sysRepo.AssertCalled(t, "IsWorkerEnabled", mock.Anything)
	reminderRepo.AssertNotCalled(t, "GetDueReminders", mock.Anything, mock.Anything)
}

// Test 2: Worker processes one-time reminder with CRP
func TestProcessOneTimeReminderWithCRP(t *testing.T) {
	sysRepo := new(MockSystemStatusRepo)
	reminderRepo := new(MockReminderRepo)
	userRepo := new(MockUserRepo)
	fcmSender := new(MockFCMSender)
	schedCalc := new(MockScheduleCalc)

	ctx := context.Background()
	now := time.Now().UTC()

	// Setup reminder
	reminder := createTestReminder("r1", "u1", models.ReminderTypeOneTime)
	reminder.MaxCRP = 3
	reminder.CRPCount = 0
	reminder.LastSentAt = time.Time{}
	reminder.NextActionAt = now.Add(-1 * time.Second)

	user := createTestUser("u1", true)

	// Setup mocks
	sysRepo.On("IsWorkerEnabled", ctx).Return(true, nil)
	reminderRepo.On("GetDueReminders", ctx, mock.MatchedBy(func(t time.Time) bool {
		return t.After(now.Add(-time.Second)) && t.Before(now.Add(time.Second))
	})).Return([]*models.Reminder{reminder}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(user, nil)
	fcmSender.On("SendNotification", "token_u1", "Test Reminder", "Test Description").Return(nil)
	schedCalc.On("CalculateNextActionAt", reminder, mock.Anything).Return(now.Add(20 * time.Second))
	reminderRepo.On("Update", ctx, mock.MatchedBy(func(r *models.Reminder) bool {
		return r.ID == "r1" && r.CRPCount == 1 && r.LastSentAt.After(now.Add(-time.Second))
	})).Return(nil)
	sysRepo.On("ClearError", ctx).Return(nil)

	w := NewWorker(sysRepo, reminderRepo, userRepo, fcmSender, schedCalc, time.Second)
	w.runOnce(ctx)

	// Verify FCM was called
	fcmSender.AssertCalled(t, "SendNotification", "token_u1", "Test Reminder", "Test Description")
	// Verify DB was updated with CRPCount=1
	reminderRepo.AssertCalled(t, "Update", ctx, mock.Anything)
}

// Test 3: Recurring reminder with FRP trigger
func TestProcessRecurringReminderWithFRP(t *testing.T) {
	sysRepo := new(MockSystemStatusRepo)
	reminderRepo := new(MockReminderRepo)
	userRepo := new(MockUserRepo)
	fcmSender := new(MockFCMSender)
	schedCalc := new(MockScheduleCalc)

	ctx := context.Background()
	now := time.Now().UTC()

	// Setup recurring reminder - FRP should trigger
	reminder := createTestReminder("r1", "u1", models.ReminderTypeRecurring)
	reminder.NextRecurring = now.Add(-1 * time.Second) // FRP is due
	reminder.NextActionAt = reminder.NextRecurring
	reminder.CRPCount = 0
	reminder.RepeatStrategy = models.RepeatStrategyNone

	user := createTestUser("u1", true)
	nextRecurring := now.Add(3 * time.Minute)

	// Setup mocks
	sysRepo.On("IsWorkerEnabled", ctx).Return(true, nil)
	reminderRepo.On("GetDueReminders", ctx, mock.Anything).Return([]*models.Reminder{reminder}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(user, nil)
	fcmSender.On("SendNotification", "token_u1", "Test Reminder", "Test Description").Return(nil)
	schedCalc.On("CalculateNextRecurring", reminder, mock.Anything).Return(nextRecurring, nil)
	schedCalc.On("CalculateNextActionAt", mock.Anything, mock.Anything).Return(nextRecurring)
	reminderRepo.On("Update", ctx, mock.MatchedBy(func(r *models.Reminder) bool {
		return r.ID == "r1" && r.CRPCount == 0 && r.NextRecurring.After(now)
	})).Return(nil)
	sysRepo.On("ClearError", ctx).Return(nil)

	w := NewWorker(sysRepo, reminderRepo, userRepo, fcmSender, schedCalc, time.Second)
	w.runOnce(ctx)

	// Verify FCM was called for FRP
	fcmSender.AssertCalled(t, "SendNotification", "token_u1", "Test Reminder", "Test Description")
	// Verify next_recurring was calculated
	reminderRepo.AssertCalled(t, "Update", ctx, mock.Anything)
}

// Test 4: CRP retry after FRP
func TestProcessCRPRetryAfterFRP(t *testing.T) {
	sysRepo := new(MockSystemStatusRepo)
	reminderRepo := new(MockReminderRepo)
	userRepo := new(MockUserRepo)
	fcmSender := new(MockFCMSender)
	schedCalc := new(MockScheduleCalc)

	ctx := context.Background()
	now := time.Now().UTC()

	// Setup reminder - FRP already sent, now CRP should trigger
	reminder := createTestReminder("r1", "u1", models.ReminderTypeRecurring)
	reminder.NextRecurring = now.Add(10 * time.Second) // FRP not due yet
	reminder.LastSentAt = now.Add(-25 * time.Second)   // Sent 25s ago
	reminder.NextCRP = now.Add(-5 * time.Second)       // CRP is due (was 5s ago)
	reminder.CRPCount = 1
	reminder.MaxCRP = 3
	reminder.CRPIntervalSec = 20
	reminder.NextActionAt = reminder.NextCRP

	user := createTestUser("u1", true)

	// Setup mocks
	sysRepo.On("IsWorkerEnabled", ctx).Return(true, nil)
	reminderRepo.On("GetDueReminders", ctx, mock.Anything).Return([]*models.Reminder{reminder}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(user, nil)
	fcmSender.On("SendNotification", "token_u1", "Test Reminder", "Test Description").Return(nil)
	schedCalc.On("CanSendCRP", reminder, mock.Anything).Return(true)
	schedCalc.On("CalculateNextActionAt", mock.Anything, mock.Anything).Return(now.Add(20 * time.Second))
	reminderRepo.On("Update", ctx, mock.MatchedBy(func(r *models.Reminder) bool {
		return r.ID == "r1" && r.CRPCount == 2
	})).Return(nil)
	sysRepo.On("ClearError", ctx).Return(nil)

	w := NewWorker(sysRepo, reminderRepo, userRepo, fcmSender, schedCalc, time.Second)
	w.runOnce(ctx)

	// Verify CRP increment
	reminderRepo.AssertCalled(t, "Update", ctx, mock.MatchedBy(func(r *models.Reminder) bool {
		return r.CRPCount == 2
	}))
}

// Test 5: One-time reminder reaches quota and completes
func TestOneTimeReminderReachesQuota(t *testing.T) {
	sysRepo := new(MockSystemStatusRepo)
	reminderRepo := new(MockReminderRepo)
	userRepo := new(MockUserRepo)
	fcmSender := new(MockFCMSender)
	schedCalc := new(MockScheduleCalc)

	ctx := context.Background()
	now := time.Now().UTC()

	// Setup reminder - about to reach quota
	reminder := createTestReminder("r1", "u1", models.ReminderTypeOneTime)
	reminder.MaxCRP = 3
	reminder.CRPCount = 2 // Next will be 3rd
	reminder.LastSentAt = now.Add(-25 * time.Second)
	reminder.NextCRP = now.Add(-5 * time.Second)
	reminder.NextActionAt = reminder.NextCRP

	user := createTestUser("u1", true)

	// Setup mocks
	sysRepo.On("IsWorkerEnabled", ctx).Return(true, nil)
	reminderRepo.On("GetDueReminders", ctx, mock.Anything).Return([]*models.Reminder{reminder}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(user, nil)
	fcmSender.On("SendNotification", "token_u1", "Test Reminder", "Test Description").Return(nil)
	schedCalc.On("CanSendCRP", reminder, mock.Anything).Return(true)
	reminderRepo.On("Update", ctx, mock.MatchedBy(func(r *models.Reminder) bool {
		return r.ID == "r1" &&
			r.CRPCount == 3 &&
			r.Status == models.ReminderStatusCompleted &&
			r.NextActionAt.IsZero()
	})).Return(nil)
	sysRepo.On("ClearError", ctx).Return(nil)

	w := NewWorker(sysRepo, reminderRepo, userRepo, fcmSender, schedCalc, time.Second)
	w.runOnce(ctx)

	// Verify reminder is marked as completed
	reminderRepo.AssertCalled(t, "Update", ctx, mock.MatchedBy(func(r *models.Reminder) bool {
		return r.Status == models.ReminderStatusCompleted
	}))
}

// Test 6: CRP quota reached for recurring reminder
func TestRecurringReminderCRPQuotaReached(t *testing.T) {
	sysRepo := new(MockSystemStatusRepo)
	reminderRepo := new(MockReminderRepo)
	userRepo := new(MockUserRepo)
	fcmSender := new(MockFCMSender)
	schedCalc := new(MockScheduleCalc)

	ctx := context.Background()
	now := time.Now().UTC()

	// Setup reminder - CRP quota reached
	reminder := createTestReminder("r1", "u1", models.ReminderTypeRecurring)
	reminder.MaxCRP = 3
	reminder.CRPCount = 3
	reminder.NextRecurring = now.Add(1 * time.Minute)
	reminder.NextActionAt = now.Add(-1 * time.Second)

	// Setup mocks
	sysRepo.On("IsWorkerEnabled", ctx).Return(true, nil)
	reminderRepo.On("GetDueReminders", ctx, mock.Anything).Return([]*models.Reminder{reminder}, nil)
	schedCalc.On("CanSendCRP", reminder, mock.Anything).Return(false)
	schedCalc.On("CalculateNextActionAt", reminder, mock.Anything).Return(reminder.NextRecurring)
	reminderRepo.On("UpdateNextActionAt", ctx, "r1", reminder.NextRecurring).Return(nil)
	sysRepo.On("ClearError", ctx).Return(nil)

	w := NewWorker(sysRepo, reminderRepo, userRepo, fcmSender, schedCalc, time.Second)
	w.runOnce(ctx)

	// Verify no FCM sent
	fcmSender.AssertNotCalled(t, "SendNotification")
	// Verify next_action_at updated to next_recurring
	reminderRepo.AssertCalled(t, "UpdateNextActionAt", ctx, "r1", mock.Anything)
}

// Test 7: FCM token error - disable user FCM
func TestFCMTokenError(t *testing.T) {
	sysRepo := new(MockSystemStatusRepo)
	reminderRepo := new(MockReminderRepo)
	userRepo := new(MockUserRepo)
	fcmSender := new(MockFCMSender)
	schedCalc := new(MockScheduleCalc)

	ctx := context.Background()
	now := time.Now().UTC()

	reminder := createTestReminder("r1", "u1", models.ReminderTypeOneTime)
	reminder.NextActionAt = now.Add(-1 * time.Second)

	user := createTestUser("u1", true)

	// Setup mocks - FCM token error
	sysRepo.On("IsWorkerEnabled", ctx).Return(true, nil)
	reminderRepo.On("GetDueReminders", ctx, mock.Anything).Return([]*models.Reminder{reminder}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(user, nil)
	fcmSender.On("SendNotification", "token_u1", "Test Reminder", "Test Description").
		Return(errors.New("UNREGISTERED"))
	userRepo.On("DisableFCM", ctx, "u1").Return(nil)
	sysRepo.On("ClearError", ctx).Return(nil)

	w := NewWorker(sysRepo, reminderRepo, userRepo, fcmSender, schedCalc, time.Second)
	w.runOnce(ctx)

	// Verify user FCM was disabled
	userRepo.AssertCalled(t, "DisableFCM", ctx, "u1")
	// Verify worker continued (didn't disable itself)
	sysRepo.AssertNotCalled(t, "DisableWorker", ctx)
}

// Test 8: System FCM error - disable worker
func TestSystemFCMError(t *testing.T) {
	sysRepo := new(MockSystemStatusRepo)
	reminderRepo := new(MockReminderRepo)
	userRepo := new(MockUserRepo)
	fcmSender := new(MockFCMSender)
	schedCalc := new(MockScheduleCalc)

	ctx := context.Background()
	now := time.Now().UTC()

	reminder := createTestReminder("r1", "u1", models.ReminderTypeOneTime)
	reminder.NextActionAt = now.Add(-1 * time.Second)

	user := createTestUser("u1", true)

	// Setup mocks - System error (deadline exceeded)
	sysRepo.On("IsWorkerEnabled", ctx).Return(true, nil)
	reminderRepo.On("GetDueReminders", ctx, mock.Anything).Return([]*models.Reminder{reminder}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(user, nil)
	fcmSender.On("SendNotification", "token_u1", "Test Reminder", "Test Description").
		Return(context.DeadlineExceeded)
	sysRepo.On("DisableWorker", ctx).Return(nil)

	w := NewWorker(sysRepo, reminderRepo, userRepo, fcmSender, schedCalc, time.Second)
	w.runOnce(ctx)

	// Verify worker was disabled on system error
	sysRepo.AssertCalled(t, "DisableWorker", ctx)
}

// Test 9: User not found error
func TestUserNotFound(t *testing.T) {
	sysRepo := new(MockSystemStatusRepo)
	reminderRepo := new(MockReminderRepo)
	userRepo := new(MockUserRepo)
	fcmSender := new(MockFCMSender)
	schedCalc := new(MockScheduleCalc)

	ctx := context.Background()
	now := time.Now().UTC()

	reminder := createTestReminder("r1", "u1", models.ReminderTypeOneTime)
	reminder.NextActionAt = now.Add(-1 * time.Second)

	// Setup mocks - user not found
	sysRepo.On("IsWorkerEnabled", ctx).Return(true, nil)
	reminderRepo.On("GetDueReminders", ctx, mock.Anything).Return([]*models.Reminder{reminder}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(nil, errors.New("user not found"))
	sysRepo.On("ClearError", ctx).Return(nil)

	w := NewWorker(sysRepo, reminderRepo, userRepo, fcmSender, schedCalc, time.Second)
	w.runOnce(ctx)

	// Verify no FCM sent
	fcmSender.AssertNotCalled(t, "SendNotification")
	// Verify worker continued
	sysRepo.AssertCalled(t, "ClearError", ctx)
}

// Test 10: Snoozed reminder skipped
func TestSnoozeReminderSkipped(t *testing.T) {
	sysRepo := new(MockSystemStatusRepo)
	reminderRepo := new(MockReminderRepo)
	userRepo := new(MockUserRepo)
	fcmSender := new(MockFCMSender)
	schedCalc := new(MockScheduleCalc)

	ctx := context.Background()
	now := time.Now().UTC()

	// Setup snoozed reminder
	reminder := createTestReminder("r1", "u1", models.ReminderTypeOneTime)
	reminder.SnoozeUntil = now.Add(1 * time.Minute)
	reminder.NextActionAt = now.Add(-1 * time.Second)

	// Setup mocks
	sysRepo.On("IsWorkerEnabled", ctx).Return(true, nil)
	reminderRepo.On("GetDueReminders", ctx, mock.Anything).Return([]*models.Reminder{reminder}, nil)
	schedCalc.On("CalculateNextActionAt", reminder, mock.Anything).Return(reminder.SnoozeUntil)
	reminderRepo.On("UpdateNextActionAt", ctx, "r1", reminder.SnoozeUntil).Return(nil)
	sysRepo.On("ClearError", ctx).Return(nil)

	w := NewWorker(sysRepo, reminderRepo, userRepo, fcmSender, schedCalc, time.Second)
	w.runOnce(ctx)

	// Verify no FCM sent
	fcmSender.AssertNotCalled(t, "SendNotification")
	// Verify user repo not called (snoozed reminders skip processing)
	userRepo.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
}

// Test 11: Multiple reminders in single cycle
func TestProcessMultipleReminders(t *testing.T) {
	sysRepo := new(MockSystemStatusRepo)
	reminderRepo := new(MockReminderRepo)
	userRepo := new(MockUserRepo)
	fcmSender := new(MockFCMSender)
	schedCalc := new(MockScheduleCalc)

	ctx := context.Background()
	now := time.Now().UTC()

	// Setup 2 reminders
	reminder1 := createTestReminder("r1", "u1", models.ReminderTypeOneTime)
	reminder1.NextActionAt = now.Add(-1 * time.Second)

	reminder2 := createTestReminder("r2", "u2", models.ReminderTypeOneTime)
	reminder2.NextActionAt = now.Add(-1 * time.Second)

	user1 := createTestUser("u1", true)
	user2 := createTestUser("u2", true)

	// Setup mocks
	sysRepo.On("IsWorkerEnabled", ctx).Return(true, nil)
	reminderRepo.On("GetDueReminders", ctx, mock.Anything).Return([]*models.Reminder{reminder1, reminder2}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(user1, nil)
	userRepo.On("GetByID", ctx, "u2").Return(user2, nil)
	fcmSender.On("SendNotification", "token_u1", "Test Reminder", "Test Description").Return(nil)
	fcmSender.On("SendNotification", "token_u2", "Test Reminder", "Test Description").Return(nil)
	schedCalc.On("CalculateNextActionAt", mock.Anything, mock.Anything).Return(now.Add(20 * time.Second))
	reminderRepo.On("Update", ctx, mock.Anything).Return(nil)
	sysRepo.On("ClearError", ctx).Return(nil)

	w := NewWorker(sysRepo, reminderRepo, userRepo, fcmSender, schedCalc, time.Second)
	w.runOnce(ctx)

	// Verify both reminders processed
	assert.Equal(t, 2, len(fcmSender.Calls))
}

// Test 12: Recurring + crp_until_complete strategy
func TestRecurringCRPUntilCompleteStrategy(t *testing.T) {
	sysRepo := new(MockSystemStatusRepo)
	reminderRepo := new(MockReminderRepo)
	userRepo := new(MockUserRepo)
	fcmSender := new(MockFCMSender)
	schedCalc := new(MockScheduleCalc)

	ctx := context.Background()
	now := time.Now().UTC()

	// Setup recurring reminder with crp_until_complete
	reminder := createTestReminder("r1", "u1", models.ReminderTypeRecurring)
	reminder.RepeatStrategy = models.RepeatStrategyCRPUntilComplete
	reminder.NextRecurring = now.Add(-1 * time.Second)
	reminder.NextActionAt = reminder.NextRecurring

	user := createTestUser("u1", true)
	nextRecurring := now.Add(7 * 24 * time.Hour) // Next week

	// Setup mocks
	sysRepo.On("IsWorkerEnabled", ctx).Return(true, nil)
	reminderRepo.On("GetDueReminders", ctx, mock.Anything).Return([]*models.Reminder{reminder}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(user, nil)
	fcmSender.On("SendNotification", "token_u1", "Test Reminder", "Test Description").Return(nil)
	schedCalc.On("CalculateNextRecurring", reminder, mock.Anything).Return(nextRecurring, nil)
	schedCalc.On("CalculateNextActionAt", mock.Anything, mock.Anything).Return(nextRecurring)
	reminderRepo.On("Update", ctx, mock.MatchedBy(func(r *models.Reminder) bool {
		// For crp_until_complete, NextRecurring should NOT be recalculated on FRP
		return r.ID == "r1" && r.CRPCount == 0
	})).Return(nil)
	sysRepo.On("ClearError", ctx).Return(nil)

	w := NewWorker(sysRepo, reminderRepo, userRepo, fcmSender, schedCalc, time.Second)
	w.runOnce(ctx)

	// Verify FRP processed
	fcmSender.AssertCalled(t, "SendNotification", "token_u1", "Test Reminder", "Test Description")
}
