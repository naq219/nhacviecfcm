 thay đổi db :
 - next_crp: thay cho next_trigger_at (xoá bỏ) 
- max_retries đổi thành max_crp 
- retry_count đổi thành crp_count 
- retry_interval_sec đổi thành crp_interval_sec
- retry_until_complete đổi thành crp_until_complete
 
 QUY TRÌNH WORKER MỚI - PHIÊN BẢN HOÀN CHỈNH

PHẦN 1: KHÁI NIỆM CƠ BẢN
Các loại vòng lặp:
FRP (Father Recurrence Pattern) - Vòng lặp cha

Khi nào có: Chỉ khi type = "recurring"
Ví dụ: Nhắc uống thuốc mỗi ngày lúc 8AM, nhắc đóng học phí mỗi tháng ngày 5
Tính theo: Lịch (ngày/tuần/tháng dương lịch hoặc âm lịch)
Thời điểm kích hoạt: next_recurring

CRP (Child Repeat Pattern) - Vòng lặp con

Khi nào có: Cả one_time và recurring
Ví dụ: Nhắc lại mỗi 15 phút cho đến khi bấm "Hoàn thành"
Tính theo: Giây (có thể rất lớn: 86400 giây = 1 ngày)
Giới hạn: max_crp (0 = chỉ gửi 1 lần, >0 = gửi nhiều lần)
Đếm: crp_count (đã gửi bao nhiêu lần)

Các trường quan trọng:
gotype Reminder struct {
    // Cơ bản
    ID          string
    UserID      string
    Type        string  // "one_time" hoặc "recurring"
    Status      string  // "active", "completed", "paused"
    
    // FRP (chỉ cho recurring)
    NextRecurring      time.Time  // Chu kỳ lặp tiếp theo
    RecurrencePattern  JSON       // Cấu hình lặp: daily/weekly/monthly/...
    RepeatStrategy     string     // "none" hoặc "crp_until_complete"
    
    // CRP (cho cả one_time và recurring)
    NextCRP          time.Time  // Lần CRP tiếp theo
    CRPIntervalSec   int        // Khoảng cách giữa các CRP (giây)
    MaxCRP           int        // Giới hạn số lần CRP (0 = 1 lần)
    CRPCount         int        // Đã gửi bao nhiêu CRP
    
    // Tracking
    LastSentAt           time.Time  // Lần cuối gửi notifi (dùng chung)
    LastCRPCompletedAt   time.Time  // User complete lần CRP hiện tại
    LastCompletedAt      time.Time  // Dùng cho repeat_strategy=crp_until_complete
    
    // Optimization
    NextActionAt    time.Time  // Thời điểm gần nhất cần check reminder này
    
    // Snooze
    SnoozeUntil     time.Time  // Tạm hoãn đến khi nào
}

PHẦN 2: LOGIC WORKER CHI TIẾT
Bước 1: Lấy danh sách reminders cần xử lý
sqlSELECT * FROM reminders 
WHERE status = 'active'
AND next_action_at <= NOW()
ORDER BY next_action_at ASC
Giải thích:

Chỉ lấy reminders đang active
next_action_at đã tính sẵn thời điểm gần nhất cần notifi
Sắp xếp theo thứ tự ưu tiên (gần nhất trước)


Bước 2: Xử lý từng reminder
gofunc ProcessReminder(reminder *Reminder) {
    now := time.Now()
    
    // ========================================
    // BƯỚC 2.1: KIỂM TRA FRP (ưu tiên cao nhất)
    // ========================================
    if reminder.Type == "recurring" && !reminder.NextRecurring.IsZero() {
        if now.After(reminder.NextRecurring) || now.Equal(reminder.NextRecurring) {
            // FRP đến hạn → Gửi ngay
            ProcessFRP(reminder, now)
            return  // Dừng, chờ chu kỳ tiếp theo
        }
    }
    
    // ========================================
    // BƯỚC 2.2: KIỂM TRA CRP
    // ========================================
    if CanSendCRP(reminder, now) {
        ProcessCRP(reminder, now)
        return
    }
    
    // ========================================
    // BƯỚC 2.3: Không cần làm gì
    // ========================================
    // next_action_at đã qua nhưng không thỏa điều kiện
    // → Tính lại next_action_at
    reminder.NextActionAt = CalculateNextActionAt(reminder, now)
    Update(reminder)
}

