# 04 — Lịch nhắc định kỳ + gửi push FCM

> Mục tiêu: cứ mỗi phút, hệ thống tự quét reminder đến hạn, gửi push FCM về điện thoại và cập nhật lịch lặp (FRP/CRP).
> Yêu cầu: đã xong [doc 03](./03-database-va-api.md).
>
> ⚠️ **Đặc tả tính lịch (11 nhánh worker, âm lịch, bộ giá trị chuẩn) nằm ở [doc 04b](./04b-tinh-lich-frp-crp.md).**
> Doc này chỉ lo phần **cài đặt**: task, FCM, kill-switch, route.

---

## 1. Mô hình hoạt động

```
Cron Triggers (Cloudflare) mỗi phút
        │  bắn task "reminders:check"
        ▼
server/tasks/reminders-check.ts
        │  0. check system_status.worker_enabled  (kill-switch)
        │  1. query reminders đến hạn
        │  2. phân nhánh A1–C5 (doc 04b mục 7)
        │  3. gửi FCM → cập nhật trạng thái + next_action_at
        ▼
server/utils/schedule.ts           (tính lịch — port từ Go, doc 04b)
server/utils/calendar.ts           (âm lịch)
server/utils/fcm.ts                (gửi push — HTTP v1)
```

- **Cron đã khai báo ở doc 02** (`scheduledTasks: { '* * * * *': 'reminders:check' }`).
- **Giờ chạy theo UTC**: mọi thời gian trong DB lưu ISO 8601 UTC nên so sánh chuỗi trực tiếp được.

### ⚠️ Giới hạn: cron 1 phút vs worker Go 5 giây

Worker Go chạy mỗi `WORKER_INTERVAL` giây (**mặc định 10, đang cấu hình 5** trong `.env` của project cũ). Cloudflare Cron Triggers **nhanh nhất là 1 phút**. Hệ quả:

| Trường hợp | Ảnh hưởng |
|---|---|
| `interval_seconds` < 120 | Lệch đáng kể (vd 60s có thể thành 60–120s) |
| `crp_interval_sec` < 120 | Retry CRP bị giãn |
| `daily` / `weekly` / `monthly` | ✅ Không ảnh hưởng (độ phân giải ngày) |

Khắc phục nếu cần độ chính xác cao: chuyển `interval_seconds` ngắn sang Durable Objects Alarm hoặc Queues. Với app nhắc việc thông thường, **1 phút là đủ**.

---

## 2. Kill-switch `system_status`

Port từ Go (`internal/worker/common.go:14-21, 68-89`). Bảng đã tạo ở doc 03.

```ts
// server/utils/system-status.ts
export async function isWorkerEnabled(): Promise<boolean> {
  const db = useDb()
  const rs = await db.execute({
    sql: 'SELECT worker_enabled FROM system_status WHERE mid = 1',
  })
  return Number(rs.rows[0]?.worker_enabled) === 1
}

export async function disableWorker(errMsg: string): Promise<void> {
  const db = useDb()
  const now = new Date().toISOString()
  await db.execute({
    sql: 'UPDATE system_status SET worker_enabled = 0, last_error = ?, updated_at = ? WHERE mid = 1',
    args: [errMsg, now],
  })
}
```

### Phân biệt 2 loại lỗi — BẮT BUỘC

| Lỗi | Hành động |
|---|---|
| User không tồn tại / không có `fcm_token` / `is_fcm_active = 0` | Chỉ log, **GIỮ `worker_enabled`**, bỏ qua reminder |
| FCM trả lỗi thật (401/403/5xx/timeout) | Ghi `last_error` rồi **`worker_enabled = 0`** |

> ⚠️ Đừng tắt worker vì lỗi của 1 user — bên Go đúng là chỉ tắt với lỗi FCM thật (`common.go:68-89`).
> ⚠️ Không có cơ chế tự bật lại. Phải bật thủ công qua `PUT /api/system_status`.

