package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/stretchr/testify/assert"
)

// Helper function to create mock RequestEvent for CORS tests
func createCORSRequestEvent(method, path string) *core.RequestEvent {
	req := httptest.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()
	
	re := &core.RequestEvent{
		Event: router.Event{
			Request:  req,
			Response: recorder,
		},
	}
	
	return re
}

func TestSetCORSHeaders(t *testing.T) {
	t.Run("should set all CORS headers correctly", func(t *testing.T) {
		re := createCORSRequestEvent("GET", "/api/test")
		
		SetCORSHeaders(re)
		
		headers := re.Response.Header()
		
		assert.Equal(t, "*", headers.Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS, HEAD, PATCH", headers.Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization, X-Requested-With, Accept, Origin, Cache-Control, X-File-Name", headers.Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "Content-Length, Content-Range", headers.Get("Access-Control-Expose-Headers"))
		assert.Equal(t, "true", headers.Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "86400", headers.Get("Access-Control-Max-Age"))
	})
	
	t.Run("should not overwrite existing headers", func(t *testing.T) {
		re := createCORSRequestEvent("GET", "/api/test")
		
		// Set a custom header first
		re.Response.Header().Set("Custom-Header", "custom-value")
		
		SetCORSHeaders(re)
		
		// Custom header should still exist
		assert.Equal(t, "custom-value", re.Response.Header().Get("Custom-Header"))
		
		// CORS headers should be set
		assert.Equal(t, "*", re.Response.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestCORSMiddleware(t *testing.T) {
	t.Run("should handle OPTIONS request (preflight)", func(t *testing.T) {
		re := createCORSRequestEvent("OPTIONS", "/api/test")
		
		// Mock next handler that should not be called for OPTIONS
		nextCalled := false
		next := func(re *core.RequestEvent) error {
			nextCalled = true
			return nil
		}
		
		middleware := CORSMiddleware(next)
		err := middleware(re)
		
		assert.NoError(t, err)
		assert.False(t, nextCalled, "Next handler should not be called for OPTIONS request")
		
		// Check CORS headers are set
		headers := re.Response.Header()
		assert.Equal(t, "*", headers.Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS, HEAD, PATCH", headers.Get("Access-Control-Allow-Methods"))
	})
	
	t.Run("should call next handler for non-OPTIONS request", func(t *testing.T) {
		re := createCORSRequestEvent("GET", "/api/test")
		
		// Mock next handler
		nextCalled := false
		next := func(re *core.RequestEvent) error {
			nextCalled = true
			return nil
		}
		
		middleware := CORSMiddleware(next)
		err := middleware(re)
		
		assert.NoError(t, err)
		assert.True(t, nextCalled, "Next handler should be called for non-OPTIONS request")
		
		// Check CORS headers are set
		headers := re.Response.Header()
		assert.Equal(t, "*", headers.Get("Access-Control-Allow-Origin"))
	})
	
	t.Run("should propagate error from next handler", func(t *testing.T) {
		re := createCORSRequestEvent("POST", "/api/test")
		
		// Mock next handler that returns error
		expectedError := assert.AnError
		next := func(re *core.RequestEvent) error {
			return expectedError
		}
		
		middleware := CORSMiddleware(next)
		err := middleware(re)
		
		assert.Equal(t, expectedError, err)
		
		// Check CORS headers are still set even when error occurs
		headers := re.Response.Header()
		assert.Equal(t, "*", headers.Get("Access-Control-Allow-Origin"))
	})
	
	t.Run("should handle different HTTP methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD"}
		
		for _, method := range methods {
			t.Run("method_"+method, func(t *testing.T) {
				re := createCORSRequestEvent(method, "/api/test")
				
				nextCalled := false
				next := func(re *core.RequestEvent) error {
					nextCalled = true
					return nil
				}
				
				middleware := CORSMiddleware(next)
				err := middleware(re)
				
				assert.NoError(t, err)
				assert.True(t, nextCalled, "Next handler should be called for "+method+" request")
				
				// Check CORS headers are set
				headers := re.Response.Header()
				assert.Equal(t, "*", headers.Get("Access-Control-Allow-Origin"))
			})
		}
	})
}

// Benchmark tests
func BenchmarkSetCORSHeaders(b *testing.B) {
	re := createCORSRequestEvent("GET", "/api/test")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SetCORSHeaders(re)
	}
}

func BenchmarkCORSMiddleware_OPTIONS(b *testing.B) {
	next := func(re *core.RequestEvent) error {
		return nil
	}
	middleware := CORSMiddleware(next)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re := createCORSRequestEvent("OPTIONS", "/api/test")
		middleware(re)
	}
}

func BenchmarkCORSMiddleware_GET(b *testing.B) {
	next := func(re *core.RequestEvent) error {
		return nil
	}
	middleware := CORSMiddleware(next)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re := createCORSRequestEvent("GET", "/api/test")
		middleware(re)
	}
}