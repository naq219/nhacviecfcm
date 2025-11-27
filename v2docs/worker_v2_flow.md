# Worker V2 Flow Documentation

## Giới thiệu

**Worker V2** là hệ thống xử lý reminder notification tự động, được thiết kế để:

- ✅ Gửi notification đúng thời điểm
- ✅ Retry thông minh khi user chưa phản hồi
- ✅ Hỗ trợ chu kỳ lặp phức tạp (daily, weekly, monthly, lunar calendar)
- ✅ Chờ user complete task trước khi tiếp tục (crp_until_complete strategy)

### Kiến trúc

Worker V2 được chia thành **3 modules độc lập**:

1. **`worker_onetime_v2.go`**: Xử lý ONE-TIME reminders (Cases 1-3)
2. **`worker_loop_noUT.go`**: Xử lý RECURRING reminders - Không chờ complete (Cases 4-5)
3. **`worker_loop_UT.go`**: Xử lý RECURRING reminders - Chờ complete (Cases 6-7) *(Sẽ implement)*

Mỗi module có:
- Query riêng biệt từ `reminder_orm_repo_worker.go`
- Logic xử lý độc lập
- Chạy song song trong cùng 1 ticker

---

## Khái niệm cốt lõi

### FRP (Father Recurrence Pattern)
Chu kỳ lặp **chính** của reminder (daily, weekly, monthly...).

**Ví dụ:** 
- Nhắc uống thuốc mỗi ngày 8:00 AM
- Họp team mỗi thứ 2 lúc 9:00 AM

### CRP (Child Repeat Pattern)
Retry/Nhắc lại **trong cùng một chu kỳ FRP** nếu user chưa phản hồi.

**Ví dụ:**
- FRP: 8:00 AM mỗi ngày
- CRP: Retry mỗi 5 phút nếu user chưa complete (8:05, 8:10, 8:15...)

### repeat_strategy

| Strategy | Mô tả |
|----------|-------|
| `none` | Gửi notification theo lịch, không cần chờ user complete |
| `crp_until_complete` | **Chờ user complete** task hiện tại trước khi tiếp tục chu kỳ mới |

### is_sended_one_time

Field đặc biệt để **phân biệt lần đầu tiên** gửi notification:
- `false`: Lần đầu tiên (chưa biết `last_completed_at` có ý nghĩa chưa)
- `true`: Đã gửi ít nhất 1 lần (có thể dùng `last_completed_at > last_sent_at`)

⚠️ **LƯU Ý QUAN TRỌNG**: Sau khi set = `true`, **KHÔNG BAO GIỜ** reset về `false`.

---

## ONE TIME REMINDERS

### 1. Một lần - không CRP

**Điều kiện lấy từ DB:**
```
- type = one_time
- max_crp = 0
- status = active
- next_action_at <= now()
- is_sended_one_time = false
- snooze_until <= now()
```

**Sau khi gửi notification:**
```
- set is_sended_one_time = true
- set status = completed
- set last_completed_at = now()
- set last_sent_at = now()
- set next_action_at = NULL
```

---

### 2. Một lần - có CRP - Lần đầu

**Điều kiện lấy từ DB:**
```
- type = one_time
- max_crp > 0
- status = active
- next_action_at <= now()
- is_sended_one_time = false
- snooze_until <= now()
```

**Sau khi gửi notification:**
```
- set is_sended_one_time = true
- set last_sent_at = now()
- set next_crp = now() + crp_interval_sec
```

---

### 3. Một lần - có CRP - Retry

**Điều kiện lấy từ DB:**
```
- type = one_time
- max_crp > 0
- crp_count < max_crp
- status = active
- next_crp <= now()
- is_sended_one_time = true
- snooze_until <= now()
```

**Sau khi gửi notification:**
```
- set crp_count = crp_count + 1
- set last_sent_at = now()
- set next_crp = now() + crp_interval_sec

- if crp_count >= max_crp:
    - set status = completed
    - set last_completed_at = now()
    - set next_action_at = NULL
```

---

## RECURRING REMINDERS - Không có "Until Complete"

### 4. Lặp lại - không CRP - không crp_until_complete

**Điều kiện lấy từ DB:**
```
- type = recurring
- status = active
- max_crp = 0
- repeat_strategy = none
- next_recurring <= now()
- snooze_until <= now()
```

**Sau khi gửi notification:**
```
- set last_sent_at = now()
- set next_recurring = calculate_next_recurring()
```

---

### 5.1. Lặp lại - có CRP - không crp_until_complete - FRP Trigger

