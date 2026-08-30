# 04b — Đặc tả tính lịch FRP/CRP (port từ Go)

> File này là **đặc tả kỹ thuật** để port logic tính lịch + worker từ backend Go sang TypeScript.
> Nó là nguồn sự thật cho `server/utils/schedule.ts`, `server/utils/calendar.ts` và `server/utils/reminder-processor.ts`.
> Yêu cầu: đã đọc [doc 03](./03-database-va-api.md). Cài đặt task/FCM nằm ở [doc 04](./04-lich-nhac-va-fcm.md).

---

## 0. Nguồn sự thật — đọc cái này, ĐỪNG đọc mấy file kia

| File trong project Go `E:\PROJECT\nhacviecfcm` | Tình trạng |
|---|---|
| `internal/services/calNextTime.go` | ✅ **Nguồn sự thật** — engine tính lịch của worker |
| `internal/services/calNextTime_test.go` | ✅ 22 test case, lấy làm bộ giá trị chuẩn (mục 6) |
| `internal/services/lunar_calendar.go` | ✅ Thuật toán âm lịch tổng quát (dùng cái này) |
| `internal/services/lunar_data.go` | ⚠️ Bảng tĩnh **chỉ phủ 2024–2025** (xem mục 5) |
| `internal/worker/worker_loop_noUT.go`, `worker_loop_UT.go`, `worker_onetime_v2.go` | ✅ 11 nhánh worker (mục 7) |
| `internal/services/reminder_service.go` (`OnUserComplete`) | ✅ Logic bấm Hoàn thành (mục 8) |
| `docs/WORKER_LOGIC.md` | ❌ **DEPRECATED** — sai chu kỳ, sai query, sai xử lý lỗi |
| `v2docs/caculate_next_recurring.md` | ❌ **DEPRECATED** — mô tả code đã bị comment out |
| `internal/services/schedule_calculator.go` | ⚠️ **Engine thứ 2, KHÔNG dùng** — xem mục 1 |

### Quyết định đã chốt

1. **Engine chuẩn = `calNextTime.go`** (engine của worker).
2. **Port đầy đủ** `repeat_strategy = 'crp_until_complete'` (11 nhánh).
3. **Giữ kill-switch** `system_status`.
4. `schedule_calculator.go` **không được port** — nó chỉ dùng ở đường create/update của Go và mâu thuẫn với worker engine.

---

## 1. Vì sao bỏ `schedule_calculator.go` (và hệ quả)

Go đang có **2 engine tính lịch mâu thuẫn nhau**. Ngay cả trong Go, tạo reminder và worker chạy reminder **dùng 2 thuật toán khác nhau** — đây là nguồn bug, bản port phải gom về 1.

| | `calNextTime.go` ✅ CHỌN | `schedule_calculator.go` ❌ BỎ |
|---|---|---|
| Mỏ neo (anchor) | **`origin_time`** | `next_recurring` |
| Giờ/phút lấy từ | **`origin_time`** | `trigger_time_of_day` |
| Thứ trong tuần | **`origin_time.getUTCDay()`** | `pattern.day_of_week` |
| `solar_last_day_of_month` | ✅ | ❌ lỗi `unsupported` |
| `lunar_last_day_of_month` | ❌ (ta sẽ bổ sung) | ✅ |
| `interval <= 0` | ❌ **treo vô hạn** (mục 3.1) | ✅ tự đổi thành 1 |

**Hệ quả bắt buộc phải xử lý khi chọn `calNextTime.go`:**

- Phải có cột **`origin_time`** (bắt buộc, xem mục 2).
- Phải **tự thêm guard `interval >= 1`** (Go đang treo).
- Phải **tự bổ sung** `weekly` + `lunar` cho đường "bấm Hoàn thành" (Go đang thiếu → mục 8).
- `trigger_time_of_day` / `day_of_week` / `day_of_month` **không còn là nguồn sự thật**. Vẫn giữ trong JSON để tương thích dữ liệu cũ, nhưng **chỉ dùng lúc tạo** để suy ra `origin_time`.

---

## 2. Các cột bắt buộc (nhắc lại từ doc 03)

```sql
origin_time        TEXT NOT NULL,   -- MỎ NEO MỌI PHÉP TÍNH LẶP
is_sended_one_time INTEGER NOT NULL DEFAULT 0,
repeat_strategy    TEXT NOT NULL DEFAULT 'none',
tag                TEXT,
```

