# Calculation Logic for Next Recurring Time
*Documentation for [calculateNextRecurringTH4NoCrpNoUT](file:///d:/PROJECT/nhacviecfcm/internal/worker/worker_loop_noUT.go#217-254) and [calculateNextRecurringTH4YesCrpNoUT](file:///d:/PROJECT/nhacviecfcm/internal/worker/worker_loop_noUT.go#255-259)*

## Overview

These functions calculate the next occurrence time for recurring reminders based on the [RecurrencePattern](file:///d:/PROJECT/nhacviecfcm/internal/models/reminder.go#52-62). Both functions use the same underlying logic.

**Location**: [internal/worker/worker_loop_noUT.go](file:///d:/PROJECT/nhacviecfcm/internal/worker/worker_loop_noUT.go)

**Reference**: Based on logic from [internal/services/schedule_calculator.go](file:///d:/PROJECT/nhacviecfcm/internal/services/schedule_calculator.go)

---

## Recurrence Pattern Types

The calculation method depends on the `pattern.Type` field from [RecurrencePattern](file:///d:/PROJECT/nhacviecfcm/internal/models/reminder.go#52-62):

### 1. **Daily** (`daily`)

**Formula**: Add `interval` days repeatedly until result is after `now`

**Fields Used**:
- `interval` - Number of days between occurrences (default: 1)
- `trigger_time_of_day` - Time in HH:MM format (UTC)

**Example**:
```json
{
  "type": "daily",
  "interval": 2,
  "trigger_time_of_day": "08:00"
}
```
→ Every 2 days at 08:00 UTC

**Calculation Steps**:
1. Start from `current` (previous NextRecurring)
2. Add `interval` days repeatedly
3. Apply trigger time of day (hour:minute)
4. Find first occurrence after `now`

---

### 2. **Weekly** (`weekly`)

**Formula**: Find next target weekday, add interval weeks

**Fields Used**:
- `day_of_week` - Target weekday (0=Sunday, 1=Monday, ..., 6=Saturday)
- `interval` - Number of weeks between occurrences (default: 1)
- `trigger_time_of_day` - Time in HH:MM format (UTC)

**Example**:
```json
{
  "type": "weekly",
  "day_of_week": 1,
  "interval": 1,
  "trigger_time_of_day": "09:00"
}
```
→ Every Monday at 09:00 UTC

**Calculation Steps**:
1. Calculate days until target weekday: [(target_day - current_day + 7) % 7](file:///d:/PROJECT/nhacviecfcm/internal/worker/worker_loop_noUT.go#38-63)
2. If already on target day (0), jump full interval weeks
3. Add calculated days
4. Apply trigger time
5. Find first occurrence after `now`

---

### 3. **Monthly (Solar)** (`monthly` with `calendar_type=solar`)

**Formula**: Add `interval` months, set to `day_of_month`

**Fields Used**:
- `day_of_month` - Day of month (1-31)
- `interval` - Number of months between occurrences (default: 1)
- `trigger_time_of_day` - Time in HH:MM format (UTC)

**Example**:
```json
{
  "type": "monthly",
  "day_of_month": 15,
  "interval": 1,
  "trigger_time_of_day": "10:00"
}
```
→ 15th of every month at 10:00 UTC

**Calculation Steps**:
1. Add `interval` months to `current`
2. Set day to `day_of_month`
3. Handle overflow (e.g., Feb 31 → last day of Feb)
4. Apply trigger time
5. Find first occurrence after `now`

**Special Handling**:
- If `day_of_month > days_in_month`, use last day of month
- Example: day_of_month=31 in February → Feb 28/29

---

### 4. **Monthly (Lunar)** (`monthly` with `calendar_type=lunar`)

**Status**: Currently uses **30-day fallback**

**Reason**: Lunar calendar calculation requires `LunarCalendar` service (not available in worker)

**Current Implementation**:
```go
return now.Add(30 * 24 * time.Hour)  // Approximation
```

**Future Enhancement**: 
Should integrate with `services.LunarCalendar` for accurate lunar month calculation

---

### 5. **Interval Seconds** (`interval_seconds`)

**Formula**: Add `interval_seconds` repeatedly until result is after `now`

**Fields Used**:
- `interval_seconds` - Number of seconds between occurrences

**Example**:
```json
{
  "type": "interval_seconds",
  "interval_seconds": 3600
}
```
→ Every hour (3600 seconds)

**Calculation Steps**:
1. Convert `interval_seconds` to `time.Duration`
2. Add interval repeatedly to `current`
3. Find first occurrence after `now`

**Use Cases**:
- Short intervals (minutes, hours)
- Fixed time intervals ignoring calendar days
- Example: "Every 2 hours", "Every 30 minutes"

---

### 6. **Lunar Last Day of Month** (`lunar_last_day_of_month`)

**Status**: Currently uses **30-day fallback**

**Reason**: Requires lunar calendar service for accurate calculation

**Current Implementation**:
```go
return now.Add(30 * 24 * time.Hour)  // Approximation
```

---

## Important Notes

### Timezone Handling
- All calculations use **UTC timezone**
- `trigger_time_of_day` is expected in UTC format
- Result is returned in UTC
- Client should convert to local timezone for display

### Base Time (`current`)
- Typically set to previous [NextRecurring](file:///d:/PROJECT/nhacviecfcm/internal/models/reminder.go#209-213)
- If zero (first calculation), initialized to `now`
- Ensures consistent intervals regardless of when worker runs

### Algorithm Pattern
All helper functions follow this pattern:
```go
next := current
for !next.After(now) {
    next = next.Add(interval)  // Add appropriate interval
}
return next
```

### Error Handling
- Returns `now + 24 hours` as fallback for errors
- Logs errors via `w.logger.Errorf()`
- Never returns zero time

---

## Function Signatures

### Main Functions
```go
func (w *WorkerLoopNoUT) calculateNextRecurringTH4NoCrpNoUT(
    r *models.Reminder, 
    now time.Time
) time.Time

func (w *WorkerLoopNoUT) calculateNextRecurringTH4YesCrpNoUT(
    r *models.Reminder, 
    now time.Time
) time.Time
```

### Helper Functions
```go
func (w *WorkerLoopNoUT) calculateNextDaily(
    current time.Time, 
    pattern *models.RecurrencePattern, 
    now time.Time
) time.Time

func (w *WorkerLoopNoUT) calculateNextWeekly(...)
func (w *WorkerLoopNoUT) calculateNextSolarMonthly(...)
func (w *WorkerLoopNoUT) calculateNextIntervalSeconds(...)
```

### Utility Functions
```go
func parseTimeOfDay(timeStr string) (time.Time, error)
    // Parses "HH:MM" format (e.g., "08:30")
    // Returns time with hour and minute set, other fields zero
```

---

## Usage Context

### Case 4: Recurring - No CRP - No UT
**When**: `max_crp = 0` and `repeat_strategy = none`

**After sending notification**:
```go
nextRecurring := w.calculateNextRecurringTH4NoCrpNoUT(r, now)
r.NextRecurring = nextRecurring
r.LastSentAt = now
```

### Case 5.1: Recurring - Yes CRP - No UT (FRP Trigger)
**When**: `max_crp > 0` and `next_recurring <= now`

**After sending notification**:
```go
nextRecurring := w.calculateNextRecurringTH4YesCrpNoUT(r, now)
r.NextRecurring = nextRecurring
r.CRPCount = 0  // Reset CRP cycle
r.NextCRP = now + crp_interval_sec
r.LastSentAt = now
```

---

## Examples

### Example 1: Daily Reminder
```
Pattern: { type: "daily", interval: 1, trigger_time_of_day: "09:00" }
Current: 2025-11-20 09:00:00 UTC
Now:     2025-11-20 10:00:00 UTC
Result:  2025-11-21 09:00:00 UTC
```

### Example 2: Weekly Reminder
```
Pattern: { type: "weekly", day_of_week: 1 (Monday), trigger_time_of_day: "08:00" }
Current: 2025-11-25 08:00:00 UTC (Monday)
Now:     2025-11-25 09:00:00 UTC
Result:  2025-12-02 08:00:00 UTC (Next Monday)
```

### Example 3: Interval Seconds
```
Pattern: { type: "interval_seconds", interval_seconds: 1800 }  // 30 minutes
Current: 2025-11-27 10:00:00 UTC
Now:     2025-11-27 10:25:00 UTC
Result:  2025-11-27 10:30:00 UTC
```

---

## Migration from Old Logic

### Before (TODO Placeholder)
```go
func calculateNextRecurringTH4NoCrpNoUT(...) time.Time {
    // TODO: Implement based on RecurrencePattern
    return now.Add(24 * time.Hour)
}
```

### After (Pattern-Based Calculation)
```go
func calculateNextRecurringTH4NoCrpNoUT(...) time.Time {
    switch pattern.Type {
    case models.RecurrenceTypeDaily:
        return w.calculateNextDaily(current, pattern, now)
    case models.RecurrenceTypeWeekly:
        return w.calculateNextWeekly(current, pattern, now)
    // ... etc
    }
}
```

---

## Related Documentation
- [WORKER VER 2.md](file:///d:/PROJECT/nhacviecfcm/docs/WORKER%20VER%202.md) - Worker implementation specification
- [schedule_calculator.go](file:///d:/PROJECT/nhacviecfcm/internal/services/schedule_calculator.go) - Original calculation logic
- [API_EXAMPLE_V2.md](file:///d:/PROJECT/nhacviecfcm/v2docs/API_EXAMPLE_V2.md) - API examples for recurrence patterns
