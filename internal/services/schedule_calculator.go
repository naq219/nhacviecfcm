package services

import (
	"errors"
	"time"

	"remiaq/internal/models"
)

// ScheduleCalculator calculates next trigger times for reminders
type ScheduleCalculator struct {
	lunarCalendar *LunarCalendar
}

// NewScheduleCalculator creates a new schedule calculator
func NewScheduleCalculator(lunarCalendar *LunarCalendar) *ScheduleCalculator {
	return &ScheduleCalculator{
		lunarCalendar: lunarCalendar,
	}
}

// CalculateNextTrigger calculates the next trigger time for a reminder
func (c *ScheduleCalculator) CalculateNextTrigger(reminder *models.Reminder, fromTime time.Time) (time.Time, error) {
	if reminder.Type == models.ReminderTypeOneTime {
		return c.calculateOneTime(reminder, fromTime)
	}

	return c.calculateRecurring(reminder, fromTime)
}

// calculateOneTime calculates next trigger for one-time reminders
func (c *ScheduleCalculator) calculateOneTime(reminder *models.Reminder, fromTime time.Time) (time.Time, error) {
	// For one-time reminders, return the set trigger time
	if reminder.NextTriggerAt != nil {
		return *reminder.NextTriggerAt, nil
	}

	return fromTime, nil
}

// calculateRecurring calculates next trigger for recurring reminders
func (c *ScheduleCalculator) calculateRecurring(reminder *models.Reminder, fromTime time.Time) (time.Time, error) {
	if reminder.RecurrencePattern == nil {
		return time.Time{}, errors.New("recurrence pattern is required for recurring reminders")
	}

	pattern := reminder.RecurrencePattern

	// Handle interval-based recurrence
	if pattern.IntervalSeconds > 0 {
		return c.calculateIntervalBased(reminder, fromTime)
	}

	// Handle calendar-based recurrence
	switch pattern.Type {
	case models.RecurrenceTypeDaily:
		return c.calculateDaily(reminder, fromTime)
	case models.RecurrenceTypeWeekly:
		return c.calculateWeekly(reminder, fromTime)
	case models.RecurrenceTypeMonthly:
		return c.calculateMonthly(reminder, fromTime)
	case models.RecurrenceTypeLunarLastDayOfMonth:
		return c.calculateLunarLastDay(reminder, fromTime)
	default:
		return time.Time{}, errors.New("unsupported recurrence type")
	}
}

// calculateIntervalBased calculates next trigger based on interval
func (c *ScheduleCalculator) calculateIntervalBased(reminder *models.Reminder, fromTime time.Time) (time.Time, error) {
	interval := time.Duration(reminder.RecurrencePattern.IntervalSeconds) * time.Second

	// If base_on is completion, calculate from completion time
	if reminder.RecurrencePattern.BaseOn == models.BaseOnCompletion {
		if reminder.LastCompletedAt != nil {
			return reminder.LastCompletedAt.Add(interval), nil
		}
		// If never completed, use creation time
		return reminder.Created.Add(interval), nil
	}

	// Otherwise, calculate from last trigger time (creation-based)
	return fromTime.Add(interval), nil
}

// calculateDaily calculates next daily trigger
func (c *ScheduleCalculator) calculateDaily(reminder *models.Reminder, fromTime time.Time) (time.Time, error) {
	if reminder.TriggerTimeOfDay == "" {
		return time.Time{}, errors.New("trigger_time_of_day is required for daily recurrence")
	}

	// Parse time of day (HH:MM format)
	targetTime, err := parseTimeOfDay(reminder.TriggerTimeOfDay)
	if err != nil {
		return time.Time{}, err
	}

	// Calculate next occurrence
	next := time.Date(
		fromTime.Year(), fromTime.Month(), fromTime.Day(),
		targetTime.Hour(), targetTime.Minute(), 0, 0,
		fromTime.Location(),
	)

	// If the time has passed today, move to tomorrow
	if next.Before(fromTime) || next.Equal(fromTime) {
		next = next.Add(24 * time.Hour)
	}

	return next, nil
}

// calculateWeekly calculates next weekly trigger
func (c *ScheduleCalculator) calculateWeekly(reminder *models.Reminder, fromTime time.Time) (time.Time, error) {
	if reminder.TriggerTimeOfDay == "" {
		return time.Time{}, errors.New("trigger_time_of_day is required for weekly recurrence")
	}

	pattern := reminder.RecurrencePattern
	targetWeekday := time.Weekday(pattern.DayOfWeek)

	// Parse time of day
	targetTime, err := parseTimeOfDay(reminder.TriggerTimeOfDay)
	if err != nil {
		return time.Time{}, err
	}

	// Find next occurrence of target weekday
	daysUntilTarget := (int(targetWeekday) - int(fromTime.Weekday()) + 7) % 7
	if daysUntilTarget == 0 {
		// It's the target day, check if time has passed
		next := time.Date(
			fromTime.Year(), fromTime.Month(), fromTime.Day(),
			targetTime.Hour(), targetTime.Minute(), 0, 0,
			fromTime.Location(),
		)
		if next.After(fromTime) {
			return next, nil
		}
		daysUntilTarget = 7
	}

	next := fromTime.Add(time.Duration(daysUntilTarget) * 24 * time.Hour)
	next = time.Date(
		next.Year(), next.Month(), next.Day(),
		targetTime.Hour(), targetTime.Minute(), 0, 0,
		next.Location(),
	)

	return next, nil
}