- **`origin_time`**: mốc gốc của lịch lặp. Mọi lần tính sau này đều chạy từ đây (không chạy từ lần gửi gần nhất) → **lịch không bị trôi** dù cron chạy trễ.
- **`is_sended_one_time`**: phân biệt "gửi lần đầu" và "đang retry CRP". Thiếu cột này là mất nhánh C1–C5 và retry của one_time.
- Tạo reminder: nếu client không gửi `origin_time`, **server tự set = `next_action_at`** (với `recurring`) hoặc `= now` (với `one_time`).

---

## 3. `server/utils/schedule.ts` — các hàm thuần

Toàn bộ tính bằng **UTC** (`getUTCHours`…). Không dùng giờ local.

### 3.1 Guard chống treo — BẮT BUỘC

> 🐞 **Bug có thật bên Go**: `calcNextDailyTime` và `calcNextSolarMonthly` **không** kiểm tra `interval <= 0`.
> Test `TestCalcNextDaily_ZeroInterval` (`calNextTime_test.go:460`) đang **treo vô hạn** tại `calNextTime.go:141`.
> Vì `RecurrencePattern.interval` có `omitempty`, client không gửi `interval` → `0` → treo.
> `calcNextWeekly` có guard (`calNextTime.go:225-227`), `calcNextIntervalSeconds` có guard (`:77-79`) — chỉ daily/monthly là thiếu.

```ts
// Chặn vòng lặp vô hạn: interval <= 0 → coi như 1
export function normalizeInterval(n: number | undefined, fallback = 1): number {
  const v = Math.floor(n ?? 0)
  return v > 0 ? v : fallback
}

// Chặn treo Worker nếu dữ liệu sai (Cloudflare sẽ kill task nếu quá lâu)
const MAX_ITER = 10_000
function guard(i: number, where: string): void {
  if (i > MAX_ITER) throw new Error(`Vượt quá ${MAX_ITER} vòng lặp khi tính ${where}`)
}
```

### 3.2 Helper

```ts
function hhmmOf(origin: Date): { h: number; m: number } {
  return { h: origin.getUTCHours(), m: origin.getUTCMinutes() }
}

function daysInMonthUTC(year: number, monthIdx0: number): number {
  return new Date(Date.UTC(year, monthIdx0 + 1, 0)).getUTCDate()
}

function withTimeUTC(d: Date, h: number, m: number): Date {
  return new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth(), d.getUTCDate(), h, m, 0, 0))
}

function addDaysUTC(d: Date, n: number): Date {
  return new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth(), d.getUTCDate() + n,
    d.getUTCHours(), d.getUTCMinutes(), 0, 0))
}
```

### 3.3 `daily` — port `calcNextDailyTime` (`calNextTime.go:119`)

```ts
export function nextDaily(origin: Date, lastTime: Date, now: Date, rawInterval: number): Date {
  const interval = normalizeInterval(rawInterval)
  const { h, m } = hhmmOf(origin)
  let next = addDaysUTC(withTimeUTC(lastTime, h, m), interval)
  let i = 0
  while (next.getTime() <= now.getTime()) { guard(++i, 'daily'); next = addDaysUTC(next, interval) }
  return next
}
```

### 3.4 `weekly` — port `calcNextWeekly` (`calNextTime.go:213`)

> ⚠️ Thứ **lấy từ `origin_time`**, KHÔNG dùng `pattern.day_of_week` (Go ghi rõ ở comment dòng 212).

```ts
export function nextWeekly(origin: Date, lastTime: Date, now: Date, rawInterval: number): Date {
  const interval = normalizeInterval(rawInterval)
  const { h, m } = hhmmOf(origin)
  const target = origin.getUTCDay()          // 0 = Chủ nhật
  let next = withTimeUTC(lastTime, h, m)
  let daysAhead = (target - next.getUTCDay() + 7) % 7
  if (daysAhead === 0) {
    if (next.getTime() > now.getTime()) return next   // hôm nay đúng thứ, chưa tới giờ
    daysAhead = 7 * interval                          // đã qua giờ → tuần sau
  }
  next = addDaysUTC(next, daysAhead)
  let i = 0
  while (next.getTime() <= now.getTime()) { guard(++i, 'weekly'); next = addDaysUTC(next, 7 * interval) }
  return next
}
```