Bước 2.1: Xử lý FRP (Recurring đến hạn)
gofunc ProcessFRP(reminder *Reminder, now time.Time) {
    fmt.Printf("📅 FRP triggered for reminder %s\n", reminder.ID)
    
    // 1. Gửi notification
    SendNotification(reminder)
    
    // 2. Cập nhật tracking
    reminder.LastSentAt = now
    
    // 3. Reset CRP cho chu kỳ mới
    reminder.CRPCount = 0
    reminder.NextCRP = reminder.NextRecurring
    
    // 4. Tính next_recurring tiếp theo (tùy theo repeat_strategy)
    if reminder.RepeatStrategy == "none" {
        // Tự động tính theo lịch, không phụ thuộc user complete
        reminder.NextRecurring = CalculateNextRecurring(reminder, now)
    } else if reminder.RepeatStrategy == "crp_until_complete" {
        // Chờ user complete mới tính chu kỳ tiếp theo
        // → Giữ nguyên next_recurring
        fmt.Printf("⏸️  Waiting for user to complete before next FRP cycle\n")
    }
    
    // 5. Tính next_action_at
    reminder.NextActionAt = CalculateNextActionAt(reminder, now)
    
    // 6. Lưu database
    Update(reminder)
    
    fmt.Printf("✅ FRP processed. Next FRP: %s\n", reminder.NextRecurring)
}
Giải thích:

FRP đến hạn = Đã đến thời điểm lặp lại trong lịch (ví dụ: mỗi tháng ngày 5)
Luôn gửi notification ngay lập tức khi FRP đến hạn
Reset CRP vì đây là chu kỳ mới
Tính next_recurring:

repeat_strategy = "none": Tự động cộng pattern (ví dụ: +1 tháng)
repeat_strategy = "crp_until_complete": Đợi user complete




Bước 2.2: Kiểm tra điều kiện gửi CRP
gofunc CanSendCRP(reminder *Reminder, now time.Time) bool {
    // Điều kiện 1: Chưa đạt giới hạn
    if reminder.MaxCRP > 0 && reminder.CRPCount >= reminder.MaxCRP {
        fmt.Printf("❌ CRP limit reached (%d/%d)\n", reminder.CRPCount, reminder.MaxCRP)
        return false
    }
    
    // Điều kiện 2: Đã đủ thời gian từ lần gửi trước
    if reminder.LastSentAt.IsZero() {
        // Chưa gửi lần nào → OK
        fmt.Printf("✅ First CRP, can send\n")
        return true
    }
    
    timeSinceLastSent := now.Sub(reminder.LastSentAt)
    requiredInterval := time.Duration(reminder.CRPIntervalSec) * time.Second
    
    if timeSinceLastSent >= requiredInterval {
        fmt.Printf("✅ CRP interval met (%.0fs >= %.0fs)\n", 
            timeSinceLastSent.Seconds(), requiredInterval.Seconds())
        return true
    }
    
    fmt.Printf("⏳ CRP interval not met yet (%.0fs < %.0fs)\n", 
        timeSinceLastSent.Seconds(), requiredInterval.Seconds())
    return false
}
Giải thích:

Điều kiện 1: Kiểm tra quota (nếu max_crp = 0 thì chỉ gửi 1 lần)
Điều kiện 2: Kiểm tra khoảng thời gian từ lần gửi trước
Chỉ gửi khi CẢ 2 điều kiện đều thỏa


