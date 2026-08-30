# SenNote — Tài liệu xây dựng fullstack bằng Nuxt + Turso + Cloudflare

SenNote là ứng dụng nhắc nhở (reminder) được **viết lại hoàn toàn** bằng **Nuxt fullstack**:
- **API** (Nitro server) + **frontend** (Vue 3) nằm chung 1 dự án, chung 1 lần deploy.
- Database: **Turso (libSQL)** — SQLite chạy trên cloud, không cần máy chủ.
- Deploy lên **Cloudflare Pages** tại **https://sennote.pages.dev**.

Backend Go/PocketBase cũ — **đường dẫn: `E:\PROJECT\nhacviecfcm`** — chỉ còn dùng để **tham khảo logic** (âm lịch, tính lịch lặp FRP/CRP, schema) và dữ liệu cũ cần migrate. Xem thêm mục "Ngoài ra" bên dưới.

Tài liệu viết cho người **chưa có kinh nghiệm** với Nuxt, Turso lẫn Cloudflare — đọc lần lượt, làm theo từng bước là xong.

## Đọc theo thứ tự

| STT | File | Nội dung |
|-----|------|----------|
| 1 | [01-tong-quan-fullstack.md](./01-tong-quan-fullstack.md) | Kiến trúc mới, vì sao Nuxt fullstack, cái gì phải port từ Go |
| 2 | [02-khoi-tao-du-an.md](./02-khoi-tao-du-an.md) | Tạo dự án Nuxt, cài auth-utils + Turso, cấu hình, chạy thử |
| 3 | [03-database-va-api.md](./03-database-va-api.md) | Schema database, đăng ký/đăng nhập, CRUD reminder |
| 4 | **[04b-tinh-lich-frp-crp.md](./04b-tinh-lich-frp-crp.md)** | **Đặc tả tính lịch: port `calNextTime.go`, 11 nhánh worker, âm lịch, bộ giá trị chuẩn** |
| 5 | [04-lich-nhac-va-fcm.md](./04-lich-nhac-va-fcm.md) | Cài đặt job định kỳ, gửi push FCM, kill-switch |
| 6 | [05-deploy-cloudflare-pages.md](./05-deploy-cloudflare-pages.md) | Đưa lên sennote.pages.dev, cấu hình secret, kiểm tra |

> ⚠️ Làm **04b trước 04**. 04b là phần khó nhất toàn bộ app (tính lịch), 04 chỉ là phần ráp nối.

## Ngoài ra

- [../AGENTS.md](../AGENTS.md) — quy tắc code của dự án (cho người lẫn AI coding assistant). Đọc 1 lần trước khi viết code.
- **Project cũ để tham khảo — `E:\PROJECT\nhacviecfcm`**:
  - Schema DB: `migrations/1671631110_init_schema.go` ⚠️ **KHÔNG dùng `docs/DATABASE_SCHEMA.md`** (thiếu 3 cột thật)
  - Logic worker (FRP/CRP, khi nào gửi): `internal/worker/worker_loop_noUT.go`, `worker_loop_UT.go`, `worker_onetime_v2.go` ⚠️ **KHÔNG dùng `docs/WORKER_LOGIC.md`**
  - Tính lịch: `internal/services/calNextTime.go` + `calNextTime_test.go` ⚠️ **KHÔNG dùng `v2docs/caculate_next_recurring.md`**
  - Âm lịch: sinh bảng bằng `cmd/gen_lunar_table` (chạy thuật toán `lunar_calendar.go`) — 🚨 **KHÔNG copy `lunar_data.go`** (bảng chỉ 2024–2025)
  - Cấu hình chạy dev: `.env.example`, `cmd/server/main.go`

## Lệnh thường dùng

```bash
npm run dev        # dev server http://localhost:3000
npm run typecheck  # vue-tsc
npm test           # 38 test giá trị chuẩn cho logic tính lịch
npm run build      # build ra dist/ (preset cloudflare-pages)
```

Gia hạn bảng âm lịch (hiện phủ 2026-08 → 2029-08), chạy trong project Go:

```bash
go run ./cmd/gen_lunar_table -from 2029-01-01 -to 2032-12-31 \
  -out "E:\PROJECT\nhacviecfcm\nhacviecnuxt\server\utils\lunar-table.ts"
```

> ℹ️ 3 tài liệu Go ở trên đã **lỗi thời** (mô tả code đã bị comment out / thiếu cột).
> Bảng đối chiếu đầy đủ: [04b mục 0](./04b-tinh-lich-frp-crp.md#0-nguồn-sự-thật--đọc-cái-này-đừng-đọc-mấy-file-kia).

## Bản đồ nhanh

```
Trình duyệt → https://sennote.pages.dev
                │  Nuxt chạy trên Cloudflare Pages (SSR + API)
                ├─ app/        (frontend Vue)
                ├─ server/api  (REST API)
                ├─ server/tasks (cron kiểm tra reminder)
                ▼
          Turso (SQLite cloud)          FCM HTTP v1 (push về điện thoại)
```