### 3.5 `monthly` (dương) — port `calcNextSolarMonthly` (`calNextTime.go:148`)

Điểm quan trọng: **normalize về ngày 1 trước khi cộng tháng** để tránh tràn ngày (vd 31/01 + 1 tháng). Đây là bug đã được fix ở commit `379bf7a`, bản port phải làm theo.

```ts
export function nextSolarMonthly(origin: Date, lastTime: Date, now: Date, rawInterval: number): Date {
  const interval = normalizeInterval(rawInterval)
  const { h, m } = hhmmOf(origin)
  const originDay = origin.getUTCDate()

  // Cộng tháng từ ngày 1, rồi kẹp ngày về min(originDay, số ngày tháng đó)
  const step = (from: Date): Date => {
    const y = from.getUTCFullYear()
    const mo = from.getUTCMonth()
    const tmp = new Date(Date.UTC(y, mo + interval, 1, h, m))
    const last = daysInMonthUTC(tmp.getUTCFullYear(), tmp.getUTCMonth())
    return new Date(Date.UTC(tmp.getUTCFullYear(), tmp.getUTCMonth(), Math.min(originDay, last), h, m))
  }

  let next = step(new Date(Date.UTC(lastTime.getUTCFullYear(), lastTime.getUTCMonth(), 1)))
  let i = 0
  while (next.getTime() <= now.getTime()) { guard(++i, 'monthly'); next = step(next) }
  return next
}
```

### 3.6 `solar_last_day_of_month` — port `calcNextSolarLastDayOfMonth` (`calNextTime.go:90`)

```ts
export function nextSolarLastDayOfMonth(now: Date): Date {
  const lastThis = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() + 1, 0,
    now.getUTCHours(), now.getUTCMinutes(), now.getUTCSeconds(), now.getUTCMilliseconds()))
  if (now.getUTCDate() < lastThis.getUTCDate()) return lastThis
  return new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() + 2, 0,
    now.getUTCHours(), now.getUTCMinutes(), now.getUTCSeconds(), now.getUTCMilliseconds()))
}
```

### 3.7 `interval_seconds` — port `calcNextIntervalSeconds` (`calNextTime.go:75`)

Dùng nhân chia thay vì vòng lặp để tránh treo khi hệ thống downtime dài:

```ts
export function nextIntervalSeconds(origin: Date, now: Date, intervalSeconds: number): Date {
  if (intervalSeconds <= 0) throw new Error('interval_seconds must be > 0')
  const stepMs = intervalSeconds * 1000
  const k = Math.floor((now.getTime() - origin.getTime()) / stepMs) + 1
  return new Date(origin.getTime() + k * stepMs)
}
```

### 3.8 Dispatcher — port `Tinhtoan_NextRecurringV2` (`calNextTime.go:34`)

```ts
export function nextRecurring(r: Reminder, now: Date): Date {
  const p = r.recurrence_pattern
  if (!p) throw new Error('recurrence_pattern required for recurring reminder')
  if (!r.origin_time) throw new Error('origin_time required')

  const origin = new Date(r.origin_time)

  // Âm lịch: calendar_type = 'lunar' thắng pattern.type (đúng theo Go)
  if (r.calendar_type === 'lunar') return nextLunarMonthly(origin, now)

  switch (p.type) {
    case 'interval_seconds':         return nextIntervalSeconds(origin, now, p.interval_seconds ?? 0)
    case 'daily':                    return nextDaily(origin, origin, now, p.interval)
    case 'weekly':                   return nextWeekly(origin, origin, now, p.interval)
    case 'monthly':                  return nextSolarMonthly(origin, origin, now, p.interval)
    case 'solar_last_day_of_month':  return nextSolarLastDayOfMonth(now)
    case 'lunar_last_day_of_month':  return nextLunarLastDayOfMonth(origin, now)  // MỞ RỘNG, xem mục 5
    default: throw new Error(`unsupported recurrence type: ${p.type}`)
  }
}
```

> ℹ️ Worker Go luôn truyền `lastTime = origin_time`. Vì vậy lịch **không bao giờ trôi**: gửi trễ 3 tiếng thì lần sau vẫn đúng mốc gốc, không bị dời.

