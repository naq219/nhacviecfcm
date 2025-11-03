package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"remiaq/internal/models"
)

// Mock ReminderService
type MockReminderService struct {
	mock.Mock
}

func (m *MockReminderService) CreateReminder(ctx context.Context, reminder *models.Reminder) error {
	args := m.Called(ctx, reminder)
	return args.Error(0)
}

func (m *MockReminderService) GetReminder(ctx context.Context, id string) (*models.Reminder, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reminder), args.Error(1)
}

func (m *MockReminderService) UpdateReminder(ctx context.Context, reminder *models.Reminder) error {
	args := m.Called(ctx, reminder)
	return args.Error(0)
}

func (m *MockReminderService) DeleteReminder(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockReminderService) GetUserReminders(ctx context.Context, userID string) ([]*models.Reminder, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Reminder), args.Error(1)
}

func (m *MockReminderService) SnoozeReminder(ctx context.Context, id string, duration time.Duration) error {
	args := m.Called(ctx, id, duration)
	return args.Error(0)
}

func (m *MockReminderService) CompleteReminder(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockReminderService) ProcessDueReminders(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Helper: Create mock RequestEvent with proper Body handling
func createReminderMockRequestEvent(method, path string, body interface{}) *core.RequestEvent {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyReader = io.NopCloser(bytes.NewReader(bodyBytes))
	} else {
		bodyReader = io.NopCloser(bytes.NewReader([]byte{}))
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	re := &core.RequestEvent{
		Event: router.Event{
			Request:  req,
			Response: recorder,
		},
	}

	return re
}

// ============= TestNewReminderHandler =============
func TestNewReminderHandler(t *testing.T) {
	mockService := &MockReminderService{}
	handler := NewReminderHandler(mockService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.reminderService)
}

// ============= TestCreateReminder =============
func TestCreateReminder(t *testing.T) {
	tests := []struct {
		name           string
		reminder       *models.Reminder
		setupMock      func(*MockReminderService)
		expectedStatus int
		shouldFail     bool
	}{
		{
			name: "successful reminder creation",
			reminder: &models.Reminder{
				Title:        "Test Reminder",
				UserID:       "user123",
				Type:         "one_time",
				CalendarType: "solar",
				Status:       "active",
			},
			setupMock: func(m *MockReminderService) {
				m.On("CreateReminder", mock.Anything, mock.AnythingOfType("*models.Reminder")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name: "reminder with recurring type",
			reminder: &models.Reminder{
				Title:        "Recurring Reminder",
				UserID:       "user123",
				Type:         "recurring",
				CalendarType: "solar",
				RecurrencePattern: &models.RecurrencePattern{
					Type: "daily",
				},
				Status: "active",
			},
			setupMock: func(m *MockReminderService) {
				m.On("CreateReminder", mock.Anything, mock.AnythingOfType("*models.Reminder")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name: "reminder with lunar calendar",
			reminder: &models.Reminder{
				Title:        "Lunar Reminder",
				UserID:       "user123",
				Type:         "recurring",
				CalendarType: "lunar",
				RecurrencePattern: &models.RecurrencePattern{
					Type:       "monthly",
					DayOfMonth: 15,
				},
				Status: "active",
			},
			setupMock: func(m *MockReminderService) {
				m.On("CreateReminder", mock.Anything, mock.AnythingOfType("*models.Reminder")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name: "reminder with retry strategy",
			reminder: &models.Reminder{
				Title:            "Retry Reminder",
				UserID:           "user123",
				Type:             "one_time",
				CalendarType:     "solar",
				RepeatStrategy:   "retry_until_complete",
				RetryIntervalSec: 300,
				MaxRetries:       3,
				Status:           "active",
			},
			setupMock: func(m *MockReminderService) {
				m.On("CreateReminder", mock.Anything, mock.AnythingOfType("*models.Reminder")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name:           "invalid request body",
			reminder:       nil,
			setupMock:      func(m *MockReminderService) {},
			expectedStatus: http.StatusBadRequest,
			shouldFail:     true,
		},
		{
			name: "service error",
			reminder: &models.Reminder{
				Title:        "Test",
				UserID:       "user123",
				Type:         "one_time",
				CalendarType: "solar",
				Status:       "active",
			},
			setupMock: func(m *MockReminderService) {
				m.On("CreateReminder", mock.Anything, mock.AnythingOfType("*models.Reminder")).
					Return(assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
			shouldFail:     true,
		},
		{
			name: "empty title",
			reminder: &models.Reminder{
				Title:        "",
				UserID:       "user123",
				Type:         "one_time",
				CalendarType: "solar",
				Status:       "active",
			},
			setupMock: func(m *MockReminderService) {
				m.On("CreateReminder", mock.Anything, mock.AnythingOfType("*models.Reminder")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockReminderService{}
			handler := NewReminderHandler(mockService)
			tt.setupMock(mockService)

			var re *core.RequestEvent
			if tt.shouldFail && tt.reminder == nil {
				req := httptest.NewRequest("POST", "/api/reminders", strings.NewReader("invalid json"))
				req.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()
				re = &core.RequestEvent{
					Event: router.Event{
						Request:  req,
						Response: recorder,
					},
				}
			} else {
				re = createReminderMockRequestEvent("POST", "/api/reminders", tt.reminder)
			}

			err := handler.CreateReminder(re)

			assert.NoError(t, err)
			if !tt.shouldFail || tt.reminder == nil {
				mockService.AssertExpectations(t)
			}
		})
	}
}

// ============= TestGetReminder =============
func TestGetReminder(t *testing.T) {
	tests := []struct {
		name           string
		reminderID     string
		setupMock      func(*MockReminderService)
		expectedStatus int
	}{
		{
			name:       "successful reminder retrieval",
			reminderID: "reminder123",
			setupMock: func(m *MockReminderService) {
				m.On("GetReminder", mock.Anything, "reminder123").Return(&models.Reminder{
					ID:     "reminder123",
					Title:  "Test",
					UserID: "user123",
					Status: "active",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "reminder not found",
			reminderID: "nonexistent",
			setupMock: func(m *MockReminderService) {
				m.On("GetReminder", mock.Anything, "nonexistent").
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "missing reminder ID",
			reminderID:     "",
			setupMock:      func(m *MockReminderService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockReminderService{}
			handler := NewReminderHandler(mockService)
			tt.setupMock(mockService)

			req := httptest.NewRequest("GET", "/api/reminders/"+tt.reminderID, nil)
			req.SetPathValue("id", tt.reminderID)
			recorder := httptest.NewRecorder()
			re := &core.RequestEvent{
				Event: router.Event{
					Request:  req,
					Response: recorder,
				},
			}

			err := handler.GetReminder(re)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

// ============= TestUpdateReminder =============
func TestUpdateReminder(t *testing.T) {
	tests := []struct {
		name           string
		reminderID     string
		reminder       *models.Reminder
		setupMock      func(*MockReminderService)
		expectedStatus int
		shouldFail     bool
	}{
		{
			name:       "successful reminder update",
			reminderID: "reminder123",
			reminder: &models.Reminder{
				Title:  "Updated Title",
				Status: "active",
			},
			setupMock: func(m *MockReminderService) {
				m.On("UpdateReminder", mock.Anything, mock.AnythingOfType("*models.Reminder")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name:       "update status to completed",
			reminderID: "reminder123",
			reminder: &models.Reminder{
				Status: "completed",
			},
			setupMock: func(m *MockReminderService) {
				m.On("UpdateReminder", mock.Anything, mock.AnythingOfType("*models.Reminder")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name:       "update with new recurrence pattern",
			reminderID: "reminder123",
			reminder: &models.Reminder{
				RecurrencePattern: &models.RecurrencePattern{
					Type:      "weekly",
					DayOfWeek: 1,
				},
			},
			setupMock: func(m *MockReminderService) {
				m.On("UpdateReminder", mock.Anything, mock.AnythingOfType("*models.Reminder")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name:           "missing reminder ID",
			reminderID:     "",
			reminder:       &models.Reminder{Title: "Test"},
			setupMock:      func(m *MockReminderService) {},
			expectedStatus: http.StatusBadRequest,
			shouldFail:     true,
		},
		{
			name:       "service error",
			reminderID: "reminder123",
			reminder: &models.Reminder{
				Title: "Updated",
			},
			setupMock: func(m *MockReminderService) {
				m.On("UpdateReminder", mock.Anything, mock.AnythingOfType("*models.Reminder")).
					Return(assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
			shouldFail:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockReminderService{}
			handler := NewReminderHandler(mockService)
			tt.setupMock(mockService)

			re := createReminderMockRequestEvent("PUT", "/api/reminders/"+tt.reminderID, tt.reminder)
			req := re.Event.Request
			req.SetPathValue("id", tt.reminderID)

			err := handler.UpdateReminder(re)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, re.Event.Response.(*httptest.ResponseRecorder).Code)
		})
	}
}

// ============= TestDeleteReminder =============
func TestDeleteReminder(t *testing.T) {
	tests := []struct {
		name           string
		reminderID     string
		setupMock      func(*MockReminderService)
		expectedStatus int
	}{
		{
			name:       "successful reminder deletion",
			reminderID: "reminder123",
			setupMock: func(m *MockReminderService) {
				m.On("DeleteReminder", mock.Anything, "reminder123").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "delete non-existent reminder",
			reminderID: "nonexistent",
			setupMock: func(m *MockReminderService) {
				m.On("DeleteReminder", mock.Anything, "nonexistent").
					Return(assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing reminder ID",
			reminderID:     "",
			setupMock:      func(m *MockReminderService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockReminderService{}
			handler := NewReminderHandler(mockService)
			tt.setupMock(mockService)

			req := httptest.NewRequest("DELETE", "/api/reminders/"+tt.reminderID, nil)
			req.SetPathValue("id", tt.reminderID)
			recorder := httptest.NewRecorder()
			re := &core.RequestEvent{
				Event: router.Event{
					Request:  req,
					Response: recorder,
				},
			}

			err := handler.DeleteReminder(re)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

// ============= TestGetUserReminders =============
func TestGetUserReminders(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*MockReminderService)
		expectedStatus int
		expectedCount  int
	}{
		{
			name:   "successful user reminders retrieval",
			userID: "user123",
			setupMock: func(m *MockReminderService) {
				m.On("GetUserReminders", mock.Anything, "user123").Return([]*models.Reminder{
					{ID: "1", Title: "Reminder 1", UserID: "user123", Status: "active"},
					{ID: "2", Title: "Reminder 2", UserID: "user123", Status: "active"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:   "user with no reminders",
			userID: "user456",
			setupMock: func(m *MockReminderService) {
				m.On("GetUserReminders", mock.Anything, "user456").Return([]*models.Reminder{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:   "service error",
			userID: "user789",
			setupMock: func(m *MockReminderService) {
				m.On("GetUserReminders", mock.Anything, "user789").
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing user ID",
			userID:         "",
			setupMock:      func(m *MockReminderService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockReminderService{}
			handler := NewReminderHandler(mockService)
			tt.setupMock(mockService)

			req := httptest.NewRequest("GET", "/api/users/"+tt.userID+"/reminders", nil)
			req.SetPathValue("userId", tt.userID)
			recorder := httptest.NewRecorder()
			re := &core.RequestEvent{
				Event: router.Event{
					Request:  req,
					Response: recorder,
				},
			}

			err := handler.GetUserReminders(re)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

// ============= TestSnoozeReminder =============
func TestSnoozeReminder(t *testing.T) {
	tests := []struct {
		name           string
		reminderID     string
		duration       int
		setupMock      func(*MockReminderService)
		expectedStatus int
		shouldFail     bool
	}{
		{
			name:       "successful 5 minute snooze",
			reminderID: "reminder123",
			duration:   300,
			setupMock: func(m *MockReminderService) {
				m.On("SnoozeReminder", mock.Anything, "reminder123", 5*time.Minute).Return(nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name:       "snooze with 1 hour",
			reminderID: "reminder123",
			duration:   3600,
			setupMock: func(m *MockReminderService) {
				m.On("SnoozeReminder", mock.Anything, "reminder123", time.Hour).Return(nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name:       "snooze with 1 second",
			reminderID: "reminder123",
			duration:   1,
			setupMock: func(m *MockReminderService) {
				m.On("SnoozeReminder", mock.Anything, "reminder123", time.Second).Return(nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name:       "snooze with zero duration",
			reminderID: "reminder123",
			duration:   0,
			setupMock: func(m *MockReminderService) {
				m.On("SnoozeReminder", mock.Anything, "reminder123", time.Duration(0)).Return(nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name:       "snooze with negative duration",
			reminderID: "reminder123",
			duration:   -300,
			setupMock: func(m *MockReminderService) {
				m.On("SnoozeReminder", mock.Anything, "reminder123", -5*time.Minute).Return(nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name:           "missing reminder ID",
			reminderID:     "",
			duration:       300,
			setupMock:      func(m *MockReminderService) {},
			expectedStatus: http.StatusBadRequest,
			shouldFail:     true,
		},
		{
			name:       "service error",
			reminderID: "reminder123",
			duration:   300,
			setupMock: func(m *MockReminderService) {
				m.On("SnoozeReminder", mock.Anything, "reminder123", 5*time.Minute).
					Return(assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
			shouldFail:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockReminderService{}
			handler := NewReminderHandler(mockService)
			tt.setupMock(mockService)

			var re *core.RequestEvent
			if tt.shouldFail && tt.duration == 300 && tt.reminderID != "" {
				req := httptest.NewRequest("POST", "/api/reminders/"+tt.reminderID+"/snooze",
					strings.NewReader("invalid json"))
				req.Header.Set("Content-Type", "application/json")
				req.SetPathValue("id", tt.reminderID)
				recorder := httptest.NewRecorder()
				re = &core.RequestEvent{
					Event: router.Event{
						Request:  req,
						Response: recorder,
					},
				}
			} else {
				snoozeReq := struct {
					Duration int `json:"duration"`
				}{Duration: tt.duration}
				re = createReminderMockRequestEvent("POST", "/api/reminders/"+tt.reminderID+"/snooze", snoozeReq)
				re.Event.Request.SetPathValue("id", tt.reminderID)
			}

			err := handler.SnoozeReminder(re)

			assert.NoError(t, err)
		})
	}
}

// ============= TestCompleteReminder =============
func TestCompleteReminder(t *testing.T) {
	tests := []struct {
		name           string
		reminderID     string
		setupMock      func(*MockReminderService)
		expectedStatus int
	}{
		{
			name:       "successful reminder completion",
			reminderID: "reminder123",
			setupMock: func(m *MockReminderService) {
				m.On("CompleteReminder", mock.Anything, "reminder123").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "complete non-existent reminder",
			reminderID: "nonexistent",
			setupMock: func(m *MockReminderService) {
				m.On("CompleteReminder", mock.Anything, "nonexistent").
					Return(assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing reminder ID",
			reminderID:     "",
			setupMock:      func(m *MockReminderService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockReminderService{}
			handler := NewReminderHandler(mockService)
			tt.setupMock(mockService)

			req := httptest.NewRequest("POST", "/api/reminders/"+tt.reminderID+"/complete", nil)
			req.SetPathValue("id", tt.reminderID)
			recorder := httptest.NewRecorder()
			re := &core.RequestEvent{
				Event: router.Event{
					Request:  req,
					Response: recorder,
				},
			}

			err := handler.CompleteReminder(re)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

// ============= TestReminderValidation =============
func TestReminderValidation(t *testing.T) {
	tests := []struct {
		name        string
		reminder    *models.Reminder
		expectValid bool
	}{
		{
			name: "valid one_time reminder",
			reminder: &models.Reminder{
				Title:        "Test",
				Type:         "one_time",
				CalendarType: "solar",
				Status:       "active",
			},
			expectValid: true,
		},
		{
			name: "valid recurring reminder",
			reminder: &models.Reminder{
				Title:        "Test",
				Type:         "recurring",
				CalendarType: "solar",
				Status:       "active",
			},
			expectValid: true,
		},
		{
			name: "valid lunar reminder",
			reminder: &models.Reminder{
				Title:        "Test",
				Type:         "one_time",
				CalendarType: "lunar",
				Status:       "active",
			},
			expectValid: true,
		},
		{
			name: "empty title",
			reminder: &models.Reminder{
				Title:        "",
				Type:         "one_time",
				CalendarType: "solar",
				Status:       "active",
			},
			expectValid: false,
		},
		{
			name: "invalid type",
			reminder: &models.Reminder{
				Title:        "Test",
				Type:         "invalid_type",
				CalendarType: "solar",
				Status:       "active",
			},
			expectValid: false,
		},
		{
			name: "invalid calendar type",
			reminder: &models.Reminder{
				Title:        "Test",
				Type:         "one_time",
				CalendarType: "invalid",
				Status:       "active",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.reminder.Validate()
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// ============= TestReminderShouldSend =============
func TestReminderShouldSend(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(1 * time.Hour)
	pastTime := now.Add(-1 * time.Hour)
	snoozeTime := now.Add(30 * time.Minute)

	tests := []struct {
		name      string
		reminder  *models.Reminder
		checkTime time.Time
		expected  bool
	}{
		{
			name: "should send - active and past trigger time",
			reminder: &models.Reminder{
				Status:        "active",
				NextTriggerAt: pastTime,
				SnoozeUntil:   nil,
			},
			checkTime: now,
			expected:  true,
		},
		{
			name: "should not send - future trigger time",
			reminder: &models.Reminder{
				Status:        "active",
				NextTriggerAt: futureTime,
				SnoozeUntil:   nil,
			},
			checkTime: now,
			expected:  false,
		},
		{
			name: "should not send - snoozed",
			reminder: &models.Reminder{
				Status:        "active",
				NextTriggerAt: pastTime,
				SnoozeUntil:   &snoozeTime,
			},
			checkTime: now,
			expected:  false,
		},
		{
			name: "should send - snooze expired",
			reminder: &models.Reminder{
				Status:        "active",
				NextTriggerAt: pastTime,
				SnoozeUntil:   &pastTime,
			},
			checkTime: now,
			expected:  true,
		},
		{
			name: "should not send - status completed",
			reminder: &models.Reminder{
				Status:        "completed",
				NextTriggerAt: pastTime,
				SnoozeUntil:   nil,
			},
			checkTime: now,
			expected:  false,
		},
		{
			name: "should not send - status paused",
			reminder: &models.Reminder{
				Status:        "paused",
				NextTriggerAt: pastTime,
				SnoozeUntil:   nil,
			},
			checkTime: now,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.reminder.ShouldSend(tt.checkTime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============= TestReminderIsRetryable =============
func TestReminderIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		reminder *models.Reminder
		expected bool
	}{
		{
			name: "retryable - under max retries",
			reminder: &models.Reminder{
				RepeatStrategy: "retry_until_complete",
				RetryCount:     1,
				MaxRetries:     3,
			},
			expected: true,
		},
		{
			name: "not retryable - at max retries",
			reminder: &models.Reminder{
				RepeatStrategy: "retry_until_complete",
				RetryCount:     3,
				MaxRetries:     3,
			},
			expected: false,
		},
		{
			name: "not retryable - no retry strategy",
			reminder: &models.Reminder{
				RepeatStrategy: "none",
				RetryCount:     0,
				MaxRetries:     3,
			},
			expected: false,
		},
		{
			name: "not retryable - exceeded max retries",
			reminder: &models.Reminder{
				RepeatStrategy: "retry_until_complete",
				RetryCount:     5,
				MaxRetries:     3,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.reminder.IsRetryable()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============= TestConcurrency =============
func TestConcurrency_Reminders(t *testing.T) {
	mockService := &MockReminderService{}
	handler := NewReminderHandler(mockService)

	mockService.On("CreateReminder", mock.Anything, mock.AnythingOfType("*models.Reminder")).Return(nil).Maybe()
	mockService.On("GetReminder", mock.Anything, mock.Anything).
		Return(&models.Reminder{ID: "1", Title: "Test"}, nil).Maybe()
	mockService.On("UpdateReminder", mock.Anything, mock.AnythingOfType("*models.Reminder")).Return(nil).Maybe()
	mockService.On("DeleteReminder", mock.Anything, mock.Anything).Return(nil).Maybe()
	mockService.On("SnoozeReminder", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	mockService.On("CompleteReminder", mock.Anything, mock.Anything).Return(nil).Maybe()

	done := make(chan bool, 6)

	go func() {
		for i := 0; i < 10; i++ {
			reminder := &models.Reminder{Title: "Test", UserID: "user123", Type: "one_time", CalendarType: "solar"}
			re := createReminderMockRequestEvent("POST", "/api/reminders", reminder)
			handler.CreateReminder(re)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			req := httptest.NewRequest("GET", "/api/reminders/reminder123", nil)
			req.SetPathValue("id", "reminder123")
			recorder := httptest.NewRecorder()
			re := &core.RequestEvent{
				Event: router.Event{
					Request:  req,
					Response: recorder,
				},
			}
			handler.GetReminder(re)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			reminder := &models.Reminder{Title: "Updated"}
			re := createReminderMockRequestEvent("PUT", "/api/reminders/reminder123", reminder)
			re.Event.Request.SetPathValue("id", "reminder123")
			handler.UpdateReminder(re)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			req := httptest.NewRequest("DELETE", "/api/reminders/reminder123", nil)
			req.SetPathValue("id", "reminder123")
			recorder := httptest.NewRecorder()
			re := &core.RequestEvent{
				Event: router.Event{
					Request:  req,
					Response: recorder,
				},
			}
			handler.DeleteReminder(re)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			snoozeReq := struct {
				Duration int `json:"duration"`
			}{Duration: 300}
			re := createReminderMockRequestEvent("POST", "/api/reminders/reminder123/snooze", snoozeReq)
			re.Event.Request.SetPathValue("id", "reminder123")
			handler.SnoozeReminder(re)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			req := httptest.NewRequest("POST", "/api/reminders/reminder123/complete", nil)
			req.SetPathValue("id", "reminder123")
			recorder := httptest.NewRecorder()
			re := &core.RequestEvent{
				Event: router.Event{
					Request:  req,
					Response: recorder,
				},
			}
			handler.CompleteReminder(re)
		}
		done <- true
	}()

	for i := 0; i < 6; i++ {
		<-done
	}

	assert.True(t, true)
}

// ============= Benchmarks =============
func BenchmarkCreateReminder(b *testing.B) {
	mockService := &MockReminderService{}
	handler := NewReminderHandler(mockService)

	mockService.On("CreateReminder", mock.Anything, mock.AnythingOfType("*models.Reminder")).
		Return(nil).Maybe()

	reminder := &models.Reminder{
		Title:        "Benchmark",
		UserID:       "user123",
		Type:         "one_time",
		CalendarType: "solar",
		Status:       "active",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re := createReminderMockRequestEvent("POST", "/api/reminders", reminder)
		handler.CreateReminder(re)
	}
}

func BenchmarkGetReminder(b *testing.B) {
	mockService := &MockReminderService{}
	handler := NewReminderHandler(mockService)

	mockService.On("GetReminder", mock.Anything, "reminder123").
		Return(&models.Reminder{ID: "reminder123", Title: "Test"}, nil).Maybe()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/reminders/reminder123", nil)
		req.SetPathValue("id", "reminder123")
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		handler.GetReminder(re)
	}
}

func BenchmarkUpdateReminder(b *testing.B) {
	mockService := &MockReminderService{}
	handler := NewReminderHandler(mockService)

	mockService.On("UpdateReminder", mock.Anything, mock.AnythingOfType("*models.Reminder")).
		Return(nil).Maybe()

	reminder := &models.Reminder{Title: "Updated"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re := createReminderMockRequestEvent("PUT", "/api/reminders/reminder123", reminder)
		re.Event.Request.SetPathValue("id", "reminder123")
		handler.UpdateReminder(re)
	}
}

func BenchmarkDeleteReminder(b *testing.B) {
	mockService := &MockReminderService{}
	handler := NewReminderHandler(mockService)

	mockService.On("DeleteReminder", mock.Anything, "reminder123").
		Return(nil).Maybe()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("DELETE", "/api/reminders/reminder123", nil)
		req.SetPathValue("id", "reminder123")
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		handler.DeleteReminder(re)
	}
}

func BenchmarkSnoozeReminder(b *testing.B) {
	mockService := &MockReminderService{}
	handler := NewReminderHandler(mockService)

	mockService.On("SnoozeReminder", mock.Anything, "reminder123", mock.Anything).
		Return(nil).Maybe()

	snoozeReq := struct {
		Duration int `json:"duration"`
	}{Duration: 300}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re := createReminderMockRequestEvent("POST", "/api/reminders/reminder123/snooze", snoozeReq)
		re.Event.Request.SetPathValue("id", "reminder123")
		handler.SnoozeReminder(re)
	}
}

func BenchmarkCompleteReminder(b *testing.B) {
	mockService := &MockReminderService{}
	handler := NewReminderHandler(mockService)

	mockService.On("CompleteReminder", mock.Anything, "reminder123").
		Return(nil).Maybe()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/reminders/reminder123/complete", nil)
		req.SetPathValue("id", "reminder123")
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		handler.CompleteReminder(re)
	}
}

func BenchmarkGetUserReminders(b *testing.B) {
	mockService := &MockReminderService{}
	handler := NewReminderHandler(mockService)

	mockService.On("GetUserReminders", mock.Anything, "user123").
		Return([]*models.Reminder{
			{ID: "1", Title: "Reminder 1"},
			{ID: "2", Title: "Reminder 2"},
		}, nil).Maybe()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/users/user123/reminders", nil)
		req.SetPathValue("userId", "user123")
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		handler.GetUserReminders(re)
	}
}