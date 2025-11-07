package main

import (
	"context"
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase"
	pbRepo "remiaq/internal/repository/pocketbase"
	"remiaq/internal/services"
)

func main() {
	// Khởi tạo PocketBase app
	app := pocketbase.New()

	// Khởi tạo repository và service
	reminderRepo := pbRepo.NewReminderORMRepo(app)
	userRepo := pbRepo.NewUserORMRepo(app)
	schedCalculator := services.NewScheduleCalculator(services.NewLunarCalendar())
	service := services.NewReminderService(reminderRepo, userRepo, nil, schedCalculator)

	// Kiểm tra reminder trước khi complete
	fmt.Println("=== TRƯỚC KHI COMPLETE ===")
	reminder, err := reminderRepo.GetByID(context.Background(), "kp80h5istz5ob3y")
	if err != nil {
		log.Fatalf("Lỗi khi lấy reminder: %v", err)
	}
	fmt.Printf("ID: %s\n", reminder.ID)
	fmt.Printf("Status: %s\n", reminder.Status)
	fmt.Printf("LastCompletedAt: %s\n", reminder.LastCompletedAt)
	fmt.Printf("NextTriggerAt: %s\n", reminder.NextTriggerAt)
	fmt.Printf("Type: %s\n", reminder.Type)

	// Thực hiện complete reminder
	fmt.Println("\n=== THỰC HIỆN COMPLETE ===")
	err = service.CompleteReminder(context.Background(), "kp80h5istz5ob3y")
	if err != nil {
		log.Fatalf("Lỗi khi complete reminder: %v", err)
	}
	fmt.Println("Complete reminder thành công")

	// Kiểm tra reminder sau khi complete
	fmt.Println("\n=== SAU KHI COMPLETE ===")
	reminder, err = reminderRepo.GetByID(context.Background(), "kp80h5istz5ob3y")
	if err != nil {
		log.Fatalf("Lỗi khi lấy reminder sau khi complete: %v", err)
	}
	fmt.Printf("ID: %s\n", reminder.ID)
	fmt.Printf("Status: %s\n", reminder.Status)
	fmt.Printf("LastCompletedAt: %s\n", reminder.LastCompletedAt)
	fmt.Printf("NextTriggerAt: %s\n", reminder.NextTriggerAt)
	fmt.Printf("Type: %s\n", reminder.Type)

	fmt.Println("\n=== TEST HOÀN TẤT ===")
}