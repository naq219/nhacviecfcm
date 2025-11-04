Tuyệt! Dưới đây là **toàn bộ tài liệu đặc tả hệ thống** được viết lại **từ đầu**, **cập nhật đầy đủ theo tất cả quyết định và thay đổi gần đây của bạn**, bao gồm:

- Dùng **PocketBase**
- **1 user → 1 FCM token** (ghi đè khi đăng nhập)
- **Không hỗ trợ DST**, **ưu tiên đơn giản**
- Hỗ trợ **lịch Dương / Âm**, **cuối tháng âm**
- **Snooze** (hoãn nhắc)
- **Nhắc theo interval bằng giây** (`interval_seconds`)
- **Xử lý lỗi FCM** (toàn cục + theo user)
- **Worker dùng API PocketBase**, không truy cập SQLite trực tiếp
- **Chưa phát triển nhắc theo vị trí**

---

# Đặc Tả Hệ Thống: Ứng Dụng Nhắc Nhở với FCM + PocketBase  
**(Phiên bản tối giản – cập nhật 02/11/2025)**

---

## 1. Mục tiêu

Xây dựng ứng dụng nhắc nhở cho phép:
- Người dùng **đăng ký / đăng nhập**.
- Mỗi người dùng chỉ lưu **1 token FCM** (thiết bị mới ghi đè).
- Gửi thông báo FCM khi đến hạn.
- Hỗ trợ:
  - Nhắc **một lần** hoặc **định kỳ**.
  - **Lịch Dương** hoặc **lịch Âm** (kể cả “cuối tháng âm”).
  - **Nhắc lại trong kỳ** nếu chưa hoàn thành.
  - **Snooze** (hoãn nhắc).
  - **Lặp theo khoảng thời gian** (vd: mỗi 7 giờ).
- Tự động **xử lý lỗi FCM** và **báo trạng thái dịch vụ**.

---

## 2. Công nghệ

| Thành phần | Công nghệ |
|-----------|----------|
| Backend + Auth + DB | **PocketBase** |
| Gửi thông báo | **Firebase Cloud Messaging (FCM)** |
| Worker | Script bên ngoài (Python/Go), gọi **PocketBase REST API**, chạy mỗi phút |
| Client | Mobile hoặc Web |

> 💡 **Không hỗ trợ DST**, tất cả thời gian lưu theo **UTC**.

---

## 3. Mô hình dữ liệu

### 3.1. `musers` (mở rộng)

| Trường | Kiểu | Mô tả |
|-------|------|------|
| `fcm_token` | text | Token FCM hiện tại |
| `is_fcm_active` | bool | `true` = có thể nhận FCM |

---

### 3.2. `reminders`

| Trường | Kiểu | Mô tả |
|-------|------|------|
| `user` | relation | |
| `title` | text | |
| `calendar_type` | text | `"solar"` / `"lunar"` |
| `type` | text | `"one_time"` / `"recurring"` |
| `repeat_strategy` | text | `"none"` / `"retry_until_complete"` |
| `retry_interval_sec` | number | Khoảng cách nhắc lại (nếu có) |
| `max_retries` | number | Số lần nhắc lại tối đa |
| `trigger_time_of_day` | text | `"HH:MM"` (UTC) — **chỉ dùng nếu lặp theo lịch** |
| `recurrence_pattern` | json | Xem mục 4 |
| `next_trigger_at` | date-time | UTC — thời điểm gửi tiếp theo |
| `last_completed_at` | date-time | |
| `snooze_until` | date-time | Thời điểm hết hoãn |
| `status` | text | `"active"`, `"completed"`, `"cancelled"` |
| `created` | date-time | |

---

### 3.3. `system_status` (1 bản ghi, `mid = 1`)

| Trường | Kiểu | Mô tả |
|-------|------|------|
| `worker_enabled` | bool | `true` = worker đang hoạt động |
| `last_error` | text | Nội dung lỗi |
| `error_at` | date-time | |

---

## 4. Cấu hình nhắc định kỳ (`recurrence_pattern`)

### 4.1. Lặp theo lịch (dùng `trigger_time_of_day`)
```json
{ "type": "daily" }
{ "type": "weekly", "days_of_week": ["mon", "wed"] }
{ "type": "monthly", "day_of_month": 15 }
{ "type": "yearly", "month": 12, "day_of_month_yearly": 23 }
{ "type": "lunar_last_day_of_month" }
```