### Route quản trị

`server/api/system_status.get.ts` và `server/api/system_status.put.ts`:

```ts
// .put.ts
export default defineEventHandler(async (event) => {
  await requireUserSession(event)                 // ⚠️ Go đang để hở, bản mới BẮT BUỘC auth
  const { worker_enabled, last_error } = await readBody(event)
  const db = useDb()
  const now = new Date().toISOString()
  await db.execute({
    sql: 'UPDATE system_status SET worker_enabled = ?, last_error = ?, updated_at = ? WHERE mid = 1',
    args: [worker_enabled ? 1 : 0, last_error ?? '', now],
  })
  return { ok: true }
})
```

> ⚠️ Go dùng Go zero-value nên `PUT {}` sẽ **tắt worker** (`system_status_handler.go:88-90`). Bản TS đọc tường minh `worker_enabled` → không bị lỗi này.

---

## 3. Gửi FCM bằng HTTP v1 (không dùng firebase-admin)

`firebase-admin` cần Node API (filesystem, crypto node) → **không chạy trên Cloudflare Workers**. Gọi thẳng REST API của Firebase bằng Web Crypto.

### 3.1 Lấy thông tin service account

1. Firebase Console → Project → ⚙️ **Project settings** → **Service accounts** → **Generate new private key**.
2. Lấy 3 giá trị vào `.env` (**NHỚ prefix `NUXT_`**):
   - `NUXT_FCM_CLIENT_EMAIL` = `client_email`
   - `NUXT_FCM_PRIVATE_KEY` = `private_key`
   - `NUXT_FCM_PROJECT_ID` = `project_id`

> ⚠️ **Cách đặt `NUXT_FCM_PRIVATE_KEY`:** key chứa `\n`. Phải đặt **trong ngoặc kép** để dotenv chuyển `\n` thành xuống dòng thật:
> ```env
> NUXT_FCM_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkq...\n-----END PRIVATE KEY-----\n"
> ```

### 3.1.1 Web có nhận được push không? — CÓ, và không cần sửa phần gửi

FCM **không phân biệt nền tảng**. Trình duyệt đăng ký qua Firebase JS SDK sẽ nhận một
FCM registration token, gửi về **đúng cùng endpoint** đang dùng:

```
POST /v1/projects/{id}/messages:send   { message: { token, notification: {...} } }
```

Token của Android và của Chrome khác nhau nhưng FCM tự phân giải → `sendFcm()` dùng chung.

Chỉ có 2 điểm khác:

| | Cần |
|---|---|
| Android / iOS | `notification: {title, body}` — đủ |
| Web | muốn bấm mở link thì thêm `webpush.fcmOptions.link` |

Gửi cả hai cùng lúc: Android/iOS đọc `notification`, web đọc thêm `webpush`, mỗi bên bỏ qua phần của bên kia.

### 3.2 `server/utils/fcm.ts`

Chữ ký hàm (đã thêm tham số `link` cho web):

```ts
export async function sendFcm(
  token: string,
  title: string,
  body?: string | null,
  link?: string,          // web: bấm vào thông báo thì mở link này
): Promise<FcmResult>
```

