# RemiAq API Demo - Các Trường Hợp Thực Tế

## 🎯 Mục Đích
Tài liệu này cung cấp các test case thực tế để client developers có thể test API RemiAq với các tình huống thường gặp.

## 🔐 Authentication

### 1. Đăng Ký User Mới
```bash
curl -X POST http://localhost:8090/api/collections/musers/records \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo_user@example.com",
    "password": "123123123",
    "passwordConfirm": "123123123"
  }'
```

### 2. Đăng Nhập
```bash
curl -X POST http://localhost:8090/api/collections/musers/auth-with-password \
  -H "Content-Type: application/json" \
  -d '{
    "identity": "demo_user@example.com",
    "password": "123123123"
  }'
```

**Lưu token từ response để dùng cho các API sau:**
```javascript
const token = "JWT_TOKEN_FROM_RESPONSE";
```

## 📋 Các Test Case Thực Tế

### 🎯 1. One-Time Reminders (Nhắc Một Lần)

#### 1.1. Gửi Ngay Lập Tức
```bash
curl -X POST http://localhost:8090/api/reminders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Họp khẩn cấp",
    "description": "Họp với team ngay lập tức",
    "type": "one_time",
    "calendar_type": "solar",
    "max_crp": 0,
    "crp_interval_sec": 0,
    "status": "active"
  }'
```

#### 1.2. Gửi Vào Thời Gian Cụ Thể
```bash
curl -X POST http://localhost:8090/api/reminders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Cuộc họp 3h chiều",
    "description": "Họp với khách hàng lúc 3PM",
    "type": "one_time",
    "calendar_type": "solar",
    "next_recurring": "2025-11-08T15:00:00Z",
    "max_crp": 0,
    "crp_interval_sec": 0,
    "status": "active"
  }'
```

#### 1.3. One-Time với Retry (Thử Lại)
```bash
curl -X POST http://localhost:8090/api/reminders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Nhắc quan trọng có retry",
    "description": "Sẽ thử lại 3 lần nếu fail",
    "type": "one_time",
    "calendar_type": "solar",
    "max_crp": 3,
    "crp_interval_sec": 300,  # 5 phút
    "status": "active"
  }'
```

### 🔁 2. Recurring Reminders (Nhắc Định Kỳ)

#### 2.1. Hàng Ngày - 8:00 AM
```bash
curl -X POST http://localhost:8090/api/reminders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Uống thuốc buổi sáng",
    "description": "Uống vitamin D mỗi sáng",
    "type": "recurring",
    "calendar_type": "solar",
    "repeat_strategy": "none",
    "recurrence_pattern": {
      "type": "daily",
      "interval": 1,
      "trigger_time_of_day": "08:00"
    },
    "max_crp": 1,
    "crp_interval_sec": 0,
    "status": "active"
  }'
```

#### 2.2. Hàng Tuần - Thứ 2, 9:00 AM
```bash
curl -X POST http://localhost:8090/api/reminders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Họp team hàng tuần",
    "description": "Họp thứ 2 hàng tuần",
    "type": "recurring",
    "calendar_type": "solar",
    "repeat_strategy": "none",
    "recurrence_pattern": {
      "type": "weekly",
      "interval": 1,
      "day_of_week": 1,  # Thứ 2
      "trigger_time_of_day": "09:00"
    },
    "max_crp": 1,
    "crp_interval_sec": 0,
    "status": "active"
  }'
```

#### 2.3. Hàng Tháng - Ngày 15, 10:00 AM
```bash
curl -X POST http://localhost:8090/api/reminders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Trả tiền nhà",
    "description": "Trả tiền thuê nhà ngày 15 hàng tháng",
    "type": "recurring",
    "calendar_type": "solar",
    "repeat_strategy": "none",
    "recurrence_pattern": {
      "type": "monthly",
      "interval": 1,
      "day_of_month": 15,
      "trigger_time_of_day": "10:00"
    },
    "max_crp": 1,
    "crp_interval_sec": 0,
    "status": "active"
  }'
```

