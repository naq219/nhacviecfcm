package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"remiaq/internal/middleware"
	"remiaq/internal/models"
	"remiaq/internal/utils"

	"github.com/pocketbase/pocketbase/core"
)

// ReminderServiceInterface defines the interface for reminder service
type ReminderServiceInterface interface {
	CreateReminder(ctx context.Context, reminder *models.Reminder) error
	GetReminder(ctx context.Context, id string) (*models.Reminder, error)
	UpdateReminder(ctx context.Context, reminder *models.Reminder) error
	DeleteReminder(ctx context.Context, id string) error
	GetUserReminders(ctx context.Context, userID string) ([]*models.Reminder, error)
	SnoozeReminder(ctx context.Context, id string, duration time.Duration) error
	CompleteReminder(ctx context.Context, id string) error
	ProcessDueReminders(ctx context.Context) error
}

// ReminderHandler handles reminder HTTP requests
type ReminderHandler struct {
	reminderService ReminderServiceInterface
}

// NewReminderHandler creates a new reminder handler
func NewReminderHandler(reminderService ReminderServiceInterface) *ReminderHandler {
	return &ReminderHandler{
		reminderService: reminderService,
	}
}

// CreateReminder handles POST /api/reminders
func (h *ReminderHandler) CreateReminder(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	authRecord := re.Auth
	if authRecord == nil {
		return utils.SendError(re, 401, "Unauthorized", errors.New("user not authenticated"))
	}

	var reminder models.Reminder
	if err := json.NewDecoder(re.Request.Body).Decode(&reminder); err != nil {
		return utils.SendError(re, 400, "Invalid request body", err)
	}

	reminder.UserID = authRecord.Id

	// Create reminder
	if err := h.reminderService.CreateReminder(re.Request.Context(), &reminder); err != nil {
		return utils.SendError(re, 400, "Failed to create reminder", err)
	}

	return utils.SendSuccess(re, "Reminder created successfully", reminder)
}

// GetReminder handles GET /api/reminders/:id
func (h *ReminderHandler) GetReminder(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	id := re.Request.PathValue("id")
	if id == "" {
		return utils.SendError(re, 400, "Reminder ID is required", nil)
	}

	reminder, err := h.reminderService.GetReminder(re.Request.Context(), id)
	if err != nil {
		return utils.SendError(re, 404, "Reminder not found", err)
	}

	return utils.SendSuccess(re, "", reminder)
}

// UpdateReminder handles PUT /api/reminders/:id
func (h *ReminderHandler) UpdateReminder(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	id := re.Request.PathValue("id")
	if id == "" {
		return utils.SendError(re, 400, "Reminder ID is required", nil)
	}

	var reminder models.Reminder
	if err := json.NewDecoder(re.Request.Body).Decode(&reminder); err != nil {
		return utils.SendError(re, 400, "Invalid request body", err)
	}

	reminder.ID = id

	if err := h.reminderService.UpdateReminder(re.Request.Context(), &reminder); err != nil {
		return utils.SendError(re, 400, "Failed to update reminder", err)
	}

	return utils.SendSuccess(re, "Reminder updated successfully", reminder)
}

// DeleteReminder handles DELETE /api/reminders/:id
func (h *ReminderHandler) DeleteReminder(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	id := re.Request.PathValue("id")
	if id == "" {
		return utils.SendError(re, 400, "Reminder ID is required", nil)
	}

	if err := h.reminderService.DeleteReminder(re.Request.Context(), id); err != nil {
		return utils.SendError(re, 400, "Failed to delete reminder", err)
	}

	return utils.SendSuccess(re, "Reminder deleted successfully", nil)
}

// GetUserReminders handles GET /api/users/:userId/reminders
func (h *ReminderHandler) GetUserReminders(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	userID := re.Request.PathValue("userId")
	if userID == "" {
		return utils.SendError(re, 400, "User ID is required", nil)
	}

	reminders, err := h.reminderService.GetUserReminders(re.Request.Context(), userID)
	if err != nil {
		return utils.SendError(re, 400, "Failed to get reminders", err)
	}

	return utils.SendSuccess(re, "", reminders)
}

// GetCurrentUserReminders handles GET /api/reminders/mine
func (h *ReminderHandler) GetCurrentUserReminders(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	authRecord := re.Auth
	if authRecord == nil {
		return utils.SendError(re, 401, "Unauthorized", errors.New("user not authenticated"))
	}

	reminders, err := h.reminderService.GetUserReminders(re.Request.Context(), authRecord.Id)
	if err != nil {
		return utils.SendError(re, 400, "Failed to get reminders", err)
	}

	return utils.SendSuccess(re, "", reminders)
}

// SnoozeReminder handles POST /api/reminders/:id/snooze
func (h *ReminderHandler) SnoozeReminder(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	id := re.Request.PathValue("id")
	if id == "" {
		return utils.SendError(re, 400, "Reminder ID is required", nil)
	}

	var req struct {
		Duration int `json:"duration"` // Duration in seconds
	}

	if err := json.NewDecoder(re.Request.Body).Decode(&req); err != nil {
		return utils.SendError(re, 400, "Invalid request body", err)
	}

	duration := time.Duration(req.Duration) * time.Second
	if err := h.reminderService.SnoozeReminder(re.Request.Context(), id, duration); err != nil {
		return utils.SendError(re, 400, "Failed to snooze reminder", err)
	}

	return utils.SendSuccess(re, "Reminder snoozed successfully", nil)
}

// CompleteReminder handles POST /api/reminders/:id/complete
func (h *ReminderHandler) CompleteReminder(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	id := re.Request.PathValue("id")
	if id == "" {
		return utils.SendError(re, 400, "Reminder ID is required", nil)
	}

	if err := h.reminderService.CompleteReminder(re.Request.Context(), id); err != nil {
		return utils.SendError(re, 400, "Failed to complete reminder", err)
	}

	return utils.SendSuccess(re, "Reminder completed successfully", nil)
}