### 3.9 `nextActionAt` — port `CalculateNextActionAt` (`schedule_calculator.go:29`)

> 🐞 **2 bug trong bản draft cũ của tài liệu** (sửa ở đây):
> 1. Đưa `now` vào danh sách ứng viên → `next_action_at` luôn ≤ hiện tại → reminder **bị xử lý lại mỗi phút vô hạn**.
> 2. Lấy `Math.min` với `snooze_until` **đã hết hạn** → kẹt trong quá khứ.
>
> Bên Go dùng `IsSnoozeUntilActive(now)` để **short-circuit**, và bỏ qua snooze đã qua (`schedule_calculator.go:33-35`).

```ts
export function nextActionAt(r: Reminder, now: Date): string | null {
  // 1. Snooze còn hiệu lực → ưu tiên tuyệt đối
  if (r.snooze_until && Date.parse(r.snooze_until) > now.getTime()) return r.snooze_until

  // 2. Ngược lại: min(next_recurring, next_crp) — snooze đã hết hạn thì BỎ QUA
  const candidates: number[] = []
  if (r.next_recurring) candidates.push(Date.parse(r.next_recurring))
  if (r.max_crp > 0 && r.next_crp) candidates.push(Date.parse(r.next_crp))
  if (candidates.length === 0) return null

  return new Date(Math.min(...candidates)).toISOString()
}
```

---

## 4. CRP (nhắc lại)

```ts
// Thời điểm retry kế tiếp. null = hết lượt retry.
export function nextCrp(r: Reminder, now: Date): string | null {
  if (r.max_crp <= 0) return null
  if (r.crp_count >= r.max_crp) return null
  return new Date(now.getTime() + r.crp_interval_sec * 1000).toISOString()
}
```

Quy ước: `max_crp = 0` → không retry. `crp_interval_sec <= 0` → fallback **60 giây** (như `worker_onetime_v2.go:156-160`).

---

## 5. `server/utils/calendar.ts` — âm lịch

### Cách làm đã chốt: sinh bảng từ code Go, KHÔNG copy `lunar_data.go`

`internal/services/lunar_data.go` chứa `SolarToLunarMap` **chỉ phủ 2024-01-01 → 2025-12-31**
(731 ngày) → đã hết hạn từ 2026. Không được copy bảng đó.

Thay vào đó, có generator **`E:\PROJECT\nhacviecfcm\cmd\gen_lunar_table`** chạy **thuật toán gốc**
(`internal/services/lunar_calendar.go` — Jean Meeus / Hồ Ngọc Đức, timezone +07) để sinh bảng mới:

```bash
cd E:\PROJECT\nhacviecfcm
go run ./cmd/gen_lunar_table -from 2026-08-01 -to 2029-08-31 \
  -out "E:\PROJECT\nhacviecfcm\nhacviecnuxt\server\utils\lunar-table.ts"
```

Kết quả hiện tại (đã có sẵn trong repo):

| | |
|---|---|
| `SOLAR_TO_LUNAR` | 1.127 ngày (2026-08-01 → 2029-08-31) |
| `LUNAR_TO_SOLAR` | 3.330 khoá (năm âm 2024 → 2032) |
| Năm nhuận phát hiện | âm 2025 nhuận tháng 6, 2028 nhuận tháng 5, 2031 nhuận tháng 3 |

**Mốc kiểm chứng đã đối chiếu đúng** (nằm trong `server/utils/schedule.test.ts`):

| Âm | Dương |
|---|---|
| 1/1/2026 | 2026-02-17 (Tết Bính Ngọ) |
| 1/1/2027 | 2027-02-06 (Tết Đinh Mùi) |
| 1/1/2028 | 2028-01-26 (Tết Mậu Thân) |
| 1/1/2029 | 2029-02-13 (Tết Kỷ Dậu) |
| 1/1/2030 | 2030-02-02 (Tết Canh Tuất) |

### ⚠️ Lỗi đã gặp khi sinh bảng (đừng lặp lại)

`ConvertLunar2Solar` **không từ chối tháng nhuận ở năm KHÔNG nhuận** — nó chỉ kiểm tra khi
`b11 - a11 > 365`. Hệ quả: gọi với `lunarLeap = 1` cho năm thường sẽ trả về **trùng ngày** với
tháng không nhuận (vd `2026-1-1-true` → `2026-02-17`, sai).
Generator xử lý bằng cách quét một khoảng rộng hơn để **phát hiện tháng nhuận thật sự**, rồi chỉ
sinh khoá `-true` cho đúng `(năm, tháng)` đó.

