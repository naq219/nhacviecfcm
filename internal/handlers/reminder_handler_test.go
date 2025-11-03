package handlers

import (
	"bytes"
	"context"
	"encoding/json"
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

// Helper function to create mock RequestEvent for reminder handler
func createReminderRequestEvent(method, path string, body interface{}) *core.RequestEvent {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(bodyBytes)
	} else {
		bodyReader = bytes.NewReader([]byte{})
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

func TestNewReminderHandler(t *testing.T) {
	mockService := &MockReminderService{}
	handler := NewReminderHandler(mockService)
	
	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.reminderService)
}

func TestReminderHandler_CreateReminder(t *testing.T) {
	t.Run("successful reminder creation", func(t *testing.T) {
		mockService := &MockReminderService{}
		handler := NewReminderHandler(mockService)
		
		reminder := &models.Reminder{
			Title:       "Test Reminder",
			Description: "Test Description",
			UserID:      "user123",
		}
		
		mockService.On("CreateReminder", mock.Anything, mock.AnythingOfType("*models.Reminder")).Return(nil)
		
		re := createReminderRequestEvent("POST", "/api/reminders", reminder)
		
		err := handler.CreateReminder(re)
		
		assert.NoError(t, err)
		mockService.AssertExpectations(t)
	})
	
	t.Run("invalid request body", func(t *testing.T) {
		mockService := &MockReminderService{}
		handler := NewReminderHandler(mockService)
		
		req := httptest.NewRequest("POST", "/api/reminders", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		err := handler.CreateReminder(re)
		
		assert.NoError(t, err) // Handler returns error via response
		assert.Equal(t, 400, recorder.Code)
	})
	
	t.Run("service error", func(t *testing.T) {
		mockService := &MockReminderService{}
		handler := NewReminderHandler(mockService)
		
		reminder := &models.Reminder{
			Title:  "Test Reminder",
			UserID: "user123",
		}
		
		mockService.On("CreateReminder", mock.Anything, mock.AnythingOfType("*models.Reminder")).Return(assert.AnError)
		
		re := createReminderRequestEvent("POST", "/api/reminders", reminder)
		
		err := handler.CreateReminder(re)
		
		assert.NoError(t, err) // Handler returns error via response
		mockService.AssertExpectations(t)
	})
}

func TestReminderHandler_GetReminder(t *testing.T) {
	t.Run("successful reminder retrieval", func(t *testing.T) {
		mockService := &MockReminderService{}
		handler := NewReminderHandler(mockService)
		
		expectedReminder := &models.Reminder{
			ID:     "reminder123",
			Title:  "Test Reminder",
			UserID: "user123",
		}
		
		mockService.On("GetReminder", mock.Anything, "reminder123").Return(expectedReminder, nil)
		
		req := httptest.NewRequest("GET", "/api/reminders/reminder123", nil)
		req.SetPathValue("id", "reminder123")
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		err := handler.GetReminder(re)
		
		assert.NoError(t, err)
		mockService.AssertExpectations(t)
	})
	
	t.Run("missing reminder ID", func(t *testing.T) {
		mockService := &MockReminderService{}
		handler := NewReminderHandler(mockService)
		
		req := httptest.NewRequest("GET", "/api/reminders/", nil)
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		err := handler.GetReminder(re)
		
		assert.NoError(t, err) // Handler returns error via response
		assert.Equal(t, 400, recorder.Code)
	})
	
	t.Run("reminder not found", func(t *testing.T) {
		mockService := &MockReminderService{}
		handler := NewReminderHandler(mockService)
		
		mockService.On("GetReminder", mock.Anything, "nonexistent").Return(nil, assert.AnError)
		
		req := httptest.NewRequest("GET", "/api/reminders/nonexistent", nil)
		req.SetPathValue("id", "nonexistent")
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		err := handler.GetReminder(re)
		
		assert.NoError(t, err) // Handler returns error via response
		assert.Equal(t, 404, recorder.Code)
		mockService.AssertExpectations(t)
	})
}

func TestReminderHandler_UpdateReminder(t *testing.T) {
	t.Run("successful reminder update", func(t *testing.T) {
		mockService := &MockReminderService{}
		handler := NewReminderHandler(mockService)
		
		reminder := &models.Reminder{
			Title:       "Updated Reminder",
			Description: "Updated Description",
		}
		
		mockService.On("UpdateReminder", mock.Anything, mock.AnythingOfType("*models.Reminder")).Return(nil)
		
		req := httptest.NewRequest("PUT", "/api/reminders/reminder123", nil)
		req.SetPathValue("id", "reminder123")
		
		bodyBytes, _ := json.Marshal(reminder)
		req.Body = &testReadCloser{bytes.NewReader(bodyBytes)}
		req.Header.Set("Content-Type", "application/json")
		
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		err := handler.UpdateReminder(re)
		
		assert.NoError(t, err)
		mockService.AssertExpectations(t)
	})
	
	t.Run("missing reminder ID", func(t *testing.T) {
		mockService := &MockReminderService{}
		handler := NewReminderHandler(mockService)
		
		req := httptest.NewRequest("PUT", "/api/reminders/", nil)
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		err := handler.UpdateReminder(re)
		
		assert.NoError(t, err) // Handler returns error via response
		assert.Equal(t, 400, recorder.Code)
	})
}

func TestReminderHandler_DeleteReminder(t *testing.T) {
	t.Run("successful reminder deletion", func(t *testing.T) {
		mockService := &MockReminderService{}
		handler := NewReminderHandler(mockService)
		
		mockService.On("DeleteReminder", mock.Anything, "reminder123").Return(nil)
		
		req := httptest.NewRequest("DELETE", "/api/reminders/reminder123", nil)
		req.SetPathValue("id", "reminder123")
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		err := handler.DeleteReminder(re)
		
		assert.NoError(t, err)
		mockService.AssertExpectations(t)
	})
	
	t.Run("missing reminder ID", func(t *testing.T) {
		mockService := &MockReminderService{}
		handler := NewReminderHandler(mockService)
		
		req := httptest.NewRequest("DELETE", "/api/reminders/", nil)
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		err := handler.DeleteReminder(re)
		
		assert.NoError(t, err) // Handler returns error via response
		assert.Equal(t, 400, recorder.Code)
	})
}

func TestReminderHandler_GetUserReminders(t *testing.T) {
	t.Run("successful user reminders retrieval", func(t *testing.T) {
		mockService := &MockReminderService{}
		handler := NewReminderHandler(mockService)
		
		expectedReminders := []*models.Reminder{
			{ID: "1", Title: "Reminder 1", UserID: "user123"},
			{ID: "2", Title: "Reminder 2", UserID: "user123"},
		}
		
		mockService.On("GetUserReminders", mock.Anything, "user123").Return(expectedReminders, nil)
		
		req := httptest.NewRequest("GET", "/api/users/user123/reminders", nil)
		req.SetPathValue("userId", "user123")
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		err := handler.GetUserReminders(re)
		
		assert.NoError(t, err)
		mockService.AssertExpectations(t)
	})
	
	t.Run("missing user ID", func(t *testing.T) {
		mockService := &MockReminderService{}
		handler := NewReminderHandler(mockService)
		
		req := httptest.NewRequest("GET", "/api/users//reminders", nil)
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		err := handler.GetUserReminders(re)
		
		assert.NoError(t, err) // Handler returns error via response
		assert.Equal(t, 400, recorder.Code)
	})
}

func TestReminderHandler_SnoozeReminder(t *testing.T) {
	t.Run("successful reminder snooze", func(t *testing.T) {
		mockService := &MockReminderService{}
		handler := NewReminderHandler(mockService)
		
		snoozeReq := struct {
			Duration int `json:"duration"`
		}{Duration: 300} // 5 minutes
		
		mockService.On("SnoozeReminder", mock.Anything, "reminder123", 5*time.Minute).Return(nil)
		
		req := httptest.NewRequest("POST", "/api/reminders/reminder123/snooze", nil)
		req.SetPathValue("id", "reminder123")
		
		bodyBytes, _ := json.Marshal(snoozeReq)
		req.Body = &testReadCloser{bytes.NewReader(bodyBytes)}
		req.Header.Set("Content-Type", "application/json")
		
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		err := handler.SnoozeReminder(re)
		
		assert.NoError(t, err)
		mockService.AssertExpectations(t)
	})
	
	t.Run("missing reminder ID", func(t *testing.T) {
		mockService := &MockReminderService{}
		handler := NewReminderHandler(mockService)
		
		req := httptest.NewRequest("POST", "/api/reminders//snooze", nil)
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		err := handler.SnoozeReminder(re)
		
		assert.NoError(t, err) // Handler returns error via response
		assert.Equal(t, 400, recorder.Code)
	})
}

func TestReminderHandler_CompleteReminder(t *testing.T) {
	t.Run("successful reminder completion", func(t *testing.T) {
		mockService := &MockReminderService{}
		handler := NewReminderHandler(mockService)
		
		mockService.On("CompleteReminder", mock.Anything, "reminder123").Return(nil)
		
		req := httptest.NewRequest("POST", "/api/reminders/reminder123/complete", nil)
		req.SetPathValue("id", "reminder123")
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		err := handler.CompleteReminder(re)
		
		assert.NoError(t, err)
		mockService.AssertExpectations(t)
	})
	
	t.Run("missing reminder ID", func(t *testing.T) {
		mockService := &MockReminderService{}
		handler := NewReminderHandler(mockService)
		
		req := httptest.NewRequest("POST", "/api/reminders//complete", nil)
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		err := handler.CompleteReminder(re)
		
		assert.NoError(t, err) // Handler returns error via response
		assert.Equal(t, 400, recorder.Code)
	})
}

// Helper type for testing request body
type testReadCloser struct {
	*bytes.Reader
}

func (t *testReadCloser) Close() error {
	return nil
}

// Benchmark tests
func BenchmarkReminderHandler_CreateReminder(b *testing.B) {
	mockService := &MockReminderService{}
	handler := NewReminderHandler(mockService)
	
	reminder := &models.Reminder{
		Title:  "Benchmark Reminder",
		UserID: "user123",
	}
	
	mockService.On("CreateReminder", mock.Anything, mock.AnythingOfType("*models.Reminder")).Return(nil).Maybe()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re := createReminderRequestEvent("POST", "/api/reminders", reminder)
		handler.CreateReminder(re)
	}
}

func BenchmarkReminderHandler_GetReminder(b *testing.B) {
	mockService := &MockReminderService{}
	handler := NewReminderHandler(mockService)
	
	expectedReminder := &models.Reminder{
		ID:     "reminder123",
		Title:  "Test Reminder",
		UserID: "user123",
	}
	
	mockService.On("GetReminder", mock.Anything, "reminder123").Return(expectedReminder, nil).Maybe()
	
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