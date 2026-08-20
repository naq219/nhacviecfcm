package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Reminder struct {
	Title        string    `json:"title"`
	Type         string    `json:"type"`
	NextActionAt time.Time `json:"next_action_at"`
}

func main() {
	jsonStr := `{"title":"android2","type":"one_time","calendar_type":"solar","next_action_at":"2025-11-24T13:51:35.436Z","repeat_strategy":"none","max_crp":0,"crp_interval_sec":300,"status":"active"}`

	var tempReminder struct {
		Reminder
		ForTestSeconds int `json:"for_test"`
	}

	if err := json.NewDecoder(strings.NewReader(jsonStr)).Decode(&tempReminder); err != nil {
		fmt.Printf("Error decoding: %v\n", err)
		return
	}

	fmt.Printf("Decoded NextActionAt: %v\n", tempReminder.Reminder.NextActionAt)
	fmt.Printf("Decoded Title: %s\n", tempReminder.Reminder.Title)
	fmt.Printf("ForTestSeconds: %d\n", tempReminder.ForTestSeconds)
}