### API của `calendar.ts`

```ts
solarToLunar(date: Date): LunarDate | null   // null = ngoài phạm vi bảng
lunarToSolar(year, month, day, isLeapMonth?): Date | null  // null = ngày âm không tồn tại
daysInLunarMonth(year, month, isLeapMonth?): number        // 29 | 30 | 0
findNextLunarMonthly(now, lunarDay): Date | null
findNextLunarLastDayOfMonth(now): Date | null
lunarTableRange(): { from, to } | null
```

> ⚠️ Hết hạn bảng → `solarToLunar()` trả `null` → `nextLunarMonthly()` **ném lỗi có chủ đích**
> (thà báo lỗi còn hơn tính sai). Gia hạn bằng 1 lệnh `go run` ở trên.

### `nextLunarMonthly` — port `calcNextLunarMonthly` (`calNextTime.go:260`)

```ts
export function nextLunarMonthly(origin: Date, now: Date): Date {
  const vnOrigin = shiftToVN(origin)                 // +07:00, key "YYYY-MM-DD"
  const lunarOrigin = solarToLunar(vnOrigin)         // ngày âm của origin
  const { h, m } = hhmmOf(origin)                    // GIỜ/PHÚT lấy từ origin (UTC)

  const nextSolar = findNextLunarMonthly(now, lunarOrigin.day)  // ngày dương của ngày âm đó, tháng sau
  return new Date(Date.UTC(nextSolar.getUTCFullYear(), nextSolar.getUTCMonth(),
    nextSolar.getUTCDate(), h, m, 0, 0))
}
```

> ⚠️ Go đổi sang VN+07 chỉ để **tra ngày âm**, rồi ghép lại với giờ/phút của `origin` theo UTC (`calNextTime.go:267-289`). Port phải giữ nguyên sự "không nhất quán" này nếu muốn khớp dữ liệu cũ, hoặc quyết định chuẩn hoá (khuyến nghị: **đổi key theo UTC luôn** và ghi rõ trong ghi chú migrate).

### `nextLunarLastDayOfMonth` — **MỞ RỘNG** (Go chưa có ở worker engine)

`lunar_last_day_of_month` hiện chỉ tồn tại ở engine bị loại (`schedule_calculator.go:139`). Vì port đầy đủ, bản TS cần tự viết: tìm ngày âm 30 (hoặc ngày cuối tháng âm nếu tháng thiếu) của tháng âm kế tiếp rồi đổi sang dương.

---

## 6. Bộ giá trị chuẩn (golden tests) — lấy từ `calNextTime_test.go`

21/22 test đang pass. `TestCalcNextDaily_ZeroInterval` **treo** (xem mục 3.1). Tất cả dùng **UTC**.

