package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"remiaq/internal/models"
)

// Mock SystemStatusRepository
type MockSystemStatusRepository struct {
	mock.Mock
}

func (m *MockSystemStatusRepository) Get(ctx context.Context) (*models.SystemStatus, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SystemStatus), args.Error(1)
}

func (m *MockSystemStatusRepository) EnableWorker(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSystemStatusRepository) DisableWorker(ctx context.Context, reason string) error {
	args := m.Called(ctx, reason)
	return args.Error(0)
}

func (m *MockSystemStatusRepository) UpdateError(ctx context.Context, errorMsg string) error {
	args := m.Called(ctx, errorMsg)
	return args.Error(0)
}

func (m *MockSystemStatusRepository) ClearError(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSystemStatusRepository) IsWorkerEnabled(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

// Helper function to create mock RequestEvent for system status handler
func createSystemStatusRequestEvent(method, path string, body interface{}) *core.RequestEvent {
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

func TestNewSystemStatusHandler(t *testing.T) {
	mockRepo := &MockSystemStatusRepository{}
	handler := NewSystemStatusHandler(mockRepo)
	
	assert.NotNil(t, handler)
	assert.Equal(t, mockRepo, handler.repo)
}

func TestSystemStatusHandler_GetSystemStatus(t *testing.T) {
	t.Run("successful system status retrieval", func(t *testing.T) {
		mockRepo := &MockSystemStatusRepository{}
		handler := NewSystemStatusHandler(mockRepo)
		
		expectedStatus := &models.SystemStatus{
			ID:            1,
			WorkerEnabled: true,
		}
		
		mockRepo.On("Get", mock.Anything).Return(expectedStatus, nil)
		
		re := createSystemStatusRequestEvent("GET", "/api/system_status", nil)
		
		err := handler.GetSystemStatus(re)
		
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("repository error", func(t *testing.T) {
		mockRepo := &MockSystemStatusRepository{}
		handler := NewSystemStatusHandler(mockRepo)
		
		mockRepo.On("Get", mock.Anything).Return(nil, assert.AnError)
		
		re := createSystemStatusRequestEvent("GET", "/api/system_status", nil)
		
		err := handler.GetSystemStatus(re)
		
		assert.NoError(t, err) // Handler returns error via response
		mockRepo.AssertExpectations(t)
	})
}

func TestSystemStatusHandler_PutSystemStatus(t *testing.T) {
	t.Run("enable worker successfully", func(t *testing.T) {
		mockRepo := &MockSystemStatusRepository{}
		handler := NewSystemStatusHandler(mockRepo)
		
		requestBody := map[string]interface{}{
			"worker_enabled": true,
		}
		
		expectedStatus := &models.SystemStatus{
			ID:            1,
			WorkerEnabled: true,
		}
		
		mockRepo.On("EnableWorker", mock.Anything).Return(nil)
		mockRepo.On("ClearError", mock.Anything).Return(nil)
		mockRepo.On("Get", mock.Anything).Return(expectedStatus, nil)
		
		re := createSystemStatusRequestEvent("PUT", "/api/system_status", requestBody)
		
		err := handler.PutSystemStatus(re)
		
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("disable worker successfully", func(t *testing.T) {
		mockRepo := &MockSystemStatusRepository{}
		handler := NewSystemStatusHandler(mockRepo)
		
		requestBody := map[string]interface{}{
			"worker_enabled": false,
		}
		
		expectedStatus := &models.SystemStatus{
			ID:            1,
			WorkerEnabled: false,
		}
		
		mockRepo.On("DisableWorker", mock.Anything, "manually disabled").Return(nil)
		mockRepo.On("Get", mock.Anything).Return(expectedStatus, nil)
		
		re := createSystemStatusRequestEvent("PUT", "/api/system_status", requestBody)
		
		err := handler.PutSystemStatus(re)
		
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("disable worker with custom error message", func(t *testing.T) {
		mockRepo := &MockSystemStatusRepository{}
		handler := NewSystemStatusHandler(mockRepo)
		
		requestBody := map[string]interface{}{
			"worker_enabled": false,
			"last_error":     "Custom error message",
		}
		
		expectedStatus := &models.SystemStatus{
			ID:            1,
			WorkerEnabled: false,
		}
		
		mockRepo.On("DisableWorker", mock.Anything, "Custom error message").Return(nil)
		mockRepo.On("Get", mock.Anything).Return(expectedStatus, nil)
		
		re := createSystemStatusRequestEvent("PUT", "/api/system_status", requestBody)
		
		err := handler.PutSystemStatus(re)
		
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("update error message only", func(t *testing.T) {
		mockRepo := &MockSystemStatusRepository{}
		handler := NewSystemStatusHandler(mockRepo)
		
		requestBody := map[string]interface{}{
			"last_error": "New error message",
		}
		
		expectedStatus := &models.SystemStatus{
			ID:            1,
			WorkerEnabled: true,
		}
		
		mockRepo.On("UpdateError", mock.Anything, "New error message").Return(nil)
		mockRepo.On("Get", mock.Anything).Return(expectedStatus, nil)
		
		re := createSystemStatusRequestEvent("PUT", "/api/system_status", requestBody)
		
		err := handler.PutSystemStatus(re)
		
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("invalid request body", func(t *testing.T) {
		mockRepo := &MockSystemStatusRepository{}
		handler := NewSystemStatusHandler(mockRepo)
		
		req := httptest.NewRequest("PUT", "/api/system_status", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		err := handler.PutSystemStatus(re)
		
		assert.NoError(t, err) // Handler returns error via response
		assert.Equal(t, 400, recorder.Code)
	})
	
	t.Run("no fields to update", func(t *testing.T) {
		mockRepo := &MockSystemStatusRepository{}
		handler := NewSystemStatusHandler(mockRepo)
		
		requestBody := map[string]interface{}{}
		
		re := createSystemStatusRequestEvent("PUT", "/api/system_status", requestBody)
		
		err := handler.PutSystemStatus(re)
		
		assert.NoError(t, err) // Handler returns error via response
	})
	
	t.Run("enable worker with error message", func(t *testing.T) {
		mockRepo := &MockSystemStatusRepository{}
		handler := NewSystemStatusHandler(mockRepo)
		
		requestBody := map[string]interface{}{
			"worker_enabled": true,
			"last_error":     "Some error",
		}
		
		expectedStatus := &models.SystemStatus{
			ID:            1,
			WorkerEnabled: true,
		}
		
		mockRepo.On("EnableWorker", mock.Anything).Return(nil)
		mockRepo.On("UpdateError", mock.Anything, "Some error").Return(nil)
		mockRepo.On("Get", mock.Anything).Return(expectedStatus, nil)
		
		re := createSystemStatusRequestEvent("PUT", "/api/system_status", requestBody)
		
		err := handler.PutSystemStatus(re)
		
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

// Benchmark tests
func BenchmarkSystemStatusHandler_GetSystemStatus(b *testing.B) {
	mockRepo := &MockSystemStatusRepository{}
	handler := NewSystemStatusHandler(mockRepo)
	
	expectedStatus := &models.SystemStatus{
		ID:            1,
		WorkerEnabled: true,
	}
	
	mockRepo.On("Get", mock.Anything).Return(expectedStatus, nil).Maybe()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re := createSystemStatusRequestEvent("GET", "/api/system_status", nil)
		handler.GetSystemStatus(re)
	}
}

func BenchmarkSystemStatusHandler_PutSystemStatus(b *testing.B) {
	mockRepo := &MockSystemStatusRepository{}
	handler := NewSystemStatusHandler(mockRepo)
	
	requestBody := map[string]interface{}{
		"worker_enabled": true,
	}
	
	expectedStatus := &models.SystemStatus{
		ID:            1,
		WorkerEnabled: true,
	}
	
	mockRepo.On("EnableWorker", mock.Anything).Return(nil).Maybe()
	mockRepo.On("ClearError", mock.Anything).Return(nil).Maybe()
	mockRepo.On("Get", mock.Anything).Return(expectedStatus, nil).Maybe()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re := createSystemStatusRequestEvent("PUT", "/api/system_status", requestBody)
		handler.PutSystemStatus(re)
	}
}