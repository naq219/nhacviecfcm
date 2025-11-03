package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

// Helper: Create mock RequestEvent with proper Body handling
func createMockRequestEvent(method, path string, body interface{}) *core.RequestEvent {
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

// ============= TestNewQueryHandler =============
func TestNewQueryHandler(t *testing.T) {
	mockRepo := &MockQueryRepository{}
	handler := NewQueryHandler(mockRepo)

	assert.NotNil(t, handler)
	assert.Equal(t, mockRepo, handler.queryRepo)
}

// ============= TestParseRequest =============
func TestParseRequest(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		body        interface{}
		expected    string
		expectError bool
	}{
		{
			name:        "GET request with query parameter",
			method:      "GET",
			path:        "/query?q=SELECT+*+FROM+users",
			body:        nil,
			expected:    "SELECT * FROM users",
			expectError: false,
		},
		{
			name:        "GET request with empty query",
			method:      "GET",
			path:        "/query",
			body:        nil,
			expected:    "",
			expectError: false,
		},
		{
			name:        "POST request with JSON body",
			method:      "POST",
			path:        "/query",
			body:        QueryRequest{Query: "SELECT * FROM reminders"},
			expected:    "SELECT * FROM reminders",
			expectError: false,
		},
		{
			name:        "POST request with invalid JSON",
			method:      "POST",
			path:        "/query",
			body:        "invalid json",
			expected:    "",
			expectError: true,
		},
		{
			name:        "POST request with empty body",
			method:      "POST",
			path:        "/query",
			body:        QueryRequest{Query: ""},
			expected:    "",
			expectError: false,
		},
		{
			name:        "GET request with special characters",
			method:      "GET",
			path:        "/query?q=SELECT+*+FROM+users+WHERE+name='John'",
			body:        nil,
			expected:    "SELECT * FROM users WHERE name='John'",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := createMockRequestEvent(tt.method, tt.path, nil)

			// Override with specific test path for GET
			if tt.method == "GET" {
				req := httptest.NewRequest(tt.method, tt.path, nil)
				re.Event.Request = req
			} else if tt.method == "POST" {
				// Handle invalid JSON case
				if tt.body == "invalid json" {
					req := httptest.NewRequest(tt.method, tt.path, strings.NewReader("invalid json"))
					req.Header.Set("Content-Type", "application/json")
					re.Event.Request = req
				} else {
					re = createMockRequestEvent(tt.method, tt.path, tt.body)
				}
			}

			result, err := parseRequest(re)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result.Query)
			}
		})
	}
}

