# RemiAq - Essential Documentation

---

## 📌 1. README.md

```markdown
# RemiAq - Smart Reminder & Lunar Calendar System

## 🎯 Overview

RemiAq là ứng dụng nhắc nhở thông minh với:
- ✅ **Firebase Cloud Messaging (FCM)** - Push notifications
- ✅ **Lịch Dương & Âm** - Support Solar & Lunar calendar
- ✅ **FRP+CRP Logic** - Smart retry system
- ✅ **Snooze** - Hoãn nhắc nhở
- ✅ **Interval-based** - Nhắc theo thời gian cố định (3 phút, 1 giờ, 20 ngày...)

## 🏗️ Architecture

```
┌─────────────┐
│   Mobile    │
│   Client    │
└──────┬──────┘
       │
       ├─ POST /api/reminders (Create)
       ├─ GET /api/reminders/mine (List)
       ├─ PUT /api/reminders/{id} (Update)
       ├─ POST /api/reminders/{id}/complete (Complete)
       ├─ POST /api/reminders/{id}/snooze (Snooze)
       │
       ▼
┌──────────────────┐     ┌──────────────┐
│  Backend API     │────▶│  PocketBase  │
│  (Go)            │     │  (Database)  │
└──────┬───────────┘     └──────────────┘
       │
       │ (Every 60s)
       ▼
┌──────────────────┐
│  Worker Process  │ ◀── Check next_action_at
│  (FCM Sender)    │     Send notifications
└──────────────────┘     Update DB
```

## 🚀 Quick Start

```bash
# Clone
git clone <repo>
cd remiaq

# Run
go run ./cmd/server serve

# Server at http://localhost:8090
```

## 📚 Documentation

- [API_DOCUMENTATION.md](./docs/API_DOCUMENTATION.md) - For Mobile Dev
- [WORKER_LOGIC.md](./docs/WORKER_LOGIC.md) - For Backend Dev
- [DATABASE_SCHEMA.md](./docs/DATABASE_SCHEMA.md) - DB Overview
- [Postman Collection](./v3_nhacviecfcm_postman.json) - API Testing

## 📦 Tech Stack

| Component | Tech |
|-----------|------|
| Backend | Go 1.21+ |
| Database | PocketBase (SQLite) |
| Auth | PocketBase Auth |
| Notifications | Firebase Cloud Messaging |
| Calendar | Custom Lunar Calendar Lib |

## 🔧 Environment Setup

```bash
# .env
POCKETBASE_ADDR=127.0.0.1:8090
FCM_CREDENTIALS=firebase-credentials.json
WORKER_INTERVAL=60
```

## 📝 API Quick Example

```bash
# Login
curl -X POST http://localhost:8090/api/collections/musers/auth-with-password \
  -H "Content-Type: application/json" \
  -d '{"identity":"test@example.com","password":"123123123"}'

# Create reminder (daily at 8 AM)
curl -X POST http://localhost:8090/api/reminders \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title":"Uống thuốc",
    "type":"recurring",
    "recurrence_pattern":{"type":"daily","trigger_time_of_day":"08:00"},
    "max_crp":1,
    "crp_interval_sec":0,
    "status":"active"
  }'
```

## 🎯 Key Concepts

### FRP (Father Recurrence Pattern)
Lặp lại theo **lịch** (calendar-based):
- Daily (mỗi ngày)
- Weekly (mỗi tuần)
- Monthly (mỗi tháng)
- Lunar last day (cuối tháng Âm)
- Interval seconds (mỗi X giây/phút/giờ/ngày)

### CRP (Child Repeat Pattern)
**Retry** nếu gửi thất bại:
- max_crp: Số lần retry tối đa
- crp_interval_sec: Khoảng cách giữa các retry

Ví dụ:
- max_crp=3, crp_interval_sec=300 → Gửi 3 lần, mỗi 5 phút

---

## 2. API_DOCUMENTATION.md


## 3. DATABASE_SCHEMA.md


## 4. WORKER_LOGIC.md



---

**Đây là 4 tài liệu CORE!** 📚

Các bạn có thể:
1. Copy markdown vào từng file docs/
2. Cập nhật thông tin (URLs, port, etc)
3. Thêm screenshots nếu cần

**Tài liệu phụ có thể viết sau** (DEPLOYMENT, ARCHITECTURE, etc)