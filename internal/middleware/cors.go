package middleware

import (
	"github.com/pocketbase/pocketbase/core"
)

// SetCORSHeaders sets CORS headers for the response
func SetCORSHeaders(re *core.RequestEvent) {
	re.Response.Header().Set("Access-Control-Allow-Origin", "*")
	re.Response.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD, PATCH")
	re.Response.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, Cache-Control, X-File-Name")
	re.Response.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range")
	re.Response.Header().Set("Access-Control-Allow-Credentials", "true")
	re.Response.Header().Set("Access-Control-Max-Age", "86400")
}

// CORSMiddleware is a middleware that handles CORS
func CORSMiddleware(next func(*core.RequestEvent) error) func(*core.RequestEvent) error {
	return func(re *core.RequestEvent) error {
		SetCORSHeaders(re)

		// Handle preflight
		if re.Request.Method == "OPTIONS" {
			return re.NoContent(204)
		}

		return next(re)
	}
}
