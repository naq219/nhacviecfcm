package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"remiaq/config"
	"remiaq/internal/handlers"
	"remiaq/internal/middleware"
	pbRepo "remiaq/internal/repository/pocketbase"
	"remiaq/internal/services"
	"remiaq/internal/worker"

	// Swagger docs
	_ "remiaq/docs"

	// Import migrations package để PocketBase load migrations
	_ "remiaq/migrations"
)

//	@title			RemiAq API
//	@version		1.0
//	@description	RemiAq - Reminder & Lunar Calendar API
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8090
//	@basePath	/
//	@schemes	http https

//	@securityDefinitions.basic	BasicAuth
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization

//	@externalDocs.description	OpenAPI
//	@externalDocs.url			https://swagger.io/resources/open-api/

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set PocketBase server address
	os.Setenv("PB_ADDR", cfg.ServerAddr)
	// Disable debug logging to reduce SQL query logs
	os.Setenv("PB_DEBUG", "false")

	// Create PocketBase instance
	app := pocketbase.New()

	// Initialize repositories (using ORM implementations)
	reminderRepo := pbRepo.NewReminderORMRepo(app)
	userRepo := pbRepo.NewUserORMRepo(app)
	queryRepo := pbRepo.NewQueryRepo(app)

	// Initialize services
	var fcmService *services.FCMService
	if _, err := os.Stat(cfg.FCMCredentials); err == nil {
		fcmService, err = services.NewFCMService(cfg.FCMCredentials)
		if err != nil {
			log.Printf("Warning: Failed to initialize FCM service: %v", err)
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
	sysRepo := pbRepo.NewSystemStatusORMRepo(app)
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
		//	@Summary		Health check
		//	@Description	Kiểm tra server có chạy hay không
		//	@Tags			system
		//	@Produce		plain
		//	@Success		200	{string}	string	"RemiAq API is running!"
		//	@Router			/hello [get]
		se.Router.GET("/hello", func(re *core.RequestEvent) error {
			middleware.SetCORSHeaders(re)
			return re.String(200, "RemiAq API is running!")
		})

		// Raw SQL query endpoints
		se.Router.GET("/api/rquery", queryHandler.HandleSelect)
		se.Router.POST("/api/rquery", queryHandler.HandleSelect)

		se.Router.GET("/api/rinsert", queryHandler.HandleInsert)
		se.Router.POST("/api/rinsert", queryHandler.HandleInsert)

		se.Router.GET("/api/rupdate", queryHandler.HandleUpdate)
		se.Router.PUT("/api/rupdate", queryHandler.HandleUpdate)

		se.Router.GET("/api/rdelete", queryHandler.HandleDelete)
		se.Router.DELETE("/api/rdelete", queryHandler.HandleDelete)

		// // --- Temporary/Public endpoints ---
		// tmpApi := se.Router.Group("/api/tmp")
		// tmpApi.POST("/reminders", reminderHandler.CreateReminder)
		// tmpApi.GET("/reminders/{id}", reminderHandler.GetReminder)
		// tmpApi.PUT("/reminders/{id}", reminderHandler.UpdateReminder)
		// tmpApi.DELETE("/reminders/{id}", reminderHandler.DeleteReminder)
		// tmpApi.GET("/users/{userId}/reminders", reminderHandler.GetUserReminders)
		// tmpApi.POST("/reminders/{id}/snooze", reminderHandler.SnoozeReminder)
		// tmpApi.POST("/reminders/{id}/complete", reminderHandler.CompleteReminder)

		// --- Auth-protected endpoints (PocketBase built-in auth) ---
		api := se.Router.Group("/api")
		api.Bind(apis.RequireAuth())
		api.POST("/reminders", reminderHandler.CreateReminder)
		api.GET("/reminders/mine", reminderHandler.GetCurrentUserReminders) // New route
		api.GET("/reminders/{id}", reminderHandler.GetReminder)
		api.PUT("/reminders/{id}", reminderHandler.UpdateReminder)
		api.DELETE("/reminders/{id}", reminderHandler.DeleteReminder)
		api.GET("/users/{userId}/reminders", reminderHandler.GetUserReminders)
		api.POST("/reminders/{id}/snooze", reminderHandler.SnoozeReminder)
		api.POST("/reminders/{id}/complete", reminderHandler.CompleteReminder)

		// System status API
		se.Router.GET("/api/system_status", sysHandler.GetSystemStatus)
		se.Router.PUT("/api/system_status", sysHandler.PutSystemStatus)

		// HTML test pages
		se.Router.GET("/test/system-status", func(re *core.RequestEvent) error {
			middleware.SetCORSHeaders(re)
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
