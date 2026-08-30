# 03 — Database schema + API auth + CRUD reminder

> Mục tiêu: đăng ký/đăng nhập được bằng tài khoản thật, tạo/sửa/xoá reminder thật, trang login + danh sách chạy ngon ở local.
> Yêu cầu: đã xong [doc 02](./02-khoi-tao-du-an.md).

---

## Bước 1 — Tạo bảng trong Turso

Tạo file `db/schema.sql`:

```sql
-- Bảng người dùng
CREATE TABLE IF NOT EXISTS users (
  id            TEXT PRIMARY KEY,          -- sinh bằng crypto.randomUUID()
  email         TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,             -- từ hashPassword() của auth-utils
  fcm_token     TEXT,                      -- token FCM của mobile (1 user 1 token, ghi đè)
  is_fcm_active INTEGER NOT NULL DEFAULT 1,
  created_at    TEXT NOT NULL,             -- ISO 8601 (UTC)
  updated_at    TEXT NOT NULL
);

-- Thiết bị nhận thông báo (1 user NHIỀU thiết bị: Android + web + iOS)
-- Lý do tách: trước đây users.fcm_token chỉ chứa 1 token, thiết bị đăng ký sau
-- sẽ GHI ĐÈ thiết bị trước (mở web là mất thông báo trên Android).
CREATE TABLE IF NOT EXISTS devices (
  id         TEXT PRIMARY KEY,
  user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token      TEXT NOT NULL,
  platform   TEXT NOT NULL DEFAULT 'android',  -- 'android' | 'ios' | 'web'
  is_active  INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_devices_token ON devices(token);
CREATE INDEX IF NOT EXISTS idx_devices_user ON devices(user_id, is_active);

-- Bảng nhắc nhở
CREATE TABLE IF NOT EXISTS reminders (
  id                    TEXT PRIMARY KEY,
  user_id               TEXT NOT NULL REFERENCES users(id),
  title                 TEXT NOT NULL,
  description           TEXT,
  tag                   TEXT,
  type                  TEXT NOT NULL,            -- 'one_time' | 'recurring'
  status                TEXT NOT NULL DEFAULT 'active', -- 'active'|'completed'|'paused'
  recurrence_pattern    TEXT,                     -- JSON (xem bên dưới)
  repeat_strategy       TEXT NOT NULL DEFAULT 'none',   -- 'none'|'crp_until_complete'
  calendar_type         TEXT NOT NULL DEFAULT 'solar',  -- 'solar'|'lunar'
  origin_time           TEXT,                     -- MỎ NEO TÍNH LỊCH LẶP (xem chú thích)
  next_recurring        TEXT,                     -- ISO 8601 UTC
  next_crp              TEXT,
  max_crp               INTEGER NOT NULL DEFAULT 0,
  crp_count             INTEGER NOT NULL DEFAULT 0,
  crp_interval_sec      INTEGER NOT NULL DEFAULT 0,
  is_sended_one_time    INTEGER NOT NULL DEFAULT 0,
  next_action_at        TEXT,                     -- thời điểm gần nhất cần xử lý
  last_sent_at          TEXT,
  last_completed_at     TEXT,
  last_crp_completed_at TEXT,
  snooze_until          TEXT,
  created_at            TEXT NOT NULL,
  updated_at            TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_reminders_next_action ON reminders(next_action_at, status);
CREATE INDEX IF NOT EXISTS idx_reminders_user ON reminders(user_id, status);

-- Kill-switch cho cron (port từ system_status của Go)
CREATE TABLE IF NOT EXISTS system_status (
  mid            INTEGER PRIMARY KEY CHECK (mid = 1),
  worker_enabled INTEGER NOT NULL DEFAULT 0,
  last_error     TEXT,
  created_at     TEXT NOT NULL,
  updated_at     TEXT NOT NULL
);
INSERT OR IGNORE INTO system_status (mid, worker_enabled, last_error, created_at, updated_at)
VALUES (1, 1, '', datetime('now'), datetime('now'));
```

