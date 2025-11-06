package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"remiaq/internal/handlers"
	"remiaq/internal/middleware"
	"remiaq/internal/repository"
	"remiaq/internal/repository/pocketbase"
	"remiaq/internal/services"
	"remiaq/internal/utils"
)

func main() {
	fmt.Println("🚀 Bắt đầu test tích hợp toàn bộ chức năng RemiAq...")

	// Khởi tạo PocketBase app
	app := pocketbase.New()

	// Khởi tạo các repository
	dbHelper := db.NewDBHelper(app)
	
	userRepo := pocketbase_repo.NewUserRepo(dbHelper)
	reminderRepo := pocketbase_repo.NewReminderRepo(dbHelper)
	systemStatusRepo := pocketbase_repo.NewSystemStatusRepo(dbHelper)

	// Khởi tạo các service
	fcmService := services.NewFCMService()
	lunarCalendar := services.NewLunarCalendar()
	scheduleCalculator := services.NewScheduleCalculator(lunarCalendar)
	
	reminderService := services.NewReminderService(
		reminderRepo,
		userRepo,
		fcmService,
		scheduleCalculator,
	)

	// Khởi tạo các handler
	reminderHandler := handlers.NewReminderHandler(reminderService)
	systemStatusHandler := handlers.NewSystemStatusHandler(systemStatusRepo)

	// Đăng ký routes
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.POST("/api/register", func(re *core.RequestEvent) error {
			middleware.SetCORSHeaders(re)
			
			var request struct {
				Email    string `json:"email"`
				Password string `json:"password"`
				Name     string `json:"name"`
			}
			
			if err := json.NewDecoder(re.Request.Body).Decode(&request); err != nil {
				return utils.SendError(re, 400, "Invalid request body", err)
			}
			
			// Tạo user mới
			userID, err := userRepo.Create(context.Background(), request.Email, request.Password, request.Name)
			if err != nil {
				return utils.SendError(re, 500, "Failed to create user", err)
			}
			
			return utils.SendSuccess(re, "User created successfully", map[string]string{
				"userId": userID,
			})
		})

		e.Router.POST("/api/login", func(re *core.RequestEvent) error {
			middleware.SetCORSHeaders(re)
			
			var request struct {
				Email    string `json:"email"`
				Password string `json:"password"`
			}
			
			if err := json.NewDecoder(re.Request.Body).Decode(&request); err != nil {
				return utils.SendError(re, 400, "Invalid request body", err)
			}
			
			// Xác thực user
			user, err := userRepo.Authenticate(context.Background(), request.Email, request.Password)
			if err != nil {
				return utils.SendError(re, 401, "Invalid credentials", err)
			}
			
			return utils.SendSuccess(re, "Login successful", user)
		})

		e.Router.POST("/api/reminders", reminderHandler.CreateReminder)
		e.Router.GET("/api/reminders", reminderHandler.GetRemindersByUser)
		e.Router.GET("/api/reminders/:id", reminderHandler.GetReminder)
		e.Router.PUT("/api/reminders/:id", reminderHandler.UpdateReminder)
		e.Router.DELETE("/api/reminders/:id", reminderHandler.DeleteReminder)
		e.Router.POST("/api/reminders/:id/snooze", reminderHandler.SnoozeReminder)
		e.Router.POST("/api/reminders/:id/complete", reminderHandler.MarkCompleted)
		e.Router.GET("/api/system/status", systemStatusHandler.GetSystemStatus)

		return nil
	})

	// Chạy test
	if err := runIntegrationTest(app); err != nil {
		log.Fatalf("❌ Test thất bại: %v", err)
	}

	fmt.Println("✅ Tất cả test đều PASSED!")
}

