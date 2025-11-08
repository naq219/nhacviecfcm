# 📝 CODE CHANGES - REMIAQ WORKER V2

## PHASE 1: MIGRATION & MODELS

### 1️⃣ **migrations/1671631110_init_schema.go** (TẠO LẠI)

**Thay đổi chính:**
- Xóa: `next_trigger_at`, `trigger_time_of_day`
- Thêm: `next_recurring`, `next_crp`, `next_action_at`
- Rename: `retry_*` → `crp_*`, `max_retries` → `max_crp`
- Xóa field không dùng

---

### 2️⃣ **internal/models/reminder.go** (SỬA)

**Fields thay đổi:**
```go
// Xóa:
- NextTriggerAt     string
- TriggerTimeOfDay  string

// Thêm/Rename:
+ NextRecurring       time.Time   // FRP: Chu kỳ lặp tiếp theo
+ NextCRP            time.Time   // CRP: Lần nhắc lại tiếp theo
+ NextActionAt       time.Time   // Thời điểm gần nhất cần check (optimization)
+ CRPCount           int         // Số lần CRP đã gửi
+ MaxCRP             int         // Giới hạn CRP (0 = 1 lần, >0 = multiple)
+ CRPIntervalSec     int         // Khoảng cách CRP (giây)
+ LastCRPCompletedAt time.Time   // Thời điểm user complete lần CRP hiện tại

// Rename:
- RetryIntervalSec → CRPIntervalSec
- MaxRetries → MaxCRP
- RetryCount → CRPCount
- RepeatStrategy: "retry_until_complete" → "crp_until_complete"
```

**Hàm xóa:**
- `IsRetryable()` (logic cũ không dùng)
- `ShouldSend()` (thay bằng worker logic mới)

---

### 3️⃣ **internal/db/mapper.go** (GIỮ NGUYÊN)
Không cần sửa, mapper generic vẫn hoạt động với field mới.

---

## PHASE 2: REPOSITORY & ORM

### 4️⃣ **internal/repository/pocketbase/reminder_orm_repo.go** (SỬA)

**Hàm cần sửa:**

1. **`recordToReminder()`** & **`reminderToRecord()`**
   - Map field cũ → mới
   - Handle `NextRecurring`, `NextCRP`, `NextActionAt`, `CRPCount`, `MaxCRP`, `CRPIntervalSec`

2. **`GetDueReminders()`** ❌ **XÓA & TẠO MỚI**
   ```
   Query cũ: next_trigger_at <= now
   Query mới: next_action_at <= now AND snooze_until IS NULL OR snooze_until <= now
   ```

3. **Hàm utility cũ cần sửa:**
   - `UpdateNextTrigger()` → ❌ Xóa (logic cũ)
   - `IncrementRetryCount()` → ❌ Xóa (thay bằng mới trong worker)
   - Giữ: `UpdateSnooze()`, `UpdateStatus()`, `MarkCompleted()`, `UpdateLastSent()`

4. **Hàm mới cần thêm:**
   ```go
   // Update CRP tracking
   UpdateCRPCount(ctx, id, crpCount int)
   
   // Update FRP tracking
   UpdateNextRecurring(ctx, id, nextRecurring time.Time)
   UpdateNextCRP(ctx, id, nextCRP time.Time)
   
   // Update next_action_at (critical)
   UpdateNextActionAt(ctx, id, nextActionAt time.Time)
   ```

---

## PHASE 3: SERVICES & CALCULATOR ⭐ **CHÍNH**

### 5️⃣ **internal/services/schedule_calculator.go** (SỬA LỚN)

**Xóa:**
- `calculateOneTime()` (logic cũ)
- `calculateRecurring()` (logic cũ)
- `calculateIntervalBased()` (logic cũ)
- Tất cả logic interval_seconds cũ

**Thêm hàm mới (theo WORKER VER 2):**

1. **`CalculateNextActionAt(reminder, now)`**
   - Return thời điểm gần nhất trong {snooze_until, next_recurring, next_crp}
   - Logic: lấy MIN time nếu có

2. **`CalculateNextRecurring(reminder, now)`**
   - Tính chu kỳ FRP tiếp theo từ `NextRecurring` hiện tại
   - Cộng pattern (daily/weekly/monthly/lunar)
   - Tìm bội số đầu tiên > now

3. **`CanSendCRP(reminder, now)`** → bool
   - Check: MaxCRP > 0 && CRPCount >= MaxCRP? → false
   - Check: now >= (LastSentAt + CRPIntervalSec)? → true

4. **`CalculateNextCRP(lastSentAt, crpIntervalSec)`**
   - Return: lastSentAt + duration(crpIntervalSec)

**Giữ lại:**
- Lunar calendar logic (không thay đổi)
- `parseTimeOfDay()`

---

### 6️⃣ **internal/services/reminder_service.go** (SỬA LỚNLỚN)