```ts
const TOKEN_URI = 'https://oauth2.googleapis.com/token'

function pemToArrayBuffer(pem: string): ArrayBuffer {
  const b64 = pem
    .replace(/-----BEGIN PRIVATE KEY-----/, '')
    .replace(/-----END PRIVATE KEY-----/, '')
    .replace(/\\n/g, '')   // nếu còn literal "\n" do .env không unescape
    .replace(/\s/g, '')    // xoá khoảng trắng / xuống dòng thật
  const bin = atob(b64)
  const buf = new Uint8Array(bin.length)
  for (let i = 0; i < bin.length; i++) buf[i] = bin.charCodeAt(i)
  return buf.buffer
}

function b64url(bytes: Uint8Array): string {
  let s = ''
  for (const b of bytes) s += String.fromCharCode(b)
  return btoa(s).replace(/=+$/, '').replace(/\+/g, '-').replace(/\//g, '_')
}

function encObj(o: object): string {
  return btoa(JSON.stringify(o)).replace(/=+$/, '').replace(/\+/g, '-').replace(/\//g, '_')
}

// Lấy access token OAuth2 bằng JWT (RS256) ký bằng Web Crypto
async function getAccessToken(): Promise<string> {
  const cfg = useRuntimeConfig()
  const now = Math.floor(Date.now() / 1000)
  const header = { alg: 'RS256', typ: 'JWT' }
  const claims = {
    iss: cfg.fcmClientEmail,
    scope: 'https://www.googleapis.com/auth/firebase.messaging',
    aud: TOKEN_URI,
    iat: now,
    exp: now + 3600,
  }
  const unsigned = `${encObj(header)}.${encObj(claims)}`

  const key = await crypto.subtle.importKey(
    'pkcs8',
    pemToArrayBuffer(cfg.fcmPrivateKey),
    { name: 'RSASSA-PKCS1-v1_5', hash: 'SHA-256' },
    false,
    ['sign'],
  )
  const sig = await crypto.subtle.sign('RSASSA-PKCS1-v1_5', key, new TextEncoder().encode(unsigned))
  const assertion = `${unsigned}.${b64url(new Uint8Array(sig))}`

  const res = await fetch(TOKEN_URI, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'urn:ietf:params:oauth:grant-type:jwt-bearer',
      assertion,
    }),
  })
  const data = await res.json()
  return data.access_token as string
}

export type FcmResult =
  | { ok: true }
  | { ok: false; unregistered: boolean; message: string }

export async function sendFcm(
  token: string,
  title: string,
  body?: string | null,
  link?: string,
): Promise<FcmResult> {
  const cfg = useRuntimeConfig()
  const accessToken = await getAccessToken()

  const message: Record<string, unknown> = {
    token,
    notification: { title, body: body ?? '' },
  }
  if (link) {
    message.webpush = { fcmOptions: { link } }
  }

  const res = await fetch(
    `https://fcm.googleapis.com/v1/projects/${cfg.fcmProjectId}/messages:send`,
    {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ message }),
    },
  )

  if (res.ok) return { ok: true, unregistered: false, message: '' }

  // Phân biệt token chết (dọn được) và lỗi hệ thống (tắt worker)
  const text = await res.text()
  const unregistered = /UNREGISTERED|NOT_FOUND|INVALID_ARGUMENT/.test(text)
  return { ok: false, unregistered, message: text.slice(0, 500) }
}
```

> ℹ️ Bên Go **không hề dọn token chết** (`DisableFCM`/`SetFCMError` có nhưng không bao giờ được gọi). Bản TS trả thêm `unregistered` để có thể xử lý (mục 4, bước 4).

---

## 4. `server/utils/reminder-processor.ts`

Xử lý **1** reminder: gửi tới **mọi** thiết bị của user, rồi cập nhật trạng thái.
Được gọi chung bởi task (và API nếu cần).

```ts
/**
 * Link mở khi user bấm vào thông báo trên web (Android/iOS bỏ qua trường này).
 * Hiện chưa có trang chi tiết reminder → trỏ về trang chủ.
 * Khi thêm `app/pages/reminders/[id].vue` thì đổi thành `/reminders/${id}`.
 */
function webLink(): string {
  return '/'
}

/** Công tắc tổng trên bảng users */
async function isMasterSwitchOn(userId: string): Promise<boolean> {
  const db = useDb()
  const rs = await db.execute({
    sql: 'SELECT is_fcm_active FROM users WHERE id = ?',
    args: [userId],
  })
  // Không có row → coi như bật (tránh mất thông báo do thiếu dữ liệu)
  if (rs.rows.length === 0) return true
  return Number(rs.rows[0]?.is_fcm_active) === 1
}