**Điều kiện lấy từ DB:**
```
- type = recurring
- status = active
- max_crp > 0
- repeat_strategy = none
- next_recurring <= now()
- snooze_until <= now()
```

**Sau khi gửi notification:**
```
- set last_sent_at = now()
- set next_recurring = calculate_next_recurring()
- set crp_count = 0
- set next_crp = now() + crp_interval_sec
```

**Giải thích:** Đây là trigger của chu kỳ FRP mới. Reset CRP counter và bắt đầu chu kỳ retry mới.

---

### 5.2. Lặp lại - có CRP - không crp_until_complete - CRP Retry

**Điều kiện lấy từ DB:**
```
- type = recurring
- status = active
- max_crp > 0
- repeat_strategy = none
- next_recurring > now()    ← Chưa đến chu kỳ FRP mới
- crp_count < max_crp
- next_crp <= now()
- snooze_until <= now()
```

**Sau khi gửi notification:**
```
- set crp_count = crp_count + 1
- set last_sent_at = now()
- set next_crp = now() + crp_interval_sec
```

**Giải thích:** Gửi retry trong cùng chu kỳ FRP. Không advance `next_recurring`.

---

## RECURRING REMINDERS - Có "Until Complete"

### 6.1. Lặp lại - không CRP - có crp_until_complete - Lần đầu

**Điều kiện lấy từ DB:**
```
- type = recurring
- status = active
- max_crp = 0
- repeat_strategy = crp_until_complete
- next_recurring <= now()
- is_sended_one_time = false    ← Lần đầu tiên của reminder này
- snooze_until <= now()
```

**Sau khi gửi notification:**
```
- set is_sended_one_time = true
- set last_sent_at = now()
```

**Giải thích:** 
- Lần đầu tiên gửi notification cho reminder này
- Sau đó **chờ user complete** trước khi gửi lại
- `next_recurring` KHÔNG thay đổi

---

### 6.2. Lặp lại - không CRP - có crp_until_complete - Đã complete

**Điều kiện lấy từ DB:**
```
- type = recurring
- status = active
- max_crp = 0
- repeat_strategy = crp_until_complete
- next_recurring <= now()
- is_sended_one_time = true     ← Đã gửi lần đầu rồi
- last_completed_at is valid
- last_completed_at > last_sent_at    ← User đã complete
- snooze_until <= now()
```

**Sau khi gửi notification:**
```
- set last_sent_at = now()
```

**Khi user click Complete (xử lý ở API, không phải worker):**
```
- set last_completed_at = now()
- set next_recurring = calculate_next_recurring()
```

**Giải thích:**
- Worker chỉ gửi notification khi user đã complete task trước đó
- API Complete sẽ advance `next_recurring` để bắt đầu chu kỳ mới
- `is_sended_one_time` **luôn = true** sau lần đầu, không bao giờ reset

---

### Flow Example (Case 6):

**Ngày 1 - 8:00 AM:**
- Worker check: `next_recurring = 8:00 AM`, `is_sended_one_time = false` → MATCH Case 6.1
- Gửi notification
- Set: `is_sended_one_time = true`, `last_sent_at = 8:05 AM`

**Ngày 1 - 9:00 AM:**
- Worker check: `is_sended_one_time = true`, nhưng `last_completed_at` chưa có → KHÔNG MATCH
- **Không gửi** (chờ user complete)

**Ngày 1 - 10:00 AM:**
- **User click Complete**
- API set: `last_completed_at = 10:00 AM`, `next_recurring = 8:00 AM ngày 2`

**Ngày 2 - 8:00 AM:**
- Worker check: `next_recurring = 8:00 AM`, `is_sended_one_time = true`, `last_completed_at (10:00 AM) > last_sent_at (8:05 AM)` → MATCH Case 6.2
- Gửi notification
- Set: `last_sent_at = 8:05 AM ngày 2`

**Ngày 2 - 9:00 AM:**
- **User click Complete**
- API set: `last_completed_at = 9:00 AM ngày 2`, `next_recurring = 8:00 AM ngày 3`

**Chu kỳ lặp lại...**

---


## 7. Lặp lại - có CRP - có crp_until_complete

Case này kết hợp **cả 2 loại trigger**: FRP (chu kỳ lặp) và CRP (retry).

⚠️ **Điểm khác biệt quan trọng**: CRP chỉ gửi khi user **CHƯA complete**. Nếu đã complete → dừng retry, chờ FRP tiếp theo.

---

### 7.1.1. FRP Trigger - Lần đầu tiên

