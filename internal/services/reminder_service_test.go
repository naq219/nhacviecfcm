package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"remiaq/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockReminderRepository struct {
	mock.Mock
}

func (m *MockReminderRepository) Create(ctx context.Context, reminder *models.Reminder) error {
	args := m.Called(ctx, reminder)
	return args.Error(0)
}

func (m *MockReminderRepository) GetByID(ctx context.Context, id string) (*models.Reminder, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Reminder), args.Error(1)
}

func (m *MockReminderRepository) Update(ctx context.Context, reminder *models.Reminder) error {
	args := m.Called(ctx, reminder)
	return args.Error(0)
}

func (m *MockReminderRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockReminderRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Reminder, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.Reminder), args.Error(1)
}

func (m *MockReminderRepository) GetDueReminders(ctx context.Context, now time.Time) ([]*models.Reminder, error) {
	args := m.Called(ctx, now)
	return args.Get(0).([]*models.Reminder), args.Error(1)
}

func (m *MockReminderRepository) UpdateSnooze(ctx context.Context, id string, snoozeUntil *time.Time) error {
	args := m.Called(ctx, id, snoozeUntil)
	return args.Error(0)
}

func (m *MockReminderRepository) MarkCompleted(ctx context.Context, id string, completedAt time.Time) error {
	args := m.Called(ctx, id, completedAt)
	return args.Error(0)
}

func (m *MockReminderRepository) UpdateLastSent(ctx context.Context, id string, lastSent time.Time) error {
	args := m.Called(ctx, id, lastSent)
	return args.Error(0)
}

func (m *MockReminderRepository) IncrementRetryCount(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockReminderRepository) UpdateNextTrigger(ctx context.Context, id string, nextTrigger time.Time) error {
	args := m.Called(ctx, id, nextTrigger)
	return args.Error(0)
}

func (m *MockReminderRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) DisableFCM(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) EnableFCM(ctx context.Context, userID string, token string) error {
	args := m.Called(ctx, userID, token)
	return args.Error(0)
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateFCMToken(ctx context.Context, userID, token string) error {
	args := m.Called(ctx, userID, token)
	return args.Error(0)
}

func (m *MockUserRepository) GetActiveUsers(ctx context.Context) ([]*models.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.User), args.Error(1)
}

// Test helper functions
func createTestReminder() *models.Reminder {
	now := time.Now()
	return &models.Reminder{
		ID:            "test-id",
		UserID:        "user-1",
		Title:         "Test Reminder",
		Description:   "Test Description",
		Type:          models.ReminderTypeOneTime,
		CalendarType:  models.CalendarTypeSolar,
		Status:        models.ReminderStatusActive,
		NextTriggerAt: now.Add(time.Hour),
		Created:       now,
		Updated:       now,
	}
}

func createTestUser() *models.User {
	return &models.User{
		ID:          "user-1",
		Email:       "test@example.com",
		FCMToken:    "test-token",
		IsFCMActive: true,
	}
}

func TestNewReminderService(t *testing.T) {
	reminderRepo := &MockReminderRepository{}
	userRepo := &MockUserRepository{}
	fcmService := &FCMService{}
	schedCalculator := NewScheduleCalculator(NewLunarCalendar())

	service := NewReminderService(reminderRepo, userRepo, fcmService, schedCalculator)

	assert.NotNil(t, service)
	assert.Equal(t, reminderRepo, service.reminderRepo)
	assert.Equal(t, userRepo, service.userRepo)
	assert.Equal(t, fcmService, service.fcmService)
	assert.Equal(t, schedCalculator, service.schedCalculator)
}

func TestReminderService_CreateReminder(t *testing.T) {
	t.Run("should create reminder successfully", func(t *testing.T) {
		reminderRepo := &MockReminderRepository{}
		userRepo := &MockUserRepository{}
		schedCalculator := NewScheduleCalculator(NewLunarCalendar())
		service := NewReminderService(reminderRepo, userRepo, nil, schedCalculator)

		reminder := createTestReminder()
		reminder.ID = "" // Test ID generation
		reminder.NextTriggerAt = time.Time{} // Test next trigger calculation

		reminderRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Reminder")).Return(nil)

		err := service.CreateReminder(context.Background(), reminder)

		assert.NoError(t, err)
		assert.NotEmpty(t, reminder.ID)
		assert.Equal(t, models.ReminderStatusActive, reminder.Status)
		assert.False(t, reminder.NextTriggerAt.IsZero()) // Check that NextTriggerAt was set
		reminderRepo.AssertExpectations(t)
	})

	t.Run("should return error for invalid reminder", func(t *testing.T) {
		service := NewReminderService(&MockReminderRepository{}, &MockUserRepository{}, nil, NewScheduleCalculator(NewLunarCalendar()))

		reminder := &models.Reminder{} // Invalid reminder (missing required fields)

		err := service.CreateReminder(context.Background(), reminder)

		assert.Error(t, err)
	})
}