**Xóa hàm cũ:**
- `handleOneTimeReminder()` (logic cũ)
- `handleRecurringReminder()` (logic cũ)
- `ProcessDueReminders()` (chuyển sang worker)
- `processSingleDueReminder()` (chuyển sang worker)

**Thêm/Sửa hàm:**

1. **`CreateReminder()`** - Sửa
   - Tính `next_crp` = now (lần gửi đầu tiên)
   - Nếu `type = recurring`: tính `next_recurring` từ recurrence_pattern
   - Tính `next_action_at` từ cả hai
   - `crp_count = 0`

2. **`OnUserSnooze(ctx, id, duration)`** - Sửa/Thêm
   - Set `snooze_until = now + duration`
   - Recalc `next_action_at`
   - **NOTE**: `CalculateNextActionAt()` không cần viết, chỉ gọi từ calculator

3. **`OnUserComplete(ctx, id)`** - Thêm hàm mới (QUAN TRỌNG!)
   ```
   Theo WORKER VER 2 - OnUserComplete:
   - Đặt: last_crp_completed_at = now
   
   One-time:
     - status = "completed"
   
   Recurring + repeat_strategy = "none":
     - Reset: crp_count = 0
   
   Recurring + repeat_strategy = "crp_until_complete":
     - Reset: crp_count = 0
     - last_completed_at = now
     - Tính: next_recurring mới từ now
     - next_crp = next_recurring (restart)
   
   - Recalc: next_action_at
   ```

4. **`UpdateReminder()`** - Sửa
   - Khi update recurrence_pattern: recalc next_recurring
   - Khi update status: recalc next_action_at

5. **`GetReminder()`, `DeleteReminder()`, `GetUserReminders()`** - Giữ nguyên

---

## PHASE 4: WORKER ⭐⭐⭐ **VIẾT LẠI HOÀN TOÀN**

### 7️⃣ **internal/worker/worker.go** (XÓA & TẠO MỚI)

**Logic mới (theo WORKER VER 2):**

```
Main loop (mỗi 60s):
1. Check system_status.worker_enabled? → Nếu NO → dừng
2. Query: reminders WHERE next_action_at <= NOW() ORDER BY next_action_at ASC
3. For each reminder:
   a. Kiểm tra FRP (type="recurring" && next_recurring <= now)
      - YES → ProcessFRP():
        * SendNotification()
        * last_sent_at = now
        * crp_count = 0
        * next_crp = next_recurring
        * Tính next_recurring mới (hoặc giữ nếu crp_until_complete)
        * Recalc next_action_at
        * Update DB
      
   b. Nếu FRP NO → Kiểm tra CRP (CanSendCRP())
      - YES → ProcessCRP():
        * SendNotification()
        * last_sent_at = now
        * crp_count++
        * Nếu type="one_time" && crp_count >= max_crp → status="completed"
        * Recalc next_crp & next_action_at
        * Update DB
      
   c. Nếu cả 2 NO → Recalc next_action_at
```

**Error handling:**
- System error (401, 403, timeout) → Disable worker
- Token error (UNREGISTERED) → Disable user FCM
- Record errors để monitoring

---

### 8️⃣ **internal/handlers/reminder_handler.go** (SỬA NHỎ)

**Thay đổi:**
1. Handler `SnoozeReminder()` giữ nguyên (gọi `service.OnUserSnooze()`)
2. Handler `CompleteReminder()` sửa
   - Gọi `service.OnUserComplete()` (hàm mới)
3. CreateReminder/UpdateReminder giữ nguyên logic nhưng cập nhật field

---

## PHASE 5: SWAGGER

### 9️⃣ **docs/swagger.json** (SỬA)

**Model Reminder:**
- Xóa: `next_trigger_at`, `trigger_time_of_day`
- Thêm: `next_recurring`, `next_crp`, `next_action_at`, `crp_count`, `max_crp`, `crp_interval_sec`, `last_crp_completed_at`
- Rename: `retry_*` → `crp_*`

---

## ✅ TÓMMÔ TẮT

| Phase | File | Loại | Độ phức tạp |
|-------|------|------|-----------|
| 1 | migration | Tạo mới | ⭐⭐ |
| 1 | reminder.go | Sửa struct | ⭐ |
| 2 | reminder_orm_repo.go | Sửa lớn | ⭐⭐ |
| 3 | schedule_calculator.go | Sửa lớn | ⭐⭐⭐ |
| 3 | reminder_service.go | Sửa lớn | ⭐⭐⭐ |
| 4 | worker.go | Viết lại | ⭐⭐⭐⭐ |
| 4 | reminder_handler.go | Sửa nhỏ | ⭐ |
| 5 | swagger.json | Sửa nhỏ | ⭐ |

**Bắt đầu từ Phase 1 → 5 để tránh dependency issues.**