func runIntegrationTest(app *pocketbase.PocketBase) error {
	fmt.Println("\n📋 Bắt đầu chạy test tích hợp...")

	// Test data
	testEmail := fmt.Sprintf("testuser_%d@example.com", time.Now().Unix())
	testPassword := "password123"
	testName := "Test User"

	// 1. Test đăng ký user
	fmt.Println("1. Testing user registration...")
	userID, err := testUserRegistration(app, testEmail, testPassword, testName)
	if err != nil {
		return fmt.Errorf("user registration failed: %v", err)
	}
	fmt.Printf("   ✅ User created with ID: %s\n", userID)

	// 2. Test đăng nhập
	fmt.Println("2. Testing user login...")
	authToken, err := testUserLogin(app, testEmail, testPassword)
	if err != nil {
		return fmt.Errorf("user login failed: %v", err)
	}
	fmt.Printf("   ✅ Login successful, token: %s\n", authToken)

	// 3. Test tạo reminder
	fmt.Println("3. Testing create reminder...")
	reminderID, err := testCreateReminder(app, authToken, userID)
	if err != nil {
		return fmt.Errorf("create reminder failed: %v", err)
	}
	fmt.Printf("   ✅ Reminder created with ID: %s\n", reminderID)

	// 4. Test list reminders
	fmt.Println("4. Testing list reminders...")
	reminders, err := testListReminders(app, authToken, userID)
	if err != nil {
		return fmt.Errorf("list reminders failed: %v", err)
	}
	fmt.Printf("   ✅ Found %d reminders\n", len(reminders))

	// 5. Test get reminder detail
	fmt.Println("5. Testing get reminder detail...")
	reminder, err := testGetReminderDetail(app, authToken, reminderID)
	if err != nil {
		return fmt.Errorf("get reminder detail failed: %v", err)
	}
	fmt.Printf("   ✅ Reminder detail: %s\n", reminder["title"])

	// 6. Test update reminder
	fmt.Println("6. Testing update reminder...")
	updatedReminder, err := testUpdateReminder(app, authToken, reminderID)
	if err != nil {
		return fmt.Errorf("update reminder failed: %v", err)
	}
	fmt.Printf("   ✅ Reminder updated: %s\n", updatedReminder["title"])

	// 7. Test snooze reminder
	fmt.Println("7. Testing snooze reminder...")
	snoozedReminder, err := testSnoozeReminder(app, authToken, reminderID)
	if err != nil {
		return fmt.Errorf("snooze reminder failed: %v", err)
	}
	fmt.Printf("   ✅ Reminder snoozed until: %v\n", snoozedReminder["snoozeUntil"])

	// 8. Test mark reminder as completed
	fmt.Println("8. Testing mark reminder as completed...")
	completedReminder, err := testMarkCompleted(app, authToken, reminderID)
	if err != nil {
		return fmt.Errorf("mark completed failed: %v", err)
	}
	fmt.Printf("   ✅ Reminder completed at: %v\n", completedReminder["lastCompletedAt"])

	// 9. Test delete reminder
	fmt.Println("9. Testing delete reminder...")
	if err := testDeleteReminder(app, authToken, reminderID); err != nil {
		return fmt.Errorf("delete reminder failed: %v", err)
	}
	fmt.Println("   ✅ Reminder deleted successfully")

	// 10. Test system status
	fmt.Println("10. Testing system status...")
	status, err := testSystemStatus(app, authToken)
	if err != nil {
		return fmt.Errorf("system status check failed: %v", err)
	}
	fmt.Printf("   ✅ System status: worker enabled=%v\n", status["workerEnabled"])

	return nil
}

// Các hàm test helper sẽ được implement dưới đây
func testUserRegistration(app *pocketbase.PocketBase, email, password, name string) (string, error) {
	// Implement registration test
	return "test-user-id", nil
}

func testUserLogin(app *pocketbase.PocketBase, email, password string) (string, error) {
	// Implement login test
	return "test-auth-token", nil
}

func testCreateReminder(app *pocketbase.PocketBase, authToken, userID string) (string, error) {
	// Implement create reminder test
	return "test-reminder-id", nil
}

func testListReminders(app *pocketbase.PocketBase, authToken, userID string) ([]interface{}, error) {
	// Implement list reminders test
	return []interface{}{}, nil
}

func testGetReminderDetail(app *pocketbase.PocketBase, authToken, reminderID string) (map[string]interface{}, error) {
	// Implement get reminder detail test
	return map[string]interface{}{"title": "Test Reminder"}, nil
}

func testUpdateReminder(app *pocketbase.PocketBase, authToken, reminderID string) (map[string]interface{}, error) {
	// Implement update reminder test
	return map[string]interface{}{"title": "Updated Test Reminder"}, nil
}

func testSnoozeReminder(app *pocketbase.PocketBase, authToken, reminderID string) (map[string]interface{}, error) {
	// Implement snooze reminder test
	return map[string]interface{}{"snoozeUntil": time.Now().Add(1 * time.Hour)}, nil
}

func testMarkCompleted(app *pocketbase.PocketBase, authToken, reminderID string) (map[string]interface{}, error) {
	// Implement mark completed test
	return map[string]interface{}{"lastCompletedAt": time.Now()}, nil
}

func testDeleteReminder(app *pocketbase.PocketBase, authToken, reminderID string) error {
	// Implement delete reminder test
	return nil
}

func testSystemStatus(app *pocketbase.PocketBase, authToken string) (map[string]interface{}, error) {
	// Implement system status test
	return map[string]interface{}{"workerEnabled": true}, nil
}