// calculateMonthly calculates next monthly trigger
func (c *ScheduleCalculator) calculateMonthly(reminder *models.Reminder, fromTime time.Time) (time.Time, error) {
	pattern := reminder.RecurrencePattern

	if reminder.CalendarType == models.CalendarTypeLunar {
		return c.calculateLunarMonthly(reminder, fromTime)
	}

	// Solar calendar
	if reminder.TriggerTimeOfDay == "" {
		return time.Time{}, errors.New("trigger_time_of_day is required for monthly recurrence")
	}

	targetTime, err := parseTimeOfDay(reminder.TriggerTimeOfDay)
	if err != nil {
		return time.Time{}, err
	}

	// Try current month first
	next := time.Date(
		fromTime.Year(), fromTime.Month(), pattern.DayOfMonth,
		targetTime.Hour(), targetTime.Minute(), 0, 0,
		fromTime.Location(),
	)

	// If date doesn't exist in current month or has passed, move to next month
	if next.Day() != pattern.DayOfMonth || next.Before(fromTime) || next.Equal(fromTime) {
		next = time.Date(
			fromTime.Year(), fromTime.Month()+1, pattern.DayOfMonth,
			targetTime.Hour(), targetTime.Minute(), 0, 0,
			fromTime.Location(),
		)

		// Handle year rollover
		if next.Day() != pattern.DayOfMonth {
			next = time.Date(
				fromTime.Year()+1, time.January, pattern.DayOfMonth,
				targetTime.Hour(), targetTime.Minute(), 0, 0,
				fromTime.Location(),
			)
		}
	}

	return next, nil
}

// calculateLunarMonthly calculates next lunar monthly trigger
func (c *ScheduleCalculator) calculateLunarMonthly(reminder *models.Reminder, fromTime time.Time) (time.Time, error) {
	pattern := reminder.RecurrencePattern
	targetDay := pattern.DayOfMonth

	// Convert current solar date to lunar
	lunarDate := c.lunarCalendar.SolarToLunar(fromTime)

	// Try to find the target day in current or next lunar month
	for i := 0; i < 13; i++ { // Max 13 lunar months in a year
		// Check if target day exists in current lunar month
		daysInMonth := c.lunarCalendar.GetLunarMonthDays(lunarDate.Year, lunarDate.Month)

		if targetDay <= daysInMonth {
			// Convert lunar date to solar
			solarDate := c.lunarCalendar.LunarToSolar(lunarDate.Year, lunarDate.Month, targetDay)

			// Apply time of day
			if reminder.TriggerTimeOfDay != "" {
				targetTime, _ := parseTimeOfDay(reminder.TriggerTimeOfDay)
				solarDate = time.Date(
					solarDate.Year(), solarDate.Month(), solarDate.Day(),
					targetTime.Hour(), targetTime.Minute(), 0, 0,
					solarDate.Location(),
				)
			}

			// If this date is in the future, return it
			if solarDate.After(fromTime) {
				return solarDate, nil
			}
		}

		// Move to next lunar month
		lunarDate.Month++
		if lunarDate.Month > 12 {
			lunarDate.Month = 1
			lunarDate.Year++
		}
	}

	return time.Time{}, errors.New("failed to calculate next lunar monthly trigger")
}

// calculateLunarLastDay calculates last day of lunar month
func (c *ScheduleCalculator) calculateLunarLastDay(reminder *models.Reminder, fromTime time.Time) (time.Time, error) {
	lunarDate := c.lunarCalendar.SolarToLunar(fromTime)

	// Try current month first
	daysInMonth := c.lunarCalendar.GetLunarMonthDays(lunarDate.Year, lunarDate.Month)
	solarDate := c.lunarCalendar.LunarToSolar(lunarDate.Year, lunarDate.Month, daysInMonth)

	if reminder.TriggerTimeOfDay != "" {
		targetTime, _ := parseTimeOfDay(reminder.TriggerTimeOfDay)
		solarDate = time.Date(
			solarDate.Year(), solarDate.Month(), solarDate.Day(),
			targetTime.Hour(), targetTime.Minute(), 0, 0,
			solarDate.Location(),
		)
	}

	if solarDate.After(fromTime) {
		return solarDate, nil
	}

	// Try next month
	lunarDate.Month++
	if lunarDate.Month > 12 {
		lunarDate.Month = 1
		lunarDate.Year++
	}

	daysInMonth = c.lunarCalendar.GetLunarMonthDays(lunarDate.Year, lunarDate.Month)
	solarDate = c.lunarCalendar.LunarToSolar(lunarDate.Year, lunarDate.Month, daysInMonth)

	if reminder.TriggerTimeOfDay != "" {
		targetTime, _ := parseTimeOfDay(reminder.TriggerTimeOfDay)
		solarDate = time.Date(
			solarDate.Year(), solarDate.Month(), solarDate.Day(),
			targetTime.Hour(), targetTime.Minute(), 0, 0,
			solarDate.Location(),
		)
	}

	return solarDate, nil
}

// parseTimeOfDay parses HH:MM format
func parseTimeOfDay(timeStr string) (time.Time, error) {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}, errors.New("invalid time format, expected HH:MM")
	}
	return t, nil
}