> ### ⚠️ 3 cột hay bị bỏ sót (bắt buộc phải có)
>
> | Cột | Vì sao |
> |-----|--------|
> | **`origin_time`** | **Mỏ neo của mọi phép tính lặp.** Worker tính `next_recurring` bằng cách cộng chu kỳ từ `origin_time` (không phải từ lần gửi gần nhất) → lịch **không bị trôi**. Thiếu cột này là không port được logic lịch. Chi tiết: [doc 04b mục 2](./04b-tinh-lich-frp-crp.md#2-các-cột-bắt-buộc-nhắc-lại-từ-doc-03) |
> | **`is_sended_one_time`** | Phân biệt "gửi lần đầu" và "đang retry CRP". Thiếu → mất toàn bộ `crp_until_complete` và retry của one_time |
> | **`tag`** | Trường có thật trong Go, dùng để phân loại |
>
> Lý do 3 cột này hay bị sót: tài liệu `docs/DATABASE_SCHEMA.md` của project Go **cũng thiếu chúng** (chỉ liệt kê field trong PocketBase collection). Luôn đối chiếu với `migrations/1671631110_init_schema.go`.
>
> ⚠️ `origin_time` khai báo nullable để không phá dữ liệu cũ one_time; khi migrate thì set `origin_time = next_action_at` cho các bản ghi one_time.

### Vì sao cần bảng `devices`

App Android cũ chỉ có 1 token trên `users.fcm_token`. Khi thêm web, thiết bị đăng ký sau sẽ **ghi đè** thiết bị trước → mở web là mất thông báo trên điện thoại. Bảng `devices` cho phép 1 user có nhiều thiết bị cùng nhận.

`users.fcm_token` + `users.is_fcm_active` được **giữ lại** để không broke app Android đang chạy:
- `users.is_fcm_active` trở thành **công tắc tổng** (user tắt hẳn thông báo mọi thiết bị).
- `users.fcm_token` vẫn được gửi kèm (khử trùng với `devices`) → **không cần migrate dữ liệu**.

Khi app Android cập nhật sang `POST /api/devices`, cột `fcm_token` có thể bỏ.

Áp dụng vào Turso (chạy 1 lần). Chọn 1 trong 2 cách:

**Cách 1 (khuyến nghị) — chạy SQL trên web dashboard:**
1. Vào https://platform.turso.tech → mở database `sennote` → mục **SQL/Console** (tab truy vấn SQL).
2. Dán toàn bộ nội dung `db/schema.sql` vào → chạy.
3. Mở lại tab đó gõ `SELECT name FROM sqlite_master WHERE type='table'` → thấy `users`, `reminders`.

**Cách 2 — nếu có Turso CLI (WSL):** lưu ý PowerShell KHÔNG hỗ trợ toán tử `<` nên không dùng được `turso db shell sennote < db/schema.sql`. Dùng pipe hoặc `cmd`:

```powershell
Get-Content db/schema.sql -Raw | turso db shell sennote
# hoặc: cmd /c "turso db shell sennote < db/schema.sql"
```

> ℹ️ Muốn kiểm tra bảng trong CLI: `turso db shell sennote` → gõ `.tables` → thấy `users`, `reminders`. Trong web console thì dùng `SELECT` như ở trên.

> ⚠️ **Chuẩn thời gian:** mọi cột datetime lưu ISO 8601 UTC (vd `2026-08-28T07:00:00.000Z`). Vì cron trên Cloudflare chạy theo giờ UTC, giữ nguyên 1 chuẩn này để so sánh bằng chuỗi cho đơn giản (`'2026-08-28T07:00:00.000Z' <= 'now'`).

### `recurrence_pattern` (JSON)

```json
{"type":"daily","interval":1,"trigger_time_of_day":"08:00"}
{"type":"interval_seconds","interval_seconds":180}
{"type":"weekly","interval":1,"day_of_week":1,"trigger_time_of_day":"09:00"}
{"type":"monthly","interval":1,"day_of_month":5,"trigger_time_of_day":"10:00"}
{"type":"solar_last_day_of_month","trigger_time_of_day":"18:00"}
{"type":"lunar_last_day_of_month","trigger_time_of_day":"18:00"}
```

> ⚠️ **`trigger_time_of_day`, `day_of_week`, `day_of_month` KHÔNG phải nguồn sự thật khi tính lịch.**
> Worker engine (đã chọn làm chuẩn, xem [doc 04b mục 1](./04b-tinh-lich-frp-crp.md#1-vì-sao-bỏ-schedule_calculatorgo-và-hệ-quả)) lấy **giờ/phút và thứ từ `origin_time`**, bỏ qua 3 trường trên.
> Chúng chỉ được dùng **lúc tạo** để suy ra `origin_time` nếu client không gửi. Vẫn giữ trong JSON để tương thích dữ liệu cũ.

> ⚠️ `type` có **6** giá trị. `solar_last_day_of_month` thường bị sót vì `docs/DATABASE_SCHEMA.md` của Go không liệt kê.

## Bước 2 — Types dùng chung

Tạo `shared/types/index.ts`:

```ts
export interface User {
  id: string
  email: string
  fcm_token?: string
  is_fcm_active?: boolean
}

// Dữ liệu public trả về cho client (KHÔNG gồm password_hash)
export interface PublicUser {
  id: string
  email: string
}

export type ReminderType = 'one_time' | 'recurring'
export type ReminderStatus = 'active' | 'completed' | 'paused'
export type RepeatStrategy = 'none' | 'crp_until_complete'
export type RecurrenceType =
  | 'daily' | 'weekly' | 'monthly' | 'interval_seconds'
  | 'solar_last_day_of_month' | 'lunar_last_day_of_month'

export interface RecurrencePattern {
  type: RecurrenceType
  interval?: number            // ngày / tuần / tháng tuỳ type; <= 0 hoặc thiếu → 1
  interval_seconds?: number    // chỉ dùng cho interval_seconds; bắt buộc > 0
  day_of_week?: number         // chỉ để tương thích; worker lấy thứ từ origin_time
  day_of_month?: number        // chỉ để tương thích; worker lấy ngày từ origin_time
  calendar_type?: 'solar' | 'lunar'
  trigger_time_of_day?: string // chỉ để tương thích; worker lấy giờ từ origin_time
}

export interface Reminder {
  id: string
  user_id: string
  title: string
  description?: string
  tag?: string
  type: ReminderType
  status: ReminderStatus
  recurrence_pattern?: RecurrencePattern
  repeat_strategy: RepeatStrategy
  calendar_type: 'solar' | 'lunar'
  origin_time: string | null        // MỎ NEO TÍNH LỊCH LẶP
  next_recurring?: string | null
  next_crp?: string | null
  max_crp: number
  crp_count: number
  crp_interval_sec: number
  is_sended_one_time: 0 | 1
  next_action_at?: string | null
  last_sent_at?: string | null
  last_completed_at?: string | null
  last_crp_completed_at?: string | null
  snooze_until?: string | null
  created_at: string
  updated_at: string
}

// Payload tạo/sửa: client chỉ gửi các trường này
export type ReminderPayload = Pick<Reminder,
  | 'title' | 'description' | 'tag' | 'type'
  | 'recurrence_pattern' | 'repeat_strategy' | 'calendar_type'
  | 'origin_time' | 'next_action_at'
  | 'max_crp' | 'crp_interval_sec'
> & { for_test?: number }   // giây — tiện test cron (giống Go)

export interface SystemStatus {
  mid: 1
  worker_enabled: 0 | 1
  last_error: string | null
  updated_at: string
}
```

> ⚠️ **`ReminderPayload` phải có `next_action_at`, `origin_time` và `repeat_strategy`.**
> - `next_action_at`: ở Go đây là input **bắt buộc** khi tạo (`reminder_handler.go:317-319`). Thiếu → cron không bao giờ nhắc.
> - `origin_time`: bắt buộc với recurring. Thiếu → `nextRecurring()` ném lỗi.
> - `repeat_strategy`: thiếu → mất tính năng "nhắc đến khi hoàn thành".
> - `for_test`: nếu > 0 thì server ghi đè `next_action_at = now + for_test` giây (port từ `reminder_handler.go:91-96`).

> ℹ️ `shared/types/` được Nuxt auto-import type cho cả server lẫn client. Tuy vậy, trong tài liệu này ta **import tường minh** bằng alias `#shared/types` cho rõ ràng, dễ truy vết (alias `#shared` do Nuxt tự cấu hình sẵn).

### Thêm type cho session user (auth.d.ts)

`useUserSession()`/`requireUserSession()` trả `session.user` nhưng kiểu mặc định là object rỗng — cần khai báo để truy cập `session.user.id` không lỗi TypeScript. Tạo `shared/types/auth.d.ts`:

```ts
// Khai báo type cho session user của nuxt-auth-utils
declare module '#auth-utils' {
  interface User {
    id: string
    email: string
  }
}

export {}
```

## Bước 3 — Auth: đăng ký + đăng nhập + đăng xuất

### 3.1 Service `server/utils/auth.ts`

```ts
// ⚠️ KHÔNG dùng hashPassword/verifyPassword của nuxt-auth-utils.
// Hai hàm đó dùng scrypt của node:crypto và BỊ HỎNG trên Cloudflare Workers:
// tạo hash được nhưng verify luôn trả false → không đăng nhập được.
// Dùng Web Crypto PBKDF2 thay thế: server/utils/password.ts
import { makePasswordHash, verifyPasswordHash } from './password'

// Tạo user mới. Trả null nếu email đã tồn tại.
export async function createUser(email: string, password: string) {
  const db = useDb()
  const exist = await db.execute({
    sql: 'SELECT id FROM users WHERE email = ?',
    args: [email],
  })
  if (exist.rows.length > 0) return null

  const user = {
    id: crypto.randomUUID(),
    email,
    password_hash: await makePasswordHash(password),
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  }
  await db.execute({
    sql: 'INSERT INTO users (id, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?)',
    args: [user.id, user.email, user.password_hash, user.created_at, user.updated_at],
  })
  return user
}

// Kiểm tra email + password. Trả null nếu sai.
export async function verifyUser(email: string, password: string) {
  const db = useDb()
  const rs = await db.execute({ sql: 'SELECT * FROM users WHERE email = ?', args: [email] })
  const row = rs.rows[0] as any
  if (!row) return null
  const ok = await verifyPasswordHash(row.password_hash as string, password)
  if (!ok) return null
  return { id: row.id as string, email: row.email as string }
}
```

### 3.2 Route `server/api/auth/register.post.ts`

```ts
export default defineEventHandler(async (event) => {
  const { email, password } = await readBody(event)
  if (!email || !password || String(password).length < 8) {
    throw createError({ statusCode: 400, message: 'Email và mật khẩu (≥ 8 ký tự) là bắt buộc' })
  }
  const user = await createUser(String(email), String(password))
  if (!user) {
    throw createError({ statusCode: 409, message: 'Email đã tồn tại' })
  }
  await setUserSession(event, { user: { id: user.id, email: user.email } })
  return { user: { id: user.id, email: user.email } }
})
```

### 3.3 Route `server/api/auth/login.post.ts`

```ts
export default defineEventHandler(async (event) => {
  const { email, password } = await readBody(event)
  const user = await verifyUser(String(email ?? ''), String(password ?? ''))
  if (!user) {
    throw createError({ statusCode: 401, message: 'Sai email hoặc mật khẩu' })
  }
  await setUserSession(event, { user })
  return { user }
})
```

### 3.4 Route `server/api/auth/logout.post.ts`

```ts
export default defineEventHandler(async (event) => {
  await clearUserSession(event)
  return { ok: true }
})
```

### 3.5 Route kiểm tra session hiện tại

```ts
// server/api/auth/me.get.ts
export default defineEventHandler(async (event) => {
  const session = await getUserSession(event)
  return { user: session.user ?? null }
})
```

## Bước 4 — Reminder: service + CRUD

### 4.1 Service `server/utils/reminders.ts`

```ts
import type { Reminder, ReminderPayload } from '#shared/types'

function mapRow(row: any): Reminder {
  return {
    ...row,
    recurrence_pattern: row.recurrence_pattern ? JSON.parse(row.recurrence_pattern) : undefined,
  }
}

// Lấy danh sách reminder của 1 user. status lọc theo trạng thái (mặc định 'active').
export async function listReminders(userId: string, status?: ReminderStatus): Promise<Reminder[]> {
  const db = useDb()
  const rs = await db.execute({
    // NULLS LAST: reminder chưa có lịch (next_action_at NULL) xếp cuối thay vì đầu
    sql: `SELECT * FROM reminders
          WHERE user_id = ? AND (? IS NULL OR status = ?)
          ORDER BY next_action_at IS NULL, next_action_at ASC`,
    args: [userId, status ?? null, status ?? null],
  })
  return rs.rows.map(mapRow)
}

// Tạo reminder.
export async function createReminder(userId: string, data: ReminderPayload): Promise<Reminder> {
  const db = useDb()
  const id = crypto.randomUUID()
  const now = new Date()
  const nowIso = now.toISOString()

  // for_test: tiện test cron — đẩy next_action_at về tương lai gần
  let nextActionAt = data.next_action_at ?? null
  if (data.for_test && data.for_test > 0) {
    nextActionAt = new Date(now.getTime() + data.for_test * 1000).toISOString()
  }

  // origin_time: mỏ neo tính lịch. Client không gửi → suy từ next_action_at (hoặc now)
  const originTime = data.origin_time ?? nextActionAt ?? nowIso

  const row = {
    id,
    user_id: userId,
    title: data.title,
    description: data.description ?? null,
    tag: data.tag ?? null,
    type: data.type,
    status: 'active' as ReminderStatus,
    recurrence_pattern: data.recurrence_pattern ? JSON.stringify(data.recurrence_pattern) : null,
    repeat_strategy: data.repeat_strategy ?? 'none',
    calendar_type: data.calendar_type ?? 'solar',
    origin_time: originTime,
    max_crp: data.max_crp ?? 0,
    crp_interval_sec: data.crp_interval_sec ?? 0,
    is_sended_one_time: 0,
    // ⚠️ next_recurring CHỈ dành cho recurring. Go từ chối tạo one_time có next_recurring
    //    (reminder_handler.go:355-357). one_time chỉ dùng next_action_at + next_crp.
    next_recurring: data.type === 'recurring' ? nextActionAt : null,
    next_crp: nextActionAt,
    next_action_at: nextActionAt,
    created_at: nowIso,
    updated_at: nowIso,
  }

  await db.execute({
    sql: `INSERT INTO reminders
      (id, user_id, title, description, tag, type, status, recurrence_pattern,
       repeat_strategy, calendar_type, origin_time, max_crp, crp_interval_sec,
       is_sended_one_time, next_recurring, next_crp, next_action_at, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
    args: [row.id, row.user_id, row.title, row.description, row.tag, row.type, row.status,
      row.recurrence_pattern, row.repeat_strategy, row.calendar_type, row.origin_time,
      row.max_crp, row.crp_interval_sec, row.is_sended_one_time,
      row.next_recurring, row.next_crp, row.next_action_at, nowIso, nowIso],
  })
  return mapRow(row)
}

// ⚠️ BẢO MẬT: các hàm dưới đây đều nhận thêm userId và WHERE kèm user_id = ?
// để user A KHÔNG đọc/sửa/xoá được reminder của user B (lỗi IDOR).

// Trả null nếu không tồn tại HOẶC không thuộc về userId.
export async function getReminder(id: string, userId: string): Promise<Reminder | null> {
  const db = useDb()
  const rs = await db.execute({
    sql: 'SELECT * FROM reminders WHERE id = ? AND user_id = ?',
    args: [id, userId],
  })
  return rs.rows[0] ? mapRow(rs.rows[0]) : null
}

// Cập nhật reminder. Chỉ ghi đè các trường được gửi (merge "có giá trị thì lấy"),
// giống Go (reminder_service.go:103-129) — KHÔNG ghi đè bằng ?? default
// (nếu không, client quên gửi description sẽ bị xoá mất).
export async function updateReminder(id: string, userId: string, data: Partial<ReminderPayload>): Promise<Reminder | null> {
  const db = useDb()
  const current = await getReminder(id, userId)
  if (!current) return null

  const merged: Reminder = {
    ...current,
    ...(data.title !== undefined ? { title: data.title } : {}),
    ...(data.description !== undefined ? { description: data.description } : {}),
    ...(data.tag !== undefined ? { tag: data.tag } : {}),
    ...(data.type !== undefined ? { type: data.type } : {}),
    ...(data.recurrence_pattern !== undefined ? { recurrence_pattern: data.recurrence_pattern } : {}),
    ...(data.repeat_strategy !== undefined ? { repeat_strategy: data.repeat_strategy } : {}),
    ...(data.calendar_type !== undefined ? { calendar_type: data.calendar_type } : {}),
    ...(data.origin_time !== undefined ? { origin_time: data.origin_time } : {}),
    ...(data.max_crp !== undefined ? { max_crp: data.max_crp } : {}),
    ...(data.crp_interval_sec !== undefined ? { crp_interval_sec: data.crp_interval_sec } : {}),
  }

  // Đổi pattern/origin → phải tính lại lịch, nếu không cron vẫn chạy theo lịch cũ
  const patternChanged =
    data.recurrence_pattern !== undefined ||
    data.origin_time !== undefined ||
    data.calendar_type !== undefined ||
    data.next_action_at !== undefined

  const now = new Date()
  if (merged.type === 'recurring' && patternChanged) {
    merged.next_recurring = nextRecurring(merged, now).toISOString()
  }
  if (data.next_action_at !== undefined) merged.next_action_at = data.next_action_at
  if (patternChanged) merged.next_action_at = nextActionAt(merged, now)
  merged.updated_at = now.toISOString()

  await db.execute({
    sql: `UPDATE reminders SET title = ?, description = ?, tag = ?, type = ?,
            recurrence_pattern = ?, repeat_strategy = ?, calendar_type = ?, origin_time = ?,
            max_crp = ?, crp_interval_sec = ?, next_recurring = ?, next_action_at = ?, updated_at = ?
          WHERE id = ? AND user_id = ?`,
    args: [
      merged.title, merged.description ?? null, merged.tag ?? null, merged.type,
      merged.recurrence_pattern ? JSON.stringify(merged.recurrence_pattern) : null,
      merged.repeat_strategy, merged.calendar_type, merged.origin_time,
      merged.max_crp, merged.crp_interval_sec,
      merged.next_recurring ?? null, merged.next_action_at ?? null,
      merged.updated_at, id, userId,
    ],
  })
  return getReminder(id, userId)
}

// Trả false nếu không tồn tại hoặc không thuộc về userId.
export async function deleteReminder(id: string, userId: string): Promise<boolean> {
  const db = useDb()
  const rs = await db.execute({
    sql: 'DELETE FROM reminders WHERE id = ? AND user_id = ?',
    args: [id, userId],
  })
  return rs.rowsAffected > 0
}

// Lưu token FCM của thiết bị (1 user 1 token, ghi đè)
export async function updateFcmToken(userId: string, token: string): Promise<void> {
  const db = useDb()
  await db.execute({
    sql: 'UPDATE users SET fcm_token = ?, is_fcm_active = 1, updated_at = ? WHERE id = ?',
    args: [token, new Date().toISOString(), userId],
  })
}
```

> ℹ️ `nextRecurring()` và `nextActionAt()` được định nghĩa ở [doc 04b](./04b-tinh-lich-frp-crp.md#3-serverutilsschedulets--các-hàm-thuần).
> Nếu muốn làm doc 03 chạy được trước khi có doc 04b, tạm thời bỏ khối `patternChanged` và ghi chú TODO.

### Validate khi tạo (port từ `reminder_handler.go:306-364`)

| Trường | Rule |
|---|---|
| `title` | bắt buộc, không rỗng |
| `type` | bắt buộc, `one_time` \| `recurring` |
| `calendar_type` | mặc định `solar` ⚠️ **bên Go bắt buộc phải gửi** do `Validate()` chạy trước khi gán default (`reminder_service.go:38` vs `:51`). Bản TS gán default trước → dễ dùng hơn |
| `next_action_at` | bắt buộc (trừ khi dùng `for_test`) |
| nếu `type = recurring` | `recurrence_pattern` bắt buộc; `origin_time` nên có (nếu thiếu, server tự = `next_action_at`) |
| nếu `max_crp > 0` | `crp_interval_sec` phải > 0 |
| `interval` | `<= 0` hoặc thiếu → tự hiểu là 1 (**bắt buộc**, nếu không worker treo — xem doc 04b mục 3.1) |
| `interval_seconds` | phải > 0 |

> ⚠️ `snooze` và `complete` đụng tới logic tính lịch → nằm ở [doc 04 mục 6](./04-lich-nhac-va-fcm.md#6-route-snooze-và-complete).

### Route đăng ký thiết bị (đa thiết bị)

`server/api/devices/index.post.ts` — **dùng cho web và app mới**:

```ts
export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  const { token, platform } = await readBody(event)   // platform: 'android' | 'ios' | 'web'

  if (!token || typeof token !== 'string') {
    throw createError({ statusCode: 400, message: 'token là bắt buộc' })
  }
  await registerDevice(session.user.id, token, platform ?? 'web')
  return { ok: true, platform }
})
```

| Route | Mục đích |
|---|---|
| `POST /api/devices` | Đăng ký / làm mới 1 thiết bị (upsert theo token) |
| `GET /api/devices` | Liệt kê thiết bị của mình (token đã che bớt) |
| `DELETE /api/devices` | Gỡ thiết bị (body `{ token }`) — gọi khi logout |

### ⚠️ Route CŨ vẫn giữ (đừng xoá)

`PUT /api/users/fcm-token` — **app Android hiện tại đang dùng**. Route này vẫn ghi
`users.fcm_token` và đồng thời upsert vào bảng `devices`, nên dữ liệu cũ tự đồng bộ
**mà không cần chạy migration**. Chỉ xoá sau khi app Android đã chuyển sang `POST /api/devices`.

> ℹ️ Go có route này (`main.go:191`) nhưng **không set CORS header** (`user_handler.go:39-61`) — bản Nuxt cùng origin nên không cần lo.

### Service `server/utils/devices.ts`

```ts
registerDevice(userId, token, platform)  // upsert theo token
unregisterDevice(userId, token)          // xoá
deactivateToken(token)                   // token chết → chỉ tắt token đó (KHÔNG tắt cả user)
listDevices(userId)                      // danh sách thiết bị
listActiveTokens(userId)                 // token đang hoạt động, GỘP cả users.fcm_token, khử trùng
```

### 4.2 Routes `server/api/reminders/`

`index.get.ts` — danh sách (chỉ của chính mình):

```ts
export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  return listReminders(session.user.id)
})
```

`index.post.ts` — tạo:

```ts
export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  const body = await readBody(event)
  const reminder = await createReminder(session.user.id, body)
  return reminder
})
```

`[id].get.ts` — chi tiết:

```ts
export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  const id = getRouterParam(event, 'id')!
  const r = await getReminder(id, session.user.id)
  if (!r) throw createError({ statusCode: 404, message: 'Không tìm thấy reminder' })
  return r
})
```

`[id].put.ts` — sửa:

```ts
export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  const id = getRouterParam(event, 'id')!
  const body = await readBody(event)
  const r = await updateReminder(id, session.user.id, body)
  if (!r) throw createError({ statusCode: 404, message: 'Không tìm thấy reminder' })
  return r
})
```

`[id].delete.ts` — xoá:

```ts
export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  const id = getRouterParam(event, 'id')!
  const ok = await deleteReminder(id, session.user.id)
  if (!ok) throw createError({ statusCode: 404, message: 'Không tìm thấy reminder' })
  return { ok: true }
})
```

> ℹ️ `requireUserSession(event)` tự trả HTTP 401 nếu chưa đăng nhập. Route nào cần login thì mở đầu bằng dòng này.

## Bước 5 — Client: lớp `app/api/` + composable + trang

### 5.1 `app/api/auth.ts`

```ts
export const apiRegister = (email: string, password: string) =>
  $fetch('/api/auth/register', { method: 'POST', body: { email, password } })

