package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock QueryRepository
type MockQueryRepository struct {
	mock.Mock
}

func (m *MockQueryRepository) ExecuteSelect(ctx context.Context, query string) ([]map[string]interface{}, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockQueryRepository) ExecuteInsert(ctx context.Context, query string) (int64, int64, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockQueryRepository) ExecuteUpdate(ctx context.Context, query string) (int64, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQueryRepository) ExecuteDelete(ctx context.Context, query string) (int64, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(int64), args.Error(1)
}

// Helper function to create mock RequestEvent
func createMockRequestEvent(method, path string, body interface{}) *core.RequestEvent {
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
	
	// Create RequestEvent with embedded router.Event
	re := &core.RequestEvent{
		Event: router.Event{
			Request:  req,
			Response: recorder,
		},
	}
	
	return re
}

func TestNewQueryHandler(t *testing.T) {
	mockRepo := &MockQueryRepository{}
	handler := NewQueryHandler(mockRepo)
	
	assert.NotNil(t, handler)
	assert.Equal(t, mockRepo, handler.queryRepo)
}

func TestParseRequest(t *testing.T) {
	t.Run("GET request with query parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/query?q=SELECT+*+FROM+users", nil)
		re := &core.RequestEvent{
			Event: router.Event{
				Request: req,
			},
		}
		
		result, err := parseRequest(re)
		
		assert.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users", result.Query)
	})
	
	t.Run("POST request with JSON body", func(t *testing.T) {
		body := QueryRequest{Query: "SELECT * FROM reminders"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/query", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		re := &core.RequestEvent{
			Event: router.Event{
				Request: req,
			},
		}
		
		result, err := parseRequest(re)
		
		assert.NoError(t, err)
		assert.Equal(t, "SELECT * FROM reminders", result.Query)
	})
	
	t.Run("POST request with invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/query", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		re := &core.RequestEvent{
			Event: router.Event{
				Request: req,
			},
		}
		
		_, err := parseRequest(re)
		
		assert.Error(t, err)
	})
	
	t.Run("GET request without query parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/query", nil)
		re := &core.RequestEvent{
			Event: router.Event{
				Request: req,
			},
		}
		
		result, err := parseRequest(re)
		
		assert.NoError(t, err)
		assert.Equal(t, "", result.Query)
	})
}

