# Worker V2 - API Examples

Tài liệu này chứa các ví dụ **API calls** để tạo và quản lý reminders cho từng scenario trong Worker V2.

**Base URL**: `http://localhost:8090/api`

**Authentication**: Bearer token (JWT từ PocketBase)

---

## Table of Contents

1. [Authentication](#authentication)
2. [ONE-TIME Reminders](#one-time-reminders)
   - [Case 1: ONE-TIME - No CRP](#case-1-one-time---no-crp)
   - [Case 2 & 3: ONE-TIME - With CRP](#case-2--3-one-time---with-crp)
3. [RECURRING - No Until Complete](#recurring---no-until-complete)
   - [Case 4: RECURRING - No CRP - No Until Complete](#case-4-recurring---no-crp---no-until-complete)
   - [Case 5: RECURRING - With CRP - No Until Complete](#case-5-recurring---with-crp---no-until-complete)
4. [RECURRING - Until Complete](#recurring---until-complete)
   - [Case 6: RECURRING - No CRP - Until Complete](#case-6-recurring---no-crp---until-complete)
   - [Case 7: RECURRING - With CRP - Until Complete](#case-7-recurring---with-crp---until-complete)
5. [Reminder Management](#reminder-management)
   - [Complete Reminder](#complete-reminder)
   - [Snooze Reminder](#snooze-reminder)
   - [Get Reminder](#get-reminder)
   - [Update Reminder](#update-reminder)
   - [Delete Reminder](#delete-reminder)

---

## Authentication

### Login

```http
POST /api/collections/users/auth-with-password
Content-Type: application/json

{
  "identity": "user@example.com",
  "password": "your_password"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "record": {
    "id": "user_abc123",
    "email": "user@example.com",
    ...
  }
}
```

**Usage:** Dùng `token` trong header cho các request sau:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

## ONE-TIME Reminders

### Case 1: ONE-TIME - No CRP

**Scenario:** Nhắc họp khách hàng vào 14:00 ngày 2025-11-27, không retry.

```http
POST /api/reminders
Authorization: Bearer {your_token}
Content-Type: application/json

{
  "title": "Họp khách hàng",
  "description": "Họp với khách hàng ABC về dự án mới",
  "type": "one_time",
  "calendar_type": "solar",
  "next_action_at": "2025-11-27T14:00:00Z",
  "max_crp": 0,
  "status": "active",
  "tag": "work"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "reminder_xyz789",
    "title": "Họp khách hàng",
    "type": "one_time",
    "next_action_at": "2025-11-27T14:00:00Z",
    "max_crp": 0,
    "is_sended_one_time": false,
    "status": "active",
    ...
  }
}
```

**Worker behavior:**
- ⏰ **14:00**: Gửi notification → Set `status = completed`
- ✅ **Không gửi lại**

---

### Case 2 & 3: ONE-TIME - With CRP

**Scenario:** Nhắc nộp báo cáo 17:00, retry mỗi 10 phút tối đa 3 lần.

```http
POST /api/reminders
Authorization: Bearer {your_token}
Content-Type: application/json

{
  "title": "Nộp báo cáo tháng 11",
  "description": "Deadline báo cáo doanh thu tháng 11",
  "type": "one_time",
  "calendar_type": "solar",
  "next_action_at": "2025-11-27T17:00:00Z",
  "max_crp": 3,
  "crp_interval_sec": 600,
  "status": "active",
  "tag": "work"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "reminder_abc456",
    "title": "Nộp báo cáo tháng 11",
    "type": "one_time",
    "next_action_at": "2025-11-27T17:00:00Z",
    "max_crp": 3,
    "crp_interval_sec": 600,
    "crp_count": 0,
    "is_sended_one_time": false,
    "status": "active",
    ...
  }
}
```

**Worker behavior:**
- ⏰ **17:00**: FRP - Gửi notification
- ⏰ **17:10**: CRP #1 - Retry
- ⏰ **17:20**: CRP #2 - Retry
- ⏰ **17:30**: CRP #3 - Retry → Set `status = completed`

---

## RECURRING - No Until Complete

### Case 4: RECURRING - No CRP - No Until Complete

**Scenario:** Nhắc uống nước mỗi 2 giờ, bắt đầu từ 8:00 sáng.

```http
POST /api/reminders
Authorization: Bearer {your_token}
Content-Type: application/json

{
  "title": "Uống nước",
  "description": "Uống 1 cốc nước để giữ sức khỏe",
  "type": "recurring",
  "calendar_type": "solar",
  "recurrence_pattern": {
    "type": "interval_seconds",
    "interval_seconds": 7200
  },
  "next_recurring": "2025-11-27T08:00:00Z",
  "max_crp": 0,
  "repeat_strategy": "none",
  "status": "active",
  "tag": "health"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "reminder_health01",
    "title": "Uống nước",
    "type": "recurring",
    "recurrence_pattern": {
      "type": "interval_seconds",
      "interval_seconds": 7200
    },
    "next_recurring": "2025-11-27T08:00:00Z",
    "max_crp": 0,
    "repeat_strategy": "none",
    "status": "active",
    ...
  }
}
```

**Worker behavior:**
- ⏰ **08:00**: Gửi → `next_recurring = 10:00`
- ⏰ **10:00**: Gửi → `next_recurring = 12:00`
- ⏰ **12:00**: Gửi → `next_recurring = 14:00`
- ♾️ **Lặp lại mãi mãi** (không cần user complete)

---

### Case 5: RECURRING - With CRP - No Until Complete

**Scenario:** Họp team mỗi thứ 2 lúc 9:00, retry mỗi 5 phút tối đa 2 lần.

```http
POST /api/reminders
Authorization: Bearer {your_token}
Content-Type: application/json

{
  "title": "Họp team weekly",
  "description": "Họp review tiến độ tuần",
  "type": "recurring",
  "calendar_type": "solar",
  "recurrence_pattern": {
    "type": "weekly",
    "day_of_week": 1,
    "trigger_time_of_day": "09:00"
  },
  "next_recurring": "2025-12-02T09:00:00Z",
  "max_crp": 2,
  "crp_interval_sec": 300,
  "repeat_strategy": "none",
  "status": "active",
  "tag": "work"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "reminder_meeting01",
    "title": "Họp team weekly",
    "type": "recurring",
    "recurrence_pattern": {
      "type": "weekly",
      "day_of_week": 1,
      "trigger_time_of_day": "09:00"
    },
    "next_recurring": "2025-12-02T09:00:00Z",
    "max_crp": 2,
    "crp_interval_sec": 300,
    "repeat_strategy": "none",
    "status": "active",
    ...
  }
}
```

**Worker behavior (Thứ 2):**
- ⏰ **09:00**: FRP - Gửi → `next_recurring = 09:00 (thứ 2 tuần sau)`, `crp_count = 0`
- ⏰ **09:05**: CRP #1 - Retry
- ⏰ **09:10**: CRP #2 - Retry
- 🛑 **09:15**: Hết quota

**Thứ 2 tuần sau:** Lặp lại với `crp_count = 0` (reset)

---

## RECURRING - Until Complete

### Case 6: RECURRING - No CRP - Until Complete

**Scenario:** Nhắc tập gym mỗi tối 18:00, chỉ tiếp tục khi user đã complete.

```http
POST /api/reminders
Authorization: Bearer {your_token}
Content-Type: application/json

{
  "title": "Tập gym",
  "description": "Tập gym 1 giờ",
  "type": "recurring",
  "calendar_type": "solar",
  "recurrence_pattern": {
    "type": "daily",
    "trigger_time_of_day": "18:00"
  },
  "next_recurring": "2025-11-27T18:00:00Z",
  "max_crp": 0,
  "repeat_strategy": "crp_until_complete",
  "status": "active",
  "tag": "health"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "reminder_gym01",
    "title": "Tập gym",
    "type": "recurring",
    "recurrence_pattern": {
      "type": "daily",
      "trigger_time_of_day": "18:00"
    },
    "next_recurring": "2025-11-27T18:00:00Z",
    "max_crp": 0,
    "repeat_strategy": "crp_until_complete",
    "is_sended_one_time": false,
    "status": "active",
    ...
  }
}
```

**Worker behavior:**

**Ngày 1:**
- ⏰ **18:00**: Gửi (lần đầu) → `is_sended_one_time = true`, `last_sent_at = 18:00`
- 🚫 **18:01 - 23:59**: Không gửi (chờ user complete)

**Ngày 2:**
- 🚫 **Cả ngày**: Không gửi (user chưa complete ngày 1)

**Ngày 2 - 20:00**: User complete ngày 1
```http
POST /api/reminders/reminder_gym01/complete
Authorization: Bearer {your_token}
```
→ `last_completed_at = 20:00 (ngày 2)`, `next_recurring = 18:00 (ngày 3)`

**Ngày 3:**
- ⏰ **18:00**: Gửi (đã complete) → `last_sent_at = 18:00 (ngày 3)`

**Lưu ý:** Reminder sẽ **stuck** nếu user không complete!

---

### Case 7: RECURRING - With CRP - Until Complete

**Scenario:** Nhắc uống thuốc mỗi ngày 8:00, retry mỗi 5 phút tối đa 3 lần, chỉ tiếp tục khi đã complete.

```http
POST /api/reminders
Authorization: Bearer {your_token}
Content-Type: application/json

{
  "title": "Uống thuốc huyết áp",
  "description": "Uống 1 viên thuốc huyết áp",
  "type": "recurring",
  "calendar_type": "solar",
  "recurrence_pattern": {
    "type": "daily",
    "trigger_time_of_day": "08:00"
  },
  "next_recurring": "2025-11-27T08:00:00Z",
  "max_crp": 3,
  "crp_interval_sec": 300,
  "repeat_strategy": "crp_until_complete",
  "status": "active",
  "tag": "health"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "reminder_medicine01",
    "title": "Uống thuốc huyết áp",
    "type": "recurring",
    "recurrence_pattern": {
      "type": "daily",
      "trigger_time_of_day": "08:00"
    },
    "next_recurring": "2025-11-27T08:00:00Z",
    "max_crp": 3,
    "crp_interval_sec": 300,
    "repeat_strategy": "crp_until_complete",
    "is_sended_one_time": false,
    "status": "active",
    ...
  }
}
```

**Worker behavior:**

**Ngày 1:**
- ⏰ **08:00**: FRP (lần đầu) - Gửi → `is_sended_one_time = true`, `crp_count = 0`
- ⏰ **08:05**: CRP #1 - Retry
- ⏰ **08:10**: CRP #2 - Retry

**Ngày 1 - 08:12**: User complete
```http
POST /api/reminders/reminder_medicine01/complete
Authorization: Bearer {your_token}
```
→ `last_completed_at = 08:12`, `next_recurring = 08:00 (ngày 2)`

- 🛑 **08:15**: Không gửi CRP #3 (user đã complete)

**Ngày 2:**
- ⏰ **08:00**: FRP (đã complete) - Gửi → `crp_count = 0` (reset)
- ⏰ **08:05**: CRP #1
- ⏰ **08:10**: CRP #2
- ⏰ **08:15**: CRP #3
- 🛑 **08:20**: Hết quota, chờ user complete

**Ngày 3:**
- 🚫 **Cả ngày**: Không gửi (user chưa complete ngày 2)

**Điểm khác biệt:** CRP dừng ngay khi user complete, nhưng nếu không complete thì stuck như Case 5.

---

## Reminder Management

### Complete Reminder

**Use case:** User đánh dấu đã hoàn thành task (cho recurring reminders với `crp_until_complete`).

```http
POST /api/reminders/{reminder_id}/complete
Authorization: Bearer {your_token}
Content-Type: application/json
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "reminder_abc123",
    "last_completed_at": "2025-11-27T08:12:00Z",
    "next_recurring": "2025-11-28T08:00:00Z",
    ...
  }
}
```

**Effect:**
- Set `last_completed_at = now()`
- Advance `next_recurring` (cho recurring reminders)
- Worker sẽ resume gửi notification cho chu kỳ mới

---

### Snooze Reminder

**Use case:** Hoãn reminder 30 phút.

```http
POST /api/reminders/{reminder_id}/snooze
Authorization: Bearer {your_token}
Content-Type: application/json

{
  "snooze_minutes": 30
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "reminder_abc123",
    "snooze_until": "2025-11-27T08:42:00Z",
    ...
  }
}
```

**Effect:**
- Set `snooze_until = now() + 30 minutes`
- Worker sẽ **không gửi** notification cho đến khi `now() > snooze_until`

---

### Get Reminder

```http
GET /api/reminders/{reminder_id}
Authorization: Bearer {your_token}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "reminder_abc123",
    "title": "Uống thuốc",
    "type": "recurring",
    "next_recurring": "2025-11-27T08:00:00Z",
    "next_crp": "2025-11-27T08:05:00Z",
    "crp_count": 1,
    "max_crp": 3,
    "is_sended_one_time": true,
    "last_sent_at": "2025-11-27T08:00:00Z",
    "last_completed_at": "2025-11-26T08:10:00Z",
    "status": "active",
    ...
  }
}
```

---

### Update Reminder

**Use case:** Thay đổi giờ uống thuốc từ 8:00 → 9:00.

```http
PUT /api/reminders/{reminder_id}
Authorization: Bearer {your_token}
Content-Type: application/json

{
  "recurrence_pattern": {
    "type": "daily",
    "trigger_time_of_day": "09:00"
  },
  "next_recurring": "2025-11-27T09:00:00Z"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "reminder_abc123",
    "recurrence_pattern": {
      "type": "daily",
      "trigger_time_of_day": "09:00"
    },
    "next_recurring": "2025-11-27T09:00:00Z",
    ...
  }
}
```

---

### Delete Reminder

```http
DELETE /api/reminders/{reminder_id}
Authorization: Bearer {your_token}
```

**Response:**
```json
{
  "success": true,
  "message": "Reminder deleted successfully"
}
```

---

## Field Reference

### Required Fields for All Reminders

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Tiêu đề reminder |
| `type` | string | `one_time` hoặc `recurring` |
| `calendar_type` | string | `solar` hoặc `lunar` |
| `status` | string | `active`, `completed`, `paused` |

### ONE-TIME Specific

| Field | Type | Description |
|-------|------|-------------|
| `next_action_at` | datetime | Thời điểm gửi notification (UTC) |

### RECURRING Specific

| Field | Type | Description |
|-------|------|-------------|
| `recurrence_pattern` | object | Pattern lặp lại (xem bên dưới) |
| `next_recurring` | datetime | Thời điểm FRP tiếp theo (UTC) |
| `repeat_strategy` | string | `none` hoặc `crp_until_complete` |

### CRP (Optional)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `max_crp` | int | 0 | Số lần retry tối đa (0 = không retry) |
| `crp_interval_sec` | int | 300 | Khoảng thời gian giữa mỗi retry (giây) |

### Recurrence Pattern Types

**Daily:**
```json
{
  "type": "daily",
  "trigger_time_of_day": "08:00"
}
```

**Weekly:**
```json
{
  "type": "weekly",
  "day_of_week": 1,
  "trigger_time_of_day": "09:00"
}
```
**Note:** `day_of_week`: 0=Sunday, 1=Monday, ..., 6=Saturday

**Monthly:**
```json
{
  "type": "monthly",
  "day_of_month": 15,
  "trigger_time_of_day": "10:00"
}
```

**Interval Seconds:**
```json
{
  "type": "interval_seconds",
  "interval_seconds": 7200
}
```
**Note:** 7200 = 2 hours

**Lunar Last Day of Month:**
```json
{
  "type": "lunar_last_day_of_month",
  "trigger_time_of_day": "18:00"
}
```

---

## Tips & Best Practices

1. **Timezone**: Luôn dùng UTC cho API. Client convert sang local timezone.

2. **Testing**: Dùng `interval_seconds` nhỏ (60s) để test nhanh.

3. **CRP Interval**: Khuyến nghị 300s (5 phút) - 600s (10 phút) để tránh spam.

4. **Max CRP**: Khuyến nghị 2-5 lần retry.

5. **Until Complete Strategy**: Chỉ dùng cho task quan trọng cần confirmation (uống thuốc, tập gym...).

6. **Snooze**: Dùng khi user bận tạm thời, sẽ gửi lại sau.

7. **Paused Status**: Dùng để tạm dừng reminder mà không xóa (ví dụ: đi du lịch).

---

## Error Handling

### Invalid datetime format

```json
{
  "success": false,
  "error": "Invalid datetime format. Use ISO 8601 (UTC): 2025-11-27T08:00:00Z"
}
```

### Missing required fields

```json
{
  "success": false,
  "error": "Missing required field: title"
}
```

### Reminder not found

```json
{
  "success": false,
  "error": "Reminder not found"
}
```

### Unauthorized

```json
{
  "success": false,
  "error": "Unauthorized. Please provide a valid token."
}
```