func TestReminderService_GetReminder(t *testing.T) {
	t.Run("should get reminder successfully", func(t *testing.T) {
		reminderRepo := &MockReminderRepository{}
		service := NewReminderService(reminderRepo, &MockUserRepository{}, nil, NewScheduleCalculator(NewLunarCalendar()))

		expectedReminder := createTestReminder()
		reminderRepo.On("GetByID", mock.Anything, "test-id").Return(expectedReminder, nil)

		result, err := service.GetReminder(context.Background(), "test-id")

		assert.NoError(t, err)
		assert.Equal(t, expectedReminder, result)
		reminderRepo.AssertExpectations(t)
	})

	t.Run("should return error when reminder not found", func(t *testing.T) {
		reminderRepo := &MockReminderRepository{}
		service := NewReminderService(reminderRepo, &MockUserRepository{}, nil, NewScheduleCalculator(NewLunarCalendar()))

		reminderRepo.On("GetByID", mock.Anything, "non-existent").Return((*models.Reminder)(nil), errors.New("not found"))

		result, err := service.GetReminder(context.Background(), "non-existent")

		assert.Error(t, err)
		assert.Nil(t, result)
		reminderRepo.AssertExpectations(t)
	})
}

func TestReminderService_CompleteReminder(t *testing.T) {
	t.Run("should complete one-time reminder", func(t *testing.T) {
		reminderRepo := &MockReminderRepository{}
		service := NewReminderService(reminderRepo, &MockUserRepository{}, nil, NewScheduleCalculator(NewLunarCalendar()))

		reminder := createTestReminder()
		reminder.Type = models.ReminderTypeOneTime

		reminderRepo.On("GetByID", mock.Anything, "test-id").Return(reminder, nil)
		reminderRepo.On("MarkCompleted", mock.Anything, "test-id", mock.AnythingOfType("time.Time")).Return(nil)

		err := service.CompleteReminder(context.Background(), "test-id")

		assert.NoError(t, err)
		reminderRepo.AssertExpectations(t)
	})

	t.Run("should handle recurring reminder", func(t *testing.T) {
		reminderRepo := &MockReminderRepository{}
		service := NewReminderService(reminderRepo, &MockUserRepository{}, nil, NewScheduleCalculator(NewLunarCalendar()))

		reminder := createTestReminder()
		reminder.Type = models.ReminderTypeRecurring

		reminderRepo.On("GetByID", mock.Anything, "test-id").Return(reminder, nil)
		reminderRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Reminder")).Return(nil)

		err := service.CompleteReminder(context.Background(), "test-id")

		assert.NoError(t, err)
		assert.NotNil(t, reminder.LastCompletedAt)
		reminderRepo.AssertExpectations(t)
	})
}

func TestReminderService_SnoozeReminder(t *testing.T) {
	t.Run("should snooze reminder successfully", func(t *testing.T) {
		reminderRepo := &MockReminderRepository{}
		service := NewReminderService(reminderRepo, &MockUserRepository{}, nil, NewScheduleCalculator(NewLunarCalendar()))

		duration := 30 * time.Minute
		reminderRepo.On("UpdateSnooze", mock.Anything, "test-id", mock.AnythingOfType("*time.Time")).Return(nil)

		err := service.SnoozeReminder(context.Background(), "test-id", duration)

		assert.NoError(t, err)
		reminderRepo.AssertExpectations(t)
	})
}

func TestReminderService_ProcessDueReminders(t *testing.T) {
	t.Run("should process due reminders successfully", func(t *testing.T) {
		reminderRepo := &MockReminderRepository{}
		userRepo := &MockUserRepository{}
		_ = NewMockFCMService() // fcmService
		service := NewReminderService(reminderRepo, userRepo, nil, NewScheduleCalculator(NewLunarCalendar()))

		_ = time.Now() // now
		reminder := createTestReminder()
		user := createTestUser()

		reminderRepo.On("GetDueReminders", mock.Anything, mock.AnythingOfType("time.Time")).Return([]*models.Reminder{reminder}, nil)
		userRepo.On("GetByID", mock.Anything, "user-1").Return(user, nil)
		// Only expect MarkCompleted for one-time reminder (no FCM service configured)
		reminderRepo.On("MarkCompleted", mock.Anything, "test-id", mock.AnythingOfType("time.Time")).Return(nil)

		err := service.ProcessDueReminders(context.Background())

		assert.NoError(t, err)
		reminderRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})

	t.Run("should handle FCM errors gracefully", func(t *testing.T) {
		reminderRepo := &MockReminderRepository{}
		userRepo := &MockUserRepository{}
		service := NewReminderService(reminderRepo, userRepo, nil, NewScheduleCalculator(NewLunarCalendar()))

		now := time.Now()
		reminder := createTestReminder()
		user := createTestUser()
		user.FCMToken = "" // Invalid token

		reminderRepo.On("GetDueReminders", mock.Anything, now).Return([]*models.Reminder{reminder}, nil)
		userRepo.On("GetByID", mock.Anything, "user-1").Return(user, nil)
		// No other expectations since FCM is not active

		err := service.ProcessDueReminders(context.Background())

		// Should not return error for individual reminder failures
		assert.NoError(t, err)
		reminderRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})
}

func TestIsTokenInvalidError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"UNREGISTERED error", errors.New("UNREGISTERED"), true},
		{"INVALID_ARGUMENT error", errors.New("INVALID_ARGUMENT"), true},
		{"NOT_FOUND error", errors.New("NOT_FOUND"), true},
		{"other error", errors.New("network error"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isTokenInvalidError(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkReminderService_CreateReminder(b *testing.B) {
	reminderRepo := &MockReminderRepository{}
	schedCalculator := NewScheduleCalculator(NewLunarCalendar())
	service := NewReminderService(reminderRepo, &MockUserRepository{}, nil, schedCalculator)

	reminderRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	// Note: schedCalculator is concrete type, no mocking needed

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reminder := createTestReminder()
		reminder.ID = ""
		reminder.NextTriggerAt = time.Time{}
		_ = service.CreateReminder(context.Background(), reminder)
	}
}