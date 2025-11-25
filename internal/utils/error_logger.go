package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var errorLogMutex sync.Mutex

// LogErrorToFile logs error messages to notifi_error.txt
func LogErrorToFile(reminderID, errorMessage string) error {
	errorLogMutex.Lock()
	defer errorLogMutex.Unlock()

	logsDir := filepath.Join(".", "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	filePath := filepath.Join(logsDir, "notifi_error.txt")
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open error log file: %w", err)
	}
	defer file.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] Reminder %s: %s\n", timestamp, reminderID, errorMessage)

	if _, err := file.WriteString(logEntry); err != nil {
		return fmt.Errorf("failed to write to error log: %w", err)
	}

	return nil
}