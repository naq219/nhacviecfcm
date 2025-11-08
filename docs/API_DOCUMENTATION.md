
```markdown
# RemiAq API Documentation

## Authentication

### Register
```
POST /api/collections/musers/records
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "passwordConfirm": "password123"
}

Response:
{
  "id": "user_id_123",
  "email": "user@example.com",
  ...
}
```

### Login
```
POST /api/collections/musers/auth-with-password
Content-Type: application/json

{
  "identity": "user@example.com",
  "password": "password123"
}

Response:
{
  "token": "JWT_TOKEN",
  "record": {
    "id": "user_id_123",
    "email": "user@example.com"
  }
}
```

## Reminders API

### Create Reminder

```
POST /api/reminders
Authorization: Bearer {token}
Content-Type: application/json

{
  "title": "string",                    ✓ Required
  "description": "string",              ✓ Optional
  "type": "one_time|recurring",         ✓ Required
  "calendar_type": "solar|lunar",       ✓ Required (default: solar)
  "status": "active|completed|paused",  ✓ Required (default: active)
  
  "next_recurring": "2025-11-08T16:00:00Z",  ✓ Optional (set when to start)
  
  "recurrence_pattern": {               ✓ For recurring only
    "type": "daily|weekly|monthly|lunar_last_day_of_month|interval_seconds",
    "interval": 1,
    "day_of_month": 5,
    "day_of_week": 1,
    "trigger_time_of_day": "08:00",
    "interval_seconds": 3600
  },
  
  "repeat_strategy": "none|crp_until_complete",  ✓ (default: none)
  
  "max_crp": 3,                         ✓ Optional (0=1 time only)
  "crp_interval_sec": 300               ✓ Optional (retry interval)
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
  "message": "",
  "data": [
    {
      "id": "reminder_1",
      "title": "...",
      ...
    }
  ]
}
```

### Update Reminder

```
PUT /api/reminders/{id}
Authorization: Bearer {token}
Content-Type: application/json

{
  "title": "New Title",
  "max_crp": 5
}
```

### Complete Reminder

```
POST /api/reminders/{id}/complete
Authorization: Bearer {token}

Effect:
- one_time: Mark as completed
- recurring + none: Reset CRP, FRP continues
- recurring + crp_until_complete: Reset CRP + recalc next_recurring
```

### Snooze Reminder

```
POST /api/reminders/{id}/snooze
Authorization: Bearer {token}
Content-Type: application/json

{
  "duration": 300  // seconds (5 minutes)
}
```

### Delete Reminder

```
DELETE /api/reminders/{id}
Authorization: Bearer {token}
```

---

## Field Definitions

### next_recurring
**Thời điểm FRP tiếp theo sẽ trigger**
- Set bởi user khi create
- Auto update khi FRP trigger
- Mỗi ngày 08:00 → next_recurring = 08:00 ngày hôm sau

### next_crp
**Thời điểm CRP tiếp theo sẽ retry**
- = next_recurring khi FRP trigger
- Auto update sau mỗi CRP: next_crp = last_sent_at + crp_interval_sec

### next_action_at
**Thời điểm gần nhất cần check reminder này**
- = MIN(snooze_until, next_recurring, next_crp)
- Worker query: WHERE next_action_at <= NOW

### Recurrence Patterns

#### Daily
```json
{
  "type": "daily",
  "interval": 1,              // Every 1, 2, 3... days
  "trigger_time_of_day": "08:00"  // UTC time HH:MM
}
```

#### Weekly
```json
{
  "type": "weekly",
  "interval": 1,
  "day_of_week": 1,           // 0=Sun, 1=Mon, ... 6=Sat
  "trigger_time_of_day": "09:00"
}
```

#### Monthly (Solar)
```json
{
  "type": "monthly",
  "interval": 1,
  "day_of_month": 5,
  "trigger_time_of_day": "10:00"
}
```

#### Interval Seconds (NEW!)
```json
{
  "type": "interval_seconds",
  "interval_seconds": 180     // 3 minutes = 180 seconds
}
```

Chuyển đổi:
- 3 phút = 180
- 1 giờ = 3600
- 1 ngày = 86400
- 20 ngày = 1728000

#### Lunar Last Day
```json
{
  "type": "lunar_last_day_of_month",
  "trigger_time_of_day": "18:00"
}
```

### Repeat Strategies

| Strategy | Meaning |
|----------|---------|
| `none` | Auto-repeat theo lịch, không cần user complete |
| `crp_until_complete` | Chờ user complete → recalc next_recurring |

---