export async function processReminder(r: Reminder): Promise<void> {
  const now = new Date()

  // 1. Công tắc tổng: user tắt hẳn thông báo → bỏ qua
  const masterOn = await isMasterSwitchOn(r.user_id)
  if (!masterOn) {
    console.warn(`[reminder] bỏ qua ${r.id}: user đã tắt thông báo`)
    return
  }

  // 2. Lấy TẤT CẢ thiết bị đang hoạt động của user
  const targets = await listActiveTokens(r.user_id)
  if (targets.length === 0) {
    console.warn(`[reminder] bỏ qua ${r.id}: user chưa đăng ký thiết bị nào`)
    return
  }

  // 3. Gửi từng thiết bị. Chỉ cần 1 thiết bị nhận được là tính thành công.
  let anyOk = false

  for (const target of targets) {
    const link = target.platform === 'web' ? webLink() : undefined
    const res = await sendFcm(target.token, r.title, r.description, link)

    if (res.ok) {
      anyOk = true
      continue
    }

    if (res.unregistered) {
      // 4. Token chết → CHỈ tắt đúng token đó (bản cũ tắt cả user)
      await deactivateToken(target.token)
      console.warn(`[reminder] token hết hạn (${target.platform}), đã tắt`)
      continue
    }

    // 5. Lỗi hệ thống FCM → kill-switch (không phải lỗi do user)
    await disableWorker(`FCM send failed for reminder ${r.id}: ${res.message}`)
    return
  }

  if (!anyOk) {
    console.warn(`[reminder] ${r.id}: không thiết bị nào nhận được (tất cả token đều chết)`)
    return
  }

  // 6. Gửi thành công ít nhất 1 thiết bị → áp dụng đúng nhánh A1–C5
  await applySent(r, now)
}
```

> ⚠️ **Khác với bản 1 token:** token hết hạn giờ chỉ tắt đúng thiết bị đó
> (`deactivateToken`), thay vì tắt thông báo của cả user. Xem [doc 03 mục devices](./03-database-va-api.md#vì-sao-cần-bảng-devices).

### `applySent` — phân nhánh theo [doc 04b mục 7](./04b-tinh-lich-frp-crp.md#7-11-nhánh-worker)

```ts
function applySent(r: Reminder, now: Date): Promise<unknown> {
  const nowIso = now.toISOString()
  const crpNext = new Date(now.getTime() + (r.crp_interval_sec || 60) * 1000).toISOString()

  // ---- Nhóm C: recurring + crp_until_complete ----
  if (r.type === 'recurring' && r.repeat_strategy === 'crp_until_complete') {
    if (r.max_crp > 0 && r.is_sended_one_time === 1 && r.crp_count < r.max_crp
        && isPast(r.next_crp, now) && !completedAfterLastSent(r)) {
      return patch(r.id, { crp_count: r.crp_count + 1, last_sent_at: nowIso, next_crp: crpNext })  // C5
    }
    if (r.max_crp > 0 && !r.is_sended_one_time) {
      return patch(r.id, { is_sended_one_time: 1, last_sent_at: nowIso, crp_count: 0, next_crp: crpNext })  // C3
    }
    if (r.max_crp > 0 && completedAfterLastSent(r)) {
      return patch(r.id, { last_sent_at: nowIso, crp_count: 0, next_crp: crpNext })  // C4
    }
    if (!r.is_sended_one_time) {
      return patch(r.id, { is_sended_one_time: 1, last_sent_at: nowIso })  // C1
    }
    return patch(r.id, { last_sent_at: nowIso })  // C2
  }

  // ---- Nhóm B: recurring + none ----
  if (r.type === 'recurring') {
    if (r.max_crp > 0 && isPast(r.next_recurring, now)) {
      const next = nextRecurring(r, now).toISOString()   // B2
      return patch(r.id, { last_sent_at: nowIso, next_recurring: next, crp_count: 0, next_crp: crpNext, next_action_at: crpNext })
    }
    if (r.max_crp > 0 && r.crp_count < r.max_crp && isPast(r.next_crp, now)) {
      return patch(r.id, { crp_count: r.crp_count + 1, last_sent_at: nowIso, next_crp: crpNext, next_action_at: crpNext })  // B3
    }
    const next = nextRecurring(r, now).toISOString()     // B1
    return patch(r.id, { last_sent_at: nowIso, next_recurring: next, next_action_at: next })
  }

  // ---- Nhóm A: one_time ----
  if (r.max_crp > 0 && !r.is_sended_one_time) {          // A2
    return patch(r.id, { is_sended_one_time: 1, last_sent_at: nowIso, next_crp: crpNext, next_action_at: crpNext })
  }
  if (r.max_crp > 0 && r.is_sended_one_time) {           // A3
    const crpCount = r.crp_count + 1
    const done = crpCount >= r.max_crp
    // ⚠️ KHÔNG truyền key nào có giá trị undefined — spread sẽ ghi đè thành NULL.
    //    Phải dựng object điều kiện như bên dưới.
    return patch(r.id, {
      crp_count: crpCount,
      last_sent_at: nowIso,
      next_crp: done ? null : crpNext,
      ...(done
        ? { status: 'completed' as const, last_completed_at: nowIso }
        : { next_action_at: crpNext }),
    })
  }
  return patch(r.id, {                                    // A1
    is_sended_one_time: 1, status: 'completed',
    last_sent_at: nowIso, last_completed_at: nowIso, next_action_at: null,
  })
}
```

Các helper dùng trong `applySent` (đặt cùng file):

```ts
import { nextActionAt, nextRecurring } from './schedule'

