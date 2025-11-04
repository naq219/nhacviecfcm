package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/apis"

	"remiaq/config"
	"remiaq/internal/handlers" // ← Đã sửa từ api/handlers
	"remiaq/internal/middleware"
	pbRepo "remiaq/internal/repository/pocketbase"
	"remiaq/internal/services"
	"remiaq/internal/worker"

	// Import migrations package để PocketBase load migrations
	_ "remiaq/migrations"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set PocketBase server address
	os.Setenv("PB_ADDR", cfg.ServerAddr)

	// Create PocketBase instance
	app := pocketbase.New()

	// Initialize repositories
	reminderRepo := pbRepo.NewReminderRepo(app)
	userRepo := pbRepo.NewUserRepo(app)
	queryRepo := pbRepo.NewQueryRepo(app)

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

	// Initialize system status repo and start background worker
	sysRepo := pbRepo.NewSystemStatusRepo(app)
	sysHandler := handlers.NewSystemStatusHandler(sysRepo)
	bgCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w := worker.NewWorker(sysRepo, reminderService, time.Duration(cfg.WorkerInterval)*time.Second)
	w.Start(bgCtx)

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

		// Auth-protected endpoints (PocketBase built-in auth)
		secure := se.Router.Group("/api/secure")
		secure.Bind(apis.RequireAuth())
		secure.POST("/reminders", reminderHandler.CreateReminder)
		secure.GET("/reminders/{id}", reminderHandler.GetReminder)
		secure.PUT("/reminders/{id}", reminderHandler.UpdateReminder)
		secure.DELETE("/reminders/{id}", reminderHandler.DeleteReminder)
		secure.GET("/users/{userId}/reminders", reminderHandler.GetUserReminders)
		secure.POST("/reminders/{id}/snooze", reminderHandler.SnoozeReminder)
		secure.POST("/reminders/{id}/complete", reminderHandler.CompleteReminder)

		// User reminders
		se.Router.GET("/api/users/{userId}/reminders", reminderHandler.GetUserReminders)

		// Reminder actions
		se.Router.POST("/api/reminders/{id}/snooze", reminderHandler.SnoozeReminder)
		se.Router.POST("/api/reminders/{id}/complete", reminderHandler.CompleteReminder)

		// System status API
		se.Router.GET("/api/system_status", sysHandler.GetSystemStatus)
		se.Router.PUT("/api/system_status", sysHandler.PutSystemStatus)

		// HTML test pages
		se.Router.GET("/test/system-status", func(re *core.RequestEvent) error {
			middleware.SetCORSHeaders(re)
			// Đọc file HTML tĩnh
			content, err := os.ReadFile("web/system_status_test.html")
			if err != nil {
				return re.String(404, "Test page not found")
			}
			re.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			return re.String(200, string(content))
		})

		// Comprehensive RemiAq test page
		se.Router.GET("/test", func(re *core.RequestEvent) error {
			middleware.SetCORSHeaders(re)
			content, err := os.ReadFile("web/remiaq_test.html")
			if err != nil {
				return re.String(404, "RemiAq test page not found")
			}
			re.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			return re.String(200, string(content))
		})

		se.Router.GET("/", func(re *core.RequestEvent) error {
			middleware.SetCORSHeaders(re)
			return re.Redirect(302, "/test")
		})

		return se.Next()
	})

	// Start server
	log.Printf("Starting RemiAq API server on %s", cfg.ServerAddr)
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
