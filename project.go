package main

import (
	"log"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"remiaq/config"
	"remiaq/internal/handlers" // ← Đã sửa từ api/handlers
	"remiaq/internal/middleware"
	pbRepo "remiaq/internal/repository/pocketbase"
	"remiaq/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set PocketBase server address
	os.Setenv("PB_ADDR", cfg.ServerAddr)

	// Create PocketBase instance
	app := pocketbase.New()

	// Initialize repositories
	reminderRepo := pbRepo.NewPocketBaseReminderRepo(app)
	userRepo := pbRepo.NewPocketBaseUserRepo(app)
	systemStatusRepo := pbRepo.NewPocketBaseSystemStatusRepo(app)
	queryRepo := pbRepo.NewPocketBaseQueryRepo(app)

	// Initialize services
	// Note: FCM service is optional, we'll initialize it with a stub for now
	var fcmService *services.FCMService
	if _, err := os.Stat(cfg.FCMCredentials); err == nil {
		fcmService, err = services.NewFCMService(cfg.FCMCredentials)
		if err != nil {
			log.Printf("Warning: Failed to initialize FCM service: %v", err)
			// Continue without FCM for development
		}
	} else {
		log.Println("Warning: FCM credentials not found, notifications disabled")
	}

	lunarCalendar := services.NewLunarCalendar()
	schedCalculator := services.NewScheduleCalculator(lunarCalendar)
	reminderService := services.NewReminderService(reminderRepo, userRepo, fcmService, schedCalculator)

	// Initialize handlers
	reminderHandler := handlers.NewReminderHandler(reminderService)
	queryHandler := handlers.NewQueryHandler(queryRepo)

	// Setup routes
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Handle preflight OPTIONS requests
		se.Router.OPTIONS("/*", func(re *core.RequestEvent) error {
			middleware.SetCORSHeaders(re)
			return re.NoContent(204)
		})

		// Health check
		se.Router.GET("/hello", func(re *core.RequestEvent) error {
			middleware.SetCORSHeaders(re)
			return re.String(200, "RemiAq API is running!")
		})

		// Raw SQL query endpoints (from original main.go)
		se.Router.GET("/api/rquery", queryHandler.HandleSelect)
		se.Router.POST("/api/rquery", queryHandler.HandleSelect)

		se.Router.GET("/api/rinsert", queryHandler.HandleInsert)
		se.Router.POST("/api/rinsert", queryHandler.HandleInsert)

		se.Router.GET("/api/rupdate", queryHandler.HandleUpdate)
		se.Router.PUT("/api/rupdate", queryHandler.HandleUpdate)

		se.Router.GET("/api/rdelete", queryHandler.HandleDelete)
		se.Router.DELETE("/api/rdelete", queryHandler.HandleDelete)

		// Reminder CRUD endpoints
		se.Router.POST("/api/reminders", reminderHandler.CreateReminder)
		se.Router.GET("/api/reminders/{id}", reminderHandler.GetReminder)
		se.Router.PUT("/api/reminders/{id}", reminderHandler.UpdateReminder)
		se.Router.DELETE("/api/reminders/{id}", reminderHandler.DeleteReminder)

		// User reminders
		se.Router.GET("/api/users/{userId}/reminders", reminderHandler.GetUserReminders)

		// Reminder actions
		se.Router.POST("/api/reminders/{id}/snooze", reminderHandler.SnoozeReminder)
		se.Router.POST("/api/reminders/{id}/complete", reminderHandler.CompleteReminder)

		return se.Next()
	})

	// Start server
	log.Printf("Starting RemiAq API server on %s", cfg.ServerAddr)
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	ServerAddr     string
	WorkerInterval int    // seconds
	FCMCredentials string // path to firebase credentials JSON
	Environment    string // development, production
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		ServerAddr:     getEnv("SERVER_ADDR", "127.0.0.1:8888"),
		WorkerInterval: 10, // 10 seconds
		FCMCredentials: getEnv("FCM_CREDENTIALS", "./firebase-credentials.json"),
		Environment:    getEnv("ENVIRONMENT", "development"),
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
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
package handlers

import (
	"encoding/json"
	"time"

	"remiaq/internal/middleware"
	"remiaq/internal/models"
	"remiaq/internal/services"
	"remiaq/internal/utils"

	"github.com/pocketbase/pocketbase/core"
)

// ReminderHandler handles reminder HTTP requests
type ReminderHandler struct {
	reminderService *services.ReminderService
}

// NewReminderHandler creates a new reminder handler
func NewReminderHandler(reminderService *services.ReminderService) *ReminderHandler {
	return &ReminderHandler{
		reminderService: reminderService,
	}
}