// iso có phải thời điểm trong quá khứ (hoặc đã tới) không; null → false
function isPast(iso: string | null | undefined, now: Date): boolean {
  if (!iso) return false
  return Date.parse(iso) <= now.getTime()
}

// user đã bấm Hoàn thành SAU lần gửi gần nhất chưa? (điều kiện nhánh C2/C4)
function completedAfterLastSent(r: Reminder): boolean {
  if (!r.last_completed_at) return false
  if (!r.last_sent_at) return true
  return Date.parse(r.last_completed_at) > Date.parse(r.last_sent_at)
}

// Ghi 1 lô thay đổi + TỰ TÍNH LẠI next_action_at
async function patch(id: string, changes: Partial<Reminder>) {
  const db = useDb()
  const now = new Date()
  const rs = await db.execute({ sql: 'SELECT * FROM reminders WHERE id = ?', args: [id] })
  const merged = { ...(rs.rows[0] as any), ...changes } as Reminder

  merged.next_action_at = nextActionAt(merged, now)
  merged.updated_at = now.toISOString()

  const keys = Object.keys(changes).concat(['next_action_at', 'updated_at'])
  const setSql = keys.map((k) => `${k} = ?`).join(', ')
  const args = keys.map((k) => (merged as any)[k] ?? null)

  await db.execute({
    sql: `UPDATE reminders SET ${setSql} WHERE id = ?`,
    args: [...args, id],
  })
}
```

> ⚠️ Khác với Go: Go chỉ cập nhật `next_action_at` ở một số nhánh (nhánh UT không cập nhật). Bản TS **cập nhật ở mọi nhánh** để chỉ cần 1 query — đơn giản hơn và không đổi hành vi, nhờ `nextActionAt` đã bỏ qua snooze hết hạn.
>
> ⚠️ `is_sended_one_time` là INTEGER 0/1 trong Turso → so sánh bằng `=== 1` hoặc truthy/falsy, đừng so `=== true`.

---

## 5. Task cron `server/tasks/reminders/check.ts`

> 🚨 **Đọc [doc 05 mục 4](./05-deploy-cloudflare-pages.md#4--cron-pages-không-tự-chạy--phải-dùng-worker-riêng) trước.**
> **Cloudflare Pages không hỗ trợ Cron Triggers** → task này **không tự chạy trên production**.
> Nó chỉ chạy khi dev (`/_nitro/tasks/reminders:check`) hoặc khi được gọi qua
> `POST /api/cron/reminders-check` do Worker `sennote-cron` gọi mỗi phút.
>
> Vì vậy logic được tách ra `server/utils/run-reminder-check.ts` để cả 2 đường cùng dùng,
> không viết 2 lần.

> ⚠️ **Đường dẫn file quyết định tên task.** Nitro đổi `/` thành `:`, nên
> `server/tasks/reminders/check.ts` → `reminders:check`.
> Đặt nhầm thành `server/tasks/reminders-check.ts` → tên `reminders-check`, và log sẽ báo
> `Scheduled task reminders:check is not defined!` + `/_nitro/tasks/reminders:check` trả 404.
>
> Kiểm tra nhanh: mở `http://localhost:3000/_nitro/tasks` phải thấy
> `"tasks": { "reminders:check": {...} }` và `scheduledTasks: [{ cron: "* * * * *", ... }]`.

