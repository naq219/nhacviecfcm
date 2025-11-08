```markdown
# Worker Processing Logic

## Overview

**Worker** là background process chạy mỗi 60 giây để:
1. Query reminders với `next_action_at <= NOW`
2. Xử lý FRP (Father Recurrence Pattern)
3. Xử lý CRP (Child Repeat Pattern)
4. Gửi FCM notifications
5. Update database

---

## Flow Chart

```
┌─────────────────────────┐
│ Every 60 seconds        │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ Check worker_enabled?   │
└────────┬────────────────┘
         │ YES
         ▼
┌─────────────────────────────────────────┐
│ Query: WHERE                            │
│   next_action_at <= NOW                 │
│   AND status = 'active'                 │
│   AND (snooze_until IS NULL OR          │
│        snooze_until <= NOW)             │
└────────┬────────────────────────────────┘
         │
         ▼
┌─────────────────────────┐
│ For each reminder:      │
└────────┬────────────────┘
         │
         ▼
    ┌─────────────────────────┐
    │ Type = recurring?       │
    └────────┬────────────────┘
             │
        YES  │  NO
         ─────────
         │       │
         ▼       ▼
     ┌─────┐ ┌─────┐
     │FRP? │ │CRP? │
     └──┬──┘ └──┬──┘
        │YES    │YES
        │       │
        ▼       ▼
    ┌──────────────┐
    │SendFCM()     │
    └───────┬──────┘
            │
            ▼
    ┌───────────────────┐
    │Update DB:         │
    │ last_sent_at      │
    │ crp_count         │
    │ next_recurring    │
    │ next_action_at    │
    └───────────────────┘
```

---

## FRP Trigger

**When**: `now >= next_recurring`

**Action**:
1. Send FCM
2. Update last_sent_at = now
3. Reset crp_count = 0
4. Tính next_recurring tiếp theo:
   - `repeat_strategy = "none"`: Auto calc
   - `repeat_strategy = "crp_until_complete"`: Wait for complete
5. Recalc next_action_at

---

## CRP Retry

**When**: `now >= last_sent_at + crp_interval_sec AND crp_count < max_crp`

**Action**:
1. Send FCM
2. Update last_sent_at = now
3. Increment crp_count++
4. If one_time AND crp_count >= max_crp:
   - Mark status = "completed"
5. Recalc next_action_at

---

## Error Handling

| Error Type | Action |
|-----------|--------|
| FCM token invalid (UNREGISTERED) | Disable user FCM |
| FCM system error (401, 403, timeout) | Disable worker, log error |
| User not found | Skip reminder |

```