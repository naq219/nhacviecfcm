# Worker V2 - Flow Examples

Tài liệu này chứa các **ví dụ flow chi tiết** cho từng loại reminder trong Worker V2.

---

## Table of Contents

1. [ONE-TIME Reminders](#one-time-reminders)
   - [Example 1: One-time no CRP](#example-1-one-time-no-crp)
   - [Example 2: One-time with CRP](#example-2-one-time-with-crp)
2. [RECURRING - No Until Complete](#recurring---no-until-complete)
   - [Example 3: Recurring no CRP](#example-3-recurring-no-crp)
   - [Example 4: Recurring with CRP](#example-4-recurring-with-crp)
3. [RECURRING - Until Complete](#recurring---until-complete)
   - [Example 5: Case 6 - No CRP](#example-5-case-6---no-crp-until-complete)
   - [Example 6: Case 7 - With CRP](#example-6-case-7---with-crp-until-complete)

---

## ONE-TIME Reminders

### Example 1: One-time no CRP

**Scenario:** Nhắc họp khách hàng vào 14:00 ngày hôm nay, không retry.

**Setup:**
```
- type: one_time
- next_action_at: 14:00
- max_crp: 0
- status: active
```

**Timeline:**

**📅 13:59 - Worker check**
```
✗ next_action_at (14:00) > now (13:59)
→ Chưa đến giờ
```

**📅 14:00 - Worker check**
```
✓ next_action_at = 14:00 (<=now)
✓ is_sended_one_time = false
✓ max_crp = 0
→ MATCH Case 1
```

**Action:**
```
→ Gửi "Họp khách hàng lúc 14:00"
→ Set: is_sended_one_time = true
       status = completed
       last_completed_at = 14:00
```

**📅 14:05 - Worker check**
```
✗ status = completed
→ Không query được (chỉ lấy status=active)
```

**Kết thúc**: Reminder completed, không bao giờ gửi lại.

---

### Example 2: One-time with CRP

**Scenario:** Nhắc nộp báo cáo 17:00, retry mỗi 10 phút tối đa 3 lần nếu chưa xong.

**Setup:**
```
- type: one_time
- next_action_at: 17:00
- max_crp: 3
- crp_interval_sec: 600 (10 phút)
- status: active
```

**Timeline:**

**📅 17:00 - FRP (Lần đầu)**
```
✓ is_sended_one_time = false
→ MATCH Case 2
```
```
→ Gửi "Đến hạn nộp báo cáo!"
→ Set: is_sended_one_time = true
       next_crp = 17:10
       crp_count = 0
```

**📅 17:10 - CRP Retry #1**
```
✓ next_crp = 17:10
✓ crp_count (0) < max_crp (3)
→ MATCH Case 3
```
```
→ Gửi "Nhắc lại: Nộp báo cáo!"
→ Set: crp_count = 1
       next_crp = 17:20
```

**📅 17:20 - CRP Retry #2**
```
→ Gửi retry lần 2
→ Set: crp_count = 2, next_crp = 17:30
```

**📅 17:30 - CRP Retry #3 (Lần cuối)**
```
→ Gửi retry lần 3
→ Set: crp_count = 3
       status = completed
```

**📅 17:40 - Worker check**
```
✗ status = completed
→ Không gửi nữa
```

**Kết thúc**: Đã gửi tổng cộng 4 lần (1 FRP + 3 CRP).

---

## RECURRING - No Until Complete

### Example 3: Recurring no CRP

**Scenario:** Nhắc uống nước mỗi 2 giờ, không retry.

**Setup:**
```
- type: recurring
- recurrence_pattern: daily, interval=2h
- next_recurring: 8:00
- max_crp: 0
- repeat_strategy: none
```

**Timeline:**

**📅 8:00 - FRP #1**
```
✓ next_recurring = 8:00
→ MATCH Case 4
```
```
→ Gửi "Đến giờ uống nước!"
→ Set: next_recurring = 10:00
```

**📅 10:00 - FRP #2**
```
→ Gửi "Đến giờ uống nước!"
→ Set: next_recurring = 12:00
```

**📅 12:00 - FRP #3**
```
→ Gửi "Đến giờ uống nước!"
→ Set: next_recurring = 14:00
```

**Kết luận:** Cứ mỗi 2 giờ gửi 1 lần, không cần user complete.

---

### Example 4: Recurring with CRP

**Scenario:** Họp team mỗi thứ 2 lúc 9:00, retry mỗi 5 phút tối đa 2 lần.

**Setup:**
```
- type: recurring
- recurrence_pattern: weekly, day_of_week=1 (Monday), time=9:00
- next_recurring: 9:00 (thứ 2)
- max_crp: 2
- crp_interval_sec: 300 (5 phút)
- repeat_strategy: none
```

**Timeline:**

**📅 Thứ 2 - 9:00 - FRP**
```
✓ next_recurring = 9:00
→ MATCH Case 5.1 (FRP Trigger)
```
```
→ Gửi "Họp team bắt đầu!"
→ Set: next_recurring = 9:00 (thứ 2 tuần sau)
       crp_count = 0
       next_crp = 9:05
```

**📅 Thứ 2 - 9:05 - CRP #1**
```
✓ next_crp = 9:05
✓ next_recurring (tuần sau) > now
→ MATCH Case 5.2 (CRP Retry)
```
```
→ Gửi "Nhắc họp team!"
→ Set: crp_count = 1
       next_crp = 9:10
```

**📅 Thứ 2 - 9:10 - CRP #2**
```
→ Gửi retry lần 2
→ Set: crp_count = 2
```

**📅 Thứ 2 - 9:15 - Worker check**
```
✗ crp_count (2) >= max_crp (2)
→ Đã hết quota, không gửi
```

**📅 Thứ 3 -> Chủ nhật**
```
→ Không gửi (chưa đến next_recurring)
```

**📅 Thứ 2 tuần sau - 9:00 - FRP mới**
```
→ Chu kỳ lặp lại, reset crp_count = 0
```

**Kết luận:** Mỗi tuần gửi 1 FRP + tối đa 2 CRP, không cần user complete.

---

## RECURRING - Until Complete

### Example 5: Case 6 - No CRP Until Complete

**Scenario:** Nhắc tập gym mỗi tối 18:00, chỉ gửi chu kỳ mới khi user đã complete.

**Setup:**
```
- type: recurring
- recurrence_pattern: daily, time=18:00
- next_recurring: 18:00 (ngày 1)
- max_crp: 0
- repeat_strategy: crp_until_complete
```

**Timeline:**

**📅 Ngày 1 - 18:00 - FRP lần đầu**
```
✓ next_recurring = 18:00
✓ is_sended_one_time = false
→ MATCH Case 6.1
```
```
→ Gửi "Đến giờ tập gym!"
→ Set: is_sended_one_time = true
       last_sent_at = 18:00
```

**📅 Ngày 1 - 19:00 - Worker check**
```
✓ next_recurring = 18:00 (vẫn là ngày 1)
✓ is_sended_one_time = true
✗ last_completed_at chưa có
→ KHÔNG MATCH (chờ user complete)
```

**📅 Ngày 1 - 20:00 - User complete!**
```
API /complete:
→ Set: last_completed_at = 20:00
       next_recurring = 18:00 (ngày 2)
```

**📅 Ngày 2 - 18:00 - FRP đã complete**
```
✓ next_recurring = 18:00
✓ is_sended_one_time = true
✓ last_completed_at (20:00 ngày 1) > last_sent_at (18:00 ngày 1)
→ MATCH Case 6.2
```
```
→ Gửi "Đến giờ tập gym!"
→ Set: last_sent_at = 18:00 (ngày 2)
```

**📅 Ngày 2 - 19:00 - Worker check**
```
✓ last_completed_at (20:00 ngày 1) <= last_sent_at (18:00 ngày 2)
→ KHÔNG GỬI (chờ user complete ngày 2)
```

**📅 Ngày 3 - 18:00 - Worker check**
```
✗ User chưa complete ngày 2
✗ last_completed_at (20:00 ngày 1) <= last_sent_at (18:00 ngày 2)
→ KHÔNG GỬI
```

**Kết quả:** Reminder "bị stuck" ở ngày 2, không tiếp tục chu kỳ mới cho đến khi user complete.

**📅 Ngày 3 - 19:00 - User complete ngày 2 (muộn)**
```
API /complete:
→ Set: last_completed_at = 19:00 (ngày 3)
       next_recurring = 18:00 (ngày 4)
```

**📅 Ngày 4 - 18:00 - FRP tiếp tục**
```
→ Gửi chu kỳ mới
```

**Kết luận:** 
- Chỉ advance khi user complete
- Nếu skip 1 ngày → không gửi cho đến khi complete
- **Chiến lược này phù hợp với task yêu cầu confirmation** (tập gym, uống thuốc...)

---

### Example 6: Case 7 - With CRP Until Complete

**Scenario:** Nhắc uống thuốc mỗi ngày 8:00, retry mỗi 5 phút tối đa 3 lần, chỉ tiếp tục khi đã complete.

**Setup:**
```
- type: recurring
- recurrence_pattern: daily, time=8:00
- next_recurring: 8:00 (ngày 1)
- max_crp: 3
- crp_interval_sec: 300 (5 phút)
- repeat_strategy: crp_until_complete
```

**Timeline:**

**📅 Ngày 1 - 8:00 - FRP lần đầu**
```
✓ is_sended_one_time = false
→ MATCH Case 7.1.1
```
```
→ Gửi "Đến giờ uống thuốc!"
→ Set: is_sended_one_time = true
       last_sent_at = 8:00
       crp_count = 0
       next_crp = 8:05
```

**📅 Ngày 1 - 8:05 - CRP #1**
```
✓ next_crp = 8:05
✓ last_completed_at chưa có
→ MATCH Case 7.2
```
```
→ Gửi "Bạn đã uống thuốc chưa?"
→ Set: crp_count = 1
       next_crp = 8:10
```

**📅 Ngày 1 - 8:10 - CRP #2**
```
→ Gửi retry
→ Set: crp_count = 2, next_crp = 8:15
```

**📅 Ngày 1 - 8:12 - User complete!**
```
API /complete:
→ Set: last_completed_at = 8:12
       next_recurring = 8:00 (ngày 2)
```

**📅 Ngày 1 - 8:15 - CRP check**
```
✓ next_crp = 8:15
✗ last_completed_at (8:12) > last_sent_at (8:10)
→ KHÔNG GỬI (user đã complete, dừng CRP)
```

**📅 Ngày 2 - 8:00 - FRP đã complete**
```
✓ is_sended_one_time = true
✓ last_completed_at (8:12 ngày 1) > last_sent_at (8:10 ngày 1)
→ MATCH Case 7.1.2
```
```
→ Gửi "Đến giờ uống thuốc!"
→ Set: last_sent_at = 8:00
       crp_count = 0       ← Reset
       next_crp = 8:05
```

**📅 Ngày 2 - 8:05 - CRP #1**
```
✓ last_completed_at (8:12 ngày 1) <= last_sent_at (8:00 ngày 2)
→ MATCH Case 7.2
```
```
→ Gửi retry
→ Set: crp_count = 1
```

**📅 Ngày 2 - 8:10, 8:15 - CRP #2, #3**
```
→ Tiếp tục gửi retry nếu chưa complete
```

**📅 Ngày 2 - 8:20 - Hết quota**
```
✗ crp_count (3) >= max_crp (3)
→ KHÔNG GỬI
```

**📅 Ngày 2 - Cả ngày**
```
→ User không complete
→ Không gửi thêm (chờ user complete)
```

**📅 Ngày 3 - 8:00 - Worker check**
```
✗ last_completed_at (8:12 ngày 1) <= last_sent_at (8:15 ngày 2)
→ KHÔNG GỬI FRP (user chưa complete ngày 2)
```

**Kết quả:** Reminder stuck ở ngày 2 cho đến khi user complete.

**📅 Ngày 3 - 10:00 - User complete ngày 2 (muộn)**
```
API /complete:
→ Set: last_completed_at = 10:00 (ngày 3)
       next_recurring = 8:00 (ngày 4)
```

**📅 Ngày 4 - 8:00 - FRP tiếp tục**
```
→ Chu kỳ mới bắt đầu
```

**Kết luận:**
- Gửi 1 FRP + tối đa 3 CRP mỗi ngày
- Nếu user complete sớm → CRP dừng ngay
- Nếu user không complete → stuck, chờ complete mới tiếp tục
- **Chiến lược này cân bằng giữa retry và chờ confirmation**

---

## Tổng kết Scenarios

| Scenario | Case | Gửi/ngày | Cần complete? | Use case |
|----------|------|----------|---------------|----------|
| Họp 1 lần | 1 | 1 | ❌ | Event, deadline |
| Nộp báo cáo w/ retry | 2 | 1 + N CRP | ❌ | Deadline quan trọng |
| Uống nước định kỳ | 4 | 1 FRP | ❌ | Habit tracking đơn giản |
| Họp team w/ retry | 5 | 1 FRP + N CRP | ❌ | Meeting reminder |
| Tập gym chờ confirm | 6 | 1 | ✅ | Task cần confirmation |
| Uống thuốc w/ retry + confirm | 7 | 1 FRP + N CRP | ✅ | Health reminder |

---

## Notes

- **UTC Timezone**: Tất cả thời gian trong examples là UTC
- **Worker interval**: Giả sử worker chạy mỗi phút (thực tế có thể khác)
- **Database**: Mỗi action đều update database ngay lập tức
- **API Complete**: Không do worker xử lý, là endpoint riêng
