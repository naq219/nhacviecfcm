package services

import (
	"errors"
	"time"

	"remiaq/internal/models"
)

// ScheduleCalculator calculates next trigger times for reminders (FRP & CRP)
type ScheduleCalculator struct {
	lunarCalendar *LunarCalendar
}

// NewScheduleCalculator creates a new schedule calculator
func NewScheduleCalculator(lunarCalendar *LunarCalendar) *ScheduleCalculator {
	return &ScheduleCalculator{
		lunarCalendar: lunarCalendar,
	}
}

// ========================================
// MAIN: CalculateNextActionAt
// ========================================

// CalculateNextActionAt calculates the nearest time to check this reminder
// Returns the minimum of: snooze_until, next_recurring, next_crp
func (c *ScheduleCalculator) CalculateNextActionAt(reminder *models.Reminder, now time.Time) time.Time {
	candidates := []time.Time{}

	// 1. If snoozed, snooze_until has highest priority
	if !reminder.SnoozeUntil.IsZero() && reminder.SnoozeUntil.After(now) {
		return reminder.SnoozeUntil
	}

	// 2. For recurring: add next_recurring
	if reminder.Type == models.ReminderTypeRecurring && !reminder.NextRecurring.IsZero() {
		candidates = append(candidates, reminder.NextRecurring)
	}

	// 3. For CRP: add next_crp if we haven't reached quota
	if reminder.MaxCRP == 0 || reminder.CRPCount < reminder.MaxCRP {
		if !reminder.LastSentAt.IsZero() {
			nextCRP := reminder.LastSentAt.Add(time.Duration(reminder.CRPIntervalSec) * time.Second)
			candidates = append(candidates, nextCRP)
		} else {
			// First CRP: send immediately
			candidates = append(candidates, now)
		}
	}

	// 4. Return the earliest candidate
	if len(candidates) == 0 {
		return time.Time{}
	}

	minTime := candidates[0]
	for _, t := range candidates[1:] {
		if t.Before(minTime) {
			minTime = t
		}
	}

	return minTime
}

// ========================================
// FRP: CalculateNextRecurring
// ========================================

// CalculateNextRecurring calculates the next recurring trigger time
// Starts from current NextRecurring and finds first occurrence after now
func (c *ScheduleCalculator) CalculateNextRecurring(reminder *models.Reminder, now time.Time) (time.Time, error) {
	if reminder.RecurrencePattern == nil {
		return time.Time{}, errors.New("recurrence_pattern required for recurring reminder")
	}
	pattern := reminder.RecurrencePattern
	current := reminder.NextRecurring

	if current.IsZero() {
		current = now
	}

	// ⭐ NEW: Handle interval_seconds
	if pattern.Type == models.RecurrenceTypeIntervalSeconds {
		return c.calculateNextIntervalSeconds(current, pattern, now)
	}

	// Existing logic (daily, weekly, monthly, lunar)
	switch pattern.Type {
	case models.RecurrenceTypeDaily:
		return c.calculateNextDaily(current, pattern, now)
	case models.RecurrenceTypeWeekly:
		return c.calculateNextWeekly(current, pattern, now)
	case models.RecurrenceTypeMonthly:
		if reminder.CalendarType == models.CalendarTypeLunar {
			return c.calculateNextLunarMonthly(current, pattern, now)
		}
		return c.calculateNextSolarMonthly(current, pattern, now)
	case models.RecurrenceTypeLunarLastDayOfMonth:
		return c.calculateNextLunarLastDay(current, pattern, now)
	default:
		return time.Time{}, errors.New("unsupported recurrence type")
	}
}

func (c *ScheduleCalculator) calculateNextIntervalSeconds(current time.Time, pattern *models.RecurrencePattern, now time.Time) (time.Time, error) {
	if pattern.IntervalSeconds <= 0 {
		return time.Time{}, errors.New("interval_seconds must be > 0")
	}
	interval := time.Duration(pattern.IntervalSeconds) * time.Second
	next := current

	// Find first occurrence after now
	for next.Before(now) || next.Equal(now) {
		next = next.Add(interval)
	}

	return next, nil
}

// calculateNextDaily: Add interval days, find first > now
func (c *ScheduleCalculator) calculateNextDaily(current time.Time, pattern *models.RecurrencePattern, now time.Time) (time.Time, error) {
	interval := pattern.Interval
	if interval <= 0 {
		interval = 1
	}

	// Get trigger time of day (HH:MM), default 00:00
	var hour, minute int
	if pattern.TriggerTimeOfDay != "" {
		t, err := parseTimeOfDay(pattern.TriggerTimeOfDay)
		if err == nil {
			hour, minute = t.Hour(), t.Minute()
		}
	}

	next := current
	for next.Before(now) || next.Equal(now) {
		next = next.AddDate(0, 0, interval)
	}

	// Apply time of day
	next = time.Date(next.Year(), next.Month(), next.Day(), hour, minute, 0, 0, next.Location())

	// Make sure it's still after now
	for !next.After(now) {
		next = next.AddDate(0, 0, interval)
	}

	return next, nil
}

// calculateNextWeekly: Find next target weekday, find first > now
func (c *ScheduleCalculator) calculateNextWeekly(current time.Time, pattern *models.RecurrencePattern, now time.Time) (time.Time, error) {
	targetWeekday := time.Weekday(pattern.DayOfWeek)
	interval := pattern.Interval
	if interval <= 0 {
		interval = 1
	}

	// Get trigger time of day
	var hour, minute int
	if pattern.TriggerTimeOfDay != "" {
		t, err := parseTimeOfDay(pattern.TriggerTimeOfDay)
		if err == nil {
			hour, minute = t.Hour(), t.Minute()
		}
	}

	next := current
	for !next.After(now) {
		// Find next target weekday
		daysUntil := (int(targetWeekday) - int(next.Weekday()) + 7) % 7
		if daysUntil == 0 {
			daysUntil = 7 * interval
		} else {
			daysUntil = daysUntil + (7 * (interval - 1))
		}
		next = next.AddDate(0, 0, daysUntil)
		next = time.Date(next.Year(), next.Month(), next.Day(), hour, minute, 0, 0, next.Location())
	}

	return next, nil
}