export const apiLogin = (email: string, password: string) =>
  $fetch('/api/auth/login', { method: 'POST', body: { email, password } })

export const apiLogout = () => $fetch('/api/auth/logout', { method: 'POST' })
```

### 5.2 `app/api/reminders.ts`

```ts
import type { Reminder, ReminderPayload } from '#shared/types'

export const apiListReminders = () => $fetch<Reminder[]>('/api/reminders')
export const apiCreateReminder = (body: ReminderPayload) =>
  $fetch<Reminder>('/api/reminders', { method: 'POST', body })
export const apiUpdateReminder = (id: string, body: ReminderPayload) =>
  $fetch<Reminder>(`/api/reminders/${id}`, { method: 'PUT', body })
export const apiDeleteReminder = (id: string) =>
  $fetch(`/api/reminders/${id}`, { method: 'DELETE' })
```

> ℹ️ `$fetch('/api/...')` gọi cùng origin → trình duyệt tự gửi session cookie, không cần tự gắn token như mô hình cũ.

### 5.3 Composable `app/composables/useReminders.ts`

```ts
import type { Reminder, ReminderPayload } from '#shared/types'
import { apiListReminders, apiCreateReminder, apiUpdateReminder, apiDeleteReminder } from '~/api/reminders'

export function useReminders() {
  const reminders = useState<Reminder[]>('reminders', () => [])
  const loading = ref(false)

  async function load() {
    loading.value = true
    reminders.value = await apiListReminders()
    loading.value = false
  }

  async function create(data: ReminderPayload) {
    const r = await apiCreateReminder(data)
    await load()
    return r
  }

  async function update(id: string, data: ReminderPayload) {
    await apiUpdateReminder(id, data)
    await load()
  }

  async function remove(id: string) {
    await apiDeleteReminder(id)
    await load()
  }

  return { reminders, loading, load, create, update, remove }
}
```

### 5.4 Route guard `app/middleware/auth.ts`

```ts
export default defineNuxtRouteMiddleware(async () => {
  const { loggedIn, ready, fetch } = useUserSession()
  // Chờ session nạp xong trước khi check — tránh bị đá về /login nhầm
  // (lúc SSR/mới vào trang, session chưa chắc đã sẵn sàng).
  if (!ready.value) await fetch()
  if (!loggedIn.value) return navigateTo('/login')
})
```

### 5.5 Trang login `app/pages/login.vue`

```vue
<script setup lang="ts">
import { apiLogin, apiRegister } from '~/api/auth'