Bước 2.3: Xử lý CRP
gofunc ProcessCRP(reminder *Reminder, now time.Time) {
    fmt.Printf("🔔 CRP triggered for reminder %s\n", reminder.ID)
    
    // 1. Gửi notification
    SendNotification(reminder)
    
    // 2. Cập nhật tracking
    reminder.LastSentAt = now
    reminder.CRPCount++
    
    fmt.Printf("📊 CRP count: %d/%d\n", reminder.CRPCount, reminder.MaxCRP)
    
    // 3. Kiểm tra nếu là one_time và đã hết quota
    if reminder.Type == "one_time" {
        if reminder.MaxCRP == 0 || reminder.CRPCount >= reminder.MaxCRP {
            fmt.Printf("🏁 One-time reminder completed\n")
            reminder.Status = "completed"
        }
    }
    
    // 4. Tính next_crp và next_action_at
    if reminder.Status != "completed" {
        reminder.NextCRP = now.Add(time.Duration(reminder.CRPIntervalSec) * time.Second)
        reminder.NextActionAt = CalculateNextActionAt(reminder, now)
    }
    
    // 5. Lưu database
    Update(reminder)
    
    fmt.Printf("✅ CRP processed. Next CRP: %s\n", reminder.NextCRP)
}
Giải thích:

Gửi notification và tăng crp_count
One-time reminder: Nếu đã hết quota → đánh dấu completed
Recurring reminder: Tiếp tục đếm CRP, chờ FRP tiếp theo hoặc user complete