#### 2.4. Theo Khoảng Thời Gian - Mỗi 2 Giờ
```bash
curl -X POST http://localhost:8090/api/reminders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Nghỉ giải lao",
    "description": "Nghỉ mắt mỗi 2 giờ",
    "type": "recurring",
    "calendar_type": "solar",
    "repeat_strategy": "none",
    "recurrence_pattern": {
      "type": "interval_seconds",
      "interval_seconds": 7200  # 2 giờ
    },
    "max_crp": 1,
    "crp_interval_sec": 0,
    "status": "active"
  }'
```

### 🌙 3. Lunar Calendar Reminders (Lịch Âm)

#### 3.1. Hàng Tháng Theo Âm Lịch
```bash
curl -X POST http://localhost:8090/api/reminders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Giỗ tổ",
    "description": "Giỗ tổ ngày 10 âm lịch",
    "type": "recurring",
    "calendar_type": "lunar",
    "repeat_strategy": "none",
    "recurrence_pattern": {
      "type": "monthly",
      "interval": 1,
      "day_of_month": 10,
      "trigger_time_of_day": "08:00"
    },
    "max_crp": 1,
    "crp_interval_sec": 0,
    "status": "active"
  }'
```

#### 3.2. Ngày Cuối Tháng Âm
```bash
curl -X POST http://localhost:8090/api/reminders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Cuối tháng âm",
    "description": "Nhắc cuối tháng âm lịch",
    "type": "recurring",
    "calendar_type": "lunar",
    "repeat_strategy": "none",
    "recurrence_pattern": {
      "type": "lunar_last_day_of_month",
      "trigger_time_of_day": "20:00"
    },
    "max_crp": 1,
    "crp_interval_sec": 0,
    "status": "active"
  }'
```

### ⚡ 4. Advanced Scenarios (Tình Huống Nâng Cao)

#### 4.1. Recurring với CRP Retry
```bash
curl -X POST http://localhost:8090/api/reminders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Nhắc quan trọng có retry",
    "description": "Sẽ thử lại nếu user không phản hồi",
    "type": "recurring",
    "calendar_type": "solar",
    "repeat_strategy": "none",
    "recurrence_pattern": {
      "type": "daily",
      "interval": 1,
      "trigger_time_of_day": "09:00"
    },
    "max_crp": 3,
    "crp_interval_sec": 600,  # 10 phút
    "status": "active"
  }'
```

#### 4.2. Paused Reminder
```bash
curl -X POST http://localhost:8090/api/reminders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Nhắc tạm dừng",
    "description": "Sẽ kích hoạt sau",
    "type": "recurring",
    "calendar_type": "solar",
    "repeat_strategy": "none",
    "recurrence_pattern": {
      "type": "daily",
      "interval": 1,
      "trigger_time_of_day": "08:00"
    },
    "max_crp": 1,
    "crp_interval_sec": 0,
    "status": "paused"  # Tạm dừng
  }'
```

### 🔍 5. Query Operations (Truy Vấn)

#### 5.1. Lấy Tất Cả Reminders Của User
```bash
curl -X GET http://localhost:8090/api/reminders/mine \
  -H "Authorization: Bearer $TOKEN"
```

#### 5.2. Lấy Reminder Theo ID
```bash
curl -X GET http://localhost:8090/api/reminders/REMINDER_ID \
  -H "Authorization: Bearer $TOKEN"
```

### ✏️ 6. Update Operations (Cập Nhật)

#### 6.1. Update Reminder
```bash
curl -X PUT http://localhost:8090/api/reminders/REMINDER_ID \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Tiêu đề đã cập nhật",
    "description": "Mô tả mới"
  }'
```

#### 6.2. Complete Reminder
```bash
curl -X POST http://localhost:8090/api/reminders/REMINDER_ID/complete \
  -H "Authorization: Bearer $TOKEN"
```