| # | type | `origin_time` | interval | lastTime | now | **Kết quả mong đợi** |
|---|------|---------------|----------|----------|-----|----------------------|
| 1 | daily | 2025-01-01 08:00 | 1 | 2025-01-05 08:00 | 2025-01-06 07:00 | **2025-01-06 08:00** |
| 2 | daily | 2025-01-01 09:30 | 2 | 2025-01-05 09:30 | 2025-01-06 10:00 | **2025-01-07 09:30** |
| 3 | daily | 2025-01-01 10:00 | 1 | 2025-01-31 10:00 | 2025-02-01 09:00 | **2025-02-01 10:00** |
| 4 | weekly | 2025-01-03 10:00 (T6) | 1 | 2025-01-06 10:00 (T2) | 2025-01-06 11:00 | **2025-01-10 10:00** (T6) |
| 5 | weekly | 2025-01-06 14:00 (T2) | 2 | 2025-01-06 14:00 | 2025-01-07 10:00 | **2025-01-20 14:00** |
| 6 | weekly | 2025-01-05 09:00 (CN) | 1 | 2025-01-11 09:00 (T7) | 2025-01-11 10:00 | **2025-01-12 09:00** (CN) |
| 7 | monthly | 2025-01-15 10:30 | 1 | 2025-01-15 10:30 | 2025-02-01 12:00 | **2025-02-15 10:30** |
| 8 | monthly | 2025-01-31 08:00 | 1 | 2025-01-31 08:00 | 2025-02-01 09:00 | **2025-02-28 08:00** ⚠️ |
| 9 | monthly | 2025-01-10 15:00 | 2 | 2025-01-10 15:00 | 2025-02-15 10:00 | **2025-03-10 15:00** |
| 10 | monthly | 2024-01-29 12:00 | 1 | 2024-01-29 12:00 | 2024-02-15 10:00 | **2024-02-29 12:00** (năm nhuận) |
| 11 | interval_seconds | — | 3600s | 2025-01-01 10:00 | 2025-01-01 10:30 | **2025-01-01 11:00** |
| 12 | interval_seconds | — | 600s | 2025-01-01 10:00 | 2025-01-01 12:00 | **2025-01-01 12:10** ⚠️ |
| 13 | solar_last_day | — | — | — | 2025-01-15 14:30 | **2025-01-31 14:30** |
| 14 | solar_last_day | — | — | — | 2025-01-31 10:00 | **2025-02-28 10:00** |
| 15 | solar_last_day | — | — | — | 2025-02-15 09:00 | **2025-02-28 09:00** |
| 16 | solar_last_day | — | — | — | 2024-02-15 11:00 | **2024-02-29 11:00** |
| 17 | tích hợp daily | 2025-01-01 08:00 | 1 | = origin | 2025-01-05 09:00 | **2025-01-06 08:00** |
| 18 | tích hợp weekly | 2025-01-01 10:00 (T4) | 1 | = origin | 2025-01-05 12:00 (CN) | **2025-01-08 10:00** (T4) |

⚠️ **#8** — tháng 2 không có ngày 31 → kẹp về ngày cuối tháng (28).
⚠️ **#12** — `After()` là **so sánh nghiêm**: 12:00 không được tính là "sau 12:00", nên kết quả là **12:10** (không phải 12:00). Test Go chỉ assert `> now` và chia hết cho 600, nên bản port phải chọn **12:10**.

> Cách lấy thêm vector: `go test ./internal/services/ -run TestCalcNext -v` (trong `E:\PROJECT\nhacviecfcm`).
> ⚠️ Đừng chạy cả package — `TestCalcNextDaily_ZeroInterval` sẽ treo. Dùng:
> `go test ./internal/services/ -run 'TestCalcNextDailyTime_|TestCalcNextWeekly|TestCalcNextSolarMonthly_|TestCalcNextIntervalSeconds_|TestCalcNextSolarLastDayOfMonth_|TestTinhtoan_NextRecurringV2_' -timeout 60s -v`

---

## 7. 11 nhánh worker

Task cron (doc 04) chọn 1 query, rồi **phân nhánh trong code**. Bảng dưới là bản port sát Go.

Query gốc: `status = 'active' AND next_action_at IS NOT NULL AND next_action_at <= :now AND (snooze_until IS NULL OR snooze_until <= :now)`

Ký hiệu: `sent` = `is_sended_one_time`, `naa` = `next_action_at`, `T()` = `nextRecurring(r, now)`.

### Nhóm A — `type = 'one_time'`

| # | Điều kiện | Sau khi gửi FCM thành công |
|---|---|---|
| **A1** | `max_crp = 0`, `sent = false` | `sent = true`, `status = 'completed'`, `last_completed_at = now`, `last_sent_at = now`, `naa = NULL` |
| **A2** | `max_crp > 0`, `sent = false` | `sent = true`, `last_sent_at = now`, `next_crp = now + crp_interval_sec`, `naa = next_crp` — **KHÔNG completed** |
| **A3** | `max_crp > 0`, `sent = true`, `crp_count < max_crp` | `crp_count++`, `last_sent_at = now`, `next_crp = now + interval`; nếu `crp_count >= max_crp` → `status = 'completed'`, `last_completed_at = now`, `naa = NULL` |

> 🐞 **Bug trong draft tài liệu cũ**: `processReminder` set `status='completed'` ngay sau lần gửi đầu → **mất hoàn toàn CRP của one_time**. Nhánh A2/A3 là hành vi đúng.

### Nhóm B — `type = 'recurring'` + `repeat_strategy = 'none'`