```ts
export default defineTask({
  meta: { name: 'reminders:check', description: 'Quét reminder đến hạn và gửi FCM' },
  async run() {
    // 0. Kill-switch
    if (!(await isWorkerEnabled())) return { result: 'worker đang tắt' }

    const db = useDb()
    const now = new Date()
    const nowIso = now.toISOString()

    // 1. Lấy lô reminder đến hạn (CÓ LIMIT — Cloudflare có giới hạn thời gian)
    const rs = await db.execute({
      sql: `SELECT * FROM reminders
             WHERE status = 'active'
               AND next_action_at IS NOT NULL
               AND next_action_at <= ?
               AND (snooze_until IS NULL OR snooze_until <= ?)
             ORDER BY next_action_at ASC
             LIMIT 100`,
      args: [nowIso, nowIso],
    })

    let sent = 0
    let failed = 0
    for (const row of rs.rows) {
      try {
        await processReminder(row as unknown as Reminder)
        sent++
      } catch (e) {
        failed++
        console.error(`[reminder] lỗi ${(row as any).id}:`, e)
      }
    }
    return { result: `Đã xử lý ${sent} reminder, ${failed} lỗi` }
  },
})
```

> ⚠️ Go **không có `LIMIT`** và load toàn bộ bảng mỗi tick. Trên Workers, backlog lớn sẽ timeout → thêm `LIMIT` + `ORDER BY next_action_at ASC`, các reminder còn lại chạy ở phút sau.
> ⚠️ Chưa có khoá chống chạy trùng. Cron overlap trên Cloudflare hiếm nhưng có thể xảy ra; nếu cần, thêm cột `locked_at` hoặc đẩy qua Queue.

---

## 6. Route `snooze` và `complete`

`server/api/reminders/[id]/snooze.post.ts`:

```ts
export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  const id = getRouterParam(event, 'id')!
  const { duration } = await readBody(event)          // giây
  const seconds = Number(duration)
  if (!Number.isFinite(seconds) || seconds <= 0 || seconds > 7 * 86400) {
    throw createError({ statusCode: 400, message: 'duration phải từ 1 giây đến 7 ngày' })
  }
  const r = await snoozeReminder(id, session.user.id, seconds)
  if (!r) throw createError({ statusCode: 404, message: 'Không tìm thấy reminder' })
  return r
})
```

`server/api/reminders/[id]/complete.post.ts`:

```ts
export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  const id = getRouterParam(event, 'id')!
  const r = await completeReminder(id, session.user.id)
  if (!r) throw createError({ statusCode: 404, message: 'Không tìm thấy reminder' })
  return r
})
```

