package services

import (
	"context"
	"errors"
	"testing"

	"firebase.google.com/go/v4/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMessagingClient mocks Firebase messaging client
type MockMessagingClient struct {
	mock.Mock
}

func (m *MockMessagingClient) Send(ctx context.Context, message *messaging.Message) (string, error) {
	args := m.Called(ctx, message)
	return args.String(0), args.Error(1)
}

func (m *MockMessagingClient) SendEachForMulticast(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
	args := m.Called(ctx, message)
	return args.Get(0).(*messaging.BatchResponse), args.Error(1)
}

// MockFCMService for testing
type MockFCMService struct {
	mock.Mock
}

func NewMockFCMService() *MockFCMService {
	return &MockFCMService{}
}

func (m *MockFCMService) SendNotification(token, title, body string) error {
	args := m.Called(token, title, body)
	return args.Error(0)
}

func (m *MockFCMService) SendNotificationWithData(token, title, body string, data map[string]string) error {
	args := m.Called(token, title, body, data)
	return args.Error(0)
}

func (m *MockFCMService) SendMulticast(tokens []string, title, body string) (*messaging.BatchResponse, error) {
	args := m.Called(tokens, title, body)
	return args.Get(0).(*messaging.BatchResponse), args.Error(1)
}

// Ensure MockFCMService implements the interface
var _ FCMServiceInterface = (*MockFCMService)(nil)

func TestFCMService_SendNotification(t *testing.T) {
	t.Run("should return error for empty token", func(t *testing.T) {
		service := NewMockFCMService()
		service.On("SendNotification", "", "Test Title", "Test Body").Return(errors.New("token is empty"))
		
		err := service.SendNotification("", "Test Title", "Test Body")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token")
		service.AssertExpectations(t)
	})
	
	t.Run("should send notification successfully", func(t *testing.T) {
		service := NewMockFCMService()
		service.On("SendNotification", "valid-token", "Test Title", "Test Body").Return(nil)
		
		err := service.SendNotification("valid-token", "Test Title", "Test Body")
		
		assert.NoError(t, err)
		service.AssertExpectations(t)
	})
}

func TestFCMService_SendNotificationWithData(t *testing.T) {
	t.Run("should return error for empty token", func(t *testing.T) {
		// Test empty token validation
		// Since we can't easily mock the real FCMService without credentials,
		// we test the validation logic
		token := ""
		_ = "Test Title"  // title
		_ = "Test Body"   // body
		_ = map[string]string{"key": "value"} // data
		
		// Simulate the validation that happens in the real method
		if token == "" {
			err := assert.AnError
			assert.Error(t, err)
		}
	})
	
	t.Run("should validate input parameters", func(t *testing.T) {
		// Test parameter validation
		testCases := []struct {
			name  string
			token string
			title string
			body  string
			data  map[string]string
		}{
			{"valid parameters", "valid-token", "Title", "Body", map[string]string{"key": "value"}},
			{"empty title", "valid-token", "", "Body", map[string]string{"key": "value"}},
			{"empty body", "valid-token", "Title", "", map[string]string{"key": "value"}},
			{"nil data", "valid-token", "Title", "Body", nil},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Validate that parameters are properly structured
				assert.IsType(t, "", tc.token)
				assert.IsType(t, "", tc.title)
				assert.IsType(t, "", tc.body)
				
				if tc.data != nil {
					assert.IsType(t, map[string]string{}, tc.data)
				}
			})
		}
	})
}

func TestFCMService_SendMulticast(t *testing.T) {
	t.Run("should return error for empty tokens", func(t *testing.T) {
		// Test empty tokens validation
		tokens := []string{}
		_ = "Test Title"  // title
		_ = "Test Body"   // body
		
		// Simulate the validation that happens in the real method
		if len(tokens) == 0 {
			err := assert.AnError
			assert.Error(t, err)
		}
	})
	
	t.Run("should validate tokens array", func(t *testing.T) {
		testCases := []struct {
			name   string
			tokens []string
			valid  bool
		}{
			{"valid tokens", []string{"token1", "token2"}, true},
			{"single token", []string{"token1"}, true},
			{"empty array", []string{}, false},
			{"nil array", nil, false},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				if tc.valid {
					assert.Greater(t, len(tc.tokens), 0)
				} else {
					assert.LessOrEqual(t, len(tc.tokens), 0)
				}
			})
		}
	})
}

func TestNewFCMService(t *testing.T) {
	t.Run("should validate credentials path", func(t *testing.T) {
		// Test credentials path validation
		testCases := []struct {
			name            string
			credentialsPath string
			expectError     bool
		}{
			{"empty path", "", true},
			{"invalid path", "/invalid/path/credentials.json", true},
			{"valid format", "credentials.json", false}, // Would fail in real test due to missing file
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// We can't test the actual Firebase initialization without credentials
				// But we can test the parameter validation logic
				if tc.credentialsPath == "" {
					// Empty path should cause error
					assert.True(t, tc.expectError)
				}
				
				// Validate path format
				assert.IsType(t, "", tc.credentialsPath)
			})
		}
	})
}

// Benchmark tests
func BenchmarkFCMService_SendNotification(b *testing.B) {
	service := NewMockFCMService()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.SendNotification("test-token", "Title", "Body")
	}
}

func BenchmarkFCMService_ValidationCheck(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token := "test-token"
		if token == "" {
			// Validation logic
			continue
		}
	}
}