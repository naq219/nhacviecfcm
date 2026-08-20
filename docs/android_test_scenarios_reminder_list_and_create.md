# Test Scenarios & Test Cases - Reminder List và Create Reminder

**Project**: RemiAq Android Client  
**Version**: 1.0  
**Last Updated**: 2025-11-23  
**Author**: QA Team

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Test Scope](#test-scope)
3. [Test Scenarios - Reminder List](#test-scenarios---reminder-list)
4. [Test Cases - Reminder List](#test-cases---reminder-list)
5. [Test Scenarios - Create Reminder](#test-scenarios---create-reminder)
6. [Test Cases - Create Reminder](#test-cases---create-reminder)
7. [Integration Test Scenarios](#integration-test-scenarios)
8. [Non-Functional Test Cases](#non-functional-test-cases)
9. [Test Data](#test-data)
10. [Defect Severity Classification](#defect-severity-classification)

---

## Overview

Tài liệu này mô tả chi tiết các test scenarios và test cases cho 2 chức năng chính:
- **Danh Sách Reminder (Reminder List)**: Hiển thị danh sách các reminders của user
- **Tạo Reminder (Create Reminder)**: Tạo mới một reminder

### References
- Technical Documentation: [remiaq_complete_technical_documentation_v4.md](file:///d:/PROJECT/NHAC_NHO_GROUP/AndroidReminaq6/docs/remiaq_complete_technical_documentation_v4.md)
- API Test Cases: [API_DEMO_TEST_CASES_v4.md](file:///d:/PROJECT/NHAC_NHO_GROUP/AndroidReminaq6/docs/API_DEMO_TEST_CASES_v4.md)
- UI Designs: 
  - [screen_list_reminder.png](file:///d:/PROJECT/NHAC_NHO_GROUP/AndroidReminaq6/docs/screen_list_reminder.png)
  - [create_reminder.png](file:///d:/PROJECT/NHAC_NHO_GROUP/AndroidReminaq6/docs/create_reminder.png)

---

## Test Scope

### In Scope
✅ Reminder List Screen functionality  
✅ Create Reminder Screen functionality  
✅ API integration (GET, POST)  
✅ UI/UX validation  
✅ Data validation  
✅ Error handling  
✅ Navigation flows  
✅ State management  

### Out of Scope
❌ Edit Reminder functionality  
❌ Delete Reminder functionality  
❌ Complete Reminder functionality  
❌ Snooze Reminder functionality  
❌ Push Notifications  
❌ Background Worker processing  

---

## Test Scenarios - Reminder List

### TS-RL-001: Display Reminder List
**Mô tả**: Kiểm tra hiển thị danh sách reminders của user đã đăng nhập

**Preconditions**: 
- User đã đăng nhập thành công
- Database có ít nhất 1 reminder của user

**Expected Result**: 
- Hiển thị danh sách reminders với đầy đủ thông tin
- Mỗi reminder item hiển thị: Title, Description, Thời gian, Loại lặp lại

---

### TS-RL-002: Empty State
**Mô tả**: Kiểm tra hiển thị khi user chưa có reminder nào

**Preconditions**: 
- User đã đăng nhập
- Database không có reminder nào của user

**Expected Result**: 
- Hiển thị empty state message
- Có button/link để tạo reminder mới

---

### TS-RL-003: Loading State
**Mô tả**: Kiểm tra trạng thái loading khi fetch data

**Preconditions**: 
- User đã đăng nhập
- Network có độ trễ

**Expected Result**: 
- Hiển thị loading indicator trong khi fetch data
- Loading indicator biến mất khi data load xong

---

### TS-RL-004: Error Handling
**Mô tả**: Kiểm tra xử lý lỗi khi fetch data thất bại

**Preconditions**: 
- User đã đăng nhập
- API endpoint không available hoặc trả về lỗi

**Expected Result**: 
- Hiển thị error message rõ ràng
- Có option để retry

---

### TS-RL-005: Refresh Data
**Mô tả**: Kiểm tra pull-to-refresh để cập nhật data

**Preconditions**: 
- User đã đăng nhập
- Đang xem reminder list

**Expected Result**: 
- User kéo xuống để refresh
- Data được reload từ server
- List cập nhật với data mới nhất

---

### TS-RL-006: Filter by Type
**Mô tả**: Kiểm tra filter reminders theo loại (one-time/recurring)

**Preconditions**: 
- User đã đăng nhập
- Database có cả one-time và recurring reminders

**Expected Result**: 
- Filter hiển thị đúng reminders theo loại đã chọn

---

### TS-RL-007: Filter by Status
**Mô tả**: Kiểm tra filter reminders theo status (active/completed/paused)

**Preconditions**: 
- User đã đăng nhập
- Database có reminders với các status khác nhau

**Expected Result**: 
- Filter hiển thị đúng reminders theo status đã chọn

---

### TS-RL-008: Navigation to Detail/Edit
**Mô tả**: Kiểm tra navigation khi click vào reminder item

**Preconditions**: 
- User đang xem reminder list
- List có ít nhất 1 reminder

**Expected Result**: 
- Click vào reminder → navigate to detail/edit screen
- Data được truyền đúng

---

### TS-RL-009: Time Format Display
**Mô tả**: Kiểm tra hiển thị thời gian đã được convert từ UTC sang Local Time

**Preconditions**: 
- User đã đăng nhập
- Database có reminders với next_action_at ở UTC

**Expected Result**: 
- Thời gian hiển thị theo device timezone
- Format: DD/MM/YYYY HH:MM hoặc theo format design

---

### TS-RL-010: Pagination/Scrolling
**Mô tả**: Kiểm tra scrolling khi có nhiều reminders

**Preconditions**: 
- User có > 20 reminders

**Expected Result**: 
- List scroll smooth
- Nếu có pagination: load thêm data khi scroll đến cuối

---

## Test Cases - Reminder List

### TC-RL-001: Verify Display of Active Reminders
**Test Scenario**: TS-RL-001  
**Priority**: High  
**Test Type**: Functional

**Preconditions**:
- User đã login với email: `test_user@example.com`
- Database có 3 active reminders của user

**Test Steps**:
1. Open app
2. Navigate to Reminder List screen
3. Verify list hiển thị

**Test Data**:
```json
{
  "reminders": [
    {
      "title": "Meeting with team",
      "type": "recurring",
      "status": "active",
      "next_action_at": "2025-11-23T09:00:00Z"
    },
    {
      "title": "Doctor appointment",
      "type": "one_time",
      "status": "active",
      "next_action_at": "2025-11-24T14:30:00Z"
    }
  ]
}
```

**Expected Result**:
- ✅ List hiển thị 3 reminders
- ✅ Mỗi item hiển thị đầy đủ: Title, Description, Next Action Time, Type
- ✅ Thời gian hiển thị theo Local Timezone
- ✅ UI match design: [screen_list_reminder.png](file:///d:/PROJECT/NHAC_NHO_GROUP/AndroidReminaq6/docs/screen_list_reminder.png)

**Actual Result**: [To be filled during testing]

---

### TC-RL-002: Verify Empty State Display
**Test Scenario**: TS-RL-002  
**Priority**: High  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Database không có reminder nào của user

**Test Steps**:
1. Open app
2. Navigate to Reminder List screen
3. Verify empty state

**Expected Result**:
- ✅ Hiển thị empty state message: "Bạn chưa có reminder nào"
- ✅ Hiển thị "Tạo Reminder" button
- ✅ Click button → navigate to Create Reminder screen

**Actual Result**: [To be filled during testing]

---

### TC-RL-003: Verify Loading State
**Test Scenario**: TS-RL-003  
**Priority**: Medium  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Simulate slow network (throttling)

**Test Steps**:
1. Open app với network throttling
2. Navigate to Reminder List screen
3. Observe loading state

**Expected Result**:
- ✅ Hiển thị loading indicator (CircularProgressIndicator)
- ✅ Loading indicator ở center screen
- ✅ Sau khi data load xong → loading indicator biến mất
- ✅ List hiển thị data

**Actual Result**: [To be filled during testing]

---

### TC-RL-004: Verify API Error Handling
**Test Scenario**: TS-RL-004  
**Priority**: High  
**Test Type**: Negative

**Preconditions**:
- User đã login
- API endpoint trả về 500 Internal Server Error

**Test Steps**:
1. Mock API response: HTTP 500
2. Open Reminder List screen
3. Observe error handling

**Expected Result**:
- ✅ Hiển thị error message: "Không thể tải dữ liệu. Vui lòng thử lại."
- ✅ Hiển thị "Thử lại" button
- ✅ Click "Thử lại" → gọi lại API

**Actual Result**: [To be filled during testing]

---

### TC-RL-005: Verify Network Error Handling
**Test Scenario**: TS-RL-004  
**Priority**: High  
**Test Type**: Negative

**Preconditions**:
- User đã login
- Device không có internet connection

**Test Steps**:
1. Turn off WiFi and Mobile Data
2. Open Reminder List screen
3. Observe error handling

**Expected Result**:
- ✅ Hiển thị error message: "Không có kết nối mạng"
- ✅ Hiển thị "Thử lại" button

**Actual Result**: [To be filled during testing]

---

### TC-RL-006: Verify Pull-to-Refresh
**Test Scenario**: TS-RL-005  
**Priority**: Medium  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Đang xem Reminder List với 2 items

**Test Steps**:
1. Navigate to Reminder List
2. Create 1 reminder mới qua API/other device
3. Kéo màn hình xuống để refresh
4. Verify list cập nhật

**Expected Result**:
- ✅ Pull-to-refresh gesture hoạt động
- ✅ Hiển thị loading indicator trong lúc refresh
- ✅ List cập nhật với 3 items (bao gồm item mới)

**Actual Result**: [To be filled during testing]

---

### TC-RL-007: Verify Time Format Conversion (UTC to Local)
**Test Scenario**: TS-RL-009  
**Priority**: High  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Device timezone: Asia/Ho_Chi_Minh (UTC+7)
- Database có reminder với `next_action_at: "2025-11-23T09:00:00Z"` (UTC)

**Test Steps**:
1. Navigate to Reminder List
2. Verify time hiển thị trên reminder item

**Expected Result**:
- ✅ Thời gian hiển thị: "23/11/2025 16:00" (UTC+7)
- ✅ KHÔNG hiển thị "09:00" (UTC time)

**Actual Result**: [To be filled during testing]

---

### TC-RL-008: Verify Reminder Type Display
**Test Scenario**: TS-RL-001  
**Priority**: Medium  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Database có:
  - 1 one-time reminder
  - 1 recurring daily reminder
  - 1 recurring weekly reminder

**Test Steps**:
1. Navigate to Reminder List
2. Verify type display cho mỗi reminder

**Expected Result**:
- ✅ One-time reminder: Hiển thị "Một lần" hoặc icon phù hợp
- ✅ Recurring daily: Hiển thị "Hàng ngày" hoặc icon repeat
- ✅ Recurring weekly: Hiển thị "Hàng tuần" hoặc icon repeat

**Actual Result**: [To be filled during testing]

---

### TC-RL-009: Verify Reminder Status Display
**Test Scenario**: TS-RL-001  
**Priority**: Medium  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Database có:
  - 1 active reminder
  - 1 paused reminder
  - 1 completed reminder

**Test Steps**:
1. Navigate to Reminder List
2. Verify status display/badge

**Expected Result**:
- ✅ Active: Hiển thị bình thường, không có badge đặc biệt
- ✅ Paused: Hiển thị badge "Tạm dừng" hoặc icon pause
- ✅ Completed: Hiển thị badge "Đã hoàn thành" hoặc strikethrough

**Actual Result**: [To be filled during testing]

---

### TC-RL-010: Verify Navigation to Create Reminder
**Test Scenario**: TS-RL-001  
**Priority**: High  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Đang ở Reminder List screen

**Test Steps**:
1. Click FAB "+" button (hoặc "Tạo Reminder" button)
2. Verify navigation

**Expected Result**:
- ✅ Navigate to Create Reminder screen
- ✅ Screen title: "Tạo Reminder"

**Actual Result**: [To be filled during testing]

---

### TC-RL-011: Verify Click on Reminder Item
**Test Scenario**: TS-RL-008  
**Priority**: Medium  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Reminder List có ít nhất 1 reminder

**Test Steps**:
1. Navigate to Reminder List
2. Click vào 1 reminder item
3. Verify navigation

**Expected Result**:
- ✅ Navigate to Reminder Detail/Edit screen
- ✅ Data của reminder được truyền đúng

**Actual Result**: [To be filled during testing]

---

### TC-RL-012: Verify Scrolling with Many Reminders
**Test Scenario**: TS-RL-010  
**Priority**: Low  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Database có 50 reminders

**Test Steps**:
1. Navigate to Reminder List
2. Scroll down danh sách
3. Verify scrolling performance

**Expected Result**:
- ✅ List scroll smooth, không lag
- ✅ Tất cả 50 items đều load và hiển thị đúng
- ✅ Scroll to bottom → không bị crash

**Actual Result**: [To be filled during testing]

---

### TC-RL-013: Verify Data Refresh After Create
**Test Scenario**: TS-RL-001  
**Priority**: High  
**Test Type**: Integration

**Preconditions**:
- User đã login
- Reminder List có 2 items

**Test Steps**:
1. Navigate to Reminder List (2 items)
2. Click "Tạo Reminder" → Create Reminder screen
3. Tạo 1 reminder mới thành công
4. Navigate back to Reminder List
5. Verify list cập nhật

**Expected Result**:
- ✅ Reminder List hiển thị 3 items
- ✅ Reminder mới tạo xuất hiện trong list
- ✅ Sắp xếp theo `next_action_at` (sắp tới nhất lên đầu)

**Actual Result**: [To be filled during testing]

---

### TC-RL-014: Verify Filter by Type
**Test Scenario**: TS-RL-006  
**Priority**: Medium  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Database có cả one-time và recurring reminders

**Test Steps**:
1. Navigate to Reminder List
2. Open filter and select Type = "Một lần"
3. Verify list items
4. Change filter to Type = "Lặp lại"
5. Verify list items

**Expected Result**:
- ✅ Type = "Một lần" → chỉ hiển thị one-time reminders
- ✅ Type = "Lặp lại" → chỉ hiển thị recurring reminders

**Actual Result**: [To be filled during testing]

---

### TC-RL-015: Verify Filter by Status
**Test Scenario**: TS-RL-007  
**Priority**: Medium  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Database có reminders với các status: active, paused, completed

**Test Steps**:
1. Navigate to Reminder List
2. Open filter and select Status = "active"
3. Verify list items
4. Change filter to Status = "paused"
5. Verify list items
6. Change filter to Status = "completed"
7. Verify list items

**Expected Result**:
- ✅ Mỗi trạng thái filter chỉ hiển thị đúng reminders thuộc trạng thái đó

**Actual Result**: [To be filled during testing]

---

### TC-RL-016: Unauthorized Without Token
**Test Scenario**: TS-RL-004  
**Priority**: High  
**Test Type**: Negative

**Preconditions**:
- App chưa có Authorization token (pb.authStore.clear())

**Test Steps**:
1. Open Reminder List screen
2. Trigger API fetch

**Expected Result**:
- ✅ API trả về 401 Unauthorized
- ✅ App hiển thị thông báo và chuyển về màn hình Login

**Actual Result**: [To be filled during testing]

---

## Test Scenarios - Create Reminder

### TS-CR-001: Create One-Time Reminder
**Mô tả**: Kiểm tra tạo một reminder chỉ chạy 1 lần

**Preconditions**: 
- User đã đăng nhập
- Đang ở Create Reminder screen

**Expected Result**: 
- Tạo thành công one-time reminder
- Redirect về Reminder List
- Reminder mới xuất hiện trong list

---

### TS-CR-002: Create Recurring Reminder - Daily
**Mô tả**: Kiểm tra tạo recurring reminder lặp hàng ngày

**Preconditions**: 
- User đã đăng nhập
- Đang ở Create Reminder screen

**Expected Result**: 
- Tạo thành công recurring daily reminder
- `recurrence_pattern.type = "daily"`
- `trigger_time_of_day` được tự động sinh từ `next_action_at`

---

### TS-CR-003: Create Recurring Reminder - Weekly
**Mô tả**: Kiểm tra tạo recurring reminder lặp hàng tuần

**Preconditions**: 
- User đã đăng nhập
- Đang ở Create Reminder screen

**Expected Result**: 
- Tạo thành công recurring weekly reminder
- `recurrence_pattern.type = "weekly"`
- `day_of_week` được set đúng

---

### TS-CR-004: Create Recurring Reminder - Monthly
**Mô tả**: Kiểm tra tạo recurring reminder lặp hàng tháng

**Preconditions**: 
- User đã đăng nhập
- Đang ở Create Reminder screen

**Expected Result**: 
- Tạo thành công recurring monthly reminder
- `recurrence_pattern.type = "monthly"`
- `day_of_month` được set đúng

---

### TS-CR-005: Create Recurring Reminder - Interval
**Mô tả**: Kiểm tra tạo recurring reminder lặp theo khoảng thời gian (giây)

**Preconditions**: 
- User đã đăng nhập
- Đang ở Create Reminder screen

**Expected Result**: 
- Tạo thành công interval reminder
- `recurrence_pattern.type = "interval_seconds"`
- `interval_seconds` được set đúng

---

### TS-CR-006: Create Lunar Calendar Reminder
**Mô tả**: Kiểm tra tạo reminder theo âm lịch

**Preconditions**: 
- User đã đăng nhập
- Đang ở Create Reminder screen

**Expected Result**: 
- Tạo thành công lunar reminder
- `calendar_type = "lunar"`

---

### TS-CR-007: Validation - Required Fields
**Mô tả**: Kiểm tra validation cho các trường bắt buộc

**Preconditions**: 
- User đang ở Create Reminder screen

**Expected Result**: 
- Không cho submit nếu thiếu: title, type, next_action_at
- Hiển thị error message rõ ràng

---

### TS-CR-008: Validation - Invalid Date/Time
**Mô tả**: Kiểm tra validation cho date/time không hợp lệ

**Preconditions**: 
- User đang ở Create Reminder screen

**Expected Result**: 
- Không cho select quá khứ (đối với one-time)
- Error message: "Thời gian phải ở tương lai"

---

### TS-CR-009: Validation - CRP Configuration
**Mô tả**: Kiểm tra validation cho CRP (Child Repeat Pattern)

**Preconditions**: 
- User đang ở Create Reminder screen
- Set `max_crp > 0`

**Expected Result**: 
- Bắt buộc phải nhập `crp_interval_sec > 0`
- Error message nếu `crp_interval_sec = 0`

---

### TS-CR-010: UI State - Loading
**Mô tả**: Kiểm tra loading state khi submit form

**Preconditions**: 
- User đã điền đầy đủ form
- Click "Tạo" button

**Expected Result**: 
- Hiển thị loading indicator
- Disable submit button để tránh double submit

---

### TS-CR-011: UI State - Success
**Mô tả**: Kiểm tra UI khi tạo reminder thành công

**Preconditions**: 
- User đã submit form hợp lệ
- API trả về success

**Expected Result**: 
- Hiển thị success toast/snackbar
- Auto navigate về Reminder List sau 1-2 giây

---

### TS-CR-012: UI State - Error
**Mô tả**: Kiểm tra UI khi tạo reminder thất bại

**Preconditions**: 
- User đã submit form
- API trả về error

**Expected Result**: 
- Hiển thị error message từ server
- Không navigate, user còn ở Create screen
- User có thể sửa và submit lại

---

### TS-CR-013: DatePicker Integration
**Mô tả**: Kiểm tra DatePicker cho chọn ngày

**Preconditions**: 
- User đang ở Create Reminder screen

**Expected Result**: 
- Click vào date field → hiển thị DatePicker
- Chọn ngày → update field
- DatePicker chỉ cho chọn ngày >= hôm nay

---

### TS-CR-014: TimePicker Integration
**Mô tả**: Kiểm tra TimePicker cho chọn giờ

**Preconditions**: 
- User đang ở Create Reminder screen

**Expected Result**: 
- Click vào time field → hiển thị TimePicker
- Chọn giờ → update field
- Format hiển thị: HH:MM (24h) hoặc hh:MM AM/PM

---

### TS-CR-015: Repeat Strategy Selection
**Mô tả**: Kiểm tra chọn repeat strategy (chỉ cho recurring)

**Preconditions**: 
- User đã chọn type = "recurring"

**Expected Result**: 
- Hiển thị RadioButton/Dropdown cho:
  - "Tự động lặp lại" (none)
  - "Chờ hoàn thành" (crp_until_complete)

---

### TS-CR-016: Calendar Type Selection
**Mô tả**: Kiểm tra chọn loại lịch (Solar/Lunar)

**Preconditions**: 
- User đang ở Create Reminder screen

**Expected Result**: 
- RadioButton/Toggle: "Dương lịch" / "Âm lịch"
- Default: "Dương lịch"

---

### TS-CR-017: Dynamic Form Based on Type
**Mô tả**: Kiểm tra form thay đổi dựa trên reminder type

**Preconditions**: 
- User đang ở Create Reminder screen

**Expected Result**: 
- Chọn "One-time" → ẩn recurrence fields
- Chọn "Recurring" → hiển thị recurrence pattern options

---

### TS-CR-018: Cancel/Back Navigation
**Mô tả**: Kiểm tra cancel và back navigation

**Preconditions**: 
- User đã nhập một số data vào form

**Expected Result**: 
- Click "Hủy" hoặc back button
- Hiển thị confirmation dialog: "Bạn có chắc muốn hủy?"
- Confirm → navigate back, data không được save

---

---

## Test Cases - Create Reminder

### TC-CR-001: Create One-Time Reminder - Valid Data
**Test Scenario**: TS-CR-001  
**Priority**: Critical  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Navigate to Create Reminder screen

**Test Steps**:
1. Enter Title: "Doctor Appointment"
2. Enter Description: "Annual health checkup"
3. Select Type: "Một lần"
4. Select Calendar Type: "Dương lịch"
5. Select Date: "2025-11-25"
6. Select Time: "14:30"
7. Set max_crp: 0 (no retry)
8. Click "Tạo" button
9. Verify API call và response

**Test Data**:
```json
{
  "title": "Doctor Appointment",
  "description": "Annual health checkup",
  "type": "one_time",
  "calendar_type": "solar",
  "next_action_at": "2025-11-25T07:30:00Z",
  "max_crp": 0,
  "crp_interval_sec": 0,
  "status": "active"
}
```

**Expected Result**:
- ✅ API POST /api/reminders được gọi với đúng payload
- ✅ Server trả về 200 OK với reminder ID
- ✅ Hiển thị success message: "Tạo reminder thành công!"
- ✅ Auto navigate về Reminder List screen
- ✅ Reminder mới xuất hiện trong list

**Actual Result**: [To be filled during testing]

---

### TC-CR-002: Create Recurring Daily Reminder
**Test Scenario**: TS-CR-002  
**Priority**: Critical  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Navigate to Create Reminder screen

**Test Steps**:
1. Enter Title: "Uống thuốc"
2. Enter Description: "Vitamin D mỗi sáng"
3. Select Type: "Lặp lại"
4. Select Calendar Type: "Dương lịch"
5. Select Recurrence: "Hàng ngày"
6. Select Time: "08:00"
7. Select Repeat Strategy: "Tự động lặp lại"
8. Set max_crp: 1
9. Click "Tạo" button

**Test Data**:
```json
{
  "title": "Uống thuốc",
  "description": "Vitamin D mỗi sáng",
  "type": "recurring",
  "calendar_type": "solar",
  "next_action_at": "2025-11-24T01:00:00Z",
  "recurrence_pattern": {
    "type": "daily",
    "interval": 1
  },
  "repeat_strategy": "none",
  "max_crp": 1,
  "crp_interval_sec": 0,
  "status": "active"
}
```

**Expected Result**:
- ✅ API call với đúng payload
- ✅ `trigger_time_of_day` KHÔNG có trong request (server tự sinh)
- ✅ Server response bao gồm `trigger_time_of_day: "08:00"`
- ✅ Success message hiển thị
- ✅ Navigate về list

**Actual Result**: [To be filled during testing]

---

### TC-CR-003: Create Recurring Weekly Reminder
**Test Scenario**: TS-CR-003  
**Priority**: High  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Navigate to Create Reminder screen

**Test Steps**:
1. Enter Title: "Họp team"
2. Select Type: "Lặp lại"
3. Select Recurrence: "Hàng tuần"
4. Select Day of Week: "Thứ 2" (Monday = 1)
5. Select Time: "09:00"
6. Set max_crp: 3
7. Set crp_interval_sec: 600 (10 phút)
8. Click "Tạo"

**Test Data**:
```json
{
  "title": "Họp team",
  "type": "recurring",
  "calendar_type": "solar",
  "next_action_at": "2025-11-24T02:00:00Z",
  "recurrence_pattern": {
    "type": "weekly",
    "interval": 1,
    "day_of_week": 1
  },
  "repeat_strategy": "none",
  "max_crp": 3,
  "crp_interval_sec": 600,
  "status": "active"
}
```

**Expected Result**:
- ✅ API call thành công
- ✅ `recurrence_pattern.day_of_week = 1`
- ✅ Success message
- ✅ Navigate về list

**Actual Result**: [To be filled during testing]

---

### TC-CR-004: Create Recurring Monthly Reminder
**Test Scenario**: TS-CR-004  
**Priority**: High  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Navigate to Create Reminder screen

**Test Steps**:
1. Enter Title: "Trả tiền nhà"
2. Select Type: "Lặp lại"
3. Select Recurrence: "Hàng tháng"
4. Select Day of Month: "15"
5. Select Time: "10:00"
6. Click "Tạo"

**Test Data**:
```json
{
  "title": "Trả tiền nhà",
  "type": "recurring",
  "calendar_type": "solar",
  "next_action_at": "2025-12-15T03:00:00Z",
  "recurrence_pattern": {
    "type": "monthly",
    "interval": 1,
    "day_of_month": 15
  },
  "repeat_strategy": "none",
  "max_crp": 1,
  "crp_interval_sec": 0,
  "status": "active"
}
```

**Expected Result**:
- ✅ API call thành công
- ✅ `recurrence_pattern.day_of_month = 15`
- ✅ Reminder được tạo cho ngày 15 tháng sau

**Actual Result**: [To be filled during testing]

---

### TC-CR-005: Create Interval Reminder (Every 2 Hours)
**Test Scenario**: TS-CR-005  
**Priority**: Medium  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Navigate to Create Reminder screen

**Test Steps**:
1. Enter Title: "Nghỉ giải lao"
2. Select Type: "Lặp lại"
3. Select Recurrence: "Theo khoảng thời gian"
4. Enter Interval: "7200" giây (2 hours)
5. Click "Tạo"

**Test Data**:
```json
{
  "title": "Nghỉ giải lao",
  "type": "recurring",
  "calendar_type": "solar",
  "next_action_at": "2025-11-23T00:00:00Z",
  "recurrence_pattern": {
    "type": "interval_seconds",
    "interval_seconds": 7200
  },
  "repeat_strategy": "none",
  "max_crp": 1,
  "crp_interval_sec": 0,
  "status": "active"
}
```

**Expected Result**:
- ✅ API call thành công
- ✅ `recurrence_pattern.interval_seconds = 7200`

**Actual Result**: [To be filled during testing]

---

### TC-CR-006: Create Lunar Calendar Reminder
**Test Scenario**: TS-CR-006  
**Priority**: Medium  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Navigate to Create Reminder screen

**Test Steps**:
1. Enter Title: "Giỗ tổ"
2. Select Type: "Lặp lại"
3. Select Calendar Type: "Âm lịch"
4. Select Recurrence: "Hàng tháng"
5. Select Day: "10"
6. Select Time: "08:00"
7. Click "Tạo"

**Test Data**:
```json
{
  "title": "Giỗ tổ",
  "type": "recurring",
  "calendar_type": "lunar",
  "next_action_at": "2025-12-10T01:00:00Z",
  "recurrence_pattern": {
    "type": "monthly",
    "interval": 1,
    "day_of_month": 10
  },
  "repeat_strategy": "none",
  "max_crp": 1,
  "crp_interval_sec": 0,
  "status": "active"
}
```

**Expected Result**:
- ✅ API call thành công
- ✅ `calendar_type = "lunar"`

**Actual Result**: [To be filled during testing]

---

### TC-CR-007: Validation - Missing Title
**Test Scenario**: TS-CR-007  
**Priority**: High  
**Test Type**: Negative

**Preconditions**:
- User đang ở Create Reminder screen

**Test Steps**:
1. Leave Title field empty
2. Fill description: "Test description"
3. Select type, date, time
4. Click "Tạo"

**Expected Result**:
- ✅ Form validation error
- ✅ Error message: "Tiêu đề là bắt buộc"
- ✅ Focus trở về Title field
- ✅ KHÔNG gọi API

**Actual Result**: [To be filled during testing]

---

### TC-CR-008: Validation - Missing Type
**Test Scenario**: TS-CR-007  
**Priority**: High  
**Test Type**: Negative

**Preconditions**:
- User đang ở Create Reminder screen

**Test Steps**:
1. Enter Title: "Test"
2. KHÔNG chọn Type (one-time/recurring)
3. Click "Tạo"

**Expected Result**:
- ✅ Validation error: "Vui lòng chọn loại reminder"
- ✅ KHÔNG gọi API

**Actual Result**: [To be filled during testing]

---

### TC-CR-009: Validation - Missing Next Action At
**Test Scenario**: TS-CR-007  
**Priority**: High  
**Test Type**: Negative

**Preconditions**:
- User đang ở Create Reminder screen

**Test Steps**:
1. Enter Title: "Test"
2. Select Type: "Một lần"
3. KHÔNG chọn date/time
4. Click "Tạo"

**Expected Result**:
- ✅ Validation error: "Vui lòng chọn thời gian"
- ✅ KHÔNG gọi API

**Actual Result**: [To be filled during testing]

---

### TC-CR-010: Validation - Past Date for One-Time
**Test Scenario**: TS-CR-008  
**Priority**: High  
**Test Type**: Negative

**Preconditions**:
- User đang ở Create Reminder screen
- Hôm nay: 2025-11-23

**Test Steps**:
1. Enter Title: "Test"
2. Select Type: "Một lần"
3. Select Date: "2025-11-22" (yesterday)
4. Click "Tạo"

**Expected Result**:
- ✅ Validation error: "Thời gian phải ở tương lai"
- ✅ KHÔNG gọi API

**Actual Result**: [To be filled during testing]

---

### TC-CR-011: Validation - CRP Configuration Invalid
**Test Scenario**: TS-CR-009  
**Priority**: Medium  
**Test Type**: Negative

**Preconditions**:
- User đang ở Create Reminder screen

**Test Steps**:
1. Enter Title: "Test"
2. Select Type: "Một lần"
3. Set max_crp: 3
4. Set crp_interval_sec: 0 (invalid)
5. Click "Tạo"

**Expected Result**:
- ✅ Validation error: "Nếu có retry, phải set khoảng thời gian > 0"
- ✅ KHÔNG gọi API

**Actual Result**: [To be filled during testing]

---

### TC-CR-012: Validation - Recurring Without Pattern
**Test Scenario**: TS-CR-007  
**Priority**: High  
**Test Type**: Negative

**Preconditions**:
- User đang ở Create Reminder screen

**Test Steps**:
1. Enter Title: "Test"
2. Select Type: "Lặp lại"
3. KHÔNG chọn recurrence pattern (daily/weekly/etc)
4. Click "Tạo"

**Expected Result**:
- ✅ Validation error: "Vui lòng chọn kiểu lặp lại"
- ✅ KHÔNG gọi API

**Actual Result**: [To be filled during testing]

---

### TC-CR-013: Loading State During Submit
**Test Scenario**: TS-CR-010  
**Priority**: Medium  
**Test Type**: Functional

**Preconditions**:
- User đã fill form hợp lệ
- Simulate slow network

**Test Steps**:
1. Fill form với valid data
2. Click "Tạo"
3. Observe UI during API call

**Expected Result**:
- ✅ "Tạo" button disabled
- ✅ Loading indicator hiển thị
- ✅ User không thể click lại button (prevent double submit)
- ✅ Sau khi API response → loading biến mất

**Actual Result**: [To be filled during testing]

---

### TC-CR-014: Success State After Create
**Test Scenario**: TS-CR-011  
**Priority**: High  
**Test Type**: Functional

**Preconditions**:
- User đã fill form hợp lệ
- API sẵn sàng

**Test Steps**:
1. Fill form với valid data
2. Click "Tạo"
3. API trả về success (200 OK)
4. Observe UI

**Expected Result**:
- ✅ Hiển thị success Snackbar: "Tạo reminder thành công!"
- ✅ Sau 1-2 giây, auto navigate về Reminder List
- ✅ Reminder mới xuất hiện trong list

**Actual Result**: [To be filled during testing]

---

### TC-CR-015: Error State - API Error
**Test Scenario**: TS-CR-012  
**Priority**: High  
**Test Type**: Negative

**Preconditions**:
- User đã fill form
- Mock API trả về 400 Bad Request

**Test Steps**:
1. Fill form
2. Click "Tạo"
3. API trả về error:
```json
{
  "success": false,
  "message": "Validation failed: title is required"
}
```

**Expected Result**:
- ✅ Error Snackbar: "Validation failed: title is required"
- ✅ Không navigate, user còn ở Create screen
- ✅ Form data vẫn giữ nguyên
- ✅ User có thể sửa và submit lại

**Actual Result**: [To be filled during testing]

---

### TC-CR-016: Error State - Network Error
**Test Scenario**: TS-CR-012  
**Priority**: High  
**Test Type**: Negative

**Preconditions**:
- User đã fill form
- Turn off internet

**Test Steps**:
1. Fill form
2. Turn off WiFi/Mobile Data
3. Click "Tạo"

**Expected Result**:
- ✅ Error Snackbar: "Không có kết nối mạng. Vui lòng kiểm tra và thử lại."
- ✅ Không navigate
- ✅ Form data giữ nguyên

**Actual Result**: [To be filled during testing]

---

### TC-CR-017: DatePicker - Valid Date Selection
**Test Scenario**: TS-CR-013  
**Priority**: Medium  
**Test Type**: Functional

**Preconditions**:
- User đang ở Create Reminder screen

**Test Steps**:
1. Click vào Date field
2. DatePicker dialog xuất hiện
3. Select date: "2025-11-30"
4. Click "OK"

**Expected Result**:
- ✅ DatePicker hiển thị đúng
- ✅ Date field update: "30/11/2025"
- ✅ DatePicker chỉ cho chọn ngày >= hôm nay

**Actual Result**: [To be filled during testing]

---

### TC-CR-018: TimePicker - Valid Time Selection
**Test Scenario**: TS-CR-014  
**Priority**: Medium  
**Test Type**: Functional

**Preconditions**:
- User đang ở Create Reminder screen

**Test Steps**:
1. Click vào Time field
2. TimePicker dialog xuất hiện
3. Select time: "14:30"
4. Click "OK"

**Expected Result**:
- ✅ TimePicker hiển thị đúng
- ✅ Time field update: "14:30" hoặc "2:30 PM"

**Actual Result**: [To be filled during testing]

---

### TC-CR-019: Repeat Strategy Selection
**Test Scenario**: TS-CR-015  
**Priority**: Medium  
**Test Type**: Functional

**Preconditions**:
- User đã chọn Type: "Lặp lại"

**Test Steps**:
1. Observe Repeat Strategy options
2. Select "Chờ hoàn thành"
3. Verify selection

**Expected Result**:
- ✅ Repeat Strategy field hiển thị
- ✅ Options:
  - "Tự động lặp lại" (none)
  - "Chờ hoàn thành" (crp_until_complete)
- ✅ Selection được hiển thị đúng

**Actual Result**: [To be filled during testing]

---

### TC-CR-020: Calendar Type Toggle
**Test Scenario**: TS-CR-016  
**Priority**: Low  
**Test Type**: Functional

**Preconditions**:
- User đang ở Create Reminder screen

**Test Steps**:
1. Default calendar type: "Dương lịch"
2. Toggle to "Âm lịch"
3. Verify change

**Expected Result**:
- ✅ Default: "Dương lịch" selected
- ✅ Toggle to "Âm lịch" → selection update
- ✅ UI hiển thị rõ ràng option nào đang chọn

**Actual Result**: [To be filled during testing]

---

### TC-CR-021: Dynamic Form - One-Time vs Recurring
**Test Scenario**: TS-CR-017  
**Priority**: High  
**Test Type**: Functional

**Preconditions**:
- User đang ở Create Reminder screen

**Test Steps**:
1. Default or select Type: "Một lần"
2. Observe form fields
3. Change Type to: "Lặp lại"
4. Observe form changes

**Expected Result**:
- ✅ Type = "Một lần":
  - Ẩn Recurrence Pattern fields
  - Ẩn Repeat Strategy
- ✅ Type = "Lặp lại":
  - Hiển thị Recurrence Pattern (Daily/Weekly/Monthly/Interval)
  - Hiển thị Repeat Strategy

**Actual Result**: [To be filled during testing]

---

### TC-CR-022: Cancel Button
**Test Scenario**: TS-CR-018  
**Priority**: Medium  
**Test Type**: Functional

**Preconditions**:
- User đã nhập: Title: "Test", Description: "Test desc"

**Test Steps**:
1. Click "Hủy" button
2. Observe confirmation dialog
3. Click "Xác nhận" trong dialog

**Expected Result**:
- ✅ Confirmation dialog: "Bạn có chắc muốn hủy?"
- ✅ Options: "Xác nhận" / "Tiếp tục chỉnh sửa"
- ✅ Click "Xác nhận" → navigate back
- ✅ Data không được save

**Actual Result**: [To be filled during testing]

---

### TC-CR-023: Back Button with Unsaved Data
**Test Scenario**: TS-CR-018  
**Priority**: Medium  
**Test Type**: Functional

**Preconditions**:
- User đã nhập data vào form

**Test Steps**:
1. Press system back button (Android)
2. Observe confirmation dialog

**Expected Result**:
- ✅ Confirmation dialog xuất hiện
- ✅ Click "Xác nhận" → back
- ✅ Click "Tiếp tục chỉnh sửa" → stay on screen

**Actual Result**: [To be filled during testing]

---

### TC-CR-024: Character Limit - Title
**Test Scenario**: TS-CR-007  
**Priority**: Low  
**Test Type**: Functional

**Preconditions**:
- User đang ở Create Reminder screen

**Test Steps**:
1. Enter Title: 200 characters
2. Try to enter more

**Expected Result**:
- ✅ Max length: 200 characters (hoặc theo design)
- ✅ Character counter hiển thị: "150/200"
- ✅ Không cho nhập quá max length

**Actual Result**: [To be filled during testing]

---

### TC-CR-025: Character Limit - Description
**Test Scenario**: TS-CR-007  
**Priority**: Low  
**Test Type**: Functional

**Preconditions**:
- User đang ở Create Reminder screen

**Test Steps**:
1. Enter Description: 500 characters
2. Try to enter more

**Expected Result**:
- ✅ Max length: 500 characters (hoặc theo design)
- ✅ Không cho nhập quá

**Actual Result**: [To be filled during testing]

---

### TC-CR-026: UI Match Design Specification
**Test Scenario**: TS-CR-001  
**Priority**: High  
**Test Type**: Visual

**Preconditions**:
- User đang ở Create Reminder screen

**Test Steps**:
1. Open Create Reminder screen
2. Compare với design: [create_reminder.png](file:///d:/PROJECT/NHAC_NHO_GROUP/AndroidReminaq6/docs/create_reminder.png)
3. Verify all UI elements

**Expected Result**:
- ✅ Layout match design
- ✅ Colors, fonts, spacing đúng spec
- ✅ All fields hiển thị đúng vị trí
- ✅ Buttons: "Tạo", "Hủy" đúng style

**Actual Result**: [To be filled during testing]

---

### TC-CR-027: Keyboard Behavior
**Test Scenario**: TS-CR-001  
**Priority**: Low  
**Test Type**: Functional

**Preconditions**:
- User đang ở Create Reminder screen

**Test Steps**:
1. Click vào Title field
2. Keyboard hiển thị
3. Enter text
4. Click outside field

**Expected Result**:
- ✅ Keyboard hiển thị khi focus vào text field
- ✅ Keyboard ẩn khi click outside hoặc tap "Done"
- ✅ Screen scroll để field không bị che bởi keyboard

**Actual Result**: [To be filled during testing]

---

### TC-CR-028: CRP Fields Visibility Based on max_crp
**Test Scenario**: TS-CR-017  
**Priority**: Medium  
**Test Type**: Functional

**Preconditions**:
- User đang ở Create Reminder screen

**Test Steps**:
1. Observe initial state (max_crp = 0 by default)
2. Change max_crp to 3
3. Observe CRP Interval field

**Expected Result**:
- ✅ max_crp = 0 → CRP Interval field disabled hoặc ẩn
- ✅ max_crp > 0 → CRP Interval field hiển thị và bắt buộc

**Actual Result**: [To be filled during testing]

---

---

## Integration Test Scenarios
### TC-CR-029: Negative - Client Sends trigger_time_of_day (Recurring)
**Test Scenario**: TS-CR-002  
**Priority**: High  
**Test Type**: Negative

**Preconditions**:
- User đã login
- Create Recurring Daily reminder

**Test Steps**:
1. Fill form với Recurrence = Daily, Time = 08:00
2. Client cố tình gửi thêm `recurrence_pattern.trigger_time_of_day = "08:00"` trong payload
3. Submit

**Expected Result**:
- ✅ Server trả về 400 Bad Request
- ✅ Error: "client không cần gửi trigger_time_of_day"
- ✅ App hiển thị lỗi và không tạo reminder

**Actual Result**: [To be filled during testing]

---

### TC-CR-030: Create Recurring with repeat_strategy = crp_until_complete
**Test Scenario**: TS-CR-002  
**Priority**: High  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Navigate to Create Reminder screen

**Test Steps**:
1. Enter Title: "Theo dõi công việc quan trọng"
2. Type: "Lặp lại"
3. Recurrence: "Hàng ngày" (08:00)
4. Repeat Strategy: "Chờ hoàn thành" (crp_until_complete)
5. Submit

**Test Data**:
```json
{
  "title": "Theo dõi công việc quan trọng",
  "type": "recurring",
  "calendar_type": "solar",
  "next_action_at": "2025-11-24T01:00:00Z",
  "recurrence_pattern": { "type": "daily", "interval": 1 },
  "repeat_strategy": "crp_until_complete",
  "max_crp": 3,
  "crp_interval_sec": 300,
  "status": "active"
}
```

**Expected Result**:
- ✅ API tạo thành công, trả về `repeat_strategy = crp_until_complete`
- ✅ Response bao gồm `trigger_time_of_day: "08:00"`

**Actual Result**: [To be filled during testing]

---

### TC-CR-031: Create Lunar Reminder - Last Day of Lunar Month
**Test Scenario**: TS-CR-006  
**Priority**: Medium  
**Test Type**: Functional

**Preconditions**:
- User đã login
- Navigate to Create Reminder screen

**Test Steps**:
1. Enter Title: "Cuối tháng âm"
2. Type: "Lặp lại"
3. Calendar: "Âm lịch"
4. Recurrence: "Cuối tháng âm"
5. Submit

**Test Data**:
```json
{
  "title": "Cuối tháng âm",
  "type": "recurring",
  "calendar_type": "lunar",
  "next_action_at": "2025-11-30T20:00:00Z",
  "recurrence_pattern": { "type": "lunar_last_day_of_month" },
  "repeat_strategy": "none",
  "max_crp": 1,
  "crp_interval_sec": 0,
  "status": "active"
}
```

**Expected Result**:
- ✅ API tạo thành công với `recurrence_pattern.type = "lunar_last_day_of_month"`

**Actual Result**: [To be filled during testing]

---

### TC-CR-032: QA Fast Test - for_test seconds
**Test Scenario**: TS-CR-001  
**Priority**: Low  
**Test Type**: Functional (QA-only)

**Preconditions**:
- User đã login
- Server hỗ trợ trường `for_test` (giây)

**Test Steps**:
1. Fill form One-time reminder
2. Set `for_test = 10`
3. Submit

**Expected Result**:
- ✅ Server set `next_action_at = now + 10s`
- ✅ Reminder xuất hiện trong list sau khoảng 10s

**Actual Result**: [To be filled during testing]

---

### ITS-001: End-to-End Flow - Create and View Reminder
**Priority**: Critical  
**Test Type**: Integration

**Test Steps**:
1. Login với user: `test_user@example.com`
2. Navigate to Reminder List (initially empty)
3. Click "Tạo Reminder"
4. Fill form:
   - Title: "Integration Test Reminder"
   - Type: "One-time"
   - Date: Tomorrow
   - Time: 10:00 AM
5. Click "Tạo"
6. Verify success message
7. Verify navigate back to Reminder List
8. Verify reminder xuất hiện trong list
9. Click vào reminder
10. Verify detail screen hiển thị đúng data

**Expected Result**:
- ✅ Toàn bộ flow hoạt động smooth
- ✅ Data consistent giữa Create → List → Detail

**Test Data**: 
```json
{
  "user": {
    "email": "test_user@example.com",
    "password": "123123123"
  },
  "reminder": {
    "title": "Integration Test Reminder",
    "type": "one_time",
    "next_action_at": "2025-11-24T03:00:00Z"
  }
}
```

---

### ITS-002: Multi-User Isolation
**Priority**: High  
**Test Type**: Integration

**Test Steps**:
1. Login User A: `user_a@test.com`
2. Create 2 reminders for User A
3. Logout
4. Login User B: `user_b@test.com`
5. Navigate to Reminder List
6. Verify chỉ thấy reminders của User B (empty nếu chưa có)
7. Create 1 reminder for User B
8. Logout
9. Login lại User A
10. Verify User A vẫn thấy 2 reminders của mình (KHÔNG thấy reminder của User B)

**Expected Result**:
- ✅ Data isolation hoàn toàn giữa users
- ✅ User A không thể access reminders của User B

---

### ITS-003: Offline Create and Sync
**Priority**: Medium  
**Test Type**: Integration

**Test Steps**:
1. Login
2. Navigate to Create Reminder
3. Turn off internet
4. Fill form và click "Tạo"
5. Verify error message
6. Turn on internet
7. Click "Thử lại" (nếu có) hoặc submit again
8. Verify thành công

**Expected Result**:
- ✅ Offline: Error message rõ ràng, data giữ nguyên
- ✅ Online again: Có thể submit thành công

---

### ITS-004: Rapid Create Multiple Reminders
**Priority**: Medium  
**Test Type**: Stress

**Test Steps**:
1. Login
2. Create 10 reminders liên tục (khác title)
3. Verify tất cả được tạo thành công
4. Navigate to Reminder List
5. Verify 10 reminders xuất hiện

**Expected Result**:
- ✅ Tất cả 10 reminders được tạo
- ✅ Không có duplicate, missing data

---

---

## Non-Functional Test Cases

### NF-001: Performance - List Load Time
**Test Type**: Performance  
**Priority**: Medium

**Test Steps**:
1. Setup database với 100 reminders
2. Login
3. Navigate to Reminder List
4. Measure load time

**Expected Result**:
- ✅ Load time < 2 seconds (3G network)
- ✅ Load time < 1 second (WiFi)

---

### NF-002: Performance - Create Reminder Response Time
**Test Type**: Performance  
**Priority**: Medium

**Test Steps**:
1. Fill Create Reminder form
2. Click "Tạo"
3. Measure time từ click đến success message

**Expected Result**:
- ✅ Response time < 3 seconds (3G network)
- ✅ Response time < 1 second (WiFi)

---

### NF-003: UI Responsiveness - Multiple Screen Sizes
**Test Type**: UI/UX  
**Priority**: High

**Test Devices**:
- Small phone: 5" screen
- Medium phone: 6" screen
- Large phone: 6.5" screen
- Tablet: 10" screen

**Expected Result**:
- ✅ UI responsive trên tất cả devices
- ✅ No text cutoff, no overlapping elements
- ✅ Touch targets đủ lớn (min 48dp)

---

### NF-004: Accessibility - Screen Reader
**Test Type**: Accessibility  
**Priority**: Low

**Test Steps**:
1. Enable TalkBack (Android)
2. Navigate Reminder List
3. Navigate Create Reminder
4. Use screen reader để fill form

**Expected Result**:
- ✅ Tất cả elements có content description
- ✅ Screen reader đọc được labels, values
- ✅ Form có thể điền bằng screen reader

---

### NF-005: Security - Token Expiration
**Test Type**: Security  
**Priority**: High

**Test Steps**:
1. Login
2. Wait for token expiration (thay đổi server config để token expire nhanh)
3. Try to create reminder
4. Verify behavior

**Expected Result**:
- ✅ API trả về 401 Unauthorized
- ✅ App tự động redirect to Login
- ✅ Sau khi login lại → có thể create reminder

---

### NF-006: Data Consistency - Timezone Handling
**Test Type**: Functional  
**Priority**: Critical

**Test Steps**:
1. Set device timezone: Asia/Ho_Chi_Minh (UTC+7)
2. Login
3. Create reminder: Date: 2025-11-24, Time: 15:00
4. Verify API payload: `next_action_at` should be UTC
5. Navigate to Reminder List
6. Verify reminder hiển thị: "24/11/2025 15:00" (Local time)

**Expected Result**:
- ✅ Client gửi UTC time to server
- ✅ Server lưu UTC time
- ✅ Client hiển thị Local time cho user
- ✅ Conversion chính xác

---

---

## Test Data

### Valid Test Users

```json
{
  "users": [
    {
      "email": "test_user_1@example.com",
      "password": "Password123!",
      "name": "Test User 1"
    },
    {
      "email": "test_user_2@example.com",
      "password": "Password123!",
      "name": "Test User 2"
    }
  ]
}
```

### Sample Reminders for Testing

```json
{
  "reminders": [
    {
      "title": "Daily Vitamin",
      "description": "Take vitamin D supplement",
      "type": "recurring",
      "calendar_type": "solar",
      "recurrence_pattern": {
        "type": "daily",
        "interval": 1
      },
      "next_action_at": "2025-11-24T01:00:00Z",
      "repeat_strategy": "none",
      "max_crp": 1,
      "crp_interval_sec": 0
    },
    {
      "title": "Weekly Team Meeting",
      "description": "Monday standup at 9 AM",
      "type": "recurring",
      "calendar_type": "solar",
      "recurrence_pattern": {
        "type": "weekly",
        "interval": 1,
        "day_of_week": 1
      },
      "next_action_at": "2025-11-24T02:00:00Z",
      "repeat_strategy": "none",
      "max_crp": 3,
      "crp_interval_sec": 600
    },
    {
      "title": "Doctor Appointment",
      "description": "Annual checkup",
      "type": "one_time",
      "calendar_type": "solar",
      "next_action_at": "2025-11-25T07:30:00Z",
      "max_crp": 0,
      "crp_interval_sec": 0
    }
  ]
}
```

---

## Defect Severity Classification

### Critical (P1)
- App crash khi create reminder
- Data loss (reminder tạo nhưng không lưu)
- Security issues (authorization bypass)
- Unable to login/logout

### High (P2)
- API errors không được handle
- Validation không hoạt động
- Navigation issues (stuck on screen)
- Time conversion sai (UTC/Local)

### Medium (P3)
- UI không match design
- Loading indicator không hiển thị
- Minor validation issues
- Performance issues (slow but functional)

### Low (P4)
- Cosmetic UI issues
- Typos trong text
- Missing animations
- Minor UX improvements

---

## Test Execution Checklist

- [ ] Tất cả test cases cho Reminder List đã execute
- [ ] Tất cả test cases cho Create Reminder đã execute
- [ ] Integration tests đã pass
- [ ] Non-functional tests đã execute
- [ ] Defects đã được log vào tracking system
- [ ] Critical defects đã được fix và retest
- [ ] Test report đã được generate

---

**Document End**

*Prepared by: QA Team*  
*Version: 1.0*  
*Date: 2025-11-23*
