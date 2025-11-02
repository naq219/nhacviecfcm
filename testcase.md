Dưới đây là **toàn bộ bộ test case** cho hệ thống nhắc nhở của bạn — được viết **đầy đủ, rõ ràng, ngắn gọn**, bao gồm **mọi tính năng đã thống nhất**:

- Nhắc một lần / định kỳ  
- Lịch Dương / Âm (kể cả cuối tháng âm)  
- Nhắc lại nếu chưa hoàn thành  
- Snooze  
- Interval bằng giây  
- Xử lý lỗi FCM  
- Quản lý token theo user  

Mỗi test case gồm: **ID, Mô tả, Dữ liệu đầu vào, Các bước, Kết quả mong đợi**.

---

# 🧪 Bộ Test Case – Ứng Dụng Nhắc Nhở

---

### **TC-01: Nhắc một lần (không nhắc lại)**
- **Mô tả**: Gửi 1 lần rồi hoàn thành ngay.
- **Đầu vào**:  
  `type=one_time`, `repeat_strategy=none`, `next_trigger_at=now+10s`
- **Các bước**: Worker quét → gửi FCM.
- **Kết quả**:  
  ✅ Gửi 1 lần → `status=completed`.

---

### **TC-02: Nhắc một lần + nhắc lại nếu chưa hoàn thành**
- **Mô tả**: Gửi lần đầu, không hoàn thành → nhắc lại đúng số lần.
- **Đầu vào**:  
  `type=one_time`, `repeat_strategy=retry_until_complete`,  
  `retry_interval_sec=60`, `max_retries=2`, `next_trigger_at=now+10s`
- **Các bước**:  
  1. Gửi FCM #1  
  2. 60s sau: gửi #2  
  3. 60s sau: gửi #3  
  4. Dừng
- **Kết quả**:  
  ✅ Gửi đúng 3 lần (1 + 2 nhắc lại)  
  ✅ Không gửi lần 4  
  ✅ `status` vẫn là `"active"`

---

### **TC-03: Nhắc định kỳ theo lịch – hàng ngày**
- **Mô tả**: Nhắc 8h mỗi ngày.
- **Đầu vào**:  
  `type=recurring`, `recurrence_pattern={"type":"daily"}`,  
  `trigger_time_of_day="08:00"`, `next_trigger_at=ngày mai 08:00 UTC`
- **Các bước**: Worker gửi → cập nhật `next_trigger_at = ngày kế 08:00`
- **Kết quả**:  
  ✅ Luôn gửi lúc 08:00 mỗi ngày  
  ✅ Không lệch dù có snooze/trễ

---

### **TC-04: Nhắc định kỳ theo interval – mỗi 7 giờ**
- **Mô tả**: Lần đầu 8h ngày kia, sau đó mỗi 7h.
- **Đầu vào**:  
  `type=recurring`, `recurrence_pattern={"interval_seconds":25200}`,  
  `next_trigger_at=2025-11-04 08:00:00`
- **Các bước**:  
  Gửi lúc 08:00 → 15:00 → 22:00 → 05:00...
- **Kết quả**:  
  ✅ Mỗi lần = lần trước + 7h  
  ✅ Không dùng `trigger_time_of_day`

---

### **TC-05: Lịch Âm – ngày cố định hàng tháng**
- **Mô tả**: Nhắc 15 âm hàng tháng.
- **Đầu vào**:  
  `calendar_type=lunar`, `recurrence_pattern={"type":"monthly","day_of_month":15}`
- **Các bước**: Tính ngày dương cho tháng 10, 11, 12 âm...
- **Kết quả**:  
  ✅ Ngày dương thay đổi đúng (24/11, 23/12, ...)  
  ✅ Không dùng interval

---

### **TC-06: Lịch Âm – cuối tháng âm**
- **Mô tả**: Nhắc vào ngày cuối tháng âm.
- **Đầu vào**:  
  `calendar_type=lunar`, `recurrence_pattern={"type":"lunar_last_day_of_month"}`
- **Các bước**: Tháng 10 âm có 29 ngày → gửi 29/10 âm; tháng 11 có 30 ngày → gửi 30/11 âm.
- **Kết quả**:  
  ✅ Luôn gửi vào ngày cuối cùng của tháng âm

