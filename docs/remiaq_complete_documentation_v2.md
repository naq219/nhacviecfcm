# RemiAq - Complete Technical Documentation

**Version**: 2.0  
**Last Updated**: 2025-11-08  
**Status**: Production Ready with CRP/FRP Logic

---

## 📋 Table of Contents

1. [System Overview](#system-overview)
2. [Architecture](#architecture)
3. [Database Schema](#database-schema)
4. [Worker Logic - FRP & CRP](#worker-logic---frp--crp)
5. [API Documentation](#api-documentation)
6. [Configuration](#configuration)
7. [Time Field Handling](#time-field-handling)
8. [Error Handling](#error-handling)

---

## System Overview

RemiAq is a reminder management system with support for:
- **One-time reminders** with retry logic (CRP)
- **Recurring reminders** with flexible scheduling patterns
- **Two recurrence strategies**: Auto-repeat or Complete-based
- **Firebase Cloud Messaging (FCM)** notifications
- **Lunar and Solar calendar** support
- **Background worker** processing reminders every 60 seconds

### Key Concepts

**FRP (Father Recurrence Pattern)**
- Main recurring schedule (daily, weekly, monthly, lunar, interval-based)
- Triggers according to pattern (e.g., every 3 minutes)
- Only for `type = "recurring"`

**CRP (Child Repeat Pattern)**
- Retry/notification resend within a cycle
- Triggers every `crp_interval_sec` seconds
- Limited to `max_crp` times per cycle
- Works for both one-time and recurring reminders

**Repeat Strategy**
- `none`: Auto-calculate next FRP, don't wait for user complete
- `crp_until_complete`: Wait for user to complete before calculating next FRP

---

## Architecture

```
┌─────────────────────────────────────────┐
│         PocketBase Server               │
│                                         │
│  ┌──────────────────────────────────┐  │
│  │    REST API Handlers             │  │
│  │  - CreateReminder                │  │
│  │  - UpdateReminder                │  │
│  │  - CompleteReminder              │  │
│  │  - SnoozeReminder                │  │
│  └──────────────────────────────────┘  │
│                 ↕                       │
│  ┌──────────────────────────────────┐  │
│  │    ReminderService               │  │
│  │  - Business Logic                │  │
│  │  - Validation                    │  │
│  │  - State Management              │  │
│  └──────────────────────────────────┘  │
│                 ↕                       │
│  ┌──────────────────────────────────┐  │
│  │    ScheduleCalculator            │  │
│  │  - Calculate next_recurring      │  │
│  │  - Calculate next_action_at      │  │
│  │  - Check CRP ready               │  │
│  └──────────────────────────────────┘  │
│                 ↕                       │
│  ┌──────────────────────────────────┐  │
│  │    PocketBase ORM Repository     │  │
│  │  - Database Operations           │  │
│  │  - Time Parsing (Multiple Formats)│  │
│  └──────────────────────────────────┘  │
└─────────────────────────────────────────┘
         ↕              ↕
    ┌────────┐     ┌─────────┐
    │ SQLite │     │  FCM    │
    │   DB   │     │ Service │
    └────────┘     └─────────┘

┌─────────────────────────────────────────┐
│      Background Worker (Every 60s)      │
│                                         │
│  1. Query: next_action_at <= NOW()      │
│  2. For each reminder:                  │
│     - Check FRP (priority)              │
│     - Check CRP (if quota available)    │
│     - Update DB + Send FCM              │
└─────────────────────────────────────────┘
```

---

## Database Schema

### reminders Table

```sql
CREATE TABLE reminders (
  -- Identifiers
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  
  -- Content
  title TEXT NOT NULL,
  description TEXT,
  
  -- Type & Calendar
  type TEXT NOT NULL,              -- 'one_time' | 'recurring'
  calendar_type TEXT,              -- 'solar' | 'lunar'
  status TEXT DEFAULT 'active',    -- 'active' | 'completed' | 'paused'
  
  -- FRP (Father Recurrence Pattern)
  next_recurring DATETIME,         -- Next FRP trigger time
  recurrence_pattern JSON,         -- Pattern config: {type, interval, ...}
  
  -- CRP (Child Repeat Pattern)
  next_crp DATETIME,               -- Next CRP retry time
  max_crp INT DEFAULT 0,           -- Max retries (0 = send only once)
  crp_count INT DEFAULT 0,         -- Current retry count
  crp_interval_sec INT,            -- Retry interval in seconds
  
  -- Strategies
  repeat_strategy TEXT,            -- 'none' | 'crp_until_complete'
  
  -- Optimization
  next_action_at DATETIME,         -- = MIN(snooze_until, next_recurring, next_crp)
  
  -- Tracking
  last_sent_at DATETIME,           -- Last notification sent time
  last_crp_completed_at DATETIME,  -- User completed current CRP cycle
  last_completed_at DATETIME,      -- Last completion (for repeat_strategy=crp_until_complete)
  
  -- Snooze
  snooze_until DATETIME,           -- Snoozed until when (if empty = not snoozed)
  
  -- Timestamps
  created DATETIME,
  updated DATETIME
);

-- Recommended Indexes
CREATE INDEX idx_next_action_at ON reminders(next_action_at, status);
CREATE INDEX idx_user_id ON reminders(user_id, status);
```

### recurrence_pattern JSON Examples

```json
// Daily at 08:00 UTC
{
  "type": "daily",
  "interval": 1,
  "trigger_time_of_day": "08:00"
}

// Every 3 minutes
{
  "type": "interval_seconds",
  "interval_seconds": 180
}

// Every Monday at 09:00 UTC
{
  "type": "weekly",
  "interval": 1,
  "day_of_week": 1,
  "trigger_time_of_day": "09:00"
}

// 5th of each month at 10:00 UTC
{
  "type": "monthly",
  "interval": 1,
  "day_of_month": 5,
  "trigger_time_of_day": "10:00"
}

// Last day of lunar month at 18:00 UTC
{
  "type": "lunar_last_day_of_month",
  "trigger_time_of_day": "18:00"
}
```

---

## Worker Logic - FRP & CRP

### Worker Cycle (Every 60 seconds)

```
1. Check if worker_enabled in system_status
2. Query: SELECT * FROM reminders 
          WHERE next_action_at <= NOW()
          AND status = 'active'
          AND (snooze_until IS NULL OR snooze_until <= NOW())
3. For each reminder:
   a. Check FRP (highest priority)
   b. Check CRP (if FRP not triggered)
   c. Recalculate next_action_at
4. Update DB with new times and counts
5. Send FCM notifications
```

### FRP Trigger

**When**: `now >= next_recurring`

**Action**:
```
1. Send notification
2. Update: last_sent_at = now
3. Reset: crp_count = 0, next_crp = next_recurring
4. Calculate next FRP:
   - repeat_strategy = "none": Auto-calculate (now + pattern)
   - repeat_strategy = "crp_until_complete": Keep same next_recurring
5. Recalculate: next_action_at = MIN(snooze_until, next_recurring, next_crp)
```

### CRP Trigger

**When**: `now >= next_crp AND crp_count < max_crp`

**Action**:
```
1. Send notification
2. Update: last_sent_at = now
3. Increment: crp_count++
4. Calculate: next_crp = now + crp_interval_sec
5. If one_time AND crp_count >= max_crp:
   - Set: status = "completed", next_action_at = empty
6. Recalculate: next_action_at = MIN(snooze_until, next_recurring, next_crp)
```

### CalculateNextActionAt Logic

```go
candidates := []time.Time{}

// Priority 1: If snoozed, return snooze_until immediately
if snooze_until.After(now) {
    return snooze_until
}

// Priority 2: For recurring, add next_recurring
if type == "recurring" && next_recurring.IsValid() {
    candidates.append(next_recurring)
}

// Priority 3: For CRP, add next_crp if quota available
if max_crp == 0 || crp_count < max_crp {
    if next_crp.IsValid() {
        candidates.append(next_crp)
    } else if last_sent_at.IsValid() {
        candidates.append(last_sent_at + crp_interval_sec)
    } else {
        candidates.append(now)  // First send
    }
}

// Return minimum (earliest time)
return MIN(candidates)
```

### Safety Check for Infinite Loops

If `next_action_at` is more than 1 hour in the past:
```
1. Log warning
2. Recalculate next_action_at
3. If still invalid, set to now + 1 minute (emergency fallback)
4. Update DB
```

---

## API Documentation

### Authentication

All endpoints (except `/hello` and `/api/system_status`) require:
```
Authorization: Bearer {JWT_TOKEN}
```

### Create Reminder

```
POST /api/reminders
Authorization: Bearer {token}
Content-Type: application/json

{
  "title": "string",                           ✓ Required
  "description": "string",                     ○ Optional
  "type": "one_time|recurring",                ✓ Required
  "calendar_type": "solar|lunar",              ✓ Required (default: solar)
  "status": "active|completed|paused",         ○ (default: active)
  
  "recurrence_pattern": {                      ○ Required if type=recurring
    "type": "daily|weekly|monthly|interval_seconds|lunar_last_day_of_month",
    "interval": 1,
    "day_of_month": 5,
    "day_of_week": 1,
    "trigger_time_of_day": "08:00",
    "interval_seconds": 3600
  },
  
  "repeat_strategy": "none|crp_until_complete",  ○ (default: none)
  
  "max_crp": 3,                                ○ (0 = send only once)
  "crp_interval_sec": 300                      ○ (retry interval in seconds)
}

Response:
{
  "success": true,
  "message": "Reminder created successfully",
  "data": {
    "id": "reminder_id_123",
    "next_recurring": "2025-11-08T16:00:00Z",
    "next_crp": "2025-11-08T16:00:00Z",
    "next_action_at": "2025-11-08T16:00:00Z",
    ...
  }
}
```

### Get My Reminders

```
GET /api/reminders/mine
Authorization: Bearer {token}

Response:
{
  "success": true,
  "data": [
    { reminder objects... }
  ]
}
```

### Update Reminder

```
PUT /api/reminders/{id}
Authorization: Bearer {token}
Content-Type: application/json

{
  "title": "Updated Title",
  "max_crp": 5,
  ...
}
```

### Complete Reminder

```
POST /api/reminders/{id}/complete
Authorization: Bearer {token}

Effect for one_time:
- status = 'completed'
- next_action_at = empty

Effect for recurring + repeat_strategy=none:
- crp_count = 0 (reset)
- FRP continues on schedule

Effect for recurring + repeat_strategy=crp_until_complete:
- crp_count = 0 (reset)
- next_recurring = calculated from completion time
- FRP next cycle starts from completion time
```

### Snooze Reminder

```
POST /api/reminders/{id}/snooze
Authorization: Bearer {token}
Content-Type: application/json

{
  "duration": 300  // seconds (5 minutes)
}

Effect:
- snooze_until = now + 300 seconds
- next_action_at = recalculated
- Worker will skip this reminder until snooze expires
```

### Delete Reminder

```
DELETE /api/reminders/{id}
Authorization: Bearer {token}
```

---

## Configuration

### Environment Variables (.env)

```env
# Server
PB_ADDR=127.0.0.1:8090

# Worker
WORKER_INTERVAL=60         # Seconds between worker cycles (default: 60)

# FCM
FCM_CREDENTIALS=./firebase-credentials.json

# Debug
PB_DEBUG=false
```

### Worker Initialization (main.go)

```go
w := worker.NewWorker(
    sysRepo,         // SystemStatusRepository
    reminderRepo,    // ReminderRepository
    userRepo,        // UserRepository
    fcmService,      // FCMSender
    schedCalculator, // ScheduleCalculator
    time.Duration(cfg.WorkerInterval)*time.Second,
)
w.Start(bgCtx)
```

---

## Time Field Handling

### Critical: parseTimeDB() Function

Times from database must be parsed correctly to avoid `Year()=1` errors.

**Supported Formats**:
1. RFC3339Nano: `2025-11-08T09:25:54.872123456Z`
2. RFC3339: `2025-11-08T09:25:54Z`
3. PocketBase: `2025-11-08 09:25:54.872Z` (space instead of T)
4. Without timezone: `2025-11-08 09:25:54`
5. With milliseconds: `2025-11-08 09:25:54.123`

**Implementation**:
```go
func parseTimeDB(s string) time.Time {
    if s == "" {
        return time.Time{}
    }
    
    formats := []string{
        time.RFC3339Nano,
        time.RFC3339,
        "2006-01-02 15:04:05.999Z",
        "2006-01-02 15:04:05Z",
        "2006-01-02 15:04:05",
        "2006-01-02 15:04:05.999",
    }
    
    for _, format := range formats {
        if t, err := time.Parse(format, s); err == nil {
            return t
        }
    }
    
    log.Printf("⚠️  parseTimeDB: failed to parse '%s'", s)
    return time.Time{}
}
```

### IsTimeValid() Helper

```go
func IsTimeValid(t time.Time) bool {
    return !t.IsZero() && t.Year() >= 2000
}

// Usage in Reminder model
func (r *Reminder) IsNextCRPSet() bool {
    return IsTimeValid(r.NextCRP)
}

func (r *Reminder) IsSnoozeUntilActive(now time.Time) bool {
    return IsTimeValid(r.SnoozeUntil) && r.SnoozeUntil.After(now)
}
```

---

## Error Handling

### FCM Token Errors

**Token Invalid/Unregistered**:
```
1. Mark user as FCM inactive
2. Skip notification for this user
3. Continue processing other reminders
4. Do NOT disable worker
```

**System Errors (401, 403, timeout)**:
```
1. Log error
2. Disable worker (prevents cascading failures)
3. Set system_status.last_error
4. Operator must manually re-enable
```

### Worker Error Recovery

```go
if isSystemFCMError(err) {
    // System-level FCM error
    sysRepo.DisableWorker(ctx)
    return
}

// User-level or reminder-level error
log.Printf("Error processing reminder: %v", err)
// Continue processing other reminders
```

### Infinite Loop Prevention

```go
// If next_action_at stuck in past > 1 hour
if now.Sub(reminder.NextActionAt) > time.Hour {
    log.Printf("⚠️  SAFETY: Recalculating next_action_at")
    reminder.NextActionAt = schedCalc.CalculateNextActionAt(reminder, now)
    
    if reminder.NextActionAt.IsZero() || reminder.NextActionAt.Before(now) {
        reminder.NextActionAt = now.Add(time.Minute)
        log.Printf("🚨 Emergency fallback: set to +1 minute")
    }
    
    reminderRepo.UpdateNextActionAt(ctx, reminder.ID, reminder.NextActionAt)
}
```

---

## Usage Examples

### Example 1: Daily Reminder with Retries

```json
{
  "title": "Take medication",
  "type": "recurring",
  "repeat_strategy": "none",
  "recurrence_pattern": {
    "type": "daily",
    "trigger_time_of_day": "08:00"
  },
  "max_crp": 3,
  "crp_interval_sec": 600
}
```

**Timeline**:
```
08:00 → FRP trigger, send notif (CRP 1/3)
08:10 → CRP 2/3
08:20 → CRP 3/3
Next day 08:00 → FRP trigger again (auto-repeat)
```

### Example 2: Weekly Report with Completion-Based Recurrence

```json
{
  "title": "Submit weekly report",
  "type": "recurring",
  "repeat_strategy": "crp_until_complete",
  "recurrence_pattern": {
    "type": "weekly",
    "day_of_week": 1,
    "trigger_time_of_day": "09:00"
  },
  "max_crp": 5,
  "crp_interval_sec": 3600
}
```

**Timeline**:
```
Mon 09:00 → FRP trigger (CRP 1/5)
Mon 10:00 → CRP 2/5
Mon 10:30 → User clicks COMPLETE
           → next_recurring = next Monday 10:30 (from completion time)
Next Mon 10:30 → FRP trigger from new time
```

### Example 3: One-Time with Limited Retries

```json
{
  "title": "Meeting reminder",
  "type": "one_time",
  "max_crp": 2,
  "crp_interval_sec": 300
}
```

**Timeline**:
```
14:00 → Send notif (CRP 1/2)
14:05 → Send notif (CRP 2/2)
14:05 → Status = "completed" (quota reached)
```

---

## File Structure

```
remiaq/
├── internal/
│   ├── worker/
│   │   └── worker.go              # Main worker logic (FRP/CRP processing)
│   ├── services/
│   │   ├── schedule_calculator.go # Schedule calculations
│   │   ├── reminder_service.go    # Business logic
│   │   ├── fcm_service.go         # Firebase notifications
│   │   └── lunar_calendar.go      # Lunar date calculations
│   ├── repository/
│   │   └── pocketbase/
│   │       ├── reminder_orm_repo.go   # DB operations
│   │       └── (parseTimeDB function here)
│   ├── models/
│   │   └── reminder.go            # Data models + helpers
│   ├── handlers/
│   │   └── reminder_handler.go    # HTTP handlers
│   └── middleware/
│       └── cors.go
├── cmd/
│   └── server/
│       └── main.go                # Server setup + worker init
├── config/
│   └── config.go                  # Configuration loading
└── migrations/
    └── (PocketBase migrations)
```

---

## Performance Considerations

### Database Indexes

Required for efficient worker queries:
```sql
CREATE INDEX idx_next_action_at ON reminders(next_action_at, status);
CREATE INDEX idx_user_id ON reminders(user_id, status);
```

### Worker Interval

- **Default**: 60 seconds
- **Min**: 10 seconds (high CPU usage)
- **Max**: 300 seconds (delayed notifications)
- **Recommended**: 60 seconds

### Query Optimization

Worker query:
```sql
SELECT * FROM reminders 
WHERE next_action_at <= NOW()
AND status = 'active'
AND (snooze_until IS NULL OR snooze_until <= NOW())
ORDER BY next_action_at ASC
```

Uses index on `(next_action_at, status)` for fast filtering.

---

## Testing Checklist

- [ ] Create one-time reminder, verify 3 CRP retries
- [ ] Create recurring daily reminder, verify auto-repeat
- [ ] Create recurring + crp_until_complete, complete it, verify next cycle from completion time
- [ ] Snooze reminder, verify it's skipped until snooze expires
- [ ] Delete reminder, verify it doesn't reappear
- [ ] Test with interval_seconds pattern
- [ ] Test with lunar calendar pattern
- [ ] Verify parseTimeDB handles all time formats
- [ ] Test worker safety check (next_action_at in past)
- [ ] Test FCM token error handling
- [ ] Test worker disable on system error

---

## Troubleshooting

### Problem: Worker not sending notifications

**Check**:
1. `system_status.worker_enabled = true`
2. Reminder `status = 'active'`
3. `next_action_at <= NOW()`
4. `snooze_until IS NULL` or `snooze_until <= NOW()`
5. User `is_fcm_active = true` and `fcm_token` is set

### Problem: Infinite loop / notifications every 10 seconds

**Cause**: `next_action_at` is stuck in past (likely from old data)

**Fix**:
1. Run worker with new safety check (automatically recalculates)
2. Or manually: `UPDATE reminders SET next_action_at = NOW() + INTERVAL '1 minute' WHERE id = ?`

### Problem: next_recurring not calculated on create

**Cause**: `CalculateNextRecurring()` fails or returns zero

**Check**:
1. `recurrence_pattern` is valid JSON
2. Pattern has required fields (type, interval, etc.)
3. `parseTimeDB()` correctly parses times

---

## Version History

### v2.0 (Current)
- ✅ Fixed CRP interval checking (now uses next_crp not last_sent_at)
- ✅ Fixed time parsing from database (multiple format support)
- ✅ Added safety check for infinite loops
- ✅ Improved error handling for FCM
- ✅ Better logging and debugging

### v1.0
- Basic FRP/CRP logic
- One-time and recurring reminders
- Database schema

---

## Contact & Support

For issues or questions:
1. Check logs in worker output
2. Verify database state
3. Test with simple reminder first
4. Check `.env` configuration