> # ⛔ DEPRECATED — TÀI LIỆU LỖI THỜI, ĐỪNG DỰA VÀO
>
> File này mô tả phiên bản worker cũ và **không khớp với code hiện tại**. Các sai lệch đã xác minh:
>
> | Mục trong file này | Thực tế trong code |
> |---|---|
> | "Chạy mỗi 60 giây" | `WORKER_INTERVAL` giây — mặc định 10, `.env` đang = 5 (`config/config.go:14`, `cmd/server/main.go:104-141`) |
> | 1 query duy nhất trên `next_action_at` | **3 worker × 10+ query**, partition rời nhau (`internal/worker/reminder_orm_repo_worker.go:127-330`) |
> | "FCM token invalid → Disable user FCM" | **Không có.** `DisableFCM`/`SetFCMError` tồn tại nhưng không bao giờ được gọi; cột `fcm_error` không bao giờ được ghi (`internal/worker/common.go`) |
> | Không nhắc `is_sended_one_time`, `repeat_strategy` | Đây là 2 điều kiện quan trọng nhất để phân nhánh |
>
> **Nguồn sự thật:** `internal/worker/worker_loop_noUT.go`, `worker_loop_UT.go`, `worker_onetime_v2.go`,
> `internal/services/reminder_service.go`, `internal/services/calNextTime.go`.
>
> **Đặc tả để port sang Nuxt:** `E:\PROJECT\nhacviecfcm\nhacviecnuxt\docs\04b-tinh-lich-frp-crp.md`
>
> ---
>
> *Nội dung gốc giữ lại bên dưới chỉ để tham khảo lịch sử.*

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