### 4.2. Lặp theo khoảng thời gian (không dùng `trigger_time_of_day`)
```json
{ "interval_seconds": 25200 }  // mỗi 7 giờ
```

> 💡 Với `interval_seconds`, **lần đầu tiên** được đặt bằng `next_trigger_at`, các lần sau = lần trước + `interval_seconds`.

---

## 5. Luồng xử lý chính

### 5.1. Worker (mỗi phút)
1. GET `/system_status/1` → nếu `worker_enabled == false` → **dừng**.
2. GET `/reminders?filter=status='active'&&next_trigger_at<=now&&(snooze_until IS NULL OR snooze_until<=now)`
3. Với mỗi reminder:
   - GET user → nếu `is_fcm_active == false` → bỏ qua.
   - Gửi FCM.
   - Xử lý phản hồi:
     - Lỗi hệ thống → tắt `worker_enabled`.
     - Lỗi token → tắt `is_fcm_active` của user.
   - Cập nhật `next_trigger_at` hoặc `status` theo loại nhắc.

### 5.2. Snooze
- Khi user hoãn: client gọi PATCH → cập nhật `snooze_until = NOW + X`.
- Worker **bỏ qua** reminder đó cho đến khi `snooze_until` qua.

### 5.3. Lịch Âm
- Chỉ cho phép: `monthly`, `yearly`, `lunar_last_day_of_month`.
- Không hỗ trợ `interval_seconds` với lịch Âm.

---

## 6. Xử lý lỗi FCM

| Loại lỗi | Hành động |
|--------|----------|
| **Hệ thống** (401, 403, timeout) | Đặt `worker_enabled = false` |
| **Thiết bị** (`UNREGISTERED`) | Đặt `is_fcm_active = false` |

---

## 7. Lưu ý triển khai

- **Tất cả thời gian trong DB là UTC**.
- **Client chuyển đổi múi giờ khi hiển thị**.
- **Không dùng SQLite trực tiếp** — worker chỉ gọi API.
- **PocketBase cần index** trên `(status, next_trigger_at)`.

---

## 8. API System Status

- GET `/api/system_status`
  - Trả về bản ghi singleton (`mid = 1`): `{ mid, worker_enabled, last_error, updated }`

- PUT `/api/system_status`
  - Body cho phép cập nhật:
    - `worker_enabled: boolean` (bật/tắt worker)
    - `last_error: string` (ghi chú lỗi hệ thống)
  - Hành vi:
    - Nếu `worker_enabled = true`: bật worker; nếu không có `last_error` → xóa lỗi; nếu có → cập nhật lỗi.
    - Nếu `worker_enabled = false`: tắt worker; nếu không có `last_error` → dùng mặc định "manually disabled"; nếu có → ghi lại.
    - Nếu chỉ có `last_error` (không thay đổi `worker_enabled`): cập nhật lỗi.
  - Response: `{ success, message, data: SystemStatus }`

---

## 9. API truy vấn SQL thô (Legacy)

Để đảm bảo tương thích ngược với các hệ thống cũ, ứng dụng cung cấp các endpoint cho phép thực thi các câu lệnh SQL thô. Các endpoint này được bảo vệ bởi các quy tắc validation nghiêm ngặt để ngăn chặn các truy vấn nguy hiểm.

- **GET/POST `/api/rquery`**: Thực thi các câu lệnh `SELECT`.
- **GET/POST `/api/rinsert`**: Thực thi các câu lệnh `INSERT`.
- **GET/PUT `/api/rupdate`**: Thực thi các câu lệnh `UPDATE`.
- **GET/DELETE `/api/rdelete`**: Thực thi các câu lệnh `DELETE`.

### Luồng xử lý:
1.  Client gửi request chứa câu lệnh SQL trong body (POST/PUT/DELETE) hoặc query parameter `q` (GET).
2.  Middleware xác thực câu lệnh dựa trên loại (ví dụ: chỉ cho phép `SELECT` ở endpoint `rquery`).
3.  Nếu hợp lệ, `QueryRepository` sẽ thực thi câu lệnh và trả về kết quả.

> ⚠️ **Cảnh báo**: Các endpoint này chỉ nên được sử dụng khi thực sự cần thiết và bởi các client được tin tưởng.

---

✅ Tài liệu này phản ánh **đúng thiết kế hiện tại** của bạn: **đơn giản, đủ mạnh, dễ triển khai**.

Chúc bạn code vui và hệ thống chạy mượt! 🚀