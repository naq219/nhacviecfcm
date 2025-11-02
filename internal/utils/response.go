package utils

import (
	"github.com/pocketbase/pocketbase/core"
)

// Response structures following PocketBase format

// QueryResponse for SELECT queries
type QueryResponse struct {
	Page       int                      `json:"page"`
	PerPage    int                      `json:"perPage"`
	TotalItems int                      `json:"totalItems"`
	TotalPages int                      `json:"totalPages"`
	Items      []map[string]interface{} `json:"items"`
}

// MutationResponse for INSERT/UPDATE/DELETE
type MutationResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
	LastInsertId int64 `json:"lastInsertId,omitempty"`
}

// ErrorResponse for errors
type ErrorResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// SuccessResponse for generic success
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// SendError sends an error response
func SendError(re *core.RequestEvent, status int, message string, err error) error {
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

// SendSuccess sends a success response
func SendSuccess(re *core.RequestEvent, message string, data interface{}) error {
	response := SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	return re.JSON(200, response)
}

// SendQueryResponse sends a query response (for SELECT)
func SendQueryResponse(re *core.RequestEvent, items []map[string]interface{}) error {
	totalItems := len(items)
	response := QueryResponse{
		Page:       1,
		PerPage:    totalItems,
		TotalItems: totalItems,
		TotalPages: 1,
		Items:      items,
	}
	return re.JSON(200, response)
}

// SendMutationResponse sends a mutation response (for INSERT/UPDATE/DELETE)
func SendMutationResponse(re *core.RequestEvent, rowsAffected int64, lastInsertId ...int64) error {
	response := MutationResponse{
		RowsAffected: rowsAffected,
	}

	if len(lastInsertId) > 0 {
		response.LastInsertId = lastInsertId[0]
	}

	return re.JSON(200, response)
}
