package main

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/pocketbase/dbx"
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

func main() {
	app := pocketbase.New()
	os.Setenv("PB_ADDR", "127.0.0.1:8888")

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Hello endpoint
		se.Router.GET("/hello", func(re *core.RequestEvent) error {
			return re.String(200, "Hello world!")
		})

		// SELECT queries - GET and POST
		se.Router.GET("/api/query", handleSelectQuery(app))
		se.Router.POST("/query", handleSelectQuery(app))

		// INSERT queries - GET and POST
		se.Router.GET("/insert", handleInsertQuery(app))
		se.Router.POST("/insert", handleInsertQuery(app))

		// UPDATE queries - GET and PUT
		se.Router.GET("/update", handleUpdateQuery(app))
		se.Router.PUT("/update", handleUpdateQuery(app))

		// DELETE queries - GET and DELETE
		se.Router.GET("/delete", handleDeleteQuery(app))
		se.Router.DELETE("/delete", handleDeleteQuery(app))

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

// Handler functions
func handleSelectQuery(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(re *core.RequestEvent) error {
		req, err := parseRequest(re)
		if err != nil {
			return sendErrorResponse(re, 400, "Invalid request format", err)
		}

		if err := validateSelectQuery(req.Query); err != nil {
			return sendErrorResponse(re, 400, "Query validation failed", err)
		}

		var rawResult []dbx.NullStringMap
		if err := app.DB().NewQuery(req.Query).All(&rawResult); err != nil {
			return sendErrorResponse(re, 400, "Query execution failed", err)
		}

		// Convert NullStringMap to regular map (consistent with PocketBase)
		result := make([]map[string]interface{}, 0, len(rawResult))
		for _, row := range rawResult {
			cleaned := map[string]interface{}{}
			for key, val := range row {
				if val.Valid {
					cleaned[key] = val.String
				} else {
					cleaned[key] = nil // Use nil instead of empty string for better JSON compatibility
				}
			}
			result = append(result, cleaned)
		}

		return sendQueryResponse(re, result)
	}
}

func handleInsertQuery(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(re *core.RequestEvent) error {
		req, err := parseRequest(re)
		if err != nil {
			return sendErrorResponse(re, 400, "Invalid request format", err)
		}

		if err := validateInsertQuery(req.Query); err != nil {
			return sendErrorResponse(re, 400, "Query validation failed", err)
		}

		result, err := app.DB().NewQuery(req.Query).Execute()
		if err != nil {
			return sendErrorResponse(re, 400, "Insert execution failed", err)
		}

		rowsAffected, _ := result.RowsAffected()
		lastInsertId, _ := result.LastInsertId()

		return sendMutationResponse(re, rowsAffected, lastInsertId)
	}
}

func handleUpdateQuery(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(re *core.RequestEvent) error {
		req, err := parseRequest(re)
		if err != nil {
			return sendErrorResponse(re, 400, "Invalid request format", err)
		}

		if err := validateUpdateQuery(req.Query); err != nil {
			return sendErrorResponse(re, 400, "Query validation failed", err)
		}

		result, err := app.DB().NewQuery(req.Query).Execute()
		if err != nil {
			return sendErrorResponse(re, 400, "Update execution failed", err)
		}

		rowsAffected, _ := result.RowsAffected()

		return sendMutationResponse(re, rowsAffected)
	}
}

func handleDeleteQuery(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(re *core.RequestEvent) error {
		req, err := parseRequest(re)
		if err != nil {
			return sendErrorResponse(re, 400, "Invalid request format", err)
		}

		if err := validateDeleteQuery(req.Query); err != nil {
			return sendErrorResponse(re, 400, "Query validation failed", err)
		}

		result, err := app.DB().NewQuery(req.Query).Execute()
		if err != nil {
			return sendErrorResponse(re, 400, "Delete execution failed", err)
		}

		rowsAffected, _ := result.RowsAffected()

		return sendMutationResponse(re, rowsAffected)
	}
}