**Điều kiện lấy từ DB:**
```
- type = recurring
- status = active
- max_crp > 0
- repeat_strategy = crp_until_complete
- next_recurring <= now()
- is_sended_one_time = false    ← Lần đầu tiên của reminder này
- snooze_until <= now()
```

**Sau khi gửi notification:**
```
- set is_sended_one_time = true
- set last_sent_at = now()
- set crp_count = 0
- set next_crp = now() + crp_interval_sec
```

**Giải thích:**
- Lần đầu tiên gửi notification cho reminder này
- Bắt đầu chu kỳ CRP (retry) nếu user không complete

---

### 7.1.2. FRP Trigger - User đã complete lần trước

**Điều kiện lấy từ DB:**
```
- type = recurring
- status = active
- max_crp > 0
- repeat_strategy = crp_until_complete
- next_recurring <= now()
- is_sended_one_time = true     ← Đã gửi lần đầu rồi
- last_completed_at is valid
- last_completed_at > last_sent_at    ← User đã complete chu kỳ trước
- snooze_until <= now()
```

**Sau khi gửi notification:**
```
- set last_sent_at = now()
- set crp_count = 0              ← Reset CRP counter
- set next_crp = now() + crp_interval_sec
```

**Khi user click Complete (xử lý ở API):**
```
- set last_completed_at = now()
- set next_recurring = calculate_next_recurring()
```

**Giải thích:**
- Đây là trigger của chu kỳ FRP mới
- User đã complete task trước đó → được phép gửi chu kỳ mới
- Reset CRP counter về 0 để bắt đầu chu kỳ retry mới
- API Complete sẽ advance `next_recurring`

---

### 7.2. CRP Trigger - Retry trong cùng chu kỳ

**Điều kiện lấy từ DB:**
```
- type = recurring
- status = active
- max_crp > 0
- crp_count < max_crp
- repeat_strategy = crp_until_complete
- next_recurring > now()         ← CHƯA đến chu kỳ FRP mới
- next_crp <= now()
- is_sended_one_time = true
- (last_completed_at IS NULL OR last_completed_at <= last_sent_at)  ← User CHƯA complete
- snooze_until <= now()
```

**Sau khi gửi notification:**
```
- set crp_count = crp_count + 1
- set last_sent_at = now()
- set next_crp = now() + crp_interval_sec
```

**Giải thích:**
- Gửi retry khi user **chưa complete** task hiện tại
- Điều kiện `last_completed_at <= last_sent_at` đảm bảo chỉ retry khi chưa complete
- Nếu user complete → CRP ngừng, chờ FRP tiếp theo
- Không advance `next_recurring`

---

### Flow Example - Case 7 (Chi tiết từng bước)

**Setup:**
- Reminder: Uống thuốc mỗi ngày 8:00 AM
- CRP: Retry mỗi 5 phút, tối đa 3 lần
- Strategy: `crp_until_complete`

---

**📅 Ngày 1 - 8:00 AM (FRP Trigger - Lần đầu)**

Worker check:
```
✓ next_recurring = 8:00 AM (<=now)
✓ is_sended_one_time = false
→ MATCH Case 7.1.1
```

Action:
```
→ Gửi notification "Đến giờ uống thuốc!"
→ Set: is_sended_one_time = true
       last_sent_at = 8:00 AM
       crp_count = 0
       next_crp = 8:05 AM
```

---

