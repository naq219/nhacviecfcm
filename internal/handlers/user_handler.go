package handlers

import (
	"encoding/json"
	"errors"

	"remiaq/internal/services"
	"remiaq/internal/utils"

	"github.com/pocketbase/pocketbase/core"
)

// UserHandler handles user-related API requests
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// UpdateFCMTokenRequest represents the request body for updating FCM token
type UpdateFCMTokenRequest struct {
	FCMToken string `json:"fcm_token"`
}

// UpdateFCMToken handles the API request to update a user's FCM token
// @Summary Update FCM token for the authenticated user
// @Description Update the Firebase Cloud Messaging token for the currently authenticated user.
// @Tags users
// @Accept json
// @Produce json
// @Param fcm_token body UpdateFCMTokenRequest true "FCM token object"
// @Success 200 {object} utils.SuccessResponse "FCM token updated successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request body or missing FCM token"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Router /api/users/fcm-token [put]
func (h *UserHandler) UpdateFCMToken(e *core.RequestEvent) error {
	authRecord := e.Auth
	if authRecord == nil {
		return utils.SendError(e, 401, "Unauthorized", errors.New("user not authenticated"))
	}

	// Parse request body
	var req UpdateFCMTokenRequest
	if err := json.NewDecoder(e.Request.Body).Decode(&req); err != nil {
		return utils.SendError(e, 400, "Invalid request body", err)
	}

	// Validate FCM token
	if req.FCMToken == "" {
		return utils.SendError(e, 400, "FCM token is required", nil)
	}

	// Gọi service để cập nhật FCM token
	if err := h.userService.UpdateFCMToken(e.Request.Context(), authRecord.Id, req.FCMToken); err != nil {
		return utils.SendError(e, 500, "Failed to update FCM token", err)
	}

	return utils.SendSuccess(e, "FCM token updated successfully", nil)
}
