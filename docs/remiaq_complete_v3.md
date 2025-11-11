# RemiAq - Complete Technical Documentation v3

**Version**: 3.0  
**Last Updated**: 2025-11-09  
**Status**: Production Ready - Updated with latest fixes

---

## 📋 Table of Contents

1. [System Overview](#system-overview)
2. [Architecture](#architecture)
3. [Database Schema](#database-schema)
4. [Worker Logic - FRP & CRP](#worker-logic---frp--crp)
5. [API Documentation](#api-documentation)
6. [Field Definitions](#field-definitions)
7. [Code Changes v3](#code-changes-v3)
8. [Testing Checklist](#testing-checklist)
9. [Troubleshooting](#troubleshooting)

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

---

## Worker Logic - FRP & CRP

### Worker Cycle (Every 60 seconds)

```
1. Check if worker_enabled
2. Query: SELECT * FROM reminders 
          WHERE next_action_at <= NOW()
          AND status = 'active'
          AND (snooze_until IS NULL OR snooze_until <= NOW())
3. For each reminder:
   a. processReminder()
      - Check snooze (highest priority)
      - Check FRP (if type = recurring)
      - Check CRP (if max_crp > 0)
      - Recalc next_action_at
4. Update DB + Send FCM
```

### processReminder() Flow

```go
// STEP 0: Check if paused
if status == paused → return

// STEP 1: Check if snoozed
if snooze_until > now → recalc next_action_at, return

// STEP 2: Check ONE-TIME (if type = one_time)
if type == one_time {
    if !LastSentAt.IsSet() → Send (first time)
    else if CanSendCRP() → Send (retry)
    else → recalc next_action_at
    return
}

// STEP 3: Check FRP (if type = recurring)
if CanTriggerNow(NextRecurring) {
    if repeat_strategy == crp_until_complete {
        if LastCompletedAt > LastSentAt → processFRP()
        else → fall through to CRP
    } else {
        processFRP()
    }
}

// STEP 4: Check CRP
if CanSendCRP() → processCRP()

// STEP 5: Recalc next_action_at
nextAction = CalculateNextActionAt()
```

### FRP Trigger

**When**: `now >= next_recurring`

**Action**:
```
1. Send FCM notification
2. Update: last_sent_at = now
3. Reset: crp_count = 0, next_crp = next_recurring
4. Calculate next_recurring (luôn tính, dù repeat_strategy gì)
5. Recalculate: next_action_at = MIN(snooze, next_recurring, next_crp)
6. Clear: snooze_until = empty (clear snooze after FRP)
7. Update DB
```

### CRP Trigger

**When**: `max_crp > 0 && crp_count < max_crp && now >= next_crp`

**Action**:
```
1. Send FCM notification
2. Update: last_sent_at = now
3. Increment: crp_count++
4. Calculate: next_crp = now + crp_interval_sec
5. If one_time AND crp_count >= max_crp:
   - Set: status = "completed"
   - Clear: next_action_at = empty
6. Else:
   - Recalculate: next_action_at
7. Update DB
```

### CalculateNextActionAt Logic

```
candidates = []

// 1. Check snooze (highest priority)
if snooze_until > now → return snooze_until

// 2. Add next_recurring (for recurring only)
if type == recurring && next_recurring.valid → add to candidates

// 3. Add next_crp (if max_crp > 0 AND crp_count < max_crp)
if max_crp > 0 && crp_count < max_crp:
    if next_crp.valid → add next_crp
    else if last_sent_at.valid → add last_sent_at + interval
    else → add now (first time)

// 4. Return MIN(candidates)
return MIN(candidates) or empty if no candidates
```

### CalculateNextRecurring Logic

```
// Input: current NextRecurring, now time.Time
// Output: next NextRecurring after now

current = NextRecurring or now if zero

switch pattern.type:
  case interval_seconds:
    interval = pattern.interval_seconds
    next = current
    while next <= now:
        next = next + interval
    return next

  case daily:
    interval = pattern.interval days
    hour, minute = pattern.trigger_time_of_day
    next = current
    while next <= now:
        next = next + interval days
        next.SetTime(hour, minute)
    return next

  case weekly:
    // Find next target weekday
    target_weekday = pattern.day_of_week
    interval = pattern.interval weeks
    // Similar logic to daily

  case monthly:
    // Solar: use day_of_month
    // Lunar: use lunar calendar

  case lunar_last_day_of_month:
    // Last day of lunar month
```

---

## API Documentation

### Create One-Time Reminder

```
POST /api/reminders
Authorization: Bearer {token}
Content-Type: application/json

{
  "title": "One-time reminder",
  "description": "Optional",
  "type": "one_time",
  "max_crp": 3,
  "crp_interval_sec": 20,
  "status": "active"
}

Response:
{
  "success": true,
  "data": {
    "id": "reminder_123",
    "type": "one_time",
    "next_action_at": "2025-11-09T10:00:00Z",
    "next_crp": "2025-11-09T10:00:00Z",
    "status": "active"
  }
}
```

### Create Recurring Reminder

```
POST /api/reminders
Authorization: Bearer {token}
Content-Type: application/json

{
  "title": "Recurring reminder",
  "type": "recurring",
  "calendar_type": "solar",
  "repeat_strategy": "none",
  "recurrence_pattern": {
    "type": "interval_seconds",
    "interval_seconds": 180
  },
  "max_crp": 0,
  "status": "active"
}

Response:
{
  "success": true,
  "data": {
    "id": "reminder_456",
    "type": "recurring",
    "next_recurring": "2025-11-09T10:03:00Z",
    "next_action_at": "2025-11-09T10:03:00Z",
    "status": "active"
  }
}
```

### Complete Reminder

```
POST /api/reminders/{id}/complete
Authorization: Bearer {token}

Effect:
- one_time: status = "completed"
- recurring + none: crp_count = 0, FRP continues
- recurring + crp_until_complete: crp_count = 0, recalc next_recurring
```

### Snooze Reminder

```
POST /api/reminders/{id}/snooze
Authorization: Bearer {token}
Content-Type: application/json

{
  "duration": 300  // seconds
}

Effect:
- snooze_until = now + 300
- next_action_at = snooze_until
- Worker skips until snooze expires
```

---

## Field Definitions

### NextRecurring vs NextActionAt

| Field | Purpose | Khi nào thay đổi |
|-------|---------|------------------|
| **NextRecurring** | Base time để tính lần lặp tiếp (FRP) | Chỉ khi FRP trigger |
| **NextActionAt** | Thời điểm worker query reminder | Mỗi lần xử lý (CRP, snooze, v.v.) |

**Example:**
```
12:00 - Create recurring, interval_seconds=180, max_crp=3
  NextRecurring = 12:03
  NextActionAt = 12:00 (MIN của FRP và CRP)

12:00 - FRP trigger
  NextRecurring = 12:06 (tính tiếp)
  NextCRP = 12:00:20
  NextActionAt = 12:00:20 (CRP sớm hơn)

12:00:20 - CRP 1
  LastSentAt = 12:00:20
  CRPCount = 1
  NextCRP = 12:00:40
  NextActionAt = 12:00:40

12:00:40 - CRP 2
  NextCRP = 12:01:00
  NextActionAt = 12:01:00

12:01:00 - CRP 3 (quota đầy)
  CRPCount = 3
  NextActionAt = NextRecurring(12:06) (chỉ chờ FRP tiếp)

12:06 - FRP 2
  NextRecurring = 12:09
  NextActionAt = 12:09
```

### MaxCRP Cases

| MaxCRP | Meaning | Hành động |
|--------|---------|----------|
| 0 | Gửi 1 lần only | Gửi FRP/CRP, xong |
| > 0 | Gửi tối đa X lần | Gửi FRP + CRP 1,2,...,max |

---

## Code Changes v3

### 1. ValidateData() - Kiểm tra dữ liệu trước xử lý

```go
func (r *Reminder) ValidateData() (bool, string) {
    if !IsTimeValid(r.NextActionAt) {
        return false, "NextActionAt không hợp lệ"
    }
    if r.Type != ReminderTypeOneTime && r.Type != ReminderTypeRecurring {
        return false, "Type phải là one_time hoặc recurring"
    }
    if r.Status != ReminderStatusActive && r.Status != ReminderStatusCompleted && r.Status != ReminderStatusPaused {
        return false, "Status không hợp lệ"
    }
    if r.Title == "" {
        return false, "Title không được trống"
    }
    if r.MaxCRP < 0 {
        return false, "MaxCRP không được âm"
    }
    if r.MaxCRP > 0 && r.CRPIntervalSec <= 0 {
        return false, "CRPIntervalSec phải > 0"
    }
    if r.Type == ReminderTypeRecurring && !IsTimeValid(r.NextRecurring) {
        return false, "Recurring phải có NextRecurring"
    }
    if r.UserID == "" {
        return false, "UserID không được trống"
    }
    return true, ""
}
```

### 2. processReminder() - Xử lý one-time riêng

```go
func (w *Worker) processReminder(ctx context.Context, reminder *models.Reminder, now time.Time) error {
    if valid, reason := reminder.ValidateData(); !valid {
        log.Printf("❌ Validation failed: %s", reason)
        return nil
    }

    // Check paused
    if reminder.Status == models.ReminderStatusPaused {
        return nil
    }

    // Check snooze
    if reminder.IsSnoozeUntilActive(now) {
        nextAction := w.schedCalc.CalculateNextActionAt(reminder, now)
        if !nextAction.Equal(reminder.NextActionAt) {
            _ = w.reminderRepo.UpdateNextActionAt(ctx, reminder.ID, nextAction)
        }
        return nil
    }

    // ONE-TIME handling
    if reminder.Type == models.ReminderTypeOneTime {
        if reminder.CanSendFRPOneTime() {
            return w.processCRPForOneTime(ctx, reminder, now)
        }
        if w.schedCalc.CanSendCRP(reminder, now) {
            return w.processCRPForOneTime(ctx, reminder, now)
        }
        return nil
    }

    // RECURRING FRP handling
    if reminder.CanTriggerNow(reminder.NextRecurring) {
        if reminder.RepeatStrategy == models.RepeatStrategyCRPUntilComplete {
            if reminder.LastCompletedAt.After(reminder.LastSentAt) {
                return w.processFRP(ctx, reminder, now)
            }
        } else {
            return w.processFRP(ctx, reminder, now)
        }
    }

    // CRP handling
    if w.schedCalc.CanSendCRP(reminder, now) {
        return w.processCRP(ctx, reminder, now)
    }

    // Recalc next_action_at
    nextAction := w.schedCalc.CalculateNextActionAt(reminder, now)
    if !nextAction.Equal(reminder.NextActionAt) {
        _ = w.reminderRepo.UpdateNextActionAt(ctx, reminder.ID, nextAction)
    }

    return nil
}
```

### 3. processFRP() - Update NextActionAt sau khi tính NextRecurring

```go
func (w *Worker) processFRP(ctx context.Context, reminder *models.Reminder, now time.Time) error {
    log.Printf("📅 FRP triggered for reminder %s", reminder.ID)

    if err := w.sendNotification(ctx, reminder); err != nil {
        log.Printf("❌ FRP failed, snoozing 60s: %v", err)
        reminder.SnoozeUntil = now.Add(60 * time.Second)
        reminder.NextActionAt = reminder.SnoozeUntil
        _ = w.reminderRepo.Update(ctx, reminder)
        return err
    }

    reminder.LastSentAt = now
    reminder.CRPCount = 0
    reminder.NextCRP = reminder.NextRecurring
    reminder.SnoozeUntil = time.Time{} // Clear snooze

    nextRecurring, err := w.schedCalc.CalculateNextRecurring(reminder, now)
    if err != nil {
        nextRecurring = now.Add(24 * time.Hour)
    }
    reminder.NextRecurring = nextRecurring

    // ✅ CRITICAL: Recalc NextActionAt AFTER NextRecurring updated
    reminder.NextActionAt = w.schedCalc.CalculateNextActionAt(reminder, now)

    if err := w.reminderRepo.Update(ctx, reminder); err != nil {
        return fmt.Errorf("failed to update after FRP: %w", err)
    }

    log.Printf("✅ FRP processed. Next FRP: %s", nextRecurring.Format("15:04:05"))
    return nil
}
```

### 4. CalculateNextActionAt() - CRP chỉ add khi max_crp > 0

```go
func (c *ScheduleCalculator) CalculateNextActionAt(reminder *models.Reminder, now time.Time) time.Time {
    candidates := []time.Time{}

    // 1. Snooze (highest priority)
    if reminder.IsSnoozeUntilActive(now) {
        return reminder.SnoozeUntil
    }

    // 2. NextRecurring (for recurring)
    if reminder.Type == models.ReminderTypeRecurring && reminder.IsNextRecurringSet() {
        candidates = append(candidates, reminder.NextRecurring)
    }

    // 3. NextCRP (ONLY if max_crp > 0 AND crp_count < max_crp)
    if reminder.MaxCRP > 0 && reminder.CRPCount < reminder.MaxCRP {
        if reminder.IsNextCRPSet() {
            candidates = append(candidates, reminder.NextCRP)
        } else if reminder.IsLastSentAtSet() {
            nextCRP := reminder.LastSentAt.Add(time.Duration(reminder.CRPIntervalSec) * time.Second)
            candidates = append(candidates, nextCRP)
        } else {
            candidates = append(candidates, now)
        }
    }

    // 4. Return MIN
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
```

---

## Testing Checklist

### 12 Recurring Test Cases

```
1. repeat_strategy=none, max_crp=0, interval_seconds → Auto-repeat, no retry ✅
2. repeat_strategy=none, max_crp=3, interval_seconds → Auto-repeat, retry 3x ✅
3. repeat_strategy=none, max_crp=0, daily → Auto-repeat daily ✅
4. repeat_strategy=none, max_crp=3, daily → Auto-repeat daily, retry 3x ✅
5. repeat_strategy=none, max_crp=0, weekly → Auto-repeat weekly ✅
6. repeat_strategy=none, max_crp=3, weekly → Auto-repeat weekly, retry 3x ✅
7. repeat_strategy=none, max_crp=0, monthly → Auto-repeat monthly ✅
8. repeat_strategy=crp_until_complete, max_crp=0, interval_seconds → Wait user, no retry ✅
9. repeat_strategy=crp_until_complete, max_crp=3, interval_seconds → Wait user, retry 3x ✅
10. repeat_strategy=crp_until_complete, max_crp=3, daily → Wait user daily, retry 3x ✅
11. repeat_strategy=crp_until_complete, max_crp=3, lunar_monthly → Wait user lunar ✅
12. repeat_strategy=crp_until_complete + user complete → Recalc from complete time ✅
```

### One-Time Test Cases

```
1. one_time, max_crp=0 → Send 1 time ✅
2. one_time, max_crp=3 → Send 3 times with interval ✅
3. one_time, future NextActionAt → Send at scheduled time ✅
4. one_time + user complete early → Stop immediately ✅
```

---

## Troubleshooting

### Problem: Reminder không trigger

**Check:**
1. `status = 'active'` ✅
2. `next_action_at <= NOW` ✅
3. `snooze_until` không active ✅
4. User `is_fcm_active = true` ✅
5. `ValidateData()` pass ✅

### Problem: NextActionAt sai

**Check:**
1. FRP trigger → Recalc NextActionAt? ✅
2. Clear SnoozeUntil sau FRP? ✅
3. CRP logic đúng (max_crp > 0)? ✅

### Problem: Recurring không lặp

**Check:**
1. `repeat_strategy` check? (nếu crp_until_complete, check LastCompletedAt) ✅
2. `CalculateNextRecurring()` output đúng? ✅
3. NextRecurring update sau FRP? ✅

---

## Version History

### v3.0 (Current)
- ✅ Add ValidateData() check
- ✅ Handle one-time reminder riêng
- ✅ Fix FRP trigger check with repeat_strategy
- ✅ Fix CalculateNextActionAt() - only add CRP if max_crp > 0
- ✅ Clear snooze_until after FRP
- ✅ Auto-snooze 60s if sendNotification fails
- ✅ Recalc NextActionAt after NextRecurring updated

### v2.0
- Fixed CRP interval checking
- Fixed time parsing from database
- Added safety check for infinite loops
- Better error handling

### v1.0
- Basic FRP/CRP logic
- One-time and recurring reminders
- Database schema



{
  "recurrence_pattern": {
    "type": "interval_seconds",
    "interval_seconds": 180
  }
}
json{
  "recurrence_pattern": {
    "type": "daily",
    "interval": 1,
    "trigger_time_of_day": "08:00"
  }
}
json{
  "recurrence_pattern": {
    "type": "daily",
    "interval": 2,
    "trigger_time_of_day": "09:30"
  }
}
json{
  "recurrence_pattern": {
    "type": "weekly",
    "interval": 1,
    "day_of_week": 1,
    "trigger_time_of_day": "09:00"
  }
}
json{
  "recurrence_pattern": {
    "type": "weekly",
    "interval": 2,
    "day_of_week": 3,
    "trigger_time_of_day": "14:00"
  }
}
json{
  "recurrence_pattern": {
    "type": "monthly",
    "interval": 1,
    "day_of_month": 5,
    "trigger_time_of_day": "10:00"
  }
}
json{
  "recurrence_pattern": {
    "type": "monthly",
    "interval": 1,
    "day_of_month": 15,
    "trigger_time_of_day": "18:00"
  }
}
json{
  "recurrence_pattern": {
    "type": "lunar_last_day_of_month",
    "trigger_time_of_day": "20:00"
  }
}
```

**day_of_week reference:**
```
0 = Sunday (Chủ nhật)
1 = Monday (Thứ 2)
2 = Tuesday (Thứ 3)
3 = Wednesday (Thứ 4)
4 = Thursday (Thứ 5)
5 = Friday (Thứ 6)
6 = Saturday (Thứ 7)


trigger_time_of_day sẽ được tự tạo dựa vào NextActionAt 

{
  "recurrence_pattern": {
    // ========================================
    // REQUIRED - Luôn phải có
    // ========================================
    "type": "daily|weekly|monthly|lunar_last_day_of_month|interval_seconds",
    
    // ========================================
    // OPTIONAL - Có hoặc không
    // ========================================
    "interval": 1,
    // - Default: 1
    // - Meaning: Mỗi X ngày/tuần/tháng
    // - VD: interval=2 → mỗi 2 ngày, mỗi 2 tuần
    // - Xuất hiện khi: type ∈ {daily, weekly, monthly}
    // - ❌ KHÔNG dùng: interval_seconds, lunar_last_day_of_month
    
    "trigger_time_of_day": "HH:MM",
    // - Format: "08:00", "14:30", "23:59"
    // - Default: "00:00"
    // - Meaning: Giờ trigger mỗi ngày
    // - Xuất hiện khi: type ∈ {daily, weekly, monthly, lunar_last_day_of_month}
    // - ❌ KHÔNG dùng: interval_seconds (không cần giờ cố định)
    
    "day_of_week": 0, -------------------CHƯA CÓ 
    // - Range: 0-6 (0=Sun, 1=Mon, ..., 6=Sat)
    // - Meaning: Ngày trong tuần
    // - Xuất hiện khi: type == "weekly" ✅ BẮT BUỘC
    // - ❌ KHÔNG dùng: daily, monthly, interval_seconds, lunar_*
    
    "day_of_month": 5,
    // - Range: 1-31
    // - Meaning: Ngày trong tháng
    // - Xuất hiện khi: type == "monthly" ✅ BẮT BUỘC
    // - ❌ KHÔNG dùng: daily, weekly, interval_seconds, lunar_*
    // - ⚠️ Edge case: day=31 nhưng tháng có 30 ngày → auto adjust last day
    
    "interval_seconds": 180,
    // - Range: > 0
    // - Meaning: Khoảng cách giữa các trigger (giây)
    // - VD: 180 = 3 phút, 86400 = 1 ngày
    // - Xuất hiện khi: type == "interval_seconds" ✅ BẮT BUỘC
    // - ❌ KHÔNG dùng: daily, weekly, monthly, lunar_*
    
    "calendar_type": "solar|lunar"
    // - Default: "solar" (dương lịch)
    // - Meaning: Loại lịch
    // - ✅ DÙNG CHO: type ∈ {monthly}
    // - ❌ KHÔNG dùng: daily, weekly, interval_seconds, lunar_last_day_of_month
    // - Note: lunar_last_day_of_month đã implicit lunar
  }
}