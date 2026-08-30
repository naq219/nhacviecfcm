# 02 — Khởi tạo dự án: Nuxt + auth-utils + Turso

> Mục tiêu: dev server chạy ở `http://localhost:3000`, có 1 route API `/api/hello`, đã cài đủ 3 thư viện lõi và có database Turso sẵn sàng.

---

## Bước 0 — Cài Node.js (1 lần duy nhất)

Nuxt cần **Node.js LTS ≥ 20**. Tải bản LTS tại https://nodejs.org → cài → kiểm tra:

```powershell
node -v    # v20.x.x trở lên
npm -v     # 10.x.x trở lên
```

Nếu báo `node is not recognized`: đóng PowerShell mở lại. Vẫn lỗi → cài lại Node, tick "Add to PATH".

## Bước 0.5 — Chọn vị trí đặt dự án ⚠️ đọc trước

Thư mục dự án Nuxt là thư mục **chứa `AGENTS.md` và `docs/`** (tức là thư mục bạn đang đọc tài liệu này).
Hiện nó đang nằm **bên trong** repo Go: `E:\PROJECT\nhacviecfcm\nhacviecnuxt`.

Hai lựa chọn — **khuyến nghị (A)**:

| | Cách | Ưu / nhược |
|---|---|---|
| **A** | Chuyển ra ngoài thành `E:\PROJECT\nhacviecnuxt` (ngang hàng với `nhacviecfcm`) | ✅ repo git riêng sạch, deploy lên Cloudflare gọn<br>❌ phải copy thư mục một lần |
| B | Giữ nguyên trong `nhacviecfcm\nhacviecnuxt` | ✅ không phải di chuyển<br>❌ **nested git repo**: nhớ thêm `nhacviecnuxt/` vào `.gitignore` của repo Go (xem doc 05 mục 2.1) |

> ℹ️ Toàn bộ tài liệu dùng đường dẫn **`E:\PROJECT\nhacviecnuxt`** (cách A).
> Nếu bạn chọn cách B, thay bằng `E:\PROJECT\nhacviecfcm\nhacviecnuxt` ở mọi lệnh.

## Bước 1 — Tạo dự án Nuxt

Mở PowerShell tại thư mục dự án:

```powershell
cd E:\PROJECT\nhacviecnuxt
npx nuxi@latest init . --force
```

- `npx` = chạy công cụ npm không cần cài sẵn. `nuxi init` = tạo dự án Nuxt. `.` = tạo trong thư mục hiện tại. `--force` = cho phép ghi vào thư mục đã có `docs/` + `AGENTS.md`.
- Lần đầu `npx` hỏi `Ok to proceed?` → gõ `y` Enter. Nếu hỏi package manager → chọn **npm**.
- ℹ️ Template có thể tạo file `README.md` mới — không sao, tài liệu nằm trong `docs/` và `AGENTS.md`.

## Bước 2 — Cài thư viện

```powershell
npm install
```

## Bước 3 — Chạy thử dev server

```powershell
npm run dev
```

✅ **Checkpoint:** mở http://localhost:3000 thấy trang welcome Nuxt = OK. Dev server có hot reload (sửa → lưu → tự cập nhật). Dừng bằng `Ctrl+C`.

## Bước 4 — Cài 3 thư viện lõi

```powershell
npx nuxi@latest module add auth-utils        # auth (session cookie)
npm install @libsql/client                    # database Turso
```

- `auth-utils` = `nuxt-auth-utils`: đăng nhập bằng session cookie.
  ⚠️ Chỉ dùng phần session. `hashPassword`/`verifyPassword` của nó **hỏng trên Cloudflare Workers**
  (scrypt của `node:crypto`) — dự án dùng `server/utils/password.ts` (Web Crypto PBKDF2) thay thế.
- `@libsql/client`: SDK của Turso, gọi SQLite qua HTTP (chạy được trên Cloudflare Workers).

## Bước 5 — Tạo database trên Turso

> ℹ️ Turso CLI không cài được bằng npm; trên Windows cần WSL. **Cách A (web dashboard) không cần cài gì — khuyến nghị cho người mới.** Cách B dành cho ai đã quen CLI.

### 5.1 Cách A — Tạo DB qua web dashboard (khuyến nghị)

