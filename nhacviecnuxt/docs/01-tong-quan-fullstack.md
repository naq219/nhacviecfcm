# 01 — Tổng quan kiến trúc fullstack Nuxt

> Đọc để hiểu **bức tranh lớn** và biết tại sao chọn từng công nghệ. Chưa cần chạy lệnh gì.

---

## 1. Bối cảnh

Trước giờ dự án chỉ có **backend** (Go + PocketBase + FCM), chưa có web frontend. Giờ mục tiêu là:

- **Làm lại cả API lẫn frontend bằng Nuxt** trong 1 dự án duy nhất.
- Deploy lên Cloudflare Pages tại **https://sennote.pages.dev**.
- Backend Go cũ bị **thay thế**, chỉ giữ lại để tham khảo logic + dữ liệu cũ. Đường dẫn project cũ: **`E:\PROJECT\nhacviecfcm`**.

> ⚠️ **Đừng tin `docs/DATABASE_SCHEMA.md` và `v2docs/` của project Go** — cả hai đã lỗi thời
> (thiếu 3 cột thật, mô tả code đã bị comment out). Nguồn tham khảo đúng:
> `migrations/1671631110_init_schema.go` (schema) và `internal/services/calNextTime.go` (tính lịch).

## 2. Kiến trúc mới (1 dự án = cả API + web)

```
┌─────────────────────────────────────────────────────────────┐
│ Trình duyệt                                                │
│   mở https://sennote.pages.dev                             │
└──────────────┬──────────────────────────────────────────────┘
               │ HTTP (cùng origin — không còn CORS/mixed-content)
               ▼
┌─────────────────────────────────────────────────────────────┐
│ Cloudflare Pages — chạy toàn bộ app Nuxt                    │
│                                                             │
│  app/          → frontend Vue (SSR)                         │
│  server/api/   → REST API (đăng ký, login, CRUD reminder)   │
│  server/tasks/ → cron: cứ mỗi phút kiểm tra reminder đến hạn│
│  server/utils/ → db (Turso), fcm, calendar, schedule        │
└───────┬───────────────────────────────┬─────────────────────┘
        │ SQL qua HTTP                  │ HTTPS (HTTP v1 API)
        ▼                               ▼
┌──────────────────┐          ┌──────────────────────┐
│ Turso (libSQL)   │          │ Firebase FCM         │
│ users, reminders │          │ push về mobile app   │
└──────────────────┘          └──────────────────────┘
```

**Điểm khác biệt lớn nhất so với mô hình cũ:** trước đây web (nếu có) phải gọi API của 1 backend khác qua mạng, phải lo CORS, mixed-content, token... Giờ API nằm **cùng chỗ** với web, cùng origin, đơn giản hơn nhiều.

## 3. Vì sao chọn Nuxt fullstack (thay vì giữ Go + thêm web)?

| Lợi ích | Giải thích |
|---------|-----------|
| 1 ngôn ngữ | Chỉ cần TypeScript cho cả API lẫn UI, không còn song song Go + JS |
| 1 lần deploy | Một project, một lệnh build, một nơi host — không phải quản lý 2 service |
| Types đi xuyên suốt | `shared/types` dùng chung cho cả server và client, API trả gì UI nhận đúng đó |
| Bớt vấn đề hạ tầng | Không lo CORS, không lo token khác origin, không lo cookie cross-site |

## 4. Các quyết định kỹ thuật và LÝ DO

| Quyết định | Lý do |
|------------|-------|
| **Nuxt (Nitro server)** | Một framework làm cả API lẫn UI; routing, auto-import, typed đi kèm sẵn |
| **Turso (libSQL)** | PocketBase/SQLite **không chạy được** trên Cloudflare Workers (không có filesystem). Turso là SQLite trên cloud, gọi qua HTTP, vẫn là SQL quen thuộc |
| **nuxt-auth-utils** | Auth kiểu session (cookie mã hoá = JWT), chạy tốt trên Workers. ⚠️ Chỉ dùng phần **session** — `hashPassword`/`verifyPassword` của nó dùng scrypt và **hỏng trên Workers**, đã thay bằng Web Crypto PBKDF2 |
| **Nitro tasks + Cron Triggers** | Thay thế worker Go chạy nền: Cloudflare tự bắn job mỗi phút, không cần server riêng |
| **FCM HTTP v1 API** | Thư viện `firebase-admin` không chạy trên Workers (cần Node API) → gọi thẳng REST API của Firebase bằng Web Crypto |
| **SSR bật** (mặc định) | auth-utils cần server chạy; SSR cũng giúp trang tải nhanh hơn. KHÔNG dùng `nuxt generate` |

