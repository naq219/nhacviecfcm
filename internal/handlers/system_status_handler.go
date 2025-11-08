package handlers

import (
	"context"
	"encoding/json"

	"remiaq/internal/middleware"
	"remiaq/internal/models"
	"remiaq/internal/repository"
	"remiaq/internal/utils"

	"github.com/pocketbase/pocketbase/core"
)

// SystemStatusHandler handles system status HTTP requests
type SystemStatusHandler struct {
	repo repository.SystemStatusRepository
}

// NewSystemStatusHandler creates a new system status handler
func NewSystemStatusHandler(repo repository.SystemStatusRepository) *SystemStatusHandler {
	return &SystemStatusHandler{
		repo: repo,
	}
}

// GetSystemStatus handles GET /api/system_status
// @Summary Get system status
// @Description Get current system status (worker enabled/disabled, last error)
// @Tags system
// @Produce json
// @Success 200 {object} models.SystemStatus
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/system_status [get]
func (h *SystemStatusHandler) GetSystemStatus(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	status, err := h.repo.GetSystemStatus(context.Background())
	if err != nil {
		return utils.SendError(re, 500, "Failed to get system status", err)
	}

	return utils.SendSuccess(re, "", status)
}

// PutSystemStatus handles PUT /api/system_status
// @Summary Update system status
// @Description Update system status (enable/disable worker, set error)
// @Tags system
// @Accept json
// @Produce json
// @Param body body models.SystemStatus true "System status update"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/system_status [put]
func (h *SystemStatusHandler) PutSystemStatus(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	var req struct {
		WorkerEnabled bool   `json:"worker_enabled"`
		LastError     string `json:"last_error"`
	}

	if err := json.NewDecoder(re.Request.Body).Decode(&req); err != nil {
		return utils.SendError(re, 400, "Invalid request body", err)
	}

	// Get current status first
	currentStatus, err := h.repo.GetSystemStatus(context.Background())
	if err != nil {
		return utils.SendError(re, 500, "Failed to get current status", err)
	}

	// Update based on request
	newStatus := &models.SystemStatus{
		ID:            currentStatus.ID,
		WorkerEnabled: req.WorkerEnabled,
		LastError:     req.LastError,
	}

	// If enabling worker and no error message, clear error
	if req.WorkerEnabled && req.LastError == "" {
		newStatus.LastError = ""
	}

	// If disabling worker and no error message, set default
	if !req.WorkerEnabled && req.LastError == "" {
		newStatus.LastError = "Manually disabled"
	}

	if err := h.repo.UpdateSystemStatus(context.Background(), newStatus); err != nil {
		return utils.SendError(re, 500, "Failed to update system status", err)
	}

	resp := map[string]interface{}{
		"success": true,
		"message": "System status updated",
		"data":    newStatus,
	}

	return utils.SendSuccess(re, "System status updated", resp)
}

// Helper function if needed
func (h *SystemStatusHandler) EnableWorker(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	status, err := h.repo.GetSystemStatus(context.Background())
	if err != nil {
		return utils.SendError(re, 500, "Failed to get system status", err)
	}

	status.WorkerEnabled = true
	status.LastError = ""

	if err := h.repo.UpdateSystemStatus(context.Background(), status); err != nil {
		return utils.SendError(re, 500, "Failed to enable worker", err)
	}

	return utils.SendSuccess(re, "Worker enabled", status)
}

// Helper function if needed
func (h *SystemStatusHandler) DisableWorker(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	var req struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(re.Request.Body).Decode(&req); err != nil {
		req.Reason = "Manually disabled"
	}

	if req.Reason == "" {
		req.Reason = "Manually disabled"
	}

	status, err := h.repo.GetSystemStatus(context.Background())
	if err != nil {
		return utils.SendError(re, 500, "Failed to get system status", err)
	}

	status.WorkerEnabled = false
	status.LastError = req.Reason

	if err := h.repo.UpdateSystemStatus(context.Background(), status); err != nil {
		return utils.SendError(re, 500, "Failed to disable worker", err)
	}

	return utils.SendSuccess(re, "Worker disabled", status)
}
