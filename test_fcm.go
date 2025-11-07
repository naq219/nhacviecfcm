package main

import (
	"fmt"
	"log"

	"remiaq/internal/services"
)

func main() {
	// Test FCM service
	fcmService, err := services.NewFCMService("firebase-credentials.json")
	if err != nil {
		log.Fatalf("Failed to create FCM service: %v", err)
	}

	// Thay thế bằng FCM token thật của bạn
	testToken := "fC2NUSQbSdSauB-ADEfAkW:APA91bGkbVOq5BvTywBKodHsYgOHIaZTsX6s7Fc3Rj0wR59OcF7A5RPd-7a4YwUT7cQcdvl4-XIhi6rEPaz67Hk1kpOKf8drsIw-y8VA17SgRS2yQOlCRqw"

	err = fcmService.SendNotification(testToken, "Test Title", "Test message from FCM service")
	if err != nil {
		log.Fatalf("FCM test failed: %v", err)
	}

	fmt.Println("✅ FCM test passed! Notification sent successfully")
}