---

### **TC-07: Lịch Âm – ngày không tồn tại (30/2 âm)**
- **Mô tả**: Tháng 2 âm chỉ 29 ngày.
- **Đầu vào**:  
  `calendar_type=lunar`, `recurrence_pattern={"type":"monthly","day_of_month":30}`
- **Kết quả**:  
  ✅ Bỏ qua tháng 2 âm (vì không có ngày 30)

---

### **TC-08: Snooze – hoãn nhắc**
- **Mô tả**: Hoãn 10 phút → không gửi trong thời gian hoãn.
- **Đầu vào**:  
  Reminder đến hạn, client gọi: `snooze_until=now+600`
- **Các bước**: Worker quét trong 10 phút → bỏ qua.
- **Kết quả**:  
  ✅ Không gửi FCM trong thời gian `snooze_until`  
  ✅ Gửi ngay khi `snooze_until` qua

---

### **TC-09: Lỗi FCM – token không hợp lệ**
- **Mô tả**: FCM trả về `UNREGISTERED`.
- **Các bước**: Worker gửi → nhận lỗi → gọi PATCH user.
- **Kết quả**:  
  ✅ `user.is_fcm_active = false`  
  ✅ `user.fcm_token = null`  
  ✅ Không gửi cho user này nữa

---

### **TC-10: Lỗi hệ thống – cấu hình FCM sai**
- **Mô tả**: HTTP 401 do token OAuth hết hạn.
- **Các bước**: Worker gửi → nhận 401 → gọi PATCH system_status.
- **Kết quả**:  
  ✅ `system_status.worker_enabled = false`  
  ✅ Worker dừng ở các lần quét sau

---

### **TC-11: Đa thiết bị – ghi đè token**
- **Mô tả**: Đăng nhập trên thiết bị mới → token cũ bị thay.
- **Các bước**:  
  1. Thiết bị A đăng nhập → lưu token A  
  2. Thiết bị B đăng nhập → lưu token B  
  3. Gửi FCM
- **Kết quả**:  
  ✅ Chỉ thiết bị B nhận được thông báo

---

### **TC-12: Nhắc định kỳ – base_on completion**
- **Mô tả**: Lặp từ thời điểm hoàn thành.
- **Đầu vào**:  
  `recurrence_pattern={"interval_seconds":86400,"base_on":"completion"}`
- **Các bước**:  
  1. Gửi FCM  
  2. User hoàn thành lúc 14:00  
  3. Kỳ tiếp = 14:00 + 24h
- **Kết quả**:  
  ✅ Không lên lịch tiếp theo nếu chưa hoàn thành  
  ✅ Kỳ tiếp tính từ `last_completed_at`

---

### **TC-13: Nhắc định kỳ – base_on creation**
- **Mô tả**: Lặp từ ngày gốc, bất kể hoàn thành.
- **Đầu vào**:  
  `recurrence_pattern={"interval_seconds":86400,"base_on":"creation"}`
- **Kết quả**:  
  ✅ Luôn lên lịch tiếp theo sau khi gửi  
  ✅ Không chờ user hoàn thành

---

### **TC-14: Worker – không quét khi bị tắt**
- **Mô tả**: `worker_enabled = false`
- **Các bước**: Worker khởi động → kiểm tra system_status.
- **Kết quả**:  
  ✅ Không query reminders  
  ✅ Không gửi FCM

---

### **TC-15: Năm nhuận âm – tháng thường**
- **Mô tả**: Nhắc 5/6 âm năm 2025 (năm nhuận tháng 6).
- **Kết quả**:  
  ✅ Dùng tháng 6 **thường**, không dùng tháng 6 nhuận  
  ✅ Không hỗ trợ chọn tháng nhuận

---

✅ Tổng cộng: **15 test case**, phủ toàn bộ chức năng cốt lõi.

Bạn có thể dùng bộ test này để:
- Viết unit test / integration test
- Kiểm thử thủ công
- Làm tài liệu QA

Chúc bạn triển khai và kiểm thử thành công! 🚀