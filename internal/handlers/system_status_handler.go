// system_status_handler.go
package handlers

import (
	"encoding/json"

	"remiaq/internal/middleware"
	"remiaq/internal/repository"
	"remiaq/internal/utils"

	"github.com/pocketbase/pocketbase/core"
)

// SystemStatusHandler cung cấp API GET/PUT cho system_status (singleton mid=1)
type SystemStatusHandler struct {
	repo repository.SystemStatusRepository
}

// NewSystemStatusHandler khởi tạo handler
func NewSystemStatusHandler(repo repository.SystemStatusRepository) *SystemStatusHandler {
	return &SystemStatusHandler{repo: repo}
}

// GetSystemStatus xử lý GET /api/system_status
func (h *SystemStatusHandler) GetSystemStatus(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	status, err := h.repo.Get(re.Request.Context())
	if err != nil {
		return utils.SendError(re, 500, "Failed to read system status", err)
	}
	return utils.SendSuccess(re, "", status)
}

// PutSystemStatus xử lý PUT /api/system_status
// Body cho phép cập nhật worker_enabled và/hoặc last_error
//
//	{
//	  "worker_enabled": true|false?,
//	  "last_error": "..."?
//	}
func (h *SystemStatusHandler) PutSystemStatus(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	var req struct {
		WorkerEnabled *bool   `json:"worker_enabled"`
		LastError     *string `json:"last_error"`
	}

	if err := json.NewDecoder(re.Request.Body).Decode(&req); err != nil {
		return utils.SendError(re, 400, "Invalid request body", err)
	}

	ctx := re.Request.Context()

	// Xử lý worker_enabled
	if req.WorkerEnabled != nil {
		if *req.WorkerEnabled {
			if err := h.repo.EnableWorker(ctx); err != nil {
				return utils.SendError(re, 500, "Failed to enable worker", err)
			}
			// Nếu có last_error đi kèm: cập nhật lại; nếu không: clear
			if req.LastError != nil {
				if err := h.repo.UpdateError(ctx, *req.LastError); err != nil {
					return utils.SendError(re, 500, "Failed to update error", err)
				}
			} else {
				if err := h.repo.ClearError(ctx); err != nil {
					return utils.SendError(re, 500, "Failed to clear error", err)
				}
			}
		} else {
			// disable worker; nếu không có last_error thì set lý do mặc định
			msg := "manually disabled"
			if req.LastError != nil {
				msg = *req.LastError
			}
			if err := h.repo.DisableWorker(ctx, msg); err != nil {
				return utils.SendError(re, 500, "Failed to disable worker", err)
			}
		}
	} else if req.LastError != nil {
		// Chỉ cập nhật last_error nếu không thay đổi trạng thái worker
		if err := h.repo.UpdateError(ctx, *req.LastError); err != nil {
			return utils.SendError(re, 500, "Failed to update error", err)
		}
	} else {
		return utils.SendError(re, 400, "No fields to update", nil)
	}

	status, err := h.repo.Get(ctx)
	if err != nil {
		return utils.SendError(re, 500, "Failed to read system status", err)
	}
	return utils.SendSuccess(re, "System status updated", status)
}
