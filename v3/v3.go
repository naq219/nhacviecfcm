package main

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// Request structures
type QueryRequest struct {
	Query string `json:"query"`
}

// Unified response structures (following PocketBase format)
type QueryResponse struct {
	Page       int                      `json:"page"`
	PerPage    int                      `json:"perPage"`
	TotalItems int                      `json:"totalItems"`
	TotalPages int                      `json:"totalPages"`
	Items      []map[string]interface{} `json:"items"`
}

type MutationResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
	LastInsertId int64 `json:"lastInsertId,omitempty"`
}

// Error response (following PocketBase format)
type ErrorResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// Security patterns
var (
	// Dangerous patterns to block
	dangerousPatterns = []string{
		`(?i)\bDROP\b`,
		`(?i)\bTRUNCATE\b`,
		`(?i)\bALTER\b`,
		`(?i)\bCREATE\b`,
		`(?i)\b--\b`,
		`(?i)/\*`,
		`(?i)\*/`,
		`(?i)\bEXEC\b`,
		`(?i)\bEVAL\b`,
		`(?i)\bSYSTEM\b`,
		`(?i)\bSHUTDOWN\b`,
		`(?i)\bGRANT\b`,
		`(?i)\bREVOKE\b`,
		`(?i)\bLOAD_FILE\b`,
		`(?i)\bINTO\s+OUTFILE\b`,
		`(?i)\bINTO\s+DUMPFILE\b`,
	}

	// Compiled regex patterns for validation
	updateWherePattern = regexp.MustCompile(`(?i)UPDATE\s+.*\s+WHERE\s+`)
	deleteWherePattern = regexp.MustCompile(`(?i)DELETE\s+FROM\s+.*\s+WHERE\s+`)
)

// Set CORS headers helper function
func setCORSHeaders(re *core.RequestEvent) {
	re.Response.Header().Set("Access-Control-Allow-Origin", "*")
	re.Response.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD, PATCH")
	re.Response.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, Cache-Control, X-File-Name")
	re.Response.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range")
	re.Response.Header().Set("Access-Control-Allow-Credentials", "true")
	re.Response.Header().Set("Access-Control-Max-Age", "86400")
}

func main() {
	app := pocketbase.New()
	os.Setenv("PB_ADDR", "127.0.0.1:8888")

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Handle preflight OPTIONS requests for all routes
		se.Router.OPTIONS("/*", func(re *core.RequestEvent) error {
			setCORSHeaders(re)
			return re.NoContent(204)
		})

		// Hello endpoint
		se.Router.GET("/hello", func(re *core.RequestEvent) error {
			setCORSHeaders(re)
			return re.String(200, "Hello world!")
		})

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// Utility functions
func parseRequest(re *core.RequestEvent) (*QueryRequest, error) {
	var req QueryRequest

	if re.Request.Method == "GET" {
		req.Query = re.Request.URL.Query().Get("q")
	} else {
		if err := json.NewDecoder(re.Request.Body).Decode(&req); err != nil {
			return nil, err
		}
	}

	return &req, nil
}

func validateQuery(query string) error {
	if strings.TrimSpace(query) == "" {
		return &ValidationError{"Query cannot be empty"}
	}

	// Check for dangerous patterns
	for _, pattern := range dangerousPatterns {
		if matched, _ := regexp.MatchString(pattern, query); matched {
			return &ValidationError{"Query contains potentially dangerous operations"}
		}
	}

	return nil
}

func validateSelectQuery(query string) error {
	if err := validateQuery(query); err != nil {
		return err
	}

	trimmedQuery := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(trimmedQuery, "SELECT") {
		return &ValidationError{"Only SELECT statements are allowed"}
	}

	return nil
}

func validateInsertQuery(query string) error {
	if err := validateQuery(query); err != nil {
		return err
	}

	trimmedQuery := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(trimmedQuery, "INSERT") {
		return &ValidationError{"Only INSERT statements are allowed"}
	}

	return nil
}

func validateUpdateQuery(query string) error {
	if err := validateQuery(query); err != nil {
		return err
	}

	trimmedQuery := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(trimmedQuery, "UPDATE") {
		return &ValidationError{"Only UPDATE statements are allowed"}
	}

	// Require WHERE clause for UPDATE
	if !updateWherePattern.MatchString(query) {
		return &ValidationError{"UPDATE statements must include a WHERE clause for safety"}
	}

	return nil
}

func validateDeleteQuery(query string) error {
	if err := validateQuery(query); err != nil {
		return err
	}

	trimmedQuery := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(trimmedQuery, "DELETE") {
		return &ValidationError{"Only DELETE statements are allowed"}
	}

	// Require WHERE clause for DELETE
	if !deleteWherePattern.MatchString(query) {
		return &ValidationError{"DELETE statements must include a WHERE clause for safety"}
	}

	return nil
}

// Unified response helpers (PocketBase style)
func sendErrorResponse(re *core.RequestEvent, status int, message string, err error) error {
	response := ErrorResponse{
		Code:    status,
		Message: message,
	}

	if err != nil {
		response.Data = map[string]interface{}{
			"error": err.Error(),
		}
	}

	return re.JSON(status, response)
}

func sendQueryResponse(re *core.RequestEvent, items []map[string]interface{}) error {
	totalItems := len(items)
	response := QueryResponse{
		Page:       1,
		PerPage:    totalItems, // For now, return all items in one page
		TotalItems: totalItems,
		TotalPages: 1,
		Items:      items,
	}

	return re.JSON(200, response)
}

func sendMutationResponse(re *core.RequestEvent, rowsAffected int64, lastInsertId ...int64) error {
	response := MutationResponse{
		RowsAffected: rowsAffected,
	}

	if len(lastInsertId) > 0 {
		response.LastInsertId = lastInsertId[0]
	}

	return re.JSON(200, response)
}

// Custom error type
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