// CreateReminder handles POST /api/reminders
func (h *ReminderHandler) CreateReminder(re *core.RequestEvent) error {
	middleware.SetCORSHeaders(re)

	var reminder models.Reminder
	if err := json.NewDecoder(re.Request.Body).Decode(&reminder); err != nil {
		return utils.SendError(re, 400, "Invalid request body", err)
	}

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
package middleware

import (
	"errors"
	"regexp"
	"strings"
)

// SQL validation patterns
var (
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

	updateWherePattern = regexp.MustCompile(`(?i)UPDATE\s+.*\s+WHERE\s+`)
	deleteWherePattern = regexp.MustCompile(`(?i)DELETE\s+FROM\s+.*\s+WHERE\s+`)
)

// ValidateQuery validates a SQL query for safety
func ValidateQuery(query string) error {
	if strings.TrimSpace(query) == "" {
		return errors.New("query cannot be empty")
	}

	// Check for dangerous patterns
	for _, pattern := range dangerousPatterns {
		if matched, _ := regexp.MatchString(pattern, query); matched {
			return errors.New("query contains potentially dangerous operations")
		}
	}

	return nil
}

// ValidateSelectQuery validates a SELECT query
func ValidateSelectQuery(query string) error {
	if err := ValidateQuery(query); err != nil {
		return err
	}

	trimmedQuery := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(trimmedQuery, "SELECT") {
		return errors.New("only SELECT statements are allowed")
	}

	return nil
}

// ValidateInsertQuery validates an INSERT query
func ValidateInsertQuery(query string) error {
	if err := ValidateQuery(query); err != nil {
		return err
	}

	trimmedQuery := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(trimmedQuery, "INSERT") {
		return errors.New("only INSERT statements are allowed")
	}

	return nil
}

// ValidateUpdateQuery validates an UPDATE query
func ValidateUpdateQuery(query string) error {
	if err := ValidateQuery(query); err != nil {
		return err
	}

	trimmedQuery := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(trimmedQuery, "UPDATE") {
		return errors.New("only UPDATE statements are allowed")
	}

	// Require WHERE clause for UPDATE
	if !updateWherePattern.MatchString(query) {
		return errors.New("UPDATE statements must include a WHERE clause for safety")
	}

	return nil
}

// ValidateDeleteQuery validates a DELETE query
func ValidateDeleteQuery(query string) error {
	if err := ValidateQuery(query); err != nil {
		return err
	}

	trimmedQuery := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(trimmedQuery, "DELETE") {
		return errors.New("only DELETE statements are allowed")
	}

	// Require WHERE clause for DELETE
	if !deleteWherePattern.MatchString(query) {
		return errors.New("DELETE statements must include a WHERE clause for safety")
	}

	return nil
}
package models

import (
	"time"
)

// Reminder represents a notification reminder
type Reminder struct {
	ID                string             `json:"id" db:"id"`
	UserID            string             `json:"user_id" db:"user_id"`
	Title             string             `json:"title" db:"title"`
	Description       string             `json:"description" db:"description"`
	Type              string             `json:"type" db:"type"`                   // one_time, recurring
	CalendarType      string             `json:"calendar_type" db:"calendar_type"` // solar, lunar
	NextTriggerAt     time.Time          `json:"next_trigger_at" db:"next_trigger_at"`
	TriggerTimeOfDay  string             `json:"trigger_time_of_day" db:"trigger_time_of_day"` // HH:MM format
	RecurrencePattern *RecurrencePattern `json:"recurrence_pattern" db:"recurrence_pattern"`   // JSON field
	RepeatStrategy    string             `json:"repeat_strategy" db:"repeat_strategy"`         // none, retry_until_complete
	RetryIntervalSec  int                `json:"retry_interval_sec" db:"retry_interval_sec"`
	MaxRetries        int                `json:"max_retries" db:"max_retries"`
	RetryCount        int                `json:"retry_count" db:"retry_count"`
	Status            string             `json:"status" db:"status"` // active, completed, paused
	SnoozeUntil       *time.Time         `json:"snooze_until" db:"snooze_until"`
	LastCompletedAt   *time.Time         `json:"last_completed_at" db:"last_completed_at"`
	LastSentAt        *time.Time         `json:"last_sent_at" db:"last_sent_at"`
	Created           time.Time          `json:"created" db:"created"`
	Updated           time.Time          `json:"updated" db:"updated"`
}

// RecurrencePattern defines how a reminder repeats
type RecurrencePattern struct {
	Type            string `json:"type"`                       // daily, weekly, monthly, lunar_last_day_of_month
	IntervalSeconds int    `json:"interval_seconds,omitempty"` // For interval-based recurrence
	DayOfMonth      int    `json:"day_of_month,omitempty"`     // For monthly recurrence
	DayOfWeek       int    `json:"day_of_week,omitempty"`      // For weekly recurrence (0=Sunday)
	BaseOn          string `json:"base_on,omitempty"`          // creation, completion
}

// User represents a user with FCM token
type User struct {
	ID          string    `json:"id" db:"id"`
	Email       string    `json:"email" db:"email"`
	FCMToken    string    `json:"fcm_token" db:"fcm_token"`
	IsFCMActive bool      `json:"is_fcm_active" db:"is_fcm_active"`
	Created     time.Time `json:"created" db:"created"`
	Updated     time.Time `json:"updated" db:"updated"`
}

// SystemStatus represents system configuration (singleton)
type SystemStatus struct {
	ID            int       `json:"id" db:"id"` // Always 1
	WorkerEnabled bool      `json:"worker_enabled" db:"worker_enabled"`
	LastError     string    `json:"last_error" db:"last_error"`
	Updated       time.Time `json:"updated" db:"updated"`
}

// Constants for reminder types
const (
	ReminderTypeOneTime   = "one_time"
	ReminderTypeRecurring = "recurring"
)

// Constants for calendar types
const (
	CalendarTypeSolar = "solar"
	CalendarTypeLunar = "lunar"
)

// Constants for repeat strategies
const (
	RepeatStrategyNone               = "none"
	RepeatStrategyRetryUntilComplete = "retry_until_complete"
)

// Constants for reminder status
const (
	ReminderStatusActive    = "active"
	ReminderStatusCompleted = "completed"
	ReminderStatusPaused    = "paused"
)

// Constants for recurrence pattern types
const (
	RecurrenceTypeDaily               = "daily"
	RecurrenceTypeWeekly              = "weekly"
	RecurrenceTypeMonthly             = "monthly"
	RecurrenceTypeLunarLastDayOfMonth = "lunar_last_day_of_month"
)

// Constants for base_on
const (
	BaseOnCreation   = "creation"
	BaseOnCompletion = "completion"
)

// Validate checks if reminder data is valid
func (r *Reminder) Validate() error {
	if r.Title == "" {
		return &ValidationError{Field: "title", Message: "Title is required"}
	}
	if r.Type != ReminderTypeOneTime && r.Type != ReminderTypeRecurring {
		return &ValidationError{Field: "type", Message: "Type must be one_time or recurring"}
	}
	if r.CalendarType != CalendarTypeSolar && r.CalendarType != CalendarTypeLunar {
		return &ValidationError{Field: "calendar_type", Message: "Calendar type must be solar or lunar"}
	}
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// IsRetryable checks if reminder can be retried
func (r *Reminder) IsRetryable() bool {
	return r.RepeatStrategy == RepeatStrategyRetryUntilComplete &&
		r.RetryCount < r.MaxRetries
}

// ShouldSend checks if reminder should be sent now
func (r *Reminder) ShouldSend(now time.Time) bool {
	if r.Status != ReminderStatusActive {
		return false
	}

	// Check snooze
	if r.SnoozeUntil != nil && now.Before(*r.SnoozeUntil) {
		return false
	}

	// Check trigger time
	return !now.Before(r.NextTriggerAt)
}
package repository

import (
	"context"
	"time"

	"remiaq/internal/models"
)

// ReminderRepository defines operations for reminder data access
type ReminderRepository interface {
	// CRUD operations
	Create(ctx context.Context, reminder *models.Reminder) error
	GetByID(ctx context.Context, id string) (*models.Reminder, error)
	Update(ctx context.Context, reminder *models.Reminder) error
	Delete(ctx context.Context, id string) error

	// Query operations
	GetDueReminders(ctx context.Context, beforeTime time.Time) ([]*models.Reminder, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Reminder, error)

	// Specific updates
	UpdateNextTrigger(ctx context.Context, id string, nextTrigger time.Time) error
	UpdateStatus(ctx context.Context, id string, status string) error
	IncrementRetryCount(ctx context.Context, id string) error
	UpdateSnooze(ctx context.Context, id string, snoozeUntil *time.Time) error
	MarkCompleted(ctx context.Context, id string, completedAt time.Time) error
	UpdateLastSent(ctx context.Context, id string, sentAt time.Time) error
}

// UserRepository defines operations for user data access
type UserRepository interface {
	// CRUD operations
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error

	// FCM token management
	UpdateFCMToken(ctx context.Context, userID, token string) error
	DisableFCM(ctx context.Context, userID string) error
	EnableFCM(ctx context.Context, userID string, token string) error

	// Query operations
	GetActiveUsers(ctx context.Context) ([]*models.User, error)
}

// SystemStatusRepository defines operations for system status management
type SystemStatusRepository interface {
	// Get singleton instance
	Get(ctx context.Context) (*models.SystemStatus, error)

	// Worker control
	IsWorkerEnabled(ctx context.Context) (bool, error)
	EnableWorker(ctx context.Context) error
	DisableWorker(ctx context.Context, errorMsg string) error

	// Error tracking
	UpdateError(ctx context.Context, errorMsg string) error
	ClearError(ctx context.Context) error
}

// QueryRepository defines operations for raw SQL queries (existing functionality)
type QueryRepository interface {
	// Raw query operations
	ExecuteSelect(ctx context.Context, query string) ([]map[string]interface{}, error)
	ExecuteInsert(ctx context.Context, query string) (rowsAffected int64, lastInsertId int64, err error)
	ExecuteUpdate(ctx context.Context, query string) (rowsAffected int64, err error)
	ExecuteDelete(ctx context.Context, query string) (rowsAffected int64, err error)
}
package pocketbase

import (
	"context"

	"remiaq/internal/repository"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
)

// PocketBaseQueryRepo implements QueryRepository for raw SQL queries
type PocketBaseQueryRepo struct {
	app *pocketbase.PocketBase
}

// Ensure implementation
var _ repository.QueryRepository = (*PocketBaseQueryRepo)(nil)

// NewPocketBaseQueryRepo creates a new query repository
func NewPocketBaseQueryRepo(app *pocketbase.PocketBase) repository.QueryRepository {
	return &PocketBaseQueryRepo{app: app}
}

// ExecuteSelect executes a SELECT query and returns results
func (r *PocketBaseQueryRepo) ExecuteSelect(ctx context.Context, query string) ([]map[string]interface{}, error) {
	var rawResult []dbx.NullStringMap
	if err := r.app.DB().NewQuery(query).All(&rawResult); err != nil {
		return nil, err
	}

	// Convert NullStringMap to regular map
	result := make([]map[string]interface{}, 0, len(rawResult))
	for _, row := range rawResult {
		cleaned := map[string]interface{}{}
		for key, val := range row {
			if val.Valid {
				cleaned[key] = val.String
			} else {
				cleaned[key] = nil
			}
		}
		result = append(result, cleaned)
	}

	return result, nil
}

// ExecuteInsert executes an INSERT query
func (r *PocketBaseQueryRepo) ExecuteInsert(ctx context.Context, query string) (int64, int64, error) {
	result, err := r.app.DB().NewQuery(query).Execute()
	if err != nil {
		return 0, 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	lastInsertId, _ := result.LastInsertId()

	return rowsAffected, lastInsertId, nil
}

// ExecuteUpdate executes an UPDATE query
func (r *PocketBaseQueryRepo) ExecuteUpdate(ctx context.Context, query string) (int64, error) {
	result, err := r.app.DB().NewQuery(query).Execute()
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

// ExecuteDelete executes a DELETE query
func (r *PocketBaseQueryRepo) ExecuteDelete(ctx context.Context, query string) (int64, error) {
	result, err := r.app.DB().NewQuery(query).Execute()
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}
// internal/repository/pocketbase/reminder_repo.go
package pocketbase

import (
	"context"
	"encoding/json"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/repository"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

type PocketBaseReminderRepo struct {
	app *pocketbase.PocketBase
}

var _ repository.ReminderRepository = (*PocketBaseReminderRepo)(nil)

func NewPocketBaseReminderRepo(app *pocketbase.PocketBase) repository.ReminderRepository {
	return &PocketBaseReminderRepo{app: app}
}

func (r *PocketBaseReminderRepo) Create(ctx context.Context, reminder *models.Reminder) error {
	patternJSON, _ := json.Marshal(reminder.RecurrencePattern)

	query := `
        INSERT INTO reminders (
            id, user_id, title, description, type, calendar_type,
            next_trigger_at, trigger_time_of_day, recurrence_pattern,
            repeat_strategy, retry_interval_sec, max_retries, status,
            snooze_until, last_completed_at, last_sent_at,
            created, updated
        ) VALUES (
            ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
        )
    `

	q := r.app.DB().NewQuery(query)
	q.Bind(
		reminder.ID,
		reminder.UserID,
		reminder.Title,
		reminder.Description,
		reminder.Type,
		reminder.CalendarType,
		reminder.NextTriggerAt,
		reminder.TriggerTimeOfDay,
		string(patternJSON),
		reminder.RepeatStrategy,
		reminder.RetryIntervalSec,
		reminder.MaxRetries,
		reminder.Status,
		reminder.SnoozeUntil,
		reminder.LastCompletedAt,
		reminder.LastSentAt,
		reminder.Created,
		reminder.Updated,
	)
	_, err := q.Execute()
	return err
}

func (r *PocketBaseReminderRepo) GetByID(ctx context.Context, id string) (*models.Reminder, error) {
	query := `SELECT * FROM reminders WHERE id = ? LIMIT 1`
	q := r.app.DB().NewQuery(query)
	q.Bind(id)

	var raw map[string]any
	err := q.One(&raw)
	if err != nil {
		return nil, err
	}
	return r.mapToReminder(raw)
}

func (r *PocketBaseReminderRepo) Update(ctx context.Context, reminder *models.Reminder) error {
	patternJSON, _ := json.Marshal(reminder.RecurrencePattern)

	query := `
        UPDATE reminders SET
            user_id = ?, title = ?, description = ?, type = ?, calendar_type = ?,
            next_trigger_at = ?, trigger_time_of_day = ?, recurrence_pattern = ?,
            repeat_strategy = ?, retry_interval_sec = ?, max_retries = ?, status = ?,
            snooze_until = ?, last_completed_at = ?, last_sent_at = ?,
            updated = ?
        WHERE id = ?
    `

	q := r.app.DB().NewQuery(query)
	q.Bind(
		reminder.UserID,
		reminder.Title,
		reminder.Description,
		reminder.Type,
		reminder.CalendarType,
		reminder.NextTriggerAt,
		reminder.TriggerTimeOfDay,
		string(patternJSON),
		reminder.RepeatStrategy,
		reminder.RetryIntervalSec,
		reminder.MaxRetries,
		reminder.Status,
		reminder.SnoozeUntil,
		reminder.LastCompletedAt,
		reminder.LastSentAt,
		time.Now(),
		reminder.ID,
	)
	_, err := q.Execute()
	return err
}

func (r *PocketBaseReminderRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM reminders WHERE id = ?`
	q := r.app.DB().NewQuery(query)
	q.Bind(id)
	_, err := q.Execute()
	return err
}

func (r *PocketBaseReminderRepo) GetDueReminders(ctx context.Context, beforeTime time.Time) ([]*models.Reminder, error) {
	query := `
        SELECT * FROM reminders
        WHERE next_trigger_at <= ?
          AND status = 'active'
          AND (snooze_until IS NULL OR snooze_until <= ?)
    `
	q := r.app.DB().NewQuery(query)
	q.Bind(beforeTime, beforeTime)

	var rawResults []map[string]any
	err := q.All(&rawResults)
	if err != nil {
		return nil, err
	}

	reminders := make([]*models.Reminder, 0, len(rawResults))
	for _, raw := range rawResults {
		rem, err := r.mapToReminder(raw)
		if err != nil {
			continue
		}
		reminders = append(reminders, rem)
	}
	return reminders, nil
}

func (r *PocketBaseReminderRepo) UpdateNextTrigger(ctx context.Context, id string, nextTrigger time.Time) error {
	query := `UPDATE reminders SET next_trigger_at = ?, updated = ? WHERE id = ?`
	q := r.app.DB().NewQuery(query)
	q.Bind(nextTrigger, time.Now(), id)
	_, err := q.Execute()
	return err
}

// Helper: map raw DB row → Reminder struct
func (r *PocketBaseReminderRepo) mapToReminder(raw map[string]any) (*models.Reminder, error) {
	var pattern models.RecurrencePattern
	if p, ok := raw["recurrence_pattern"].(string); ok && p != "" {
		json.Unmarshal([]byte(p), &pattern)
	}

	rem := &models.Reminder{
		ID:                getString(raw, "id"),
		UserID:            getString(raw, "user_id"),
		Title:             getString(raw, "title"),
		Description:       getString(raw, "description"),
		Type:              getString(raw, "type"),
		CalendarType:      getString(raw, "calendar_type"),
		NextTriggerAt:     getTime(raw, "next_trigger_at"),
		TriggerTimeOfDay:  getString(raw, "trigger_time_of_day"),
		RecurrencePattern: pattern,
		RepeatStrategy:    getString(raw, "repeat_strategy"),
		RetryIntervalSec:  getInt(raw, "retry_interval_sec"),
		MaxRetries:        getInt(raw, "max_retries"),
		RetryCount:        getInt(raw, "retry_count"),
		Status:            getString(raw, "status"),
		Created:           getTime(raw, "created"),
		Updated:           getTime(raw, "updated"),
	}

	if v := raw["snooze_until"]; v != nil {
		t := getTime(raw, "snooze_until")
		rem.SnoozeUntil = &t
	}
	if v := raw["last_completed_at"]; v != nil {
		t := getTime(raw, "last_completed_at")
		rem.LastCompletedAt = &t
	}
	if v := raw["last_sent_at"]; v != nil {
		t := getTime(raw, "last_sent_at")
		rem.LastSentAt = &t
	}

	return rem, nil
}

// Helper functions
func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]any, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getTime(m map[string]any, key string) time.Time {
	if v, ok := m[key].(time.Time); ok {
		return v
	}
	return time.Time{}
}
package pocketbase

import (
	"context"
	"encoding/json"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/repository"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
)

// PocketBaseSystemStatusRepo implements SystemStatusRepository
type PocketBaseSystemStatusRepo struct {
	app *pocketbase.PocketBase
}

// Ensure implementation
var _ repository.SystemStatusRepository = (*PocketBaseSystemStatusRepo)(nil)

// NewPocketBaseSystemStatusRepo creates a new system status repository
func NewPocketBaseSystemStatusRepo(app *pocketbase.PocketBase) repository.SystemStatusRepository {
	return &PocketBaseSystemStatusRepo{app: app}
}

// Get retrieves the system status (singleton, id=1)
func (r *PocketBaseSystemStatusRepo) Get(ctx context.Context) (*models.SystemStatus, error) {
	query := `SELECT * FROM system_status WHERE id = 1`

	var rawResult dbx.NullStringMap
	err := r.app.DB().NewQuery(query).One(&rawResult)
	if err != nil {
		return nil, err
	}

	return r.mapToSystemStatus(rawResult)
}

// IsWorkerEnabled checks if worker is enabled
func (r *PocketBaseSystemStatusRepo) IsWorkerEnabled(ctx context.Context) (bool, error) {
	query := `SELECT worker_enabled FROM system_status WHERE id = 1`

	var rawResult dbx.NullStringMap
	err := r.app.DB().NewQuery(query).One(&rawResult)
	if err != nil {
		return false, err
	}

	if rawResult["worker_enabled"].Valid {
		var enabled bool
		json.Unmarshal([]byte(rawResult["worker_enabled"].String), &enabled)
		return enabled, nil
	}

	return false, nil
}

// EnableWorker enables the worker
func (r *PocketBaseSystemStatusRepo) EnableWorker(ctx context.Context) error {
	query := `UPDATE system_status SET worker_enabled = TRUE, updated = ? WHERE id = 1`
	_, err := r.app.DB().NewQuery(query).Execute(time.Now().UTC())
	return err
}

// DisableWorker disables the worker with an error message
func (r *PocketBaseSystemStatusRepo) DisableWorker(ctx context.Context, errorMsg string) error {
	query := `UPDATE system_status SET worker_enabled = FALSE, last_error = ?, updated = ? WHERE id = 1`
	_, err := r.app.DB().NewQuery(query).Execute(errorMsg, time.Now().UTC())
	return err
}

// UpdateError updates the last error message
func (r *PocketBaseSystemStatusRepo) UpdateError(ctx context.Context, errorMsg string) error {
	query := `UPDATE system_status SET last_error = ?, updated = ? WHERE id = 1`
	_, err := r.app.DB().NewQuery(query).Execute(errorMsg, time.Now().UTC())
	return err
}

// ClearError clears the error message
func (r *PocketBaseSystemStatusRepo) ClearError(ctx context.Context) error {
	query := `UPDATE system_status SET last_error = '', updated = ? WHERE id = 1`
	_, err := r.app.DB().NewQuery(query).Execute(time.Now().UTC())
	return err
}

// Helper function

func (r *PocketBaseSystemStatusRepo) mapToSystemStatus(raw dbx.NullStringMap) (*models.SystemStatus, error) {
	status := &models.SystemStatus{}

	// Parse ID
	if raw["id"].Valid {
		var id int
		json.Unmarshal([]byte(raw["id"].String), &id)
		status.ID = id
	}

	// Parse worker_enabled
	if raw["worker_enabled"].Valid {
		var enabled bool
		json.Unmarshal([]byte(raw["worker_enabled"].String), &enabled)
		status.WorkerEnabled = enabled
	}

	// Last error
	status.LastError = raw["last_error"].String

	// Parse timestamp
	if raw["updated"].Valid {
		t, _ := time.Parse(time.RFC3339, raw["updated"].String)
		status.Updated = t
	}

	return status, nil
}
package pocketbase

import (
	"context"
	"encoding/json"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/repository"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
)

// PocketBaseUserRepo implements UserRepository for PocketBase
type PocketBaseUserRepo struct {
	app *pocketbase.PocketBase
}

// Ensure implementation
var _ repository.UserRepository = (*PocketBaseUserRepo)(nil)

// NewPocketBaseUserRepo creates a new user repository
func NewPocketBaseUserRepo(app *pocketbase.PocketBase) repository.UserRepository {
	return &PocketBaseUserRepo{app: app}
}

// Create inserts a new user
func (r *PocketBaseUserRepo) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, fcm_token, is_fcm_active, created, updated)
		VALUES ({:id}, {:email}, {:fcm_token}, {:is_fcm_active}, {:created}, {:updated})
	`

	_, err := r.app.DB().NewQuery(query).Bind(dbx.Params{
		"id":            user.ID,
		"email":         user.Email,
		"fcm_token":     user.FCMToken,
		"is_fcm_active": user.IsFCMActive,
		"created":       time.Now().UTC(),
		"updated":       time.Now().UTC(),
	}).Execute()

	return err
}

// GetByID retrieves a user by ID
func (r *PocketBaseUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `SELECT * FROM users WHERE id = ?`

	var rawResult dbx.NullStringMap
	err := r.app.DB().NewQuery(query).One(&rawResult, id)
	if err != nil {
		return nil, err
	}

	return r.mapToUser(rawResult)
}

// GetByEmail retrieves a user by email
func (r *PocketBaseUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT * FROM users WHERE email = ?`

	var rawResult dbx.NullStringMap
	err := r.app.DB().NewQuery(query).One(&rawResult, email)
	if err != nil {
		return nil, err
	}

	return r.mapToUser(rawResult)
}

// Update updates user information
func (r *PocketBaseUserRepo) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET email = ?, fcm_token = ?, is_fcm_active = ?, updated = ?
		WHERE id = ?
	`

	_, err := r.app.DB().NewQuery(query).Execute(
		user.Email,
		user.FCMToken,
		user.IsFCMActive,
		time.Now().UTC(),
		user.ID,
	)

	return err
}

// UpdateFCMToken updates only the FCM token
func (r *PocketBaseUserRepo) UpdateFCMToken(ctx context.Context, userID, token string) error {
	query := `UPDATE users SET fcm_token = ?, is_fcm_active = TRUE, updated = ? WHERE id = ?`
	_, err := r.app.DB().NewQuery(query).Execute(token, time.Now().UTC(), userID)
	return err
}

// DisableFCM disables FCM for a user (token invalid)
func (r *PocketBaseUserRepo) DisableFCM(ctx context.Context, userID string) error {
	query := `UPDATE users SET is_fcm_active = FALSE, fcm_token = NULL, updated = ? WHERE id = ?`
	_, err := r.app.DB().NewQuery(query).Execute(time.Now().UTC(), userID)
	return err
}

// EnableFCM re-enables FCM with a new token
func (r *PocketBaseUserRepo) EnableFCM(ctx context.Context, userID string, token string) error {
	query := `UPDATE users SET fcm_token = ?, is_fcm_active = TRUE, updated = ? WHERE id = ?`
	_, err := r.app.DB().NewQuery(query).Execute(token, time.Now().UTC(), userID)
	return err
}

// GetActiveUsers retrieves all users with active FCM
func (r *PocketBaseUserRepo) GetActiveUsers(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT * FROM users 
		WHERE is_fcm_active = TRUE 
		  AND fcm_token IS NOT NULL 
		  AND fcm_token != ''
	`

	var rawResults []dbx.NullStringMap
	err := r.app.DB().NewQuery(query).All(&rawResults)
	if err != nil {
		return nil, err
	}

	return r.mapToUsers(rawResults)
}

// Helper functions

func (r *PocketBaseUserRepo) mapToUser(raw dbx.NullStringMap) (*models.User, error) {
	user := &models.User{}

	user.ID = raw["id"].String
	user.Email = raw["email"].String
	user.FCMToken = raw["fcm_token"].String

	// Parse boolean
	if raw["is_fcm_active"].Valid {
		var val bool
		json.Unmarshal([]byte(raw["is_fcm_active"].String), &val)
		user.IsFCMActive = val
	}

	// Parse timestamps
	if raw["created"].Valid {
		t, _ := time.Parse(time.RFC3339, raw["created"].String)
		user.Created = t
	}
	if raw["updated"].Valid {
		t, _ := time.Parse(time.RFC3339, raw["updated"].String)
		user.Updated = t
	}

	return user, nil
}

func (r *PocketBaseUserRepo) mapToUsers(rawList []dbx.NullStringMap) ([]*models.User, error) {
	users := make([]*models.User, 0, len(rawList))

	for _, raw := range rawList {
		user, err := r.mapToUser(raw)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
package services

import (
	"context"
	"errors"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// FCMService handles Firebase Cloud Messaging
type FCMService struct {
	client *messaging.Client
}

// NewFCMService creates a new FCM service
func NewFCMService(credentialsPath string) (*FCMService, error) {
	ctx := context.Background()

	// Initialize Firebase app
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	// Get messaging client
	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	return &FCMService{client: client}, nil
}

// SendNotification sends a notification to a device
func (s *FCMService) SendNotification(token, title, body string) error {
	if token == "" {
		return errors.New("token is empty")
	}

	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound: "default",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}

	// Send message
	_, err := s.client.Send(context.Background(), message)
	return err
}

// SendNotificationWithData sends a notification with custom data
func (s *FCMService) SendNotificationWithData(token, title, body string, data map[string]string) error {
	if token == "" {
		return errors.New("token is empty")
	}

	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound: "default",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}

	_, err := s.client.Send(context.Background(), message)
	return err
}

// SendMulticast sends the same notification to multiple devices
func (s *FCMService) SendMulticast(tokens []string, title, body string) (*messaging.BatchResponse, error) {
	if len(tokens) == 0 {
		return nil, errors.New("no tokens provided")
	}

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound: "default",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}

	return s.client.SendEachForMulticast(context.Background(), message)
}
package services

import (
	"time"
)

// LunarDate represents a date in lunar calendar
type LunarDate struct {
	Year  int
	Month int
	Day   int
}

// LunarCalendar handles lunar calendar conversions
type LunarCalendar struct {
	// TODO: Implement full lunar calendar algorithm
}

// NewLunarCalendar creates a new lunar calendar service
func NewLunarCalendar() *LunarCalendar {
	return &LunarCalendar{}
}

// SolarToLunar converts solar date to lunar date
func (lc *LunarCalendar) SolarToLunar(solar time.Time) LunarDate {
	// TODO: Implement proper solar to lunar conversion
	// This is a stub for now
	return LunarDate{
		Year:  solar.Year(),
		Month: int(solar.Month()),
		Day:   solar.Day(),
	}
}

// LunarToSolar converts lunar date to solar date
func (lc *LunarCalendar) LunarToSolar(year, month, day int) time.Time {
	// TODO: Implement proper lunar to solar conversion
	// This is a stub for now
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// GetLunarMonthDays returns number of days in a lunar month
func (lc *LunarCalendar) GetLunarMonthDays(year, month int) int {
	// TODO: Implement proper calculation
	// Lunar months have either 29 or 30 days
	// This is a stub that returns 30 for now
	return 30
}
package services

import (
	"context"
	"errors"
	"time"

	"remiaq/internal/models"
	"remiaq/internal/repository"

	"github.com/google/uuid"
)

// ReminderService handles reminder business logic
type ReminderService struct {
	reminderRepo    repository.ReminderRepository
	userRepo        repository.UserRepository
	fcmService      *FCMService
	schedCalculator *ScheduleCalculator
}

// NewReminderService creates a new reminder service
func NewReminderService(
	reminderRepo repository.ReminderRepository,
	userRepo repository.UserRepository,
	fcmService *FCMService,
	schedCalculator *ScheduleCalculator,
) *ReminderService {
	return &ReminderService{
		reminderRepo:    reminderRepo,
		userRepo:        userRepo,
		fcmService:      fcmService,
		schedCalculator: schedCalculator,
	}
}

// CreateReminder creates a new reminder
func (s *ReminderService) CreateReminder(ctx context.Context, reminder *models.Reminder) error {
	// Validate
	if err := reminder.Validate(); err != nil {
		return err
	}

	// Generate ID if not provided
	if reminder.ID == "" {
		reminder.ID = uuid.New().String()
	}

	// Set default values
	if reminder.Status == "" {
		reminder.Status = models.ReminderStatusActive
	}
	if reminder.RepeatStrategy == "" {
		reminder.RepeatStrategy = models.RepeatStrategyNone
	}
	if reminder.CalendarType == "" {
		reminder.CalendarType = models.CalendarTypeSolar
	}

	// Calculate next trigger time if not set
	if reminder.NextTriggerAt.IsZero() {
		nextTrigger, err := s.schedCalculator.CalculateNextTrigger(reminder, time.Now())
		if err != nil {
			return err
		}
		reminder.NextTriggerAt = nextTrigger
	}

	return s.reminderRepo.Create(ctx, reminder)
}

// GetReminder retrieves a reminder by ID
func (s *ReminderService) GetReminder(ctx context.Context, id string) (*models.Reminder, error) {
	return s.reminderRepo.GetByID(ctx, id)
}

// UpdateReminder updates a reminder
func (s *ReminderService) UpdateReminder(ctx context.Context, reminder *models.Reminder) error {
	if err := reminder.Validate(); err != nil {
		return err
	}

	return s.reminderRepo.Update(ctx, reminder)
}

// DeleteReminder deletes a reminder
func (s *ReminderService) DeleteReminder(ctx context.Context, id string) error {
	return s.reminderRepo.Delete(ctx, id)
}

// GetUserReminders retrieves all reminders for a user
func (s *ReminderService) GetUserReminders(ctx context.Context, userID string) ([]*models.Reminder, error) {
	return s.reminderRepo.GetByUserID(ctx, userID)
}

// SnoozeReminder postpones a reminder
func (s *ReminderService) SnoozeReminder(ctx context.Context, id string, duration time.Duration) error {
	snoozeUntil := time.Now().Add(duration)
	return s.reminderRepo.UpdateSnooze(ctx, id, &snoozeUntil)
}

// CompleteReminder marks a reminder as completed
func (s *ReminderService) CompleteReminder(ctx context.Context, id string) error {
	reminder, err := s.reminderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()

	// For one-time reminders, mark as completed
	if reminder.Type == models.ReminderTypeOneTime {
		return s.reminderRepo.MarkCompleted(ctx, id, now)
	}

	// For recurring reminders with base_on=completion
	if reminder.RecurrencePattern != nil &&
		reminder.RecurrencePattern.BaseOn == models.BaseOnCompletion {
		// Calculate next trigger from completion time
		nextTrigger, err := s.schedCalculator.CalculateNextTrigger(reminder, now)
		if err != nil {
			return err
		}

		// Update last_completed_at and next_trigger_at
		reminder.LastCompletedAt = &now
		reminder.NextTriggerAt = nextTrigger
		return s.reminderRepo.Update(ctx, reminder)
	}

	// For other recurring reminders, just update last_completed_at
	reminder.LastCompletedAt = &now
	return s.reminderRepo.Update(ctx, reminder)
}

// ProcessDueReminders processes all reminders that are due (called by worker)
func (s *ReminderService) ProcessDueReminders(ctx context.Context) error {
	now := time.Now()

	// Get all due reminders
	reminders, err := s.reminderRepo.GetDueReminders(ctx, now)
	if err != nil {
		return err
	}

	for _, reminder := range reminders {
		// Process each reminder
		if err := s.processReminder(ctx, reminder, now); err != nil {
			// Log error but continue with other reminders
			continue
		}
	}

	return nil
}

// processReminder processes a single reminder
func (s *ReminderService) processReminder(ctx context.Context, reminder *models.Reminder, now time.Time) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, reminder.UserID)
	if err != nil {
		return err
	}

	// Check if user has active FCM
	if !user.IsFCMActive || user.FCMToken == "" {
		return errors.New("user FCM not active")
	}

	// Send notification
	err = s.fcmService.SendNotification(user.FCMToken, reminder.Title, reminder.Description)
	if err != nil {
		// Handle FCM errors
		if isTokenInvalidError(err) {
			// Disable FCM for this user
			s.userRepo.DisableFCM(ctx, user.ID)
		}
		return err
	}

	// Update last_sent_at
	s.reminderRepo.UpdateLastSent(ctx, reminder.ID, now)

	// Handle based on type
	if reminder.Type == models.ReminderTypeOneTime {
		return s.handleOneTimeReminder(ctx, reminder, now)
	} else {
		return s.handleRecurringReminder(ctx, reminder, now)
	}
}

// handleOneTimeReminder handles one-time reminder logic
func (s *ReminderService) handleOneTimeReminder(ctx context.Context, reminder *models.Reminder, now time.Time) error {
	// Check if should retry
	if reminder.RepeatStrategy == models.RepeatStrategyRetryUntilComplete && reminder.IsRetryable() {
		// Increment retry count
		s.reminderRepo.IncrementRetryCount(ctx, reminder.ID)

		// Calculate next retry time
		nextRetry := now.Add(time.Duration(reminder.RetryIntervalSec) * time.Second)
		return s.reminderRepo.UpdateNextTrigger(ctx, reminder.ID, nextRetry)
	}

	// Otherwise, mark as completed
	return s.reminderRepo.MarkCompleted(ctx, reminder.ID, now)
}

// handleRecurringReminder handles recurring reminder logic
func (s *ReminderService) handleRecurringReminder(ctx context.Context, reminder *models.Reminder, now time.Time) error {
	// Calculate next trigger
	nextTrigger, err := s.schedCalculator.CalculateNextTrigger(reminder, now)
	if err != nil {
		return err
	}

	// Update next trigger time
	return s.reminderRepo.UpdateNextTrigger(ctx, reminder.ID, nextTrigger)
}

// Helper function to check if FCM error is due to invalid token
func isTokenInvalidError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return errStr == "UNREGISTERED" ||
		errStr == "INVALID_ARGUMENT" ||
		errStr == "NOT_FOUND"
}
package services

import (
	"errors"
	"time"

	"remiaq/internal/models"
)

// ScheduleCalculator calculates next trigger times for reminders
type ScheduleCalculator struct {
	lunarCalendar *LunarCalendar
}

// NewScheduleCalculator creates a new schedule calculator
func NewScheduleCalculator(lunarCalendar *LunarCalendar) *ScheduleCalculator {
	return &ScheduleCalculator{
		lunarCalendar: lunarCalendar,
	}
}

// CalculateNextTrigger calculates the next trigger time for a reminder
func (c *ScheduleCalculator) CalculateNextTrigger(reminder *models.Reminder, fromTime time.Time) (time.Time, error) {
	if reminder.Type == models.ReminderTypeOneTime {
		return c.calculateOneTime(reminder, fromTime)
	}

	return c.calculateRecurring(reminder, fromTime)
}

// calculateOneTime calculates next trigger for one-time reminders
func (c *ScheduleCalculator) calculateOneTime(reminder *models.Reminder, fromTime time.Time) (time.Time, error) {
	// For one-time reminders, return the set trigger time
	if !reminder.NextTriggerAt.IsZero() {
		return reminder.NextTriggerAt, nil
	}

	return fromTime, nil
}

// calculateRecurring calculates next trigger for recurring reminders
func (c *ScheduleCalculator) calculateRecurring(reminder *models.Reminder, fromTime time.Time) (time.Time, error) {
	if reminder.RecurrencePattern == nil {
		return time.Time{}, errors.New("recurrence pattern is required for recurring reminders")
	}

	pattern := reminder.RecurrencePattern

	// Handle interval-based recurrence
	if pattern.IntervalSeconds > 0 {
		return c.calculateIntervalBased(reminder, fromTime)
	}

	// Handle calendar-based recurrence
	switch pattern.Type {
	case models.RecurrenceTypeDaily:
		return c.calculateDaily(reminder, fromTime)
	case models.RecurrenceTypeWeekly:
		return c.calculateWeekly(reminder, fromTime)
	case models.RecurrenceTypeMonthly:
		return c.calculateMonthly(reminder, fromTime)
	case models.RecurrenceTypeLunarLastDayOfMonth:
		return c.calculateLunarLastDay(reminder, fromTime)
	default:
		return time.Time{}, errors.New("unsupported recurrence type")
	}
}

// calculateIntervalBased calculates next trigger based on interval
func (c *ScheduleCalculator) calculateIntervalBased(reminder *models.Reminder, fromTime time.Time) (time.Time, error) {
	interval := time.Duration(reminder.RecurrencePattern.IntervalSeconds) * time.Second

	// If base_on is completion, calculate from completion time
	if reminder.RecurrencePattern.BaseOn == models.BaseOnCompletion {
		if reminder.LastCompletedAt != nil {
			return reminder.LastCompletedAt.Add(interval), nil
		}
		// If never completed, use creation time
		return reminder.Created.Add(interval), nil
	}

	// Otherwise, calculate from last trigger time (creation-based)
	return fromTime.Add(interval), nil
}

// calculateDaily calculates next daily trigger
func (c *ScheduleCalculator) calculateDaily(reminder *models.Reminder, fromTime time.Time) (time.Time, error) {
	if reminder.TriggerTimeOfDay == "" {
		return time.Time{}, errors.New("trigger_time_of_day is required for daily recurrence")
	}

	// Parse time of day (HH:MM format)
	targetTime, err := parseTimeOfDay(reminder.TriggerTimeOfDay)
	if err != nil {
		return time.Time{}, err
	}

	// Calculate next occurrence
	next := time.Date(
		fromTime.Year(), fromTime.Month(), fromTime.Day(),
		targetTime.Hour(), targetTime.Minute(), 0, 0,
		fromTime.Location(),
	)

	// If the time has passed today, move to tomorrow
	if next.Before(fromTime) || next.Equal(fromTime) {
		next = next.Add(24 * time.Hour)
	}

	return next, nil
}

// calculateWeekly calculates next weekly trigger
func (c *ScheduleCalculator) calculateWeekly(reminder *models.Reminder, fromTime time.Time) (time.Time, error) {
	if reminder.TriggerTimeOfDay == "" {
		return time.Time{}, errors.New("trigger_time_of_day is required for weekly recurrence")
	}

	pattern := reminder.RecurrencePattern
	targetWeekday := time.Weekday(pattern.DayOfWeek)

	// Parse time of day
	targetTime, err := parseTimeOfDay(reminder.TriggerTimeOfDay)
	if err != nil {
		return time.Time{}, err
	}

	// Find next occurrence of target weekday
	daysUntilTarget := (int(targetWeekday) - int(fromTime.Weekday()) + 7) % 7
	if daysUntilTarget == 0 {
		// It's the target day, check if time has passed
		next := time.Date(
			fromTime.Year(), fromTime.Month(), fromTime.Day(),
			targetTime.Hour(), targetTime.Minute(), 0, 0,
			fromTime.Location(),
		)
		if next.After(fromTime) {
			return next, nil
		}
		daysUntilTarget = 7
	}

	next := fromTime.Add(time.Duration(daysUntilTarget) * 24 * time.Hour)
	next = time.Date(
		next.Year(), next.Month(), next.Day(),
		targetTime.Hour(), targetTime.Minute(), 0, 0,
		next.Location(),
	)

	return next, nil
}

// calculateMonthly calculates next monthly trigger
func (c *ScheduleCalculator) calculateMonthly(reminder *models.Reminder, fromTime time.Time) (time.Time, error) {
	pattern := reminder.RecurrencePattern

	if reminder.CalendarType == models.CalendarTypeLunar {
		return c.calculateLunarMonthly(reminder, fromTime)
	}

	// Solar calendar
	if reminder.TriggerTimeOfDay == "" {
		return time.Time{}, errors.New("trigger_time_of_day is required for monthly recurrence")
	}

	targetTime, err := parseTimeOfDay(reminder.TriggerTimeOfDay)
	if err != nil {
		return time.Time{}, err
	}

	// Try current month first
	next := time.Date(
		fromTime.Year(), fromTime.Month(), pattern.DayOfMonth,
		targetTime.Hour(), targetTime.Minute(), 0, 0,
		fromTime.Location(),
	)

	// If date doesn't exist in current month or has passed, move to next month
	if next.Day() != pattern.DayOfMonth || next.Before(fromTime) || next.Equal(fromTime) {
		next = time.Date(
			fromTime.Year(), fromTime.Month()+1, pattern.DayOfMonth,
			targetTime.Hour(), targetTime.Minute(), 0, 0,
			fromTime.Location(),
		)

		// Handle year rollover
		if next.Day() != pattern.DayOfMonth {
			next = time.Date(
				fromTime.Year()+1, time.January, pattern.DayOfMonth,
				targetTime.Hour(), targetTime.Minute(), 0, 0,
				fromTime.Location(),
			)
		}
	}

	return next, nil
}

// calculateLunarMonthly calculates next lunar monthly trigger
func (c *ScheduleCalculator) calculateLunarMonthly(reminder *models.Reminder, fromTime time.Time) (time.Time, error) {
	pattern := reminder.RecurrencePattern
	targetDay := pattern.DayOfMonth

	// Convert current solar date to lunar
	lunarDate := c.lunarCalendar.SolarToLunar(fromTime)

	// Try to find the target day in current or next lunar month
	for i := 0; i < 13; i++ { // Max 13 lunar months in a year
		// Check if target day exists in current lunar month
		daysInMonth := c.lunarCalendar.GetLunarMonthDays(lunarDate.Year, lunarDate.Month)

		if targetDay <= daysInMonth {
			// Convert lunar date to solar
			solarDate := c.lunarCalendar.LunarToSolar(lunarDate.Year, lunarDate.Month, targetDay)

			// Apply time of day
			if reminder.TriggerTimeOfDay != "" {
				targetTime, _ := parseTimeOfDay(reminder.TriggerTimeOfDay)
				solarDate = time.Date(
					solarDate.Year(), solarDate.Month(), solarDate.Day(),
					targetTime.Hour(), targetTime.Minute(), 0, 0,
					solarDate.Location(),
				)
			}

			// If this date is in the future, return it
			if solarDate.After(fromTime) {
				return solarDate, nil
			}
		}

		// Move to next lunar month
		lunarDate.Month++
		if lunarDate.Month > 12 {
			lunarDate.Month = 1
			lunarDate.Year++
		}
	}

	return time.Time{}, errors.New("failed to calculate next lunar monthly trigger")
}

// calculateLunarLastDay calculates last day of lunar month
func (c *ScheduleCalculator) calculateLunarLastDay(reminder *models.Reminder, fromTime time.Time) (time.Time, error) {
	lunarDate := c.lunarCalendar.SolarToLunar(fromTime)

	// Try current month first
	daysInMonth := c.lunarCalendar.GetLunarMonthDays(lunarDate.Year, lunarDate.Month)
	solarDate := c.lunarCalendar.LunarToSolar(lunarDate.Year, lunarDate.Month, daysInMonth)

	if reminder.TriggerTimeOfDay != "" {
		targetTime, _ := parseTimeOfDay(reminder.TriggerTimeOfDay)
		solarDate = time.Date(
			solarDate.Year(), solarDate.Month(), solarDate.Day(),
			targetTime.Hour(), targetTime.Minute(), 0, 0,
			solarDate.Location(),
		)
	}

	if solarDate.After(fromTime) {
		return solarDate, nil
	}

	// Try next month
	lunarDate.Month++
	if lunarDate.Month > 12 {
		lunarDate.Month = 1
		lunarDate.Year++
	}

	daysInMonth = c.lunarCalendar.GetLunarMonthDays(lunarDate.Year, lunarDate.Month)
	solarDate = c.lunarCalendar.LunarToSolar(lunarDate.Year, lunarDate.Month, daysInMonth)

	if reminder.TriggerTimeOfDay != "" {
		targetTime, _ := parseTimeOfDay(reminder.TriggerTimeOfDay)
		solarDate = time.Date(
			solarDate.Year(), solarDate.Month(), solarDate.Day(),
			targetTime.Hour(), targetTime.Minute(), 0, 0,
			solarDate.Location(),
		)
	}

	return solarDate, nil
}

// parseTimeOfDay parses HH:MM format
func parseTimeOfDay(timeStr string) (time.Time, error) {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}, errors.New("invalid time format, expected HH:MM")
	}
	return t, nil
}
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