Service (`server/utils/reminders.ts`) — **nhớ kèm `user_id` chống IDOR**:

```ts
export async function snoozeReminder(id: string, userId: string, seconds: number) {
  const db = useDb()
  const now = new Date()
  const until = new Date(now.getTime() + seconds * 1000).toISOString()
  const rs = await db.execute({
    sql: `UPDATE reminders SET snooze_until = ?, next_action_at = ?, updated_at = ?
          WHERE id = ? AND user_id = ?`,
    args: [until, until, now.toISOString(), id, userId],
  })
  if (rs.rowsAffected === 0) return null
  return getReminder(id, userId)
}

export async function completeReminder(id: string, userId: string) {
  const db = useDb()
  const r = await getReminder(id, userId)
  if (!r) return null
  const now = new Date()
  const changes = applyComplete(r, now)      // xem doc 04b mục 8
  const rs = await db.execute({
    sql: `UPDATE reminders SET status = COALESCE(?, status), last_completed_at = ?,
            last_crp_completed_at = ?, crp_count = ?, next_recurring = ?, next_crp = ?,
            next_action_at = ?, updated_at = ?
          WHERE id = ? AND user_id = ?`,
    args: [
      changes.status ?? null, changes.last_completed_at ?? null,
      changes.last_crp_completed_at ?? null, changes.crp_count ?? null,
      changes.next_recurring ?? null, changes.next_crp ?? null,
      changes.next_action_at ?? null, now.toISOString(), id, userId,
    ],
  })
  if (rs.rowsAffected === 0) return null
  return getReminder(id, userId)
}
```

---

## 6.5 Thiết lập Web Push (trình duyệt)

### Bước 1 — Lấy cấu hình Firebase cho web

1. Firebase Console → **Project settings → General → Your apps** → thêm app **Web** (`</>`).
2. Copy 5 giá trị vào `.env` (đều có prefix `NUXT_PUBLIC_`).
3. Vào **Project settings → Cloud Messaging → Web push certificates** → **Generate key pair**
   → copy giá trị VAPID vào `NUXT_PUBLIC_FCM_VAPID_KEY`.

> ⚠️ Các giá trị `NUXT_PUBLIC_*` được **gửi xuống trình duyệt** — điều này bình thường và an toàn
> với Firebase web config. **KHÔNG BAO GIỜ** cho `NUXT_FCM_PRIVATE_KEY` (server) vào đây.

### Bước 2 — Service Worker

`public/firebase-messaging-sw.js` đã có sẵn. Hai điểm cần nhớ:

- Phải nằm ở **gốc domain** (`/firebase-messaging-sw.js`), không được trong `/_nuxt/`.
- Không dùng `import` (trình duyệt cũ không hỗ trợ SW module) → dùng `importScripts`.
- Cấu hình được truyền qua **query string** khi đăng ký SW (vì file tĩnh không import được `runtimeConfig`).

### Bước 3 — Bật trên giao diện

Component `PushToggle.vue` + composable `usePush()` đã có sẵn, gắn sẵn ở header trang chủ:

```ts
const { state, error, enable, disable } = usePush()
await enable()   // xin quyền → lấy token → POST /api/devices { platform: 'web' }
```

Các trạng thái: `unsupported` → `default` → `granted` → `ready`.
Nếu `denied`, trình duyệt chặn hoàn toàn — user phải bật lại ở icon khoá trên thanh địa chỉ.

### Yêu cầu & giới hạn

| | |
|---|---|
| HTTPS | Bắt buộc. `localhost` được coi là an toàn → dev OK |
| Chrome / Edge / Firefox | Hoạt động bình thường |
| Safari macOS | Hoạt động từ 16.4+ |
| **Safari iOS** | ⚠️ Chỉ hoạt động với PWA **đã thêm vào màn hình chính** (16.4+) |
| Ẩn danh | Thường bị chặn Service Worker |

