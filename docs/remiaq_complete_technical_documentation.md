# RemiAq - Complete Technical Documentation

**Version**: 4.0  
**Last Updated**: 2025-11-09  
**Status**: Production Ready

---

## 📋 Table of Contents

1. [System Overview](#system-overview)
2. [Architecture](#architecture)
3. [API Data Reception from Client](#api-data-reception-from-client)
4. [Worker Processing Logic](#worker-processing-logic)
5. [Database Schema](#database-schema)
6. [Field Definitions](#field-definitions)
7. [Testing Checklist](#testing-checklist)
8. [Troubleshooting](#troubleshooting)

---

## System Overview

RemiAq là reminder management system hỗ trợ:
- **One-time reminders**: Gửi 1 lần hoặc retry X lần rồi dừng
- **Recurring reminders**: Lặp lại theo lịch (mỗi ngày, tuần, tháng, âm lịch, hoặc interval seconds)
- **Two repeat strategies**: Auto-repeat hoặc chờ user complete
- **Firebase Cloud Messaging (FCM)**: Gửi notification qua FCM
- **Background worker**: Xử lý reminders mỗi 60 giây

### Key Concepts

**FRP (Father Recurrence Pattern)**
- Main recurring schedule (chỉ cho recurring reminders)
- Trigger theo lịch được cấu hình
- Base time để tính lần lặp tiếp theo
- Không thay đổi khi có CRP/snooze

**CRP (Child Repeat Pattern)**
- Retry/notification resend trong một chu kỳ
- Áp dụng cho cả one-time và recurring
- Limited bởi `max_crp` (0 = không retry)
- Interval tính bằng giây

**repeat_strategy**
- `none`: Auto-advance theo lịch, không phụ thuộc user complete
- `crp_until_complete`: Chờ user bấm complete mới tính lần lặp tiếp theo

---

## Architecture

```
┌────────────────────────────────────────────────────────────┐
│                 PocketBase Server                          │
│                                                            │
│  ┌──────────────────────────────────────────────────────┐  │
│  │          REST API Handlers                           │  │
│  │  - CreateReminder                                    │  │
│  │  - UpdateReminder                                    │  │
│  │  - CompleteReminder                                  │  │
│  │  - SnoozeReminder                                    │  │
│  └──────────────────────────────────────────────────────┘  │
│                        ↓                                    │
│  ┌──────────────────────────────────────────────────────┐  │
│  │          ReminderService                             │  │
│  │  - Business Logic                                    │  │
│  │  - Validation                                        │  │
│  │  - State Management                                  │  │
│  └──────────────────────────────────────────────────────┘  │
│                        ↓                                    │
│  ┌──────────────────────────────────────────────────────┐  │
│  │          ScheduleCalculator                          │  │
│  │  - CalculateNextRecurring()                          │  │
│  │  - CalculateNextActionAt()                           │  │
│  │  - CanSendCRP()                                      │  │
│  └──────────────────────────────────────────────────────┘  │
│                        ↓                                    │
│  ┌──────────────────────────────────────────────────────┐  │
│  │          ORM Repository                              │  │
│  │  - Database Operations                               │  │
│  │  - Time Parsing (Multiple Formats)                   │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────┘
           ↓                          ↓
    ┌──────────────┐        ┌──────────────────┐
    │   SQLite DB  │        │   FCM Service    │
    └──────────────┘        └──────────────────┘

┌────────────────────────────────────────────────────────────┐
│      Background Worker (Every 60 seconds)                  │
│                                                            │
│  1. Check worker_enabled                                   │
│  2. Query: WHERE next_action_at <= NOW()                   │
│  3. For each reminder:                                     │
│     a. Check snooze                                        │
│     b. Check FRP (priority)                                │
│     c. Check CRP                                           │
│     d. Recalc next_action_at                               │
│  4. Update DB + Send FCM                                   │
└────────────────────────────────────────────────────────────┘
```

---

## API Data Reception from Client

### Create Reminder API

**Endpoint**: `POST /api/reminders`  
**Authentication**: Bearer Token Required  
**Content-Type**: `application/json`

#### Request Body Structure

```json
{
  "title": "Meeting with team",
  "description": "Weekly team meeting",
  "type": "recurring",
  "calendar_type": "solar",
  "next_action_at": "2025-11-10T09:00:00Z",
  "recurrence_pattern": {
    "type": "weekly",
    "interval": 1,
    "day_of_week": 1,
    "trigger_time_of_day": ""
  },
  "repeat_strategy": "none",
  "max_crp": 3,
  "crp_interval_sec": 300,
  "for_test": 60
}
```

#### Field Processing Logic

1. **Authentication Check**: 
   - Lấy `authRecord.Id` từ request context
   - Gán `user_id = authRecord.Id`

2. **Temporary Struct Decoding**:
   ```go
   var tempReminder struct {
       models.Reminder
       ForTestSeconds int `json:"for_test"`
   }
   ```
   - Xử lý field `for_test` để test nhanh
   - Nếu `for_test > 0`: set `next_action_at = now + for_test seconds`

3. **Auto-generation of trigger_time_of_day**:
   ```go
   if reminder.Type == "recurring" && reminder.RecurrencePattern.TriggerTimeOfDay == "" {
       reminder.RecurrencePattern.TriggerTimeOfDay = reminder.NextActionAt.Format("15:04")
   }
   ```
   - Tự động generate từ `next_action_at` format "HH:MM"
   - Client KHÔNG được gửi `trigger_time_of_day` (return 400 error)

4. **Validation Rules**:
   - `title`: Required
   - `type`: Must be "one_time" or "recurring"
   - `next_action_at`: Required, must be valid time
   - `recurrence_pattern`: Required for recurring type
   - `max_crp > 0` requires `crp_interval_sec > 0`

#### Response Structure

```json
{
  "success": true,
  "message": "Reminder created successfully",
  "data": {
    "id": "reminder_123",
    "next_recurring": "2025-11-10T09:00:00Z",
    "next_crp": "2025-11-10T09:00:00Z", 
    "next_action_at": "2025-11-10T09:00:00Z",
    "trigger_time_of_day": "09:00"
  }
}
```

### Other API Endpoints

#### Get Reminder
- `GET /api/reminders/:id` - Get reminder by ID
- `GET /api/reminders/mine` - Get current user's reminders

#### Update Reminder  
- `PUT /api/reminders/:id` - Update existing reminder

#### Complete Reminder
- `POST /api/reminders/:id/complete` - Mark as completed
  - one_time: Mark completed
  - recurring + none: Reset CRP, FRP continues  
  - recurring + crp_until_complete: Reset CRP + recalc next_recurring

#### Snooze Reminder
- `POST /api/reminders/:id/snooze` - Snooze for duration
  ```json
  { "duration": 300 } // seconds
  ```

#### Delete Reminder
- `DELETE /api/reminders/:id` - Delete reminder

---

## Worker Processing Logic

### Worker Configuration
- **Interval**: 60 seconds (configurable)
- **Enabled Check**: Verifies `system_status.worker_enabled` before each run
- **Error Handling**: System errors disable worker, user errors are logged

### Worker Cycle Flow

```go
func (w *Worker) runOnce(ctx context.Context) {
    // 1. Check if worker enabled
    enabled, err := w.sysRepo.IsWorkerEnabled(ctx)
    if !enabled { return }
    
    // 2. Query due reminders
    reminders, err := w.reminderRepo.GetDueReminders(ctx, time.Now().UTC())
    
    // 3. Process each reminder
    for _, reminder := range reminders {
        w.processReminder(ctx, reminder, now)
    }
}
```

### processReminder Logic

#### Step 0: Validation & Initial Checks
```go
// Validate reminder data integrity
if valid, reason := reminder.ValidateData(); !valid {
    log.Printf("Validation failed: %s", reason)
    return nil
}

// Skip paused reminders
if reminder.Status == "paused" { return nil }

// Handle snoozed reminders
if reminder.IsSnoozeUntilActive(now) {
    nextAction := w.schedCalc.CalculateNextActionAt(reminder, now)
    w.reminderRepo.UpdateNextActionAt(ctx, reminder.ID, nextAction)
    return nil
}
```

#### Step 1: One-Time Reminder Processing
```go
if reminder.Type == "one_time" {
    if !reminder.LastSentAt.IsZero() {
        // First time sending
        return w.processCRPForOneTime(ctx, reminder, now)
    } else if w.schedCalc.CanSendCRP(reminder, now) {
        // CRP retry
        return w.processCRPForOneTime(ctx, reminder, now)
    }
    return nil
}
```

#### Step 2: FRP (Father Recurrence Pattern) Processing
```go
if reminder.CanTriggerNow(reminder.NextRecurring) {
    if reminder.RepeatStrategy == "crp_until_complete" {
        // Wait for user completion
        if reminder.LastCompletedAt.After(reminder.LastSentAt) {
            return w.processFRP(ctx, reminder, now)
        }
    } else {
        // Auto-advance strategy
        return w.processFRP(ctx, reminder, now)
    }
}
```

#### Step 3: CRP (Child Repeat Pattern) Processing  
```go
if w.schedCalc.CanSendCRP(reminder, now) {
    return w.processCRP(ctx, reminder, now)
}
```

### FRP Processing Details

```go
func (w *Worker) processFRP(ctx context.Context, reminder *models.Reminder, now time.Time) error {
    // 1. Send FCM notification
    if err := w.sendNotification(ctx, reminder); err != nil {
        // On error, snooze for 60 seconds
        reminder.SnoozeUntil = now.Add(60 * time.Second)
        reminder.NextActionAt = reminder.SnoozeUntil
        w.reminderRepo.Update(ctx, reminder)
        return err
    }
    
    // 2. Update tracking
    reminder.LastSentAt = now
    reminder.CRPCount = 0
    reminder.NextCRP = reminder.NextRecurring
    
    // 3. Calculate next FRP occurrence
    nextRecurring, err := w.schedCalc.CalculateNextRecurring(reminder, now)
    reminder.NextRecurring = nextRecurring
    
    // 4. Recalculate next action time
    reminder.NextActionAt = w.schedCalc.CalculateNextActionAt(reminder, now)
    
    // 5. Update database
    return w.reminderRepo.Update(ctx, reminder)
}
```

### CRP Processing Details

```go
func (w *Worker) processCRP(ctx context.Context, reminder *models.Reminder, now time.Time) error {
    // 1. Send FCM notification
    if err := w.sendNotification(ctx, reminder); err != nil {
        return err
    }
    
    // 2. Update tracking
    reminder.LastSentAt = now
    reminder.CRPCount++
    
    // 3. Check if reached max CRP
    if reminder.MaxCRP > 0 && reminder.CRPCount >= reminder.MaxCRP {
        if reminder.Type == "one_time" {
            reminder.Status = "completed"
        }
        // Reset for next FRP cycle
        reminder.CRPCount = 0
        reminder.NextCRP = reminder.NextRecurring
    } else {
        // Schedule next CRP
        reminder.NextCRP = now.Add(time.Duration(reminder.CRPIntervalSec) * time.Second)
    }
    
    // 4. Recalculate next action time
    reminder.NextActionAt = w.schedCalc.CalculateNextActionAt(reminder, now)
    
    // 5. Update database
    return w.reminderRepo.Update(ctx, reminder)
}
```

### FCM Notification Handling

```go
func (w *Worker) sendNotification(ctx context.Context, reminder *models.Reminder) error {
    // 1. Get user FCM token
    user, err := w.userRepo.GetByID(ctx, reminder.UserID)
    
    // 2. Send FCM notification
    err = w.fcmSender.SendNotification(
        user.FCMToken,
        reminder.Title,
        reminder.Description,
    )
    
    // 3. Handle FCM errors
    if err != nil {
        if isTokenError(err.Error()) {
            // Disable user FCM on token errors
            w.userRepo.DisableFCM(ctx, user.ID)
        } else if isSystemFCMError(err) {
            // Disable worker on system errors
            w.sysRepo.DisableWorker(ctx)
            return err
        }
    }
    
    return nil
}
```

---

## Database Schema

### reminders Table

| Field | Type | Description |
|-------|------|-------------|
| id | text | Reminder ID |
| user_id | relation | Owner |
| title | text | Reminder title |
| description | text | Details |
| type | select | `one_time` hoặc `recurring` |
| status | select | `active`, `completed`, `paused` |
| **NextRecurring** | datetime | **Thời điểm FRP tiếp theo (base để tính)** |
| recurrence_pattern | json | Pattern config (type, interval, ...) |
| repeat_strategy | select | `none` hoặc `crp_until_complete` |
| calendar_type | select | `solar` hoặc `lunar` |
| **NextCRP** | datetime | Thời điểm CRP tiếp theo |
| **NextActionAt** | datetime | **Thời điểm worker sẽ check (MIN(snooze, frp, crp))** |
| max_crp | number | Max retries (0 = no retry) |
| crp_count | number | Current retry count |
| crp_interval_sec | number | Retry interval (seconds) |
| last_sent_at | datetime | Last notification sent |
| last_completed_at | datetime | User completed |
| last_crp_completed_at | datetime | User completed current CRP cycle |
| snooze_until | datetime | Snoozed until when |
| created | datetime | Created at |
| updated | datetime | Updated at |

### system_status Table

| Field | Type | Description |
|-------|------|-------------|
| mid | number | System ID (always 1) |
| worker_enabled | boolean | Worker enabled flag |
| last_error | text | Last error message |
| updated | datetime | Last updated |

### users Table

| Field | Type | Description |
|-------|------|-------------|
| id | text | User ID |
| email | text | User email |
| fcm_token | text | FCM token for notifications |
| is_fcm_active | boolean | FCM enabled flag |
| fcm_error | text | Last FCM error |
| created | datetime | Created at |
| updated | datetime | Updated at |

---

## Field Definitions

### Reminder Type Values
- `one_time`: Reminder chỉ xảy ra một lần
- `recurring`: Reminder lặp lại theo lịch

### Calendar Type Values  
- `solar`: Dương lịch
- `lunar`: Âm lịch

### Repeat Strategy Values
- `none`: Tự động lặp lại, không phụ thuộc user action
- `crp_until_complete`: Chờ user complete mới tính lần lặp tiếp theo

### Status Values
- `active`: Reminder đang active
- `completed`: Reminder đã hoàn thành (one-time only)
- `paused`: Reminder tạm dừng

### Recurrence Pattern Types
- `daily`: Lặp hàng ngày
- `weekly`: Lặp hàng tuần  
- `monthly`: Lặp hàng tháng
- `lunar_last_day_of_month`: Ngày cuối tháng âm lịch
- `interval_seconds`: Lặp theo số giây

---

## Testing Checklist

### API Testing
- [ ] Create one-time reminder
- [ ] Create recurring reminder
- [ ] Get reminder by ID
- [ ] Get user reminders
- [ ] Update reminder
- [ ] Complete reminder
- [ ] Snooze reminder
- [ ] Delete reminder
- [ ] Error handling: invalid data
- [ ] Error handling: unauthorized access

### Worker Testing
- [ ] One-time reminder with no CRP
- [ ] One-time reminder with CRP
- [ ] Recurring reminder with none strategy
- [ ] Recurring reminder with crp_until_complete strategy
- [ ] Snooze functionality
- [ ] FCM notification sending
- [ ] Error handling: FCM token errors
- [ ] Error handling: System FCM errors

### Integration Testing
- [ ] Full reminder lifecycle
- [ ] Multiple users scenario
- [ ] High load scenario
- [ ] Database recovery scenario

---

## Troubleshooting

### Common Issues

#### Worker Not Running
- Check `system_status.worker_enabled = true`
- Verify worker interval configuration
- Check system logs for errors

#### FCM Notifications Not Sending
- Verify user `fcm_token` is valid
- Check `is_fcm_active = true`
- Review FCM configuration credentials

#### Reminders Not Triggering
- Verify `next_action_at` is set correctly
- Check reminder `status = active`
- Ensure `snooze_until` is not set in future

#### Database Time Issues
- All times stored in UTC
- Use `parseTimeDB()` function for time parsing
- Ensure timezone consistency

### Logging & Monitoring

Key log messages to monitor:
```
Worker started (interval=60s)
Worker: Processing X due reminders
Worker: FRP triggered for reminder Y
Worker: CRP triggered for reminder Z (count: A/B)
❌ Validation failed: [reason]
❌ FCM ERROR: [error details]
```

### Performance Optimization

#### Database Indexes
```sql
CREATE INDEX idx_next_action_at ON reminders(next_action_at, status);
CREATE INDEX idx_user_id ON reminders(user_id, status);
CREATE INDEX idx_snooze_until ON reminders(snooze_until);
```

#### Worker Configuration
- **Default interval**: 60 seconds
- **Minimum**: 10 seconds (high CPU usage)
- **Maximum**: 300 seconds (delayed notifications)

---

## Conclusion

RemiAq là hệ thống reminder notification hoàn chỉnh với:
- API RESTful cho client interactions
- Background worker xử lý reminders mỗi 60 giây
- Hỗ trợ cả one-time và recurring reminders
- Hai chiến lược lặp: auto-advance và wait-for-complete
- Integration với Firebase Cloud Messaging
- Database schema được tối ưu cho performance

Hệ thống được thiết kế để scalable và reliable, với comprehensive error handling và monitoring capabilities.