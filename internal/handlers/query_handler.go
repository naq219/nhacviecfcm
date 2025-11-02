package handlers

import (
	"encoding/json"

	"remiaq/internal/middleware"
	"remiaq/internal/repository"
	"remiaq/internal/utils"

	"github.com/pocketbase/pocketbase/core"
)

// QueryHandler handles raw SQL query requests
type QueryHandler struct {
	queryRepo repository.QueryRepository
}

// NewQueryHandler creates a new query handler
func NewQueryHandler(queryRepo repository.QueryRepository) *QueryHandler {
	return &QueryHandler{
		queryRepo: queryRepo,
	}
}

// QueryRequest represents a raw SQL query request
type QueryRequest struct {
	Query string `json:"query"`
}

// parseRequest parses query from request (GET or POST)
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

// HandleSelect handles SELECT queries
func (h *QueryHandler) HandleSelect(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	req, err := parseRequest(re)
	if err != nil {
		return utils.SendError(re, 400, "Invalid request format", err)
	}

	// Validate query
	if err := middleware.ValidateSelectQuery(req.Query); err != nil {
		return utils.SendError(re, 400, "Query validation failed", err)
	}

	// Execute query
	result, err := h.queryRepo.ExecuteSelect(re.Request.Context(), req.Query)
	if err != nil {
		return utils.SendError(re, 400, "Query execution failed", err)
	}

	return utils.SendQueryResponse(re, result)
}

// HandleInsert handles INSERT queries
func (h *QueryHandler) HandleInsert(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	req, err := parseRequest(re)
	if err != nil {
		return utils.SendError(re, 400, "Invalid request format", err)
	}

	// Validate query
	if err := middleware.ValidateInsertQuery(req.Query); err != nil {
		return utils.SendError(re, 400, "Query validation failed", err)
	}

	// Execute query
	rowsAffected, lastInsertId, err := h.queryRepo.ExecuteInsert(re.Request.Context(), req.Query)
	if err != nil {
		return utils.SendError(re, 400, "Insert execution failed", err)
	}

	return utils.SendMutationResponse(re, rowsAffected, lastInsertId)
}

// HandleUpdate handles UPDATE queries
func (h *QueryHandler) HandleUpdate(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	req, err := parseRequest(re)
	if err != nil {
		return utils.SendError(re, 400, "Invalid request format", err)
	}

	// Validate query
	if err := middleware.ValidateUpdateQuery(req.Query); err != nil {
		return utils.SendError(re, 400, "Query validation failed", err)
	}

	// Execute query
	rowsAffected, err := h.queryRepo.ExecuteUpdate(re.Request.Context(), req.Query)
	if err != nil {
		return utils.SendError(re, 400, "Update execution failed", err)
	}

	return utils.SendMutationResponse(re, rowsAffected)
}

// HandleDelete handles DELETE queries
func (h *QueryHandler) HandleDelete(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	req, err := parseRequest(re)
	if err != nil {
		return utils.SendError(re, 400, "Invalid request format", err)
	}

	// Validate query
	if err := middleware.ValidateDeleteQuery(req.Query); err != nil {
		return utils.SendError(re, 400, "Query validation failed", err)
	}

	// Execute query
	rowsAffected, err := h.queryRepo.ExecuteDelete(re.Request.Context(), req.Query)
	if err != nil {
		return utils.SendError(re, 400, "Delete execution failed", err)
	}

	return utils.SendMutationResponse(re, rowsAffected)
}