PHẦN 3: TÍNH TOÁN THỜI GIAN
3.1. Tính next_action_at (Thời điểm gần nhất cần check)
gofunc CalculateNextActionAt(reminder *Reminder, now time.Time) time.Time {
    candidates := []time.Time{}
    
    // ========================================
    // Ứng viên 1: Snooze (ưu tiên cao nhất)
    // ========================================
    if !reminder.SnoozeUntil.IsZero() && reminder.SnoozeUntil.After(now) {
        return reminder.SnoozeUntil
    }
    
    // ========================================
    // Ứng viên 2: FRP (nếu là recurring)
    // ========================================
    if reminder.Type == "recurring" && !reminder.NextRecurring.IsZero() {
        candidates = append(candidates, reminder.NextRecurring)
    }
    
    // ========================================
    // Ứng viên 3: CRP tiếp theo (nếu còn quota)
    // ========================================
    if reminder.MaxCRP == 0 || reminder.CRPCount < reminder.MaxCRP {
        if !reminder.LastSentAt.IsZero() {
            nextCRP := reminder.LastSentAt.Add(time.Duration(reminder.CRPIntervalSec) * time.Second)
            candidates = append(candidates, nextCRP)
        } else {
            // Chưa gửi lần nào → gửi ngay
            candidates = append(candidates, now)
        }
    }
    
    // ========================================
    // Lấy thời điểm SỚM NHẤT
    // ========================================
    if len(candidates) == 0 {
        // Không còn action nào (đã completed hoặc hết quota)
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
Giải thích:

Snooze: Nếu đang snooze → trả về snooze_until
FRP: Nếu là recurring → next_recurring là ứng viên
CRP: Nếu còn quota → last_sent_at + interval là ứng viên
Lấy MIN: Thời điểm nào sớm nhất thì đó là next_action_at


3.2. Tính next_recurring (Chu kỳ FRP tiếp theo)
gofunc CalculateNextRecurring(reminder *Reminder, now time.Time) time.Time {
    pattern := reminder.RecurrencePattern
    current := reminder.NextRecurring
    
    switch pattern.Type {
    case "daily":
        // Ví dụ: Mỗi 2 ngày
        interval := time.Duration(pattern.Interval) * 24 * time.Hour
        next := current.Add(interval)
        
        // QUAN TRỌNG: Tìm bội số đầu tiên > now
        for next.Before(now) || next.Equal(now) {
            next = next.Add(interval)
        }
        
        return next
        
    case "weekly":
        // Ví dụ: Mỗi thứ 2 hàng tuần
        // ... logic tương tự daily
        
    case "monthly":
        if pattern.CalendarType == "solar" {
            // Ví dụ: Mỗi tháng ngày 5
            next := current.AddDate(0, pattern.Interval, 0)
            
            // Tìm bội số > now
            for next.Before(now) || next.Equal(now) {
                next = next.AddDate(0, pattern.Interval, 0)
            }
            
            return next
        } else {
            // Âm lịch: cần lunar calendar library
            return CalculateLunarMonthly(current, pattern, now)
        }
        
    case "lunar_last_day_of_month":
        // Ngày cuối tháng âm
        return CalculateLunarLastDay(current, pattern, now)
    }
    
    return time.Time{}
}
```

**Giải thích**:
- **Tính từ `next_recurring` hiện tại**, không phải từ `now`
- **Cộng pattern** (ngày/tuần/tháng)
- **Nếu kết quả <= now**: Tiếp tục cộng cho đến khi tìm được bội số đầu tiên > now
- **Ví dụ**:
```
  Pattern: Mỗi tháng
  next_recurring cũ: 01/01/2025
  now: 15/03/2025
  
  Tính:
  01/01 + 1 tháng = 01/02 (< 15/03) → tiếp
  01/02 + 1 tháng = 01/03 (< 15/03) → tiếp
  01/03 + 1 tháng = 01/04 (> 15/03) → OK!
  
  Kết quả: 01/04/2025

PHẦN 4: XỬ LÝ USER ACTIONS
4.1. User bấm "Hoàn thành"
gofunc OnUserComplete(reminderID string) error {
    reminder := GetReminderByID(reminderID)
    now := time.Now()
    
    fmt.Printf("✅ User completed reminder %s\n", reminderID)
    
    // Cập nhật tracking
    reminder.LastCRPCompletedAt = now
    
    // ========================================
    // Xử lý theo loại reminder
    // ========================================
    if reminder.Type == "one_time" {
        // One-time: Đánh dấu hoàn thành
        reminder.Status = "completed"
        fmt.Printf("🏁 One-time reminder marked as completed\n")
        
    } else if reminder.Type == "recurring" {
        // Recurring: Reset CRP cho chu kỳ hiện tại
        reminder.CRPCount = 0
        fmt.Printf("🔄 CRP reset for current FRP cycle\n")
        
        // Nếu repeat_strategy = crp_until_complete
        if reminder.RepeatStrategy == "crp_until_complete" {
            // Tính chu kỳ FRP tiếp theo từ thời điểm complete
            reminder.LastCompletedAt = now
            reminder.NextRecurring = CalculateNextRecurring(reminder, now)
            reminder.NextCRP = reminder.NextRecurring
            
            fmt.Printf("📅 Next FRP calculated from completion: %s\n", reminder.NextRecurring)
        }
        // else: repeat_strategy = "none" → không làm gì
        // next_recurring vẫn tự động chạy theo lịch
    }
    
    // Tính next_action_at
    reminder.NextActionAt = CalculateNextActionAt(reminder, now)
    
    // Lưu database
    return Update(reminder)
}
Giải thích:

One-time: Complete → kết thúc reminder
Recurring + none: Complete → chỉ reset CRP, FRP vẫn chạy theo lịch
Recurring + crp_until_complete: Complete → tính chu kỳ FRP mới từ thời điểm complete


4.2. User bấm "Snooze"
gofunc OnUserSnooze(reminderID string, durationSec int) error {
    reminder := GetReminderByID(reminderID)
    now := time.Now()
    
    // Tính thời điểm hết snooze
    reminder.SnoozeUntil = now.Add(time.Duration(durationSec) * time.Second)
    
    fmt.Printf("😴 Reminder %s snoozed until %s\n", reminderID, reminder.SnoozeUntil)
    
    // Cập nhật next_action_at
    reminder.NextActionAt = CalculateNextActionAt(reminder, now)
    
    return Update(reminder)
}
```

**Giải thích**:
- Đơn giản: Đặt `snooze_until`
- Worker sẽ bỏ qua reminder này cho đến khi hết snooze
- Khi hết snooze, xử lý bình thường (check FRP/CRP)

---

## PHẦN 5: VÍ DỤ CỤ THỂ

### **Ví dụ 1: One-time reminder với CRP**
```
Reminder: Nhắc họp lúc 14:00
Type: one_time
MaxCRP: 3 (nhắc tối đa 3 lần)
CRPIntervalSec: 300 (5 phút)

Timeline:
14:00 → Gửi notifi lần 1 (CRPCount=1)
14:05 → Gửi notifi lần 2 (CRPCount=2)
14:10 → Gửi notifi lần 3 (CRPCount=3)
14:10 → Status = "completed" (hết quota)
```

---

### **Ví dụ 2: Recurring với repeat_strategy = "none"**
```
Reminder: Uống thuốc mỗi ngày 8AM
Type: recurring
RecurrencePattern: daily (interval=1)
RepeatStrategy: none (không phụ thuộc complete)
MaxCRP: 0 (chỉ notifi 1 lần mỗi ngày)

Timeline:
01/11 8:00 → Gửi notifi
01/11 9:00 → User complete
02/11 8:00 → Gửi notifi (tự động, không phụ thuộc complete)
02/11 → User KHÔNG complete
03/11 8:00 → Vẫn gửi notifi (vì repeat_strategy=none)
```

---

### **Ví dụ 3: Recurring với repeat_strategy = "crp_until_complete"**
```
Reminder: Nộp báo cáo mỗi tuần
Type: recurring
RecurrencePattern: weekly (mỗi thứ 2)
RepeatStrategy: crp_until_complete
MaxCRP: 5 (nhắc tối đa 5 lần)
CRPIntervalSec: 3600 (1 giờ)

Timeline:
Thứ 2 9:00 → Gửi notifi (FRP trigger, CRPCount=1)
Thứ 2 10:00 → Gửi notifi (CRP, CRPCount=2)
Thứ 2 11:00 → Gửi notifi (CRP, CRPCount=3)
Thứ 2 11:30 → User complete
Thứ 2 11:30 → NextRecurring = 11:30 + 7 ngày = Thứ 2 tuần sau 11:30
Thứ 2 tuần sau 11:30 → Gửi notifi (FRP trigger mới)
```

---

### **Ví dụ 4: Worker bị down, restart lại**
```
Reminder: Recurring daily, 8AM
CRPIntervalSec: 86400 (1 ngày)
NextRecurring: 01/11 8:00
LastSentAt: 01/11 8:00

Timeline:
01/11 8:00 → Gửi notifi
02/11 → Worker DOWN cả ngày
03/11 10:00 → Worker RESTART

Xử lý:
- NextRecurring = 01/11 8:00 (vẫn cũ)
- now = 03/11 10:00
- now >= NextRecurring? YES → FRP trigger
- CalculateNextRecurring():
  01/11 + 1 ngày = 02/11 (< 03/11) → tiếp
  02/11 + 1 ngày = 03/11 (< 03/11) → tiếp
  03/11 + 1 ngày = 04/11 (> 03/11) → OK!
- Gửi notifi, NextRecurring = 04/11 8:00
```

---

## PHẦN 6: TÓM TẮT FLOW
```
┌─────────────────────────────────────┐
│  Worker chạy mỗi 60 giây            │
└─────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│  Query: next_action_at <= NOW       │
└─────────────────────────────────────┘
              │
              ▼
    ┌─────────────────────┐
    │  Có reminders?      │
    └─────────────────────┘
         │          │
        YES        NO → Chờ chu kỳ tiếp
         │
         ▼
┌─────────────────────────────────────┐
│  For each reminder:                 │
└─────────────────────────────────────┘
         │
         ▼
    ┌─────────────────────┐
    │  Type = recurring?  │
    └─────────────────────┘
         │          │
        YES        NO
         │          │
         │          └─────────┐
         ▼                    ▼
┌──────────────────┐   ┌──────────────────┐
│ FRP đến hạn?     │   │ CRP OK?          │
└──────────────────┘   └──────────────────┘
    │          │           │          │
   YES        NO          YES        NO
    │          │           │          │
    │          └───────────┘          │
    ▼                      ▼          ▼
┌──────────────┐   ┌──────────────┐  Skip
│ ProcessFRP() │   │ ProcessCRP() │
└──────────────┘   └──────────────┘
    │                      │
    └──────────┬───────────┘
               ▼
    ┌─────────────────────┐
    │ Update database     │
    │ - LastSentAt        │
    │ - CRPCount          │
    │ - NextRecurring     │
    │ - NextActionAt      │
    └─────────────────────┘