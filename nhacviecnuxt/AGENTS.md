# AGENTS.md

Web **SenNote** (https://sennote.pages.dev): ứng dụng **fullstack Nuxt** (Vue 3 + Nitro server) deploy lên Cloudflare Pages. Toàn bộ API + frontend nằm trong dự án NÀY — **thay thế hoàn toàn** backend Go/PocketBase cũ, không còn phụ thuộc vào nó. Database dùng **Turso (libSQL)**. UI và comment trong code dùng tiếng Việt.

**Project cũ để tham khảo logic + migrate dữ liệu:** `E:\PROJECT\nhacviecfcm` (Go + PocketBase + FCM). Chỉ đọc, KHÔNG sửa.

**Nguồn tham khảo ĐÚNG** (đã kiểm chứng):

| Cần | Đọc file này |
|---|---|
| Schema DB | `migrations/1671631110_init_schema.go` — ⚠️ **KHÔNG đọc `docs/DATABASE_SCHEMA.md`** (thiếu `origin_time`, `is_sended_one_time`, `tag`) |
| Tính lịch lặp | `internal/services/calNextTime.go` (+ `calNextTime_test.go` = bộ giá trị chuẩn) |
| 11 nhánh worker | `internal/worker/worker_loop_noUT.go`, `worker_loop_UT.go`, `worker_onetime_v2.go` |
| Bấm Hoàn thành | `internal/services/reminder_service.go` → `OnUserComplete` |
| Âm lịch | `internal/services/lunar_calendar.go` (thuật toán) — 🚨 **KHÔNG copy `lunar_data.go`** (bảng chỉ phủ 2024–2025) |

**Tài liệu project Go đã LỖI THỜI — KHÔNG dùng:** `docs/WORKER_LOGIC.md`, `v2docs/caculate_next_recurring.md`, `docs/DATABASE_SCHEMA.md`, `internal/services/schedule_calculator.go` (engine thứ 2, đã bị loại).

Đặc tả port chi tiết: **[docs/04b-tinh-lich-frp-crp.md](./docs/04b-tinh-lich-frp-crp.md)**.

## Build & chạy

```bash
npm install          # lần đầu sau khi clone
npm run dev          # dev server → http://localhost:3000 (có CẢ API + frontend + SSR, hot reload)
npm run build        # build production ra dist/ (nitro preset cloudflare-pages)
npm run typecheck    # kiểm tra type (vue-tsc)
npm test             # 38 test giá trị chuẩn cho logic tính lịch (vitest)
npx wrangler pages deploy dist --project-name=sennote   # deploy thủ công (hoặc để Git integration tự deploy)
```

- Dev server chạy được API ngay: mở `http://localhost:3000/api/hello` để thử 1 route. Không cần chạy thêm bất kỳ process nào khác.
- Env bắt buộc (đều phải có prefix `NUXT_` — Nuxt chỉ nhận env có prefix này cho `runtimeConfig`): `NUXT_SESSION_PASSWORD` (≥ 32 ký tự, cho auth), `NUXT_TURSO_DATABASE_URL`, `NUXT_TURSO_AUTH_TOKEN`, `NUXT_FCM_CLIENT_EMAIL`, `NUXT_FCM_PRIVATE_KEY`, `NUXT_FCM_PROJECT_ID`.
- KHÔNG commit `.env` / `.dev.vars`; `.env.example` làm mẫu.
- Không có test tự động; xác minh bằng smoke test: `npm run dev` → đăng ký/login → tạo/sửa/xoá 1 reminder → F12 tab Network xem request tới `/api/...`.
- Chạy task lịch thủ công khi dev server đang chạy: mở `http://localhost:3000/_nitro/tasks/reminders:check` (hoặc `Invoke-WebRequest http://localhost:3000/_nitro/tasks/reminders:check`). Không dùng `npx nitro task run` (nó đọc `.nitro/nitro.json` mà Nuxt không ghi).
- Xem task đã đăng ký chưa: `http://localhost:3000/_nitro/tasks` → phải thấy `reminders:check` kèm `scheduledTasks: [{ cron: "* * * * *", tasks: ["reminders:check"] }]`. Nếu chỉ thấy `reminders-check` (gạch ngang) → sai đường dẫn file, xem mục Kiến trúc ở trên.
- ⚠️ Trong dev, cron **chạy thật** (croner) mỗi phút → sẽ gọi DB và báo lỗi nối Turso nếu `.env` chưa có credentials thật. Muốn tắt tạm: `PUT /api/system_status { worker_enabled: false }` (kill-switch).

## Kiến trúc

- **SSR bật** (mặc định, KHÔNG đặt `ssr: false`): app cần server chạy API (auth dùng session cookie), nên dùng `nuxt build`, KHÔNG dùng `nuxt generate`.
- **`app/`** — frontend: `pages/` (route tự động), `components/`, `composables/` (useAuth, useReminders), `api/` (hàm typed gọi `/api/...` cùng origin), `middleware/auth.ts` (route guard).
- **`server/`** — backend: `api/` (route handler: `server/api/reminders/index.get.ts`...), `utils/` (db, fcm, calendar, schedule — business logic), `tasks/` (job chạy định kỳ).
- **`shared/types/`** — type dùng chung 2 bên (`Reminder`, `User`, `RecurrencePattern`...). Import tường minh bằng alias `#shared/types` (auto-import cũng hỗ trợ, nhưng import rõ cho dễ truy vết).
- **Auth**: `nuxt-auth-utils` cho **session** (`setUserSession`/`requireUserSession`/`useUserSession()`).
  ⚠️ **KHÔNG dùng** `hashPassword`/`verifyPassword` của nó — hai hàm này dùng scrypt (`node:crypto`) và **hỏng trên Cloudflare Workers**: tạo hash được nhưng `verifyPassword` luôn trả `false` → không đăng nhập được. Đã thay bằng `server/utils/password.ts` (Web Crypto PBKDF2-SHA256, 100k vòng, salt 16 byte, format `pbkdf2-sha256$iter$salt$hash`). Mọi chỗ cần băm mật khẩu phải dùng `makePasswordHash` / `verifyPasswordHash`.
- **Database**: Turso qua `@libsql/client` (`server/utils/db.ts` là nơi DUY NHẤT tạo client).
- **🚨 Lịch nhắc: Cloudflare Pages KHÔNG chạy Nitro task.** Nitro vẫn sinh handler `scheduled()`, nhưng **Pages không hỗ trợ Cron Triggers** nên nó không bao giờ được gọi. Hậu quả: reminder đến hạn mà không bao giờ gửi.
  Cách đang dùng: Worker **`sennote-cron`** (thư mục `cron/`, có `crons = ["* * * * *"]`) gọi `POST https://sennote.pages.dev/api/cron/reminders-check` mỗi phút, bảo vệ bằng header `x-cron-secret` (env `NUXT_CRON_SECRET` trên Pages = `CRON_SECRET` trên Worker).
  Logic nằm ở `server/utils/run-reminder-check.ts` — dùng chung cho cả Nitro task và endpoint HTTP, **không viết 2 lần**.
  Triệu chứng nếu cron hỏng: reminder `next_action_at` đã qua nhưng `is_sended_one_time` vẫn 0 và `last_sent_at` null.
  - ⚠️ **Tên task suy ra từ ĐƯỜNG DẪN file** (`/` → `:`), không phải `meta.name`. File `server/tasks/reminders-check.ts` → tên `reminders-check`; muốn tên `reminders:check` (dùng trong `scheduledTasks` và URL) **phải** đặt file ở `server/tasks/reminders/check.ts`.
  - ⚠️ **`preset: 'cloudflare-pages'` chỉ đặt khi build**, không đặt khi dev (xem `nuxt.config.ts`). Đặt trong dev → Nitro dùng `cloudflare-dev` emulation, cần wrangler và `/_nitro/tasks/...` trả 404.
- **FCM**: gọi trực tiếp **HTTP v1 API** qua `server/utils/fcm.ts` (firebase-admin KHÔNG chạy trên Workers). Cùng 1 hàm cho Android / iOS / **web** — FCM không phân biệt nền tảng, chỉ thêm `webpush.fcmOptions.link` cho web.
- **Đa thiết bị**: bảng `devices` (1 user nhiều token). Gửi = fan-out qua `listActiveTokens()`, gộp cả `users.fcm_token` legacy và khử trùng. Token chết → `deactivateToken()` tắt **đúng token đó**, không tắt cả user.
- **Web Push**: `public/firebase-messaging-sw.js` (file tĩnh ở gốc, config truyền qua query) + `app/composables/usePush.ts` + component `PushToggle.vue`.
- Port logic cũ từ Go: `server/utils/calendar.ts` (âm lịch) + `server/utils/schedule.ts` (tính FRP/CRP). Giữ chúng là hàm thuần. **Đặc tả: [docs/04b](./docs/04b-tinh-lich-frp-crp.md).**
- **Kill-switch**: bảng `system_status` (singleton `mid = 1`). Task phải check `worker_enabled` trước khi chạy; lỗi FCM **hệ thống** mới tắt worker, lỗi do user (không có token) thì chỉ log và bỏ qua.

### Tầng & quy tắc phụ thuộc

```
app/composables  →  app/api  →  server/api  →  server/utils (service)
                                              server/utils/db.ts (Turso)
server/tasks  →  server/utils (dùng CHUNG service với api, không code trùng)
```

- `server/api/*.ts` chỉ là **handler mỏng**: đọc input → gọi service trong `server/utils` → trả output. KHÔNG nhồi SQL/business logic vào handler.
- `server/utils/*.ts` chứa business logic + SQL (thuần, không import Vue reactivity).

## Gotchas

- **Nitro tasks đang là experimental** → bắt buộc bật `experimental: { tasks: true }` trong `nuxt.config.ts`, và `scheduledTasks` PHẢI đặt trong khối `nitro: {}` (đặt top-level sẽ bị bỏ qua, cron không chạy).
- **Env cho `runtimeConfig` phải có prefix `NUXT_`** (vd `NUXT_TURSO_DATABASE_URL` → key `tursoDatabaseUrl`). Env không có prefix (vd `TURSO_DATABASE_URL`) sẽ KHÔNG được nạp vào config.
- **Cloudflare chỉ đọc env trong vòng đời request/task** — KHÔNG đọc `process.env` ở top-level module (global scope) khi deploy lên Workers; đọc qua `useRuntimeConfig()` hoặc `process.env` bên trong handler.
- **`runtimeConfig.public` được gửi xuống trình duyệt** — chỉ để Firebase web config + VAPID key (vốn đã an toàn khi công khai). **Tuyệt đối không** cho `fcmPrivateKey` / service account vào đây.
- **`public/firebase-messaging-sw.js` phải ở gốc domain** (không được nằm trong `/_nuxt/`) và **không được dùng `import`** — Service Worker không phải module. Cấu hình lấy qua `self.location.search`.
- **Web Push cần HTTPS**; `localhost` được coi là an toàn. Safari iOS chỉ chạy với PWA đã thêm vào màn hình chính (16.4+).
- `NUXT_SESSION_PASSWORD` phải ≥ 32 ký tự; thiếu → lỗi khi chạy.
- **Session là cookie, giới hạn 4 KB** → `setUserSession` chỉ lưu thông tin tối thiểu (id, email...), KHÔNG nhồi dữ liệu lớn vào session.
- **Mọi cột datetime lưu ISO 8601 UTC bằng `toISOString()`** (`2026-08-30T09:25:00.000Z`). Khi so sánh trong SQL **phải truyền cùng định dạng**. Dùng `datetime('now')` của SQLite (`2026-08-30 09:25:00`) sẽ so sánh SAI: tại vị trí thứ 10, `'T'` > `' '` nên mọi ISO string đều "lớn hơn" → query trả về rỗng. Luôn truyền `new Date().toISOString()` làm tham số.
- **Workers không có filesystem** → không dùng SQLite local file (`file:...`), không ghi file; mọi thứ qua Turso HTTP.
- `nuxt generate` KHÔNG dùng được (auth-utils cần server) → chỉ dùng `nuxt build`.
- Gọi API từ client: dùng `$fetch('/api/...')` (browser tự gửi cookie). Gọi trong `useAsyncData`/SSR phải dùng `useFetch` hoặc `useRequestFetch()` để forward cookie — `$fetch` thường trong `useAsyncData` sẽ KHÔNG mang cookie.
- API cùng origin với frontend → **không còn** vấn đề CORS hay mixed-content như mô hình "frontend gọi backend riêng".
- Cron chạy theo giờ **UTC** — khi so sánh thời gian nhắc nhở phải quy về UTC (lưu `next_action_at` dạng ISO/UTC).
- `next_action_at` lưu dạng ISO 8601 (text) để query so sánh được; đồng nhất 1 chuẩn, tránh dùng timestamp số.
- Deploy xong vẫn thấy bản cũ → hard refresh (Ctrl+Shift+R) và kiểm tra đúng deployment mới nhất trên dashboard Cloudflare.
- **Bảng âm lịch `server/utils/lunar-table.ts` CÓ HẠN** (hiện phủ 2026-08-01 → 2029-08-31). Hết hạn → `solarToLunar()` trả `null` và reminder âm lịch ném lỗi `vượt quá phạm vi bảng âm lịch`. Gia hạn bằng 1 lệnh (trong project Go `E:\PROJECT\nhacviecfcm`):
  ```bash
  go run ./cmd/gen_lunar_table -from 2029-01-01 -to 2032-12-31 \
    -out "E:\PROJECT\nhacviecfcm\nhacviecnuxt\server\utils\lunar-table.ts"
  ```
  File này **do máy sinh**, không sửa tay. Kiểm tra phạm vi hiện tại: xem 4 dòng comment đầu file.
- Build trên Cloudflare Pages fail vì Node cũ → set env `NODE_VERSION=20` (hoặc mới hơn) trong Pages project settings.

## Quy tắc khi port logic tính lịch từ Go

Bắt buộc đọc [docs/04b](./docs/04b-tinh-lich-frp-crp.md) trước khi viết `server/utils/schedule.ts`.

1. **Engine chuẩn là `calNextTime.go`.** Không dùng `schedule_calculator.go` (engine thứ 2 của Go, mâu thuẫn với worker).
2. **Mỏ neo là `origin_time`**, không phải `next_recurring` hay lần gửi gần nhất → lịch không trôi.
   `trigger_time_of_day` / `day_of_week` / `day_of_month` chỉ để tương thích, worker **không đọc**.
3. **Luôn guard `interval <= 0`** trước mọi vòng lặp `while (next <= now)`. Go đang **treo vô hạn**
   ở `daily` và `monthly` khi `interval = 0`, mà `interval` có `omitempty` nên client dễ quên gửi.
   Trên Workers, treo = task bị kill và thử lại → lặp vô hạn.
4. **`nextActionAt()` KHÔNG được đưa `now` vào danh sách ứng viên** và **phải bỏ qua `snooze_until`
   đã hết hạn** (short-circuit như Go). Sai 1 trong 2 → reminder bị gửi lặp mỗi phút vô hạn.
5. **Complete recurring ≠ completed.** Giữ `status = 'active'`, đẩy `next_recurring` sang chu kỳ mới.
   Set `completed` là làm chết reminder lặp.
6. **Giờ lấy từ `origin_time` theo UTC.** Âm lịch chỉ đổi sang VN+07 ở bước *tra ngày âm*, sau đó ghép
   lại với giờ UTC của origin.
7. Thêm case mới → thêm luôn vào bảng giá trị chuẩn ở [doc 04b mục 6](./docs/04b-tinh-lich-frp-crp.md#6-bộ-giá-trị-chuẩn-golden-tests--lấy-từ-calnexttime_testgo)
   và chạy `go test` bên Go để lấy kết quả mong đợi.

## Quy ước làm việc (user yêu cầu)

- Khi user nhắn bắt đầu bằng "..." → KHÔNG được viết code, chỉ thảo luận/chờ.
- Không tự commit trừ khi user yêu cầu; commit message ngắn bằng tiếng Việt (vd: `auth register login`, `api reminders crud`, `cron gui fcm`).

## Giới hạn kích thước file

- Không để 1 file vượt quá ~200–300 dòng logic thực (không tính comment/blank line). Nếu vượt, PHẢI tách trước khi viết tiếp.
- Mỗi file chỉ chịu trách nhiệm cho ĐÚNG MỘT domain/concern (Single Responsibility). Nếu file đang xử lý > 1 domain, phải đề xuất tách ngay cả khi không được yêu cầu.

## Tách để AI (và người) đọc nhanh, đỡ tốn token

- Pure function (không phụ thuộc state ngoài, không side-effect) → tách ra file `utils/` riêng, có thể đọc độc lập mà không cần đọc phần còn lại.
- Logic gọi API/network → tách thành 1 lớp riêng (`app/api/` cho client, `server/api/` cho server), không viết `$fetch`/`fetch` trực tiếp trong component hoặc business logic.
- Business logic phức tạp (nhiều state, nhiều effect liên quan) → tách thành composable (`app/composables/`) hoặc service (`server/utils/`), đặt tên theo domain, KHÔNG dồn vào 1 component/handler.
- File chính (`app.vue`, `pages/*.vue`, `server/api/*.ts`) chỉ nên "lắp ráp", không chứa logic chi tiết.

## Tái sử dụng & dễ sửa

- Không copy-paste logic giống nhau ở 2 nơi trở lên — bắt buộc trích xuất thành hàm/module dùng chung.
- Đặt tên rõ nghĩa theo domain, tránh tên chung chung (`data`, `temp`, `handle`, `item`).
- Mỗi hàm nên làm đúng 1 việc, input/output rõ ràng, tránh side-effect ẩn.
- Ưu tiên function thuần (pure) khi có thể để dễ test và dễ AI suy luận.
- Khi sửa 1 tính năng, chỉ động vào phần liên quan — không viết lại toàn bộ file nếu không cần thiết.

## Đơn giản trên hết

- Ưu tiên giải pháp đơn giản, dễ hiểu hơn là "khéo léo"/tối ưu quá mức khi không có yêu cầu hiệu năng cụ thể.
- Không over-engineer: không tạo abstraction/pattern phức tạp cho use-case chỉ dùng 1 lần.

## Quy tắc Vue 3 / Nuxt (cộng thêm với các quy tắc chung ở trên)

### Cấu trúc file .vue

- `<script setup>` trong 1 component KHÔNG vượt quá ~150 dòng logic. Nếu vượt, tách state + logic ra composable (`useXxx.ts`) trước khi viết tiếp.
- Component chỉ nên chứa: (1) gọi composable, (2) state hiển thị thuần UI, (3) template. KHÔNG chứa business logic phức tạp, KHÔNG gọi `$fetch` trực tiếp.
- Nếu `<template>` có > 1 khu vực chức năng rõ rệt (vd: thống kê + danh sách reminder + form tạo), PHẢI tách thành component con (`<StatsPanel>`, `<ReminderList>`, `<ReminderForm>`), component cha chỉ ráp qua props/emit.

### Composable hoá state & logic

- Nhóm state liên quan vào 1 composable theo domain: `useAuth()` (session/user), `useReminders()` (list + CRUD), `useSystemStatus()`.
- Composable trả về đúng thứ component cần (state readonly khi có thể, hàm hành động rõ tên).
- Pure helper (format thời gian, mô tả recurrence...) → `app/utils/*.ts` thuần, không import Vue reactivity.

### API layer

- Client: mọi request đi qua `app/api/*.ts` (`api/auth.ts`, `api/reminders.ts`) export hàm typed; component/composable chỉ gọi hàm này. KHÔNG `$fetch` rải rác trong component.
- Server: mọi route đi qua `server/api/*.ts` (handler mỏng) gọi service `server/utils/*.ts`.

### TypeScript

- Định nghĩa type/interface cho mọi dữ liệu qua API và object state phức tạp trong `shared/types/index.ts` dùng chung 2 bên, không dùng object literal ngầm định kiểu.

### Đặt tên & tổ chức thư mục

```
app/
  pages/        index.vue  login.vue          (route, chỉ ráp component)
  middleware/   auth.ts                       (route guard)
  composables/  useAuth.ts  useReminders.ts
  api/          auth.ts  reminders.ts         (typed $fetch tới /api/...)
  utils/        time.ts  recurrence.ts        (hàm thuần)
  components/   StatsPanel.vue  ReminderList.vue  ReminderCard.vue  ReminderForm.vue
  app.vue                                     (chỉ ráp layout + <NuxtPage/>)
server/
  api/          auth/register.post.ts  auth/login.post.ts  auth/logout.post.ts
                reminders/index.get.ts  reminders/index.post.ts
                reminders/[id].put.ts  reminders/[id].delete.ts  ...
                reminders/[id]/snooze.post.ts  reminders/[id]/complete.post.ts
                devices/index.post.ts  index.get.ts  index.delete.ts  (đa thiết bị)
                users/fcm-token.put.ts         (route CŨ, app Android đang dùng - đừng xoá)
                system_status.get.ts  system_status.put.ts       (kill-switch, cần auth)
  utils/        db.ts  fcm.ts  calendar.ts  schedule.ts  reminders.ts
                devices.ts             (đa thiết bị)
                reminder-processor.ts  (11 nhánh worker + fan-out nhiều thiết bị)
                system-status.ts       (kill-switch)
  tasks/        reminders/check.ts            (scheduled task — tên "reminders:check")
shared/
  types/        index.ts                      (Reminder, User, ...)
                auth.d.ts                     (khai báo session user cho #auth-utils)
docs/                                         (tài liệu)
public/                                       (file tĩnh)
nuxt.config.ts                                (preset cloudflare-pages, tasks, scheduledTasks, runtimeConfig)
```

- Dùng auto-import của Nuxt (composables, utils, server utils). Riêng `app/api/` import tường minh (`~/api/reminders`) và type dùng chung import tường minh qua `#shared/types` để dễ truy vết.

## Khi review/refactor code có sẵn

- Nếu phát hiện 1 file `.vue` > 300 dòng hoặc > 15 ref/reactive trong cùng `<script setup>`, PHẢI đề xuất kế hoạch tách trước, liệt kê rõ composable/component nào sẽ được tạo, rồi mới thực hiện — không tách âm thầm không giải thích.
- Tương tự với `server/api/*.ts`/`server/utils/*.ts` > 300 dòng: đề xuất tách service/handler trước khi làm.
