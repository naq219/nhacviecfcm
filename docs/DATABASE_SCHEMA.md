
```markdown
# Database Schema

## Collections

### musers (Auth Collection)

| Field | Type | Description |
|-------|------|-------------|
| id | text | User ID |
| email | email | Login email |
| password | password | Hashed password |
| fcm_token | text | Firebase token (1 per user, overwrites) |
| is_fcm_active | bool | Can receive FCM? |
| fcm_error | text | Last FCM error |
| created | datetime | |
| updated | datetime | |

---

### reminders (Main Collection)

#### Basic Info
| Field | Type | Description |
|-------|------|-------------|
| id | text | Reminder ID |
| user_id | relation | Owner |
| title | text | Reminder title |
| description | text | Details |
| type | select | `one_time` or `recurring` |
| status | select | `active`, `completed`, `paused` |

#### FRP (Father Recurrence Pattern)
| Field | Type | Description |
|-------|------|-------------|
| next_recurring | datetime | Next FRP trigger time |
| recurrence_pattern | json | Pattern config (type, interval, ...) |
| repeat_strategy | select | `none` or `crp_until_complete` |
| calendar_type | select | `solar` or `lunar` |

#### CRP (Child Repeat Pattern)
| Field | Type | Description |
|-------|------|-------------|
| next_crp | datetime | Next CRP retry time |
| max_crp | number | Max retries (0 = no retry) |
| crp_count | number | Current retry count |
| crp_interval_sec | number | Retry interval (seconds) |

#### Tracking
| Field | Type | Description |
|-------|------|-------------|
| next_action_at | datetime | Nearest check time = MIN(snooze, frp, crp) |
| last_sent_at | datetime | Last notification sent |
| last_completed_at | datetime | User completed |
| last_crp_completed_at | datetime | User completed current CRP cycle |
| snooze_until | datetime | Snoozed until when |

#### System
| Field | Type | Description |
|-------|------|-------------|
| created | datetime | Created at |
| updated | datetime | Updated at |

---

### system_status (Singleton: mid=1)

| Field | Type | Description |
|-------|------|-------------|
| mid | number | Always 1 (singleton) |
| worker_enabled | bool | Worker running? |
| last_error | text | Last error message |
| updated | datetime | |

---

## Indexes (Recommended)

```sql
-- For worker query
CREATE INDEX idx_reminders_next_action ON reminders(next_action_at, status);

-- For user query
CREATE INDEX idx_reminders_user ON reminders(user_id, status);
```

---

## RecurrencePattern JSON Examples

```json
// Daily at 8 AM
{"type": "daily", "interval": 1, "trigger_time_of_day": "08:00"}

// Every 3 minutes
{"type": "interval_seconds", "interval_seconds": 180}

// Every 20 days
{"type": "interval_seconds", "interval_seconds": 1728000}

// Every Monday at 9 AM
{"type": "weekly", "interval": 1, "day_of_week": 1, "trigger_time_of_day": "09:00"}

// 5th of each month at 10 AM
{"type": "monthly", "interval": 1, "day_of_month": 5, "trigger_time_of_day": "10:00"}

// Last day of lunar month at 6 PM
{"type": "lunar_last_day_of_month", "trigger_time_of_day": "18:00"}
```

---