1. Vào https://platform.turso.tech (đăng ký tài khoản miễn phí).
2. **Create database** → đặt tên `sennote` (vùng nào cũng được, chọn gần VN nhất nếu có) → tạo.
3. Trong trang database, vào mục **Tokens** (hoặc "Connect") → tạo token → **LƯU NGAY** (chỉ hiện 1 lần).
4. Ghi lại 2 giá trị:
   - URL database (dạng `libsql://sennote-....turso.io`)
   - Token vừa tạo

### 5.2 Cách B — Dùng Turso CLI (cần WSL)

Turso CLI trên Windows **bắt buộc cài qua WSL** (xem https://docs.turso.tech/cli/installation). Đại khái:

```bash
# trong WSL
curl -sSfL https://get.tur.so/install.sh | bash
turso auth login
turso db create sennote
turso db show sennote          # lấy URL
turso db tokens create sennote # tạo token — LƯU NGAY
```

### 5.3 Kết quả

Ta có:
- `TURSO_DATABASE_URL` = URL database (dùng cho biến `NUXT_TURSO_DATABASE_URL`)
- `TURSO_AUTH_TOKEN` = token (dùng cho biến `NUXT_TURSO_AUTH_TOKEN`)

## Bước 6 — File `.env`

Tạo `.env` ở gốc dự án (file này **không commit** — `.gitignore` của Nuxt đã lo sẵn):

```env
# Auth: chuỗi ngẫu nhiên ≥ 32 ký tự (tự nghĩ hoặc gõ bừa 1 chuỗi dài)
NUXT_SESSION_PASSWORD=doi-mot-chuoi-ngau-nhien-dai-hon-32-ky-tu-nhe!

# Turso — BẮT BUỘC có prefix NUXT_ (Nuxt chỉ nạp env có prefix này vào runtimeConfig)
NUXT_TURSO_DATABASE_URL=libsql://sennote-<ten-cua-ban>.turso.io
NUXT_TURSO_AUTH_TOKEN=token-vua-lay-o-buoc-5

# FCM (sẽ dùng ở doc 04 — để trống cũng chạy được bước hiện tại)
NUXT_FCM_CLIENT_EMAIL=
NUXT_FCM_PRIVATE_KEY=
NUXT_FCM_PROJECT_ID=
```

Tạo thêm `.env.example` (file này **được commit**) để người khác biết cần những biến gì — copy y nguyên `.env` nhưng **bỏ giá trị thật** (để `=` trống).

> ⚠️ Sửa `.env` xong phải `Ctrl+C` rồi `npm run dev` lại mới có tác dụng.

## Bước 7 — Cấu hình `nuxt.config.ts`

```ts
// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: ['nuxt-auth-utils'],

  app: {
    head: {
      title: 'SenNote',
      htmlAttrs: { lang: 'vi' },
    },
  },

  runtimeConfig: {
    // Các biến này chỉ đọc được ở SERVER (server/utils, server/api, server/tasks)
    tursoDatabaseUrl: '',
    tursoAuthToken: '',
    fcmClientEmail: '',
    fcmPrivateKey: '',
    fcmProjectId: '',
  },

  nitro: {
    // ⚠️ CHỈ đặt preset khi build, KHÔNG đặt khi chạy dev.
    // preset cloudflare-pages trong dev → Nitro dùng "cloudflare-dev emulation",
    // cần cài wrangler và endpoint /_nitro/tasks/... bị 404 (không test cron tay được).
    ...(process.env.NODE_ENV === 'production' ? { preset: 'cloudflare-pages' } : {}),

    experimental: {
      tasks: true,                // bật Nitro tasks (bắt buộc cho cron ở doc 04)
    },
    // Cron: mỗi phút chạy task "reminders:check" (tạo task ở doc 04).
    // PHẢI đặt TRONG khối nitro: {} — đây là option của Nitro, để top-level sẽ bị bỏ qua.
    scheduledTasks: {
      '* * * * *': 'reminders:check',
    },
  },
})
```

> ℹ️ `runtimeConfig` ở trên là biến **server-only** (không có `public:`). Nuxt tự nạp giá trị từ env trùng tên nhưng **BẮT BUỘC có prefix `NUXT_`**: key `tursoDatabaseUrl` ← env `NUXT_TURSO_DATABASE_URL`, key `fcmPrivateKey` ← env `NUXT_FCM_PRIVATE_KEY`... (prefix bỏ đi + đổi thành camelCase). Env không có `NUXT_` (vd `TURSO_DATABASE_URL`) sẽ không được nạp.

## Bước 8 — Tạo các thư mục theo quy ước

```powershell
mkdir app\pages, app\components, app\composables, app\api, app\utils, app\middleware
mkdir server\api, server\utils, server\tasks\reminders
mkdir shared\types
mkdir db
```

> ⚠️ `server\tasks\reminders` có thư mục con `reminders\` là **cố ý**: tên task của Nitro suy ra từ
> đường dẫn file (`/` → `:`). File `server/tasks/reminders/check.ts` → tên task **`reminders:check`**
> (khớp `scheduledTasks` ở Bước 7). Nếu để `server/tasks/reminders-check.ts` thì tên thành
> `reminders-check` và cron sẽ báo `Scheduled task reminders:check is not defined!`.

Cấu trúc chuẩn (xem `AGENTS.md`):

```
app/        frontend
server/     backend (api/, utils/, tasks/)
shared/     types dùng chung
```

## Bước 8.5 — Sửa `app.vue` để các trang hiển thị được

Template mặc định sau init chứa trang welcome, **không có `<NuxtPage />`** → tạo file trong `app/pages/` mà không sửa `app.vue` thì **không trang nào render**. Mở `app/app.vue` và thay toàn bộ bằng:

```vue
<template>
  <div>
    <NuxtPage />
  </div>
</template>
```

✅ **Checkpoint:** mở http://localhost:3000 → vẫn thấy trang welcome là bình thường (chưa có file trong `pages/`). Sau doc 03 tạo `pages/index.vue` + `pages/login.vue`, chúng sẽ xuất hiện tại `/` và `/login`.

## Bước 9 — Route `/api/hello` để kiểm tra API

Tạo `server/api/hello.ts`:

```ts
export default defineEventHandler(() => {
  return { hello: 'SenNote API chạy ngon!' }
})
```

✅ **Checkpoint:** mở http://localhost:3000/api/hello → thấy JSON `{"hello":"SenNote API chạy ngon!"}` = **API và frontend cùng chạy trong 1 dev server**. Đây là điểm mấu chốt của mô hình fullstack.

## Bước 10 — Test nối database

Tạo `server/utils/db.ts`:

```ts
import { createClient } from '@libsql/client'

let _db: ReturnType<typeof createClient> | null = null

// Nơi DUY NHẤT tạo connection tới Turso
export function useDb() {
  if (_db) return _db
  const cfg = useRuntimeConfig()
  _db = createClient({
    url: cfg.tursoDatabaseUrl,
    authToken: cfg.tursoAuthToken,
  })
  return _db
}
```

Tạo `server/api/ping.ts` để thử đọc database:

```ts
export default defineEventHandler(async () => {
  const db = useDb()
  const rs = await db.execute('SELECT 1 AS ok')
  return { ok: true, rows: rs.rows }
})
```

✅ **Checkpoint:** mở http://localhost:3000/api/ping → thấy `{"ok":true,"rows":[{"ok":1}]}` = đã nối được Turso. (Lỗi → kiểm tra lại 2 biến Turso trong `.env`.)

## Lỗi thường gặp

| Triệu chứng | Nguyên nhân | Cách sửa |
|-------------|-------------|----------|
| `node is not recognized` | Chưa cài Node / thiếu PATH | Cài lại Node LTS, mở PowerShell mới |
| `nuxi init` báo thư mục không trống | Quên `--force` | Thêm `--force` |
| `/api/ping` lỗi `unauthorized` | Token/URL Turso sai | Xem lại 2 biến `NUXT_TURSO_*` trong `.env`; lấy lại token trên dashboard Turso |
| `/api/ping` trả lỗi nhưng token đúng | Đặt nhầm tên biến thiếu prefix `NUXT_` | Đổi thành `NUXT_TURSO_DATABASE_URL` / `NUXT_TURSO_AUTH_TOKEN` |
| App báo lỗi `NUXT_SESSION_PASSWORD` | Thiếu hoặc quá ngắn | Đặt chuỗi ≥ 32 ký tự trong `.env`, restart |
| Sửa `.env` không ăn | Chưa restart dev server | Ctrl+C rồi `npm run dev` lại |

## Xong doc này khi nào?

- [ ] `npm run dev` chạy, mở http://localhost:3000 thấy trang Nuxt
- [ ] `app/app.vue` đã chứa `<NuxtPage />`
- [ ] `/api/hello` trả JSON
- [ ] `/api/ping` đọc được Turso
- [ ] Có `.env` đủ 6 biến (3 biến FCM có thể để trống)

→ Tiếp theo: [03 — Database schema + API auth + CRUD reminder](./03-database-va-api.md)