## 5. So sánh: cái cũ (Go) → cái mới (Nuxt)

| Phần ở backend Go cũ | Chuyển thành (ở dự án này) |
|----------------------|---------------------------|
| Routes trong `cmd/server/main.go` | `server/api/**` (1 file = 1 route) |
| `internal/repository/*` (SQL) | `server/utils/db.ts` + service `server/utils/*.ts` |
| `internal/services/*` (business logic) | `server/utils/reminders.ts`, `schedule.ts`, `calendar.ts` |
| `internal/worker/*` (3 worker chạy nền, mỗi `WORKER_INTERVAL` giây — **mặc định 10, đang = 5**) | `server/tasks/reminders-check.ts` + `scheduledTasks` (1 phút) |
| `internal/services/fcm_service.go` (gửi push) | `server/utils/fcm.ts` (HTTP v1) |
| Bảng `system_status` (kill-switch) | bảng `system_status` + `server/utils/system-status.ts` |
| Collection `musers` (auth) | bảng `users` + nuxt-auth-utils |
| Collection `reminders` | bảng `reminders` (xem doc 03 — ⚠️ có 3 cột hay bị sót) |
| Lunar calendar lib | `server/utils/calendar.ts` (port **thuật toán** `lunar_calendar.go`, không phải bảng dữ liệu) |

> ⚠️ `internal/services/fcmutils` của Go **là code chết** (gửi qua proxy `localhost:404`, nhúng sẵn
> service account base64 trong source). Đừng port — dùng `fcm_service.go` làm chuẩn.

### Cần port logic nào từ Go?

Hai phần "khó" nhất của app nhắc nhở nằm ở backend cũ, phải **viết lại bằng TypeScript** (để nguyên hàm thuần, dễ test):

1. **Âm lịch** → `server/utils/calendar.ts`.
   Port **thuật toán** trong `E:\PROJECT\nhacviecfcm\internal\services\lunar_calendar.go`.
   🚨 **Không copy `lunar_data.go`** — bảng `SolarToLunarMap` chỉ phủ **2024-01-01 → 2025-12-31** (731 ngày),
   mọi reminder âm lịch sẽ lỗi từ 2026.
2. **Tính lịch lặp FRP/CRP** (daily/weekly/monthly/lunar, retry) → `server/utils/schedule.ts`.
   Đặc tả chi tiết nằm ở **[doc 04b](./04b-tinh-lich-frp-crp.md)**, port từ
   `E:\PROJECT\nhacviecfcm\internal\services\calNextTime.go`.

Đây là phần tốn công nhất — nhưng một khi port xong, phần còn lại (CRUD + UI) chỉ là việc làm theo docs 03–05.

> ℹ️ Go có **2 engine tính lịch mâu thuẫn nhau** (`calNextTime.go` và `schedule_calculator.go`).
> Đã chọn `calNextTime.go` làm chuẩn. Lý do + hệ quả: [doc 04b mục 1](./04b-tinh-lich-frp-crp.md#1-vì-sao-bỏ-schedule_calculatorgo-và-hệ-quả).

## 6. Di chuyển dữ liệu cũ (nếu cần)

Nếu backend cũ đang có dữ liệu thật cần giữ: export PocketBase sang SQL rồi import vào Turso (chú ý: field `user_id` relation, `recurrence_pattern` JSON). Việc này làm sau khi app mới chạy được — xem ghi chú cuối doc 03. Nếu chỉ đang phát triển, bỏ qua, tạo dữ liệu mới từ đầu.

## 7. Lộ trình làm việc

1. **Doc 02** — tạo dự án Nuxt, cài auth-utils + Turso, chạy dev server có `/api/hello`.
2. **Doc 03** — tạo schema + API auth + CRUD reminder, làm trang login + danh sách.
3. **Doc 04b** — đặc tả tính lịch: port `calNextTime.go` + 11 nhánh worker + âm lịch. *(làm trước doc 04)*
4. **Doc 04** — cài đặt cron + gửi FCM + kill-switch, ráp với service ở 04b.
5. **Doc 05** — deploy lên `sennote.pages.dev`.

→ Tiếp theo: [02 — Khởi tạo dự án](./02-khoi-tao-du-an.md)
