package main

import (
	"fmt"
	"time"
	
	"remiaq/internal/models"
	"github.com/pocketbase/pocketbase/core"
)

// Hàm test để kiểm tra chuyển đổi reminder sang record
func testReminderToRecord() {
	// Tạo một reminder với next_recurring cụ thể
	reminder := &models.Reminder{
		Title:       "Test Reminder",
		Description: "Test conversion", 
		Type:        "recurring",
		NextRecurring: time.Date(2025, 11, 8, 8, 5, 0, 0, time.UTC),
	}
	
	// Tạo record
	record := core.NewRecord(&core.Collection{})
	
	// Giả lập hàm reminderToRecord (copy từ reminder_orm_repo.go)
	if !reminder.NextRecurring.IsZero() {
		record.Set("next_recurring", reminder.NextRecurring.Format(time.RFC3339Nano))
	} else {
		record.Set("next_recurring", nil)
	}
	
	// Kiểm tra kết quả
	nextRecurringValue := record.GetString("next_recurring")
	fmt.Printf("NextRecurring trong record: %s\n", nextRecurringValue)
	fmt.Printf("NextRecurring trong reminder: %s\n", reminder.NextRecurring.Format(time.RFC3339Nano))
	
	// Kiểm tra xem có giống nhau không
	if nextRecurringValue == reminder.NextRecurring.Format(time.RFC3339Nano) {
		fmt.Println("✅ Chuyển đổi thành công - next_recurring được giữ nguyên")
	} else {
		fmt.Println("❌ Chuyển đổi thất bại - next_recurring bị thay đổi")
	}
}

func main() {
	fmt.Println("Testing reminder to record conversion...")
	testReminderToRecord()
}