| # | Điều kiện | Sau khi gửi FCM thành công |
|---|---|---|
| **B1** | `max_crp = 0`, `next_recurring <= now` | `last_sent_at = now`, `next_recurring = T()`, `naa = next_recurring` |
| **B2** | `max_crp > 0`, `next_recurring <= now` | `last_sent_at = now`, `next_recurring = T()`, `crp_count = 0`, `next_crp = now + interval`, `naa = next_crp` |
| **B3** | `max_crp > 0`, `crp_count < max_crp`, `next_crp <= now`, `next_recurring > now` | `crp_count++`, `next_crp = now + interval`, `last_sent_at = now`, `naa = next_crp` |

> 🐞 **Bug trong draft tài liệu cũ**: set `next_crp = NULL` khi sang FRP mới → CRP của chu kỳ sau **không bao giờ chạy**. Phải là `now + crp_interval_sec`.

### Nhóm C — `type = 'recurring'` + `repeat_strategy = 'crp_until_complete'`

**Điểm cốt lõi:** ở nhóm C, worker **KHÔNG BAO GIỜ** tính `next_recurring`. Lịch chỉ tiến khi user bấm Hoàn thành (mục 8).

| # | Điều kiện | Sau khi gửi FCM thành công |
|---|---|---|
| **C1** | `max_crp = 0`, `sent = false`, `next_recurring <= now` | `sent = true`, `last_sent_at = now` — **`next_recurring` giữ nguyên** |
| **C2** | `max_crp = 0`, `sent = true`, `next_recurring <= now`, `last_completed_at > last_sent_at` | `last_sent_at = now` |
| **C3** | `max_crp > 0`, `sent = false`, `next_recurring <= now` | `sent = true`, `last_sent_at = now`, `crp_count = 0`, `next_crp = now + interval` |
| **C4** | `max_crp > 0`, `sent = true`, `next_recurring <= now`, `last_completed_at > last_sent_at` | `last_sent_at = now`, `crp_count = 0`, `next_crp = now + interval` |
| **C5** | `max_crp > 0`, `sent = true`, `crp_count < max_crp`, `next_crp <= now`, (`last_completed_at IS NULL` hoặc `last_completed_at <= last_sent_at`) | `crp_count++`, `last_sent_at = now`, `next_crp = now + interval` — **không bao giờ tự completed** |

Ghi chú:
- C1/C3 chỉ chạy đúng 1 lần nhờ cờ `sent = true`. Sau đó chỉ C2/C4/C5 khớp.
- C2/C4 cần `last_completed_at > last_sent_at` → sau khi gửi, điều kiện này sai → tự dừng. Đây là cơ chế "nhắc 1 lần rồi chờ complete".
- ⚠️ **Không reset `sent` khi complete** — nếu reset sẽ rơi nhầm về C1/C3.
- ⚠️ Nhánh C5 bên Go **thiếu** guard `next_recurring > now` (nhánh B3 có). Khuyến nghị **thêm guard** cho nhất quán; nếu giữ nguyên hành vi Go thì ghi chú rõ.

---

## 8. API `complete` — port `OnUserComplete` (`reminder_service.go:176`)

> 🐞 **Bug nghiêm trọng trong draft tài liệu cũ**: `completeReminder` set `status = 'completed'` cho recurring → recurring **chết hẳn sau 1 lần complete**.

```ts
export function applyComplete(r: Reminder, now: Date): Partial<Reminder> {
  if (r.type === 'one_time') {
    return {
      status: 'completed',
      last_completed_at: now.toISOString(),
      last_crp_completed_at: now.toISOString(),
      crp_count: 0,
      next_action_at: null,
    }
  }
  // recurring: GIỮ status='active', đẩy sang chu kỳ FRP kế tiếp
  const next = nextRecurringFromComplete(r, now)
  return {
    last_completed_at: now.toISOString(),
    last_crp_completed_at: now.toISOString(),
    crp_count: 9999,                    // sentinel: "đã xong CRP, chờ FRP mới"
    next_recurring: next.toISOString(),
    next_crp: next.toISOString(),
    next_action_at: next.toISOString(),
    // status GIỮ NGUYÊN
  }
}
```

### `nextRecurringFromComplete` — port `Tinhtoan_NextRecurringV2_fromapi` + MỞ RỘNG