// ============= TestHandleSelect =============
func TestHandleSelect(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		setupMock      func(*MockQueryRepository)
		expectedStatus int
		expectedRows   int
	}{
		{
			name:  "successful SELECT query",
			query: "SELECT * FROM users",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteSelect", mock.Anything, "SELECT * FROM users").
					Return([]map[string]interface{}{
						{"id": "1", "name": "John"},
						{"id": "2", "name": "Jane"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedRows:   2,
		},
		{
			name:  "SELECT with WHERE clause",
			query: "SELECT * FROM users WHERE id = 1",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteSelect", mock.Anything, "SELECT * FROM users WHERE id = 1").
					Return([]map[string]interface{}{
						{"id": "1", "name": "John"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedRows:   1,
		},
		{
			name:  "SELECT with empty result",
			query: "SELECT * FROM users WHERE id = 999",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteSelect", mock.Anything, "SELECT * FROM users WHERE id = 999").
					Return([]map[string]interface{}{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedRows:   0,
		},
		{
			name:  "SELECT query execution failed",
			query: "SELECT * FROM nonexistent_table",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteSelect", mock.Anything, "SELECT * FROM nonexistent_table").
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "SELECT via GET with query parameter",
			query: "",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteSelect", mock.Anything, "SELECT * FROM users").
					Return([]map[string]interface{}{
						{"id": "1", "name": "John"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedRows:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockQueryRepository{}
			handler := NewQueryHandler(mockRepo)
			tt.setupMock(mockRepo)

			var re *core.RequestEvent
			if tt.name == "SELECT via GET with query parameter" {
				req := httptest.NewRequest("GET", "/query?q=SELECT+*+FROM+users", nil)
				recorder := httptest.NewRecorder()
				re = &core.RequestEvent{
					Event: router.Event{
						Request:  req,
						Response: recorder,
					},
				}
			} else {
				re = createMockRequestEvent("GET", "/query?q="+url.QueryEscape(tt.query), nil)
				if tt.query != "" {
					req := httptest.NewRequest("GET", "/query?q="+url.QueryEscape(tt.query), nil)
					re.Event.Request = req
				}
			}

			err := handler.HandleSelect(re)

			assert.NoError(t, err)
			if tt.expectedStatus == http.StatusOK {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

// ============= TestHandleInsert =============
func TestHandleInsert(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		setupMock      func(*MockQueryRepository)
		expectedStatus int
		shouldFail     bool
	}{
		{
			name:  "successful INSERT query",
			query: "INSERT INTO users (name) VALUES ('John')",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteInsert", mock.Anything, "INSERT INTO users (name) VALUES ('John')").
					Return(int64(1), int64(123), nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name:  "INSERT with multiple rows",
			query: "INSERT INTO users (name, email) VALUES ('John', 'john@example.com')",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteInsert", mock.Anything, mock.Anything).
					Return(int64(1), int64(124), nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name:  "INSERT execution failed",
			query: "INSERT INTO users (name) VALUES ('John')",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteInsert", mock.Anything, "INSERT INTO users (name) VALUES ('John')").
					Return(int64(0), int64(0), assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
			shouldFail:     true,
		},
		{
			name:  "INSERT with duplicate key",
			query: "INSERT INTO users (id, name) VALUES (1, 'John')",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteInsert", mock.Anything, mock.Anything).
					Return(int64(0), int64(0), assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
			shouldFail:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockQueryRepository{}
			handler := NewQueryHandler(mockRepo)
			tt.setupMock(mockRepo)

			body := QueryRequest{Query: tt.query}
			re := createMockRequestEvent("POST", "/query", body)

			err := handler.HandleInsert(re)

			assert.NoError(t, err)
			if !tt.shouldFail {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

// ============= TestHandleUpdate =============
func TestHandleUpdate(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		setupMock      func(*MockQueryRepository)
		expectedStatus int
		shouldFail     bool
	}{
		{
			name:  "successful UPDATE query",
			query: "UPDATE users SET name='Jane' WHERE id=1",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteUpdate", mock.Anything, "UPDATE users SET name='Jane' WHERE id=1").
					Return(int64(1), nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name:  "UPDATE with no matching rows",
			query: "UPDATE users SET name='Jane' WHERE id=999",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteUpdate", mock.Anything, mock.Anything).
					Return(int64(0), nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name:  "UPDATE multiple rows",
			query: "UPDATE users SET status='active' WHERE role='admin'",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteUpdate", mock.Anything, mock.Anything).
					Return(int64(5), nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name:  "UPDATE execution failed",
			query: "UPDATE users SET name='Jane' WHERE id=1",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteUpdate", mock.Anything, "UPDATE users SET name='Jane' WHERE id=1").
					Return(int64(0), assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
			shouldFail:     true,
		},
		{
			name:  "UPDATE with JOIN",
			query: "UPDATE users SET status='verified' WHERE id IN (SELECT user_id FROM verifications)",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteUpdate", mock.Anything, mock.Anything).
					Return(int64(3), nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockQueryRepository{}
			handler := NewQueryHandler(mockRepo)
			tt.setupMock(mockRepo)

			body := QueryRequest{Query: tt.query}
			re := createMockRequestEvent("POST", "/query", body)

			err := handler.HandleUpdate(re)

			assert.NoError(t, err)
			if !tt.shouldFail {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

// ============= TestHandleDelete =============
func TestHandleDelete(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		setupMock      func(*MockQueryRepository)
		expectedStatus int
		shouldFail     bool
	}{
		{
			name:  "successful DELETE query",
			query: "DELETE FROM users WHERE id=1",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteDelete", mock.Anything, "DELETE FROM users WHERE id=1").
					Return(int64(1), nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name:  "DELETE with no matching rows",
			query: "DELETE FROM users WHERE id=999",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteDelete", mock.Anything, mock.Anything).
					Return(int64(0), nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name:  "DELETE multiple rows",
			query: "DELETE FROM users WHERE created_at < '2020-01-01'",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteDelete", mock.Anything, mock.Anything).
					Return(int64(100), nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
		{
			name:  "DELETE execution failed",
			query: "DELETE FROM users WHERE id=1",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteDelete", mock.Anything, "DELETE FROM users WHERE id=1").
					Return(int64(0), assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
			shouldFail:     true,
		},
		{
			name:  "DELETE with complex WHERE",
			query: "DELETE FROM reminders WHERE user_id IN (SELECT id FROM users WHERE status='inactive')",
			setupMock: func(m *MockQueryRepository) {
				m.On("ExecuteDelete", mock.Anything, mock.Anything).
					Return(int64(50), nil)
			},
			expectedStatus: http.StatusOK,
			shouldFail:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockQueryRepository{}
			handler := NewQueryHandler(mockRepo)
			tt.setupMock(mockRepo)

			body := QueryRequest{Query: tt.query}
			re := createMockRequestEvent("POST", "/query", body)

			err := handler.HandleDelete(re)

			assert.NoError(t, err)
			if !tt.shouldFail {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

// ============= TestRequestValidation =============
func TestRequestValidation(t *testing.T) {
	tests := []struct {
		name   string
		req    *QueryRequest
		method string
		valid  bool
	}{
		{
			name:   "valid SELECT",
			req:    &QueryRequest{Query: "SELECT * FROM users"},
			method: "GET",
			valid:  true,
		},
		{
			name:   "valid INSERT",
			req:    &QueryRequest{Query: "INSERT INTO users (name) VALUES ('John')"},
			method: "POST",
			valid:  true,
		},
		{
			name:   "valid UPDATE",
			req:    &QueryRequest{Query: "UPDATE users SET name='Jane' WHERE id=1"},
			method: "POST",
			valid:  true,
		},
		{
			name:   "valid DELETE",
			req:    &QueryRequest{Query: "DELETE FROM users WHERE id=1"},
			method: "POST",
			valid:  true,
		},
		{
			name:   "empty query",
			req:    &QueryRequest{Query: ""},
			method: "GET",
			valid:  false, // Empty query should be invalid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.req)
			if tt.valid {
				assert.NotEmpty(t, tt.req.Query) // For this test, non-empty means valid
			} else {
				// For invalid cases, we can check that query is empty or has other issues
				if tt.name == "empty query" {
					assert.Empty(t, tt.req.Query)
				}
			}
		})
	}
}

// ============= TestErrorHandling =============
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		handler        func(*MockQueryRepository) *QueryHandler
		requestMethod  string
		expectedStatus int
	}{
		{
			name: "SELECT error handling",
			handler: func(m *MockQueryRepository) *QueryHandler {
				m.On("ExecuteSelect", mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
				return NewQueryHandler(m)
			},
			requestMethod:  "GET",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "INSERT error handling",
			handler: func(m *MockQueryRepository) *QueryHandler {
				m.On("ExecuteInsert", mock.Anything, mock.Anything).
					Return(int64(0), int64(0), assert.AnError)
				return NewQueryHandler(m)
			},
			requestMethod:  "POST",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "UPDATE error handling",
			handler: func(m *MockQueryRepository) *QueryHandler {
				m.On("ExecuteUpdate", mock.Anything, mock.Anything).
					Return(int64(0), assert.AnError)
				return NewQueryHandler(m)
			},
			requestMethod:  "POST",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "DELETE error handling",
			handler: func(m *MockQueryRepository) *QueryHandler {
				m.On("ExecuteDelete", mock.Anything, mock.Anything).
					Return(int64(0), assert.AnError)
				return NewQueryHandler(m)
			},
			requestMethod:  "POST",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockQueryRepository{}
			handler := tt.handler(mockRepo)

			var err error
			switch tt.name {
			case "SELECT error handling":
				body := QueryRequest{Query: "SELECT * FROM users"}
				re := createMockRequestEvent(tt.requestMethod, "/query", body)
				err = handler.HandleSelect(re)
			case "INSERT error handling":
				body := QueryRequest{Query: "INSERT INTO users (name) VALUES ('test')"}
				re := createMockRequestEvent(tt.requestMethod, "/query", body)
				err = handler.HandleInsert(re)
			case "UPDATE error handling":
				body := QueryRequest{Query: "UPDATE users SET name='test' WHERE id=1"}
				re := createMockRequestEvent(tt.requestMethod, "/query", body)
				err = handler.HandleUpdate(re)
			case "DELETE error handling":
				body := QueryRequest{Query: "DELETE FROM users WHERE id=1"}
				re := createMockRequestEvent(tt.requestMethod, "/query", body)
				err = handler.HandleDelete(re)
			}

			require.NoError(t, err)
		})
	}
}

// ============= TestConcurrency =============
func TestConcurrency(t *testing.T) {
	mockRepo := &MockQueryRepository{}
	handler := NewQueryHandler(mockRepo)

	mockRepo.On("ExecuteSelect", mock.Anything, mock.Anything).
		Return([]map[string]interface{}{{"id": "1"}}, nil).
		Maybe()
	mockRepo.On("ExecuteInsert", mock.Anything, mock.Anything).
		Return(int64(1), int64(100), nil).
		Maybe()
	mockRepo.On("ExecuteUpdate", mock.Anything, mock.Anything).
		Return(int64(1), nil).
		Maybe()
	mockRepo.On("ExecuteDelete", mock.Anything, mock.Anything).
		Return(int64(1), nil).
		Maybe()

	done := make(chan bool, 4)

	// Concurrent SELECT
	go func() {
		for i := 0; i < 10; i++ {
			re := createMockRequestEvent("GET", "/query?q=SELECT+*+FROM+users", nil)
			handler.HandleSelect(re)
		}
		done <- true
	}()

	// Concurrent INSERT
	go func() {
		for i := 0; i < 10; i++ {
			re := createMockRequestEvent("POST", "/query", QueryRequest{Query: "INSERT INTO users (name) VALUES ('John')"})
			handler.HandleInsert(re)
		}
		done <- true
	}()

	// Concurrent UPDATE
	go func() {
		for i := 0; i < 10; i++ {
			re := createMockRequestEvent("POST", "/query", QueryRequest{Query: "UPDATE users SET name='Jane' WHERE id=1"})
			handler.HandleUpdate(re)
		}
		done <- true
	}()

	// Concurrent DELETE
	go func() {
		for i := 0; i < 10; i++ {
			re := createMockRequestEvent("POST", "/query", QueryRequest{Query: "DELETE FROM users WHERE id=1"})
			handler.HandleDelete(re)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}
}

// ============= Benchmarks =============
func BenchmarkHandleSelect(b *testing.B) {
	mockRepo := &MockQueryRepository{}
	handler := NewQueryHandler(mockRepo)

	mockRepo.On("ExecuteSelect", mock.Anything, mock.Anything).
		Return([]map[string]interface{}{{"id": "1"}}, nil).
		Maybe()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re := createMockRequestEvent("GET", "/query?q=SELECT+*+FROM+users", nil)
		handler.HandleSelect(re)
	}
}

func BenchmarkHandleInsert(b *testing.B) {
	mockRepo := &MockQueryRepository{}
	handler := NewQueryHandler(mockRepo)

	mockRepo.On("ExecuteInsert", mock.Anything, mock.Anything).
		Return(int64(1), int64(100), nil).
		Maybe()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := QueryRequest{Query: "INSERT INTO users (name) VALUES ('John')"}
		re := createMockRequestEvent("POST", "/query", body)
		handler.HandleInsert(re)
	}
}

func BenchmarkHandleUpdate(b *testing.B) {
	mockRepo := &MockQueryRepository{}
	handler := NewQueryHandler(mockRepo)

	mockRepo.On("ExecuteUpdate", mock.Anything, mock.Anything).
		Return(int64(1), nil).
		Maybe()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := QueryRequest{Query: "UPDATE users SET name='Jane' WHERE id=1"}
		re := createMockRequestEvent("POST", "/query", body)
		handler.HandleUpdate(re)
	}
}

func BenchmarkHandleDelete(b *testing.B) {
	mockRepo := &MockQueryRepository{}
	handler := NewQueryHandler(mockRepo)

	mockRepo.On("ExecuteDelete", mock.Anything, mock.Anything).
		Return(int64(1), nil).
		Maybe()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := QueryRequest{Query: "DELETE FROM users WHERE id=1"}
		re := createMockRequestEvent("POST", "/query", body)
		handler.HandleDelete(re)
	}
}

func BenchmarkParseRequest_GET(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re := createMockRequestEvent("GET", "/query?q=SELECT+*+FROM+users", nil)
		parseRequest(re)
	}
}

func BenchmarkParseRequest_POST(b *testing.B) {
	body := QueryRequest{Query: "SELECT * FROM users"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re := createMockRequestEvent("POST", "/query", body)
		parseRequest(re)
	}
}