// calculateNextSolarMonthly: Add interval months on day_of_month, find first > now
func (c *ScheduleCalculator) calculateNextSolarMonthly(current time.Time, pattern *models.RecurrencePattern, now time.Time) (time.Time, error) {
	dayOfMonth := pattern.DayOfMonth
	if dayOfMonth <= 0 {
		dayOfMonth = 1
	}
	interval := pattern.Interval
	if interval <= 0 {
		interval = 1
	}

	// Get trigger time of day
	var hour, minute int
	if pattern.TriggerTimeOfDay != "" {
		t, err := parseTimeOfDay(pattern.TriggerTimeOfDay)
		if err == nil {
			hour, minute = t.Hour(), t.Minute()
		}
	}

	next := current
	for !next.After(now) {
		next = next.AddDate(0, interval, 0)
		// Set to day_of_month (may be adjusted if day doesn't exist in month)
		next = time.Date(next.Year(), next.Month(), dayOfMonth, hour, minute, 0, 0, next.Location())
		// Re-adjust if day overflowed to next month
		if next.Day() != dayOfMonth {
			next = time.Date(next.Year(), next.Month(), 1, hour, minute, 0, 0, next.Location()).AddDate(0, 1, -1)
		}
	}

	return next, nil
}

// calculateNextLunarMonthly: Similar to solar but using lunar calendar
func (c *ScheduleCalculator) calculateNextLunarMonthly(current time.Time, pattern *models.RecurrencePattern, now time.Time) (time.Time, error) {
	dayOfMonth := pattern.DayOfMonth
	if dayOfMonth <= 0 {
		dayOfMonth = 1
	}
	interval := pattern.Interval
	if interval <= 0 {
		interval = 1
	}

	lunarDate := c.lunarCalendar.SolarToLunar(current)

	// Try up to 24 lunar months
	for i := 0; i < 24; i++ {
		daysInMonth := c.lunarCalendar.GetLunarMonthDays(lunarDate.Year, lunarDate.Month)
		if dayOfMonth > daysInMonth {
			dayOfMonth = daysInMonth
		}

		solarDate := c.lunarCalendar.LunarToSolar(lunarDate.Year, lunarDate.Month, dayOfMonth)

		// Apply time of day
		if pattern.TriggerTimeOfDay != "" {
			t, _ := parseTimeOfDay(pattern.TriggerTimeOfDay)
			solarDate = time.Date(solarDate.Year(), solarDate.Month(), solarDate.Day(), t.Hour(), t.Minute(), 0, 0, solarDate.Location())
		}

		if solarDate.After(now) {
			return solarDate, nil
		}

		// Move to next lunar month
		lunarDate.Month += interval
		if lunarDate.Month > 12 {
			lunarDate.Year += (lunarDate.Month - 1) / 12
			lunarDate.Month = ((lunarDate.Month - 1) % 12) + 1
		}
	}

	return time.Time{}, errors.New("failed to calculate next lunar monthly trigger")
}

// calculateNextLunarLastDay: Last day of lunar month, find first > now
func (c *ScheduleCalculator) calculateNextLunarLastDay(current time.Time, pattern *models.RecurrencePattern, now time.Time) (time.Time, error) {
	lunarDate := c.lunarCalendar.SolarToLunar(current)

	// Try up to 24 lunar months
	for i := 0; i < 24; i++ {
		daysInMonth := c.lunarCalendar.GetLunarMonthDays(lunarDate.Year, lunarDate.Month)
		solarDate := c.lunarCalendar.LunarToSolar(lunarDate.Year, lunarDate.Month, daysInMonth)

		// Apply time of day
		if pattern.TriggerTimeOfDay != "" {
			t, _ := parseTimeOfDay(pattern.TriggerTimeOfDay)
			solarDate = time.Date(solarDate.Year(), solarDate.Month(), solarDate.Day(), t.Hour(), t.Minute(), 0, 0, solarDate.Location())
		}

		if solarDate.After(now) {
			return solarDate, nil
		}

		// Move to next lunar month
		lunarDate.Month++
		if lunarDate.Month > 12 {
			lunarDate.Month = 1
			lunarDate.Year++
		}
	}

	return time.Time{}, errors.New("failed to calculate next lunar last day trigger")
}

// ========================================
// CRP: CanSendCRP
// ========================================

// CanSendCRP checks if we can send a CRP notification
// Returns true if: quota not reached AND enough time passed since last send
func (c *ScheduleCalculator) CanSendCRP(reminder *models.Reminder, now time.Time) bool {
	// Check quota: if MaxCRP > 0, must not exceed it
	if reminder.MaxCRP > 0 && reminder.CRPCount >= reminder.MaxCRP {
		return false
	}

	// Check interval
	if reminder.LastSentAt.IsZero() {
		// First CRP: can send
		return true
	}

	elapsed := now.Sub(reminder.LastSentAt)
	required := time.Duration(reminder.CRPIntervalSec) * time.Second

	return elapsed >= required
}

// ========================================
// HELPER
// ========================================

// parseTimeOfDay parses HH:MM format
func parseTimeOfDay(timeStr string) (time.Time, error) {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}, errors.New("invalid time format, expected HH:MM")
	}
	return t, nil
}