## 7. Test thủ công

1. **Chạy task:** khi `npm run dev` đang chạy:
   ```powershell
   Invoke-WebRequest http://localhost:3000/_nitro/tasks/reminders:check | Select-Object -ExpandProperty Content
   ```
   Không dùng `npx nitro task run` (đọc `.nitro/nitro.json` mà Nuxt không ghi).

2. **Kill-switch:** `UPDATE system_status SET worker_enabled = 0` → chạy task → trả "worker đang tắt".

3. **FCM:** cần token thật từ thiết bị:
   - Mobile: `PUT /api/users/fcm-token` (app cũ) hoặc `POST /api/devices` (app mới).
   - **Web:** đăng nhập → bấm nút **"Bật thông báo"** ở header → đồng ý quyền → token tự đăng ký.
   - Kiểm tra đã đăng ký: `GET /api/devices` (mở bằng trình duyệt đã đăng nhập).

4. **Lịch lặp:** xem checklist nghiệm thu ở [doc 04b mục 9](./04b-tinh-lich-frp-crp.md#9-checklist-nghiệm-thu).

## Lỗi thường gặp

| Triệu chứng | Nguyên nhân | Cách sửa |
|---|---|---|
| Task không chạy khi build | Thiếu `experimental.tasks` hoặc `scheduledTasks` đặt sai chỗ | `tasks` trong `nitro.experimental`, `scheduledTasks` TRONG `nitro` |
| Log báo `Scheduled task reminders:check is not defined!` | Đặt file thành `server/tasks/reminders-check.ts` | Đổi thành `server/tasks/reminders/check.ts` (có thư mục con) |
| `/_nitro/tasks/reminders:check` 404 | Sai đường dẫn file, hoặc dev đang dùng `cloudflare-dev` emulation | Xem dòng trên; đảm bảo `preset` chỉ đặt khi build (doc 02) |
| Dev log đầy lỗi nối Turso mỗi phút | Cron chạy THẬT trong dev | Điền credentials Turso thật, hoặc tắt tạm `PUT /api/system_status` |
| Task chạy nhưng không gửi gì | `worker_enabled = 0` | `UPDATE system_status SET worker_enabled = 1` |
| Task treo/không dứt | `interval <= 0` gây vòng lặp vô hạn | Dùng `normalizeInterval` (doc 04b mục 3.1) |
| Reminder bị gửi lặp mỗi phút | `nextActionAt` đưa `now` vào danh sách ứng viên | Xem doc 04b mục 3.9 |
| Snooze hết hạn mà vẫn bị gửi hoài | Lấy `min` với `snooze_until` đã qua | Xem doc 04b mục 3.9 |
| Recurring chết sau 1 lần complete | Set `status='completed'` cho recurring | Xem doc 04b mục 8 |
| Complete weekly bị gửi thêm 1 thông báo thừa | Thiếu `weekly` trong `nextRecurringFromComplete` | Xem doc 04b mục 8 |
| Âm lịch ném `vượt quá phạm vi bảng âm lịch` | Bảng `lunar-table.ts` hết hạn (hiện phủ 2026-08 → 2029-08) | Chạy lại `go run ./cmd/gen_lunar_table` (xem AGENTS.md) |
| FCM 403 | JWT/scope sai hoặc key sai format | Kiểm tra 3 biến `NUXT_FCM_*`, key phải trong ngoặc kép |
| FCM 404 | `project_id` sai | Copy đúng `project_id` trong service account |
| Lỗi ở `importKey` | Private key sai format | Đặt key trong ngoặc kép để `.env` chuyển `\n` thật |
| Giờ nhắc lệch 7 tiếng | Nhầm múi giờ | Luôn lưu/tính bằng UTC; chỉ đổi sang giờ VN khi hiển thị |

→ Tiếp theo: [05 — Deploy lên Cloudflare Pages](./05-deploy-cloudflare-pages.md)
