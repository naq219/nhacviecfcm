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