const { fetch: fetchSession } = useUserSession()
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit(action: 'login' | 'register') {
  error.value = ''
  loading.value = true
  try {
    if (action === 'login') await apiLogin(email.value, password.value)
    else await apiRegister(email.value, password.value)
    await fetchSession()      // cập nhật session sau khi login/register
    await navigateTo('/')
  } catch (e: any) {
    error.value = e?.data?.message ?? 'Có lỗi xảy ra'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main style="max-width: 320px; margin: 4rem auto; font-family: sans-serif">
    <h1>SenNote</h1>
    <form @submit.prevent="submit('login')">
      <input v-model="email" type="email" placeholder="Email" required style="display:block;width:100%;margin-bottom:8px;padding:8px">
      <input v-model="password" type="password" placeholder="Mật khẩu (≥ 8 ký tự)" required style="display:block;width:100%;margin-bottom:8px;padding:8px">
      <button :disabled="loading" style="width:100%;padding:8px">Đăng nhập</button>
      <button type="button" :disabled="loading" @click="submit('register')" style="width:100%;margin-top:6px;padding:8px">Đăng ký</button>
    </form>
    <p v-if="error" style="color:red">{{ error }}</p>
  </main>
</template>
```

### 5.6 Trang chủ `app/pages/index.vue`

```vue
<script setup lang="ts">
import { useUserSession } from '#imports'

definePageMeta({ middleware: 'auth' })

const { user, clear } = useUserSession()
const { reminders, loading, load, remove } = useReminders()

onMounted(load)

const title = ref('')
async function add() {
  if (!title.value.trim()) return
  await useReminders().create({ title: title.value.trim(), type: 'one_time' })
  title.value = ''
}
</script>

<template>
  <main style="max-width: 640px; margin: 2rem auto; font-family: sans-serif">
    <header style="display:flex;justify-content:space-between;align-items:center">
      <h1>Reminder của {{ user?.email }}</h1>
      <button @click="clear()">Đăng xuất</button>
    </header>

    <form @submit.prevent="add" style="margin: 1rem 0">
      <input v-model="title" placeholder="Tiêu đề nhắc nhở mới" style="padding:8px;width:70%">
      <button type="submit" style="padding:8px">Thêm</button>
    </form>

    <p v-if="loading">Đang tải...</p>
    <p v-else-if="reminders.length === 0">Chưa có reminder nào.</p>

    <ul style="padding:0;list-style:none">
      <li v-for="r in reminders" :key="r.id"
          style="border:1px solid #ddd;border-radius:8px;padding:12px;margin-bottom:8px;display:flex;justify-content:space-between">
        <div>
          <strong>{{ r.title }}</strong>
          <span style="margin-left:8px;color:#888">[{{ r.status }}]</span>
        </div>
        <button @click="remove(r.id)" style="color:red">Xoá</button>
      </li>
    </ul>
  </main>
</template>
```

✅ **Checkpoint tổng:**
1. Mở http://localhost:3000 → bị đá sang `/login` (guard hoạt động).
2. Đăng ký tài khoản mới → vào trang chủ → thêm vài reminder → thấy trong danh sách.
3. F5 vẫn đăng nhập (session cookie). Đăng xuất → bị đá về login.
4. Xoá 1 reminder → danh sách cập nhật.
5. F12 → Network: request đi tới `/api/...` (cùng origin), không cần token thủ công.

## Lỗi thường gặp

| Triệu chứng | Nguyên nhân | Cách sửa |
|-------------|-------------|----------|
| `/api/ping` lỗi | Chưa xong doc 02 (Turso) | Kiểm tra lại `.env` (đúng tên `NUXT_TURSO_*`) |
| Register báo lỗi ở `hashPassword` | Thiếu `NUXT_SESSION_PASSWORD` | Đặt chuỗi ≥ 32 ký tự, restart |
| 401 khi gọi `/api/reminders` | Chưa login / session mất | Login lại; F12 xem cookie `nuxt-session` có không |
| `crypto.randomUUID` lỗi | Chạy môi trường cũ | Node ≥ 20 là có sẵn; trên Workers cũng hỗ trợ |
| SQL lỗi cú pháp | Viết sai câu SQL | Chạy thử câu đó trong web console Turso để bắt lỗi |
| TS lỗi `Property 'id' does not exist on type 'User'` | Thiếu `auth.d.ts` | Tạo `shared/types/auth.d.ts` như Bước 2 |

## Ghi chú: migrate dữ liệu từ backend cũ

Khi app mới chạy ổn, nếu cần mang dữ liệu từ PocketBase cũ sang (`E:\PROJECT\nhacviecfcm`):
1. Export collection **`musers`** — ⚠️ tên collection là `musers`, **không phải `users`**
   (`v2docs/API_EXAMPLE_V2.md:37` trong project Go ghi sai).
   Sinh `id` mới, hash lại password (PocketBase dùng bcrypt, ta dùng PBKDF2-SHA256
   → **không dùng lại hash cũ**, phải yêu cầu user đặt lại mật khẩu hoặc migrate hash).
2. Export collection `reminders` → copy nguyên các field, `recurrence_pattern` vốn đã là JSON string.
   ⚠️ Đổi tên cột: `created` → `created_at`, `updated` → `updated_at` (không phải "giữ nguyên").
   ⚠️ Backfill `origin_time`: với `one_time` lấy `next_action_at`; với `recurring` lấy
   `next_action_at` (hoặc `created` nếu `next_action_at` null) — nếu bỏ trống, cron sẽ báo lỗi
   `origin_time required` cho mọi reminder recurring.
   ⚠️ Backfill `is_sended_one_time`: lấy từ PocketBase, mặc định 0.
3. Dùng 1 script tạm (`scripts/migrate.ts`) đọc dump + `insert` vào Turso qua `@libsql/client`.
4. ⚠️ **Reminder âm lịch cũ không migrate được nguyên trạng** — `lunar_data.go` của Go chỉ có dữ liệu
   2024–2025. Cần tính lại `origin_time` theo thuật toán âm lịch tổng quát (doc 04b mục 5).
5. Import xong nhớ `UPDATE system_status SET worker_enabled = 1` (mặc định schema mới là 1, nhưng
   bản Go seed sẵn là 0).

→ Tiếp theo: [04 — Lịch nhắc định kỳ + gửi FCM](./04-lich-nhac-va-fcm.md)