> 🐞 **Bug có thật bên Go** (`calNextTime.go:20-31`): hàm này **chỉ hỗ trợ daily / monthly / solar_last_day_of_month / interval_seconds** — **thiếu `weekly` và `lunar`**.
> Commit `379bf7a` đổi fallback từ `now + 24h` thành `reminder.NextRecurring` (giữ nguyên giá trị cũ, đang nằm trong quá khứ).
> **Hệ quả:** bấm Hoàn thành một reminder weekly/lunar → `next_recurring` bị giữ nguyên (quá khứ) → `next_action_at` = quá khứ → worker **gửi thêm 1 thông báo thừa ngay lập tức** rồi mới dừng.

Bản TS phải **thêm `weekly` và `lunar`**:

```ts
export function nextRecurringFromComplete(r: Reminder, now: Date): Date {
  const p = r.recurrence_pattern
  if (!p) throw new Error('recurrence_pattern required')
  const origin = r.origin_time ? new Date(r.origin_time) : now

  switch (p.type) {
    case 'daily':
      return nextDaily(origin, now, now, p.interval)
    case 'weekly':                                        // ⬅ MỞ RỘNG (Go thiếu)
      return nextWeekly(origin, now, now, p.interval)
    case 'monthly':
      return r.calendar_type === 'lunar'
        ? nextLunarMonthly(origin, now)                   // ⬅ MỞ RỘNG (Go thiếu)
        : nextSolarMonthly(origin, now, now, p.interval)
    case 'solar_last_day_of_month':
      return nextSolarLastDayOfMonth(now)
    case 'lunar_last_day_of_month':                       // ⬅ MỞ RỘNG (Go thiếu)
      return nextLunarLastDayOfMonth(origin, now)
    case 'interval_seconds':
      return nextIntervalSeconds(now, now, p.interval_seconds ?? 0)
    default:
      throw new Error(`unsupported recurrence type: ${p.type}`)
  }
}
```

Lưu ý khác biệt với `nextRecurring`:
- **`lastTime = now`** (tính từ thời điểm complete), không phải `origin_time`.
- `interval_seconds` cũng **đổi mỏ neo sang `now`**.
- Go không cần biến `from = "api_complete"` vì tách thành 2 hàm riêng — bản port **bỏ luôn field `from`** (nó vốn không có cột trong DB).

### API `snooze`

```ts
export function applySnooze(r: Reminder, now: Date, seconds: number): Partial<Reminder> {
  const until = new Date(now.getTime() + seconds * 1000)
  return { snooze_until: until.toISOString(), next_action_at: until.toISOString() }
}
```
Validate: `seconds > 0` và ≤ 86400×7 (Go hiện **không validate** → gửi số âm sẽ set `snooze_until` trong quá khứ).

---

## 9. Checklist nghiệm thu

Chạy task thủ công: `Invoke-WebRequest http://localhost:3000/_nitro/tasks/reminders:check`

- [ ] **18 golden test** mục 6 → đưa vào 1 file `app/utils/schedule.test.ts` hoặc chạy tay qua console, so kết quả từng dòng.
- [ ] `interval` bị thiếu/=`0` → **không treo**, tự hiểu là 1.
- [ ] `interval_seconds = 600` + downtime 2 tiếng → ra đúng bội của 600 và **> now**.
- [ ] Monthly `day_of_month = 31` sang tháng 2 → ra 28 (hoặc 29 năm nhuận).
- [ ] **A2/A3**: one_time + `max_crp = 3` → gửi 3 lần rồi mới completed, không completed ngay.
- [ ] **B2**: recurring + CRP → sau FRP có `next_crp = now + interval` (không phải NULL).
- [ ] **C1→C2**: `crp_until_complete` → gửi 1 lần rồi **dừng**, chờ complete.
- [ ] **Complete recurring weekly** → `status` vẫn `active`, `next_recurring` nhảy đúng 1 tuần, **không gửi thông báo thừa**.
- [ ] **Snooze** → `next_action_at = snooze_until`; sau khi hết hạn, reminder chạy lại bình thường (không lặp vô hạn).
- [ ] Âm lịch: test với năm **2026 trở lên** — phải chạy được (chứng tỏ đã port thuật toán, không copy bảng).

---

→ Quay lại [04 — Lịch nhắc + FCM](./04-lich-nhac-va-fcm.md) để cài đặt task, FCM và kill-switch.