**📅 Ngày 1 - 8:05 AM (CRP Retry #1)**

Worker check:
```
✓ next_crp = 8:05 AM (<=now)
✓ crp_count (0) < max_crp (3)
✓ last_completed_at = NULL (chưa complete)
→ MATCH Case 7.2
```

Action:
```
→ Gửi retry "Bạn đã uống thuốc chưa?"
→ Set: crp_count = 1
       last_sent_at = 8:05 AM
       next_crp = 8:10 AM
```

---

**📅 Ngày 1 - 8:07 AM (User Complete!)**

API `/complete` được gọi:
```
→ Set: last_completed_at = 8:07 AM
       next_recurring = 8:00 AM (ngày 2)
```

---

**📅 Ngày 1 - 8:10 AM (CRP Check - Không gửi)**

Worker check:
```
✓ next_crp = 8:10 AM (<=now)
✗ last_completed_at (8:07) > last_sent_at (8:05)
→ KHÔNG MATCH Case 7.2
```

**Kết quả:** Không gửi vì user đã complete. Dừng retry, chờ FRP ngày 2.

---

**📅 Ngày 2 - 8:00 AM (FRP Trigger - Đã complete)**

Worker check:
```
✓ next_recurring = 8:00 AM (<=now)
✓ is_sended_one_time = true
✓ last_completed_at (8:07 ngày 1) > last_sent_at (8:05 ngày 1)
→ MATCH Case 7.1.2
```

Action:
```
→ Gửi notification "Đến giờ uống thuốc!" (chu kỳ mới)
→ Set: last_sent_at = 8:00 AM (ngày 2)
       crp_count = 0         ← Reset
       next_crp = 8:05 AM (ngày 2)
```

---

**📅 Ngày 2 - 8:05 AM (CRP Retry #1)**

Worker check:
```
✓ next_crp = 8:05 AM (<=now)
✓ last_completed_at (8:07 ngày 1) <= last_sent_at (8:00 ngày 2)
→ MATCH Case 7.2
```

Action:
```
→ Gửi retry
→ Set: crp_count = 1
       next_crp = 8:10 AM
```

---

**📅 Ngày 2 - 8:10 AM, 8:15 AM (CRP Retry #2, #3)**

Tương tự, gửi tiếp nếu user chưa complete.

---

**📅 Ngày 2 - 8:20 AM (Hết quota CRP)**

Worker check:
```
✓ next_crp = 8:20 AM
✗ crp_count (3) >= max_crp (3)
→ KHÔNG GỬI
```

**Kết quả:** Hết 3 lần retry, dừng. Chờ user complete hoặc đến `next_recurring` ngày 3.

---

### Lưu ý quan trọng - Case 7

1. **CRP chỉ hoạt động khi user CHƯA complete**
   - Điều kiện: `last_completed_at IS NULL OR last_completed_at <= last_sent_at`
   - Nếu user complete → CRP ngừng ngay lập tức

2. **FRP reset CRP counter**
   - Mỗi chu kỳ FRP mới → `crp_count = 0`
   - User có 3 lần retry mới cho mỗi chu kỳ

3. **API Complete advance next_recurring**
   - Worker KHÔNG advance `next_recurring`
   - Chỉ API Complete mới advance để bắt đầu chu kỳ mới

4. **is_sended_one_time luôn = true**
   - Sau lần đầu tiên, field này không bao giờ reset
   - Dùng để phân biệt logic lần đầu vs các lần sau

---

## So sánh Case 6 vs Case 7

| Tính năng | Case 6 (No CRP) | Case 7 (With CRP) |
|-----------|----------------|-------------------|
| Gửi notification | 1 lần/chu kỳ | 1 FRP + tối đa N CRP |
| Retry khi chưa complete | ❌ Không | ✅ Có |
| Advance next_recurring | API Complete | API Complete |
| Reset crp_count | N/A | Mỗi FRP mới |

---

---

## Field Definitions

| Field | Type | Description |
|-------|------|-------------|
| `is_sended_one_time` | bool | Đánh dấu lần đầu tiên gửi notification cho reminder này. Một khi = true thì không bao giờ reset về false. |
| `last_sent_at` | timestamp | Thời điểm gửi notification gần nhất |
| `last_completed_at` | timestamp | Thời điểm user click Complete gần nhất |
| `next_action_at` | timestamp | Thời điểm action tiếp theo (cho one_time) |
| `next_recurring` | timestamp | Thời điểm FRP tiếp theo (cho recurring) |
| `next_crp` | timestamp | Thời điểm CRP retry tiếp theo |
| `crp_count` | int | Số lần đã retry trong chu kỳ FRP hiện tại |
| `max_crp` | int | Số lần retry tối đa cho mỗi chu kỳ FRP |
| `repeat_strategy` | string | `none` hoặc `crp_until_complete` |

---

## Worker Implementation Files

- **`worker_onetime_v2.go`**: Xử lý cases 1-3 (One Time)
- **`worker_loop_noUT.go`**: Xử lý cases 4-5 (Recurring - Không Until Complete)
- **`worker_loop_UT.go`**: Xử lý cases 6-7 (Recurring - Có Until Complete) - *Chưa implement*
- **`reminder_orm_repo_worker.go`**: Database queries cho tất cả cases

---

## Notes

1. **Timezone**: Tất cả thời gian được lưu và xử lý ở UTC
2. **Snooze**: Tất cả cases đều check `snooze_until <= now()` trước khi gửi
3. **Status**: Chỉ xử lý reminders có `status = active`
4. **Calculation Functions**: `calculate_next_recurring()` cần implement dựa trên `RecurrencePattern` (daily, weekly, monthly, lunar...)