func TestQueryHandler_HandleSelect(t *testing.T) {
	t.Run("successful SELECT query", func(t *testing.T) {
		mockRepo := &MockQueryRepository{}
		handler := NewQueryHandler(mockRepo)
		
		expectedResult := []map[string]interface{}{
			{"id": "1", "name": "John"},
			{"id": "2", "name": "Jane"},
		}
		
		mockRepo.On("ExecuteSelect", mock.Anything, "SELECT * FROM users").Return(expectedResult, nil)
		
		// Test GET request
		req := httptest.NewRequest("GET", "/query?q=SELECT+*+FROM+users", nil)
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		err := handler.HandleSelect(re)
		
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("invalid request format", func(t *testing.T) {
		mockRepo := &MockQueryRepository{}
		handler := NewQueryHandler(mockRepo)
		
		req := httptest.NewRequest("POST", "/query", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		err := handler.HandleSelect(re)
		
		assert.NoError(t, err) // Handler returns error via response, not Go error
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})
	
	t.Run("query execution failed", func(t *testing.T) {
		mockRepo := &MockQueryRepository{}
		handler := NewQueryHandler(mockRepo)
		
		mockRepo.On("ExecuteSelect", mock.Anything, "SELECT * FROM users").Return(nil, assert.AnError)
		
		req := httptest.NewRequest("GET", "/query?q=SELECT+*+FROM+users", nil)
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		err := handler.HandleSelect(re)
		
		assert.NoError(t, err) // Handler returns error via response
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestQueryHandler_HandleInsert(t *testing.T) {
	t.Run("successful INSERT query", func(t *testing.T) {
		mockRepo := &MockQueryRepository{}
		handler := NewQueryHandler(mockRepo)
		
		mockRepo.On("ExecuteInsert", mock.Anything, "INSERT INTO users (name) VALUES ('John')").Return(int64(1), int64(123), nil)
		
		body := QueryRequest{Query: "INSERT INTO users (name) VALUES ('John')"}
		re := createMockRequestEvent("POST", "/query", body)
		
		err := handler.HandleInsert(re)
		
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("insert execution failed", func(t *testing.T) {
		mockRepo := &MockQueryRepository{}
		handler := NewQueryHandler(mockRepo)
		
		mockRepo.On("ExecuteInsert", mock.Anything, "INSERT INTO users (name) VALUES ('John')").Return(int64(0), int64(0), assert.AnError)
		
		body := QueryRequest{Query: "INSERT INTO users (name) VALUES ('John')"}
		re := createMockRequestEvent("POST", "/query", body)
		
		err := handler.HandleInsert(re)
		
		assert.NoError(t, err) // Handler returns error via response
		mockRepo.AssertExpectations(t)
	})
}

func TestQueryHandler_HandleUpdate(t *testing.T) {
	t.Run("successful UPDATE query", func(t *testing.T) {
		mockRepo := &MockQueryRepository{}
		handler := NewQueryHandler(mockRepo)
		
		mockRepo.On("ExecuteUpdate", mock.Anything, "UPDATE users SET name='Jane' WHERE id=1").Return(int64(1), nil)
		
		body := QueryRequest{Query: "UPDATE users SET name='Jane' WHERE id=1"}
		re := createMockRequestEvent("POST", "/query", body)
		
		err := handler.HandleUpdate(re)
		
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("update execution failed", func(t *testing.T) {
		mockRepo := &MockQueryRepository{}
		handler := NewQueryHandler(mockRepo)
		
		mockRepo.On("ExecuteUpdate", mock.Anything, "UPDATE users SET name='Jane' WHERE id=1").Return(int64(0), assert.AnError)
		
		body := QueryRequest{Query: "UPDATE users SET name='Jane' WHERE id=1"}
		re := createMockRequestEvent("POST", "/query", body)
		
		err := handler.HandleUpdate(re)
		
		assert.NoError(t, err) // Handler returns error via response
		mockRepo.AssertExpectations(t)
	})
}

func TestQueryHandler_HandleDelete(t *testing.T) {
	t.Run("successful DELETE query", func(t *testing.T) {
		mockRepo := &MockQueryRepository{}
		handler := NewQueryHandler(mockRepo)
		
		mockRepo.On("ExecuteDelete", mock.Anything, "DELETE FROM users WHERE id=1").Return(int64(1), nil)
		
		body := QueryRequest{Query: "DELETE FROM users WHERE id=1"}
		re := createMockRequestEvent("POST", "/query", body)
		
		err := handler.HandleDelete(re)
		
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("delete execution failed", func(t *testing.T) {
		mockRepo := &MockQueryRepository{}
		handler := NewQueryHandler(mockRepo)
		
		mockRepo.On("ExecuteDelete", mock.Anything, "DELETE FROM users WHERE id=1").Return(int64(0), assert.AnError)
		
		body := QueryRequest{Query: "DELETE FROM users WHERE id=1"}
		re := createMockRequestEvent("POST", "/query", body)
		
		err := handler.HandleDelete(re)
		
		assert.NoError(t, err) // Handler returns error via response
		mockRepo.AssertExpectations(t)
	})
}

// Benchmark tests
func BenchmarkQueryHandler_HandleSelect(b *testing.B) {
	mockRepo := &MockQueryRepository{}
	handler := NewQueryHandler(mockRepo)
	
	expectedResult := []map[string]interface{}{
		{"id": "1", "name": "John"},
	}
	
	mockRepo.On("ExecuteSelect", mock.Anything, mock.Anything).Return(expectedResult, nil).Maybe()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/query?q=SELECT+*+FROM+users", nil)
		recorder := httptest.NewRecorder()
		re := &core.RequestEvent{
			Event: router.Event{
				Request:  req,
				Response: recorder,
			},
		}
		
		handler.HandleSelect(re)
	}
}

func BenchmarkParseRequest_GET(b *testing.B) {
	req := httptest.NewRequest("GET", "/query?q=SELECT+*+FROM+users", nil)
	re := &core.RequestEvent{
		Event: router.Event{
			Request: req,
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseRequest(re)
		// Reset URL for next iteration
		req.URL, _ = url.Parse("/query?q=SELECT+*+FROM+users")
	}
}

func BenchmarkParseRequest_POST(b *testing.B) {
	body := QueryRequest{Query: "SELECT * FROM users"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/query", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		re := &core.RequestEvent{
			Event: router.Event{
				Request: req,
			},
		}
		
		parseRequest(re)
	}
}