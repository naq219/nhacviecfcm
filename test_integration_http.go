package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8090/api"

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type Reminder struct {
	ID               string     `json:"id"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	UserID           string     `json:"userId"`
	NextTriggerAt    *time.Time `json:"nextTriggerAt"`
	LastCompletedAt  *time.Time `json:"lastCompletedAt"`
	LastSentAt       *time.Time `json:"lastSentAt"`
	SnoozeUntil      *time.Time `json:"snoozeUntil"`
	Created          time.Time  `json:"created"`
	Updated          time.Time  `json:"updated"`
	CalendarType     string     `json:"calendarType"`
	RecurrenceType   string     `json:"recurrenceType"`
	TriggerTimeOfDay string    `json:"triggerTimeOfDay"`
	IsActive         bool       `json:"isActive"`
}

type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func main() {
	fmt.Println("🚀 Bắt đầu test tích hợp toàn bộ chức năng RemiAq qua HTTP API...")

	// Đợi server khởi động
	fmt.Println("⏳ Đợi server khởi động (5 giây)...")
	time.Sleep(5 * time.Second)

	// Test data
	testEmail := fmt.Sprintf("testuser_%d@example.com", time.Now().Unix())
	testPassword := "password123"
	testName := "Test User"

	var userID, authToken, reminderID string
	var err error

	// 1. Test đăng ký user
	fmt.Println("\n1. Testing user registration...")
	userID, err = testUserRegistration(testEmail, testPassword, testName)
	if err != nil {
		log.Fatalf("❌ User registration failed: %v", err)
	}
	fmt.Printf("   ✅ User created with ID: %s\n", userID)

	// 2. Test đăng nhập
	fmt.Println("\n2. Testing user login...")
	authToken, err = testUserLogin(testEmail, testPassword)
	if err != nil {
		log.Fatalf("❌ User login failed: %v", err)
	}
	fmt.Printf("   ✅ Login successful, token: %s\n", authToken)

	// 3. Test tạo reminder
	fmt.Println("\n3. Testing create reminder...")
	reminderID, err = testCreateReminder(authToken, userID)
	if err != nil {
		log.Fatalf("❌ Create reminder failed: %v", err)
	}
	fmt.Printf("   ✅ Reminder created with ID: %s\n", reminderID)

	// 4. Test list reminders
	fmt.Println("\n4. Testing list reminders...")
	reminders, err := testListReminders(authToken, userID)
	if err != nil {
		log.Fatalf("❌ List reminders failed: %v", err)
	}
	fmt.Printf("   ✅ Found %d reminders\n", len(reminders))

	// 5. Test get reminder detail
	fmt.Println("\n5. Testing get reminder detail...")
	reminder, err := testGetReminderDetail(authToken, reminderID)
	if err != nil {
		log.Fatalf("❌ Get reminder detail failed: %v", err)
	}
	fmt.Printf("   ✅ Reminder detail: %s\n", reminder.Title)

	// 6. Test update reminder
	fmt.Println("\n6. Testing update reminder...")
	updatedReminder, err := testUpdateReminder(authToken, reminderID)
	if err != nil {
		log.Fatalf("❌ Update reminder failed: %v", err)
	}
	fmt.Printf("   ✅ Reminder updated: %s\n", updatedReminder.Title)

	// 7. Test snooze reminder
	fmt.Println("\n7. Testing snooze reminder...")
	snoozedReminder, err := testSnoozeReminder(authToken, reminderID)
	if err != nil {
		log.Fatalf("❌ Snooze reminder failed: %v", err)
	}
	if snoozedReminder.SnoozeUntil != nil {
		fmt.Printf("   ✅ Reminder snoozed until: %v\n", snoozedReminder.SnoozeUntil.Format(time.RFC3339))
	}

	// 8. Test mark reminder as completed
	fmt.Println("\n8. Testing mark reminder as completed...")
	completedReminder, err := testMarkCompleted(authToken, reminderID)
	if err != nil {
		log.Fatalf("❌ Mark completed failed: %v", err)
	}
	if completedReminder.LastCompletedAt != nil {
		fmt.Printf("   ✅ Reminder completed at: %v\n", completedReminder.LastCompletedAt.Format(time.RFC3339))
	}

	// 9. Test delete reminder
	fmt.Println("\n9. Testing delete reminder...")
	if err := testDeleteReminder(authToken, reminderID); err != nil {
		log.Fatalf("❌ Delete reminder failed: %v", err)
	}
	fmt.Println("   ✅ Reminder deleted successfully")

	// 10. Test system status
	fmt.Println("\n10. Testing system status...")
	_, err = testSystemStatus(authToken)
	if err != nil {
		log.Fatalf("❌ System status check failed: %v", err)
	}
	fmt.Printf("   ✅ System status check completed\n")

	fmt.Println("\n🎉 Tất cả test đều PASSED!")
	fmt.Println("✅ Toàn bộ chức năng RemiAq hoạt động tốt qua HTTP API")
}

func makeRequest(method, url, authToken string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	return http.DefaultClient.Do(req)
}

func testUserRegistration(email, password, name string) (string, error) {
	payload := map[string]interface{}{
		"email":            email,
		"password":         password,
		"passwordConfirm": password,
		"name":             name,
	}

	resp, err := makeRequest("POST", baseURL+"/collections/users/records", "", payload)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registration failed with status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if id, ok := result["id"].(string); ok {
		return id, nil
	}

	return "", fmt.Errorf("invalid response format")
}

func testUserLogin(email, password string) (string, error) {
	payload := map[string]interface{}{
		"identity": email,
		"password": password,
	}

	resp, err := makeRequest("POST", baseURL+"/collections/users/auth-with-password", "", payload)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if token, ok := result["token"].(string); ok {
		return token, nil
	}

	return "", fmt.Errorf("invalid response format: no token found")
}

func testCreateReminder(authToken, userID string) (string, error) {
	now := time.Now().Add(1 * time.Hour)
	payload := map[string]interface{}{
		"title":            "Test Reminder",
		"description":      "This is a test reminder created by integration test",
		"userId":          userID,
		"nextTriggerAt":    now.Format(time.RFC3339),
		"calendarType":     "solar",
		"recurrenceType":   "one_time",
		"triggerTimeOfDay": "09:00",
		"isActive":         true,
	}

	resp, err := makeRequest("POST", baseURL+"/reminders", authToken, payload)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("create reminder failed with status: %d", resp.StatusCode)
	}

	var result ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if data, ok := result.Data.(map[string]interface{}); ok {
		if reminderID, ok := data["id"].(string); ok {
			return reminderID, nil
		}
	}

	return "", fmt.Errorf("invalid response format")
}

func testListReminders(authToken, userID string) ([]Reminder, error) {
	resp, err := makeRequest("GET", baseURL+"/reminders?userId="+userID, authToken, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list reminders failed with status: %d", resp.StatusCode)
	}

	var result ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if data, ok := result.Data.([]interface{}); ok {
		var reminders []Reminder
		for _, item := range data {
			if reminderData, ok := item.(map[string]interface{}); ok {
				// Convert map to Reminder struct
				jsonData, _ := json.Marshal(reminderData)
				var reminder Reminder
				json.Unmarshal(jsonData, &reminder)
				reminders = append(reminders, reminder)
			}
		}
		return reminders, nil
	}

	return nil, fmt.Errorf("invalid response format")
}

func testGetReminderDetail(authToken, reminderID string) (*Reminder, error) {
	resp, err := makeRequest("GET", baseURL+"/reminders/"+reminderID, authToken, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get reminder detail failed with status: %d", resp.StatusCode)
	}

	var result ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if data, ok := result.Data.(map[string]interface{}); ok {
		jsonData, _ := json.Marshal(data)
		var reminder Reminder
		json.Unmarshal(jsonData, &reminder)
		return &reminder, nil
	}

	return nil, fmt.Errorf("invalid response format")
}

func testUpdateReminder(authToken, reminderID string) (*Reminder, error) {
	payload := map[string]interface{}{
		"title":       "Updated Test Reminder",
		"description": "This reminder has been updated by integration test",
	}

	resp, err := makeRequest("PUT", baseURL+"/reminders/"+reminderID, authToken, payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update reminder failed with status: %d", resp.StatusCode)
	}

	var result ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if data, ok := result.Data.(map[string]interface{}); ok {
		jsonData, _ := json.Marshal(data)
		var reminder Reminder
		json.Unmarshal(jsonData, &reminder)
		return &reminder, nil
	}

	return nil, fmt.Errorf("invalid response format")
}

func testSnoozeReminder(authToken, reminderID string) (*Reminder, error) {
	snoozeTime := time.Now().Add(2 * time.Hour)
	payload := map[string]interface{}{
		"snoozeUntil": snoozeTime.Format(time.RFC3339),
	}

	resp, err := makeRequest("POST", baseURL+"/reminders/"+reminderID+"/snooze", authToken, payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("snooze reminder failed with status: %d", resp.StatusCode)
	}

	var result ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if data, ok := result.Data.(map[string]interface{}); ok {
		jsonData, _ := json.Marshal(data)
		var reminder Reminder
		json.Unmarshal(jsonData, &reminder)
		return &reminder, nil
	}

	return nil, fmt.Errorf("invalid response format")
}

func testMarkCompleted(authToken, reminderID string) (*Reminder, error) {
	resp, err := makeRequest("POST", baseURL+"/reminders/"+reminderID+"/complete", authToken, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mark completed failed with status: %d", resp.StatusCode)
	}

	var result ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if data, ok := result.Data.(map[string]interface{}); ok {
		jsonData, _ := json.Marshal(data)
		var reminder Reminder
		json.Unmarshal(jsonData, &reminder)
		return &reminder, nil
	}

	return nil, fmt.Errorf("invalid response format")
}

func testDeleteReminder(authToken, reminderID string) error {
	resp, err := makeRequest("DELETE", baseURL+"/reminders/"+reminderID, authToken, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete reminder failed with status: %d", resp.StatusCode)
	}

	return nil
}

func testSystemStatus(authToken string) (map[string]interface{}, error) {
	resp, err := makeRequest("GET", baseURL+"/system/status", authToken, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("system status check failed with status: %d", resp.StatusCode)
	}

	var result ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if data, ok := result.Data.(map[string]interface{}); ok {
		return data, nil
	}

	return nil, fmt.Errorf("invalid response format")
}