#### 6.3. Snooze Reminder (Trì Hoãn)
```bash
curl -X POST http://localhost:8090/api/reminders/REMINDER_ID/snooze \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "snooze_minutes": 30  # Trì hoãn 30 phút
  }'
```

#### 6.4. Delete Reminder
```bash
curl -X DELETE http://localhost:8090/api/reminders/REMINDER_ID \
  -H "Authorization: Bearer $TOKEN"
```

## 🧪 Test Scripts (JavaScript)

### Test Helper Function
```javascript
const API_BASE = 'http://localhost:8090';
let authToken = '';
let userId = '';

// Helper function để gọi API
async function callAPI(endpoint, method = 'GET', data = null) {
  const options = {
    method,
    headers: {
      'Content-Type': 'application/json',
    },
  };

  if (authToken) {
    options.headers.Authorization = `Bearer ${authToken}`;
  }

  if (data) {
    options.body = JSON.stringify(data);
  }

  const response = await fetch(`${API_BASE}${endpoint}`, options);
  return await response.json();
}

// Test flow
async function testFlow() {
  try {
    // 1. Đăng ký user
    console.log('1. Đăng ký user...');
    const registerResponse = await callAPI('/api/collections/musers/records', 'POST', {
      email: 'test_user@example.com',
      password: '123123123',
      passwordConfirm: '123123123'
    });
    console.log('Đăng ký thành công:', registerResponse);

    // 2. Đăng nhập
    console.log('2. Đăng nhập...');
    const loginResponse = await callAPI('/api/collections/musers/auth-with-password', 'POST', {
      identity: 'test_user@example.com',
      password: '123123123'
    });
    
    authToken = loginResponse.token;
    userId = loginResponse.record.id;
    console.log('Đăng nhập thành công. Token:', authToken.substring(0, 20) + '...');

    // 3. Tạo reminder
    console.log('3. Tạo reminder...');
    const reminderResponse = await callAPI('/api/reminders', 'POST', {
      title: 'Test Reminder',
      description: 'Reminder for testing',
      type: 'one_time',
      calendar_type: 'solar',
      max_crp: 0,
      crp_interval_sec: 0,
      status: 'active'
    });
    console.log('Tạo reminder thành công:', reminderResponse);

    // 4. Lấy danh sách reminders
    console.log('4. Lấy danh sách reminders...');
    const remindersResponse = await callAPI('/api/reminders/mine');
    console.log('Danh sách reminders:', remindersResponse);

  } catch (error) {
    console.error('Lỗi:', error);
  }
}

// Chạy test
testFlow();
```

## 🚨 Common Errors (Lỗi Thường Gặp)

### 1. Authentication Errors
- `401 Unauthorized`: Token không hợp lệ hoặc hết hạn
- `403 Forbidden`: Không có quyền truy cập

### 2. Validation Errors
- `400 Bad Request`: Dữ liệu gửi lên không hợp lệ
- Thiếu trường bắt buộc (title, type, calendar_type)
- Định dạng thời gian không đúng

### 3. Resource Errors
- `404 Not Found`: Reminder không tồn tại
- `409 Conflict`: Dữ liệu bị trùng lặp

## 📊 Response Format

### Success Response
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {
    "id": "reminder_id",
    "title": "Reminder Title",
    // ... other fields
  }
}
```

### Error Response
```json
{
  "success": false,
  "message": "Error description",
  "error": "Detailed error information"
}
```

## 🔧 Testing Tips

1. **Start Server**: `go run ./cmd/server serve`
2. **Test Authentication** trước khi test reminders
3. **Check Response Status** và error messages
4. **Use Postman Collection** có sẵn trong project
5. **Monitor Logs** để debug issues

## 📝 Next Steps

1. Test tất cả các API endpoints
2. Verify FCM notifications được gửi
3. Test worker processing (chạy mỗi 60s)
4. Test các trường hợp edge cases

---

**Chúc bạn test thành công!** 🎉

*Document version: 1.0 - Updated: 2025-11-08*