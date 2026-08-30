import type { Reminder } from '#shared/types'
import { findNextLunarLastDayOfMonth, findNextLunarMonthly, solarToLunar } from './calendar'

/**
 * Tính lịch lặp FRP/CRP — port từ `internal/services/calNextTime.go`.
 * Đặc tả + bộ giá trị chuẩn: docs/04b-tinh-lich-frp-crp.md
 *
 * Quy tắc:
 *  - Mọi phép tính theo **UTC**.
 *  - Mỏ neo là `origin_time` (KHÔNG phải lần gửi gần nhất) → lịch không trôi.
 *  - Giờ/phút và thứ trong tuần lấy từ `origin_time`, bỏ qua
 *    `trigger_time_of_day` / `day_of_week` / `day_of_month`.
 *  - BẮT BUỘC guard `interval <= 0` trước mọi vòng lặp (Go đang treo ở chỗ này).
 */

/** Chặn treo Worker nếu dữ liệu sai */
const MAX_ITER = 10_000

/** interval <= 0 hoặc thiếu → coi như 1 (chống vòng lặp vô hạn) */
export function normalizeInterval(n: number | undefined, fallback = 1): number {
  const v = Math.floor(n ?? 0)
  return v > 0 ? v : fallback
}

// ---------- helper UTC ----------

function hhmmOf(origin: Date): { h: number, m: number } {
  return { h: origin.getUTCHours(), m: origin.getUTCMinutes() }
}

function daysInMonthUTC(year: number, monthIdx0: number): number {
  return new Date(Date.UTC(year, monthIdx0 + 1, 0)).getUTCDate()
}

function withTimeUTC(d: Date, h: number, m: number): Date {
  return new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth(), d.getUTCDate(), h, m, 0, 0))
}

function addDaysUTC(d: Date, n: number): Date {
  return new Date(Date.UTC(
    d.getUTCFullYear(), d.getUTCMonth(), d.getUTCDate() + n,
    d.getUTCHours(), d.getUTCMinutes(), d.getUTCSeconds(), d.getUTCMilliseconds(),
  ))
}

function guard(i: number, where: string): void {
  if (i > MAX_ITER) {
    throw new Error(`Vượt quá ${MAX_ITER} vòng lặp khi tính lịch ${where} — kiểm tra interval`)
  }
}

// ---------- daily ----------

/** Port `calcNextDailyTime` (calNextTime.go:119) */
export function nextDaily(origin: Date, lastTime: Date, now: Date, rawInterval?: number): Date {
  const interval = normalizeInterval(rawInterval)
  const { h, m } = hhmmOf(origin)

  let next = addDaysUTC(withTimeUTC(lastTime, h, m), interval)
  let i = 0
  while (next.getTime() <= now.getTime()) {
    guard(++i, 'daily')
    next = addDaysUTC(next, interval)
  }
  return next
}

// ---------- weekly ----------

/** Port `calcNextWeekly` (calNextTime.go:213) — thứ lấy từ origin_time */
export function nextWeekly(origin: Date, lastTime: Date, now: Date, rawInterval?: number): Date {
  const interval = normalizeInterval(rawInterval)
  const { h, m } = hhmmOf(origin)
  const target = origin.getUTCDay() // 0 = Chủ nhật

  let next = withTimeUTC(lastTime, h, m)
  let daysAhead = (target - next.getUTCDay() + 7) % 7

  if (daysAhead === 0) {
    if (next.getTime() > now.getTime()) return next // hôm nay đúng thứ, chưa tới giờ
    daysAhead = 7 * interval
  }
  next = addDaysUTC(next, daysAhead)

  let i = 0
  while (next.getTime() <= now.getTime()) {
    guard(++i, 'weekly')
    next = addDaysUTC(next, 7 * interval)
  }
  return next
}

// ---------- monthly (dương) ----------

/** Port `calcNextSolarMonthly` (calNextTime.go:148) — normalize về ngày 1 tránh tràn ngày */
export function nextSolarMonthly(origin: Date, lastTime: Date, now: Date, rawInterval?: number): Date {
  const interval = normalizeInterval(rawInterval)
  const { h, m } = hhmmOf(origin)
  const originDay = origin.getUTCDate()

  const step = (from: Date): Date => {
    const tmp = new Date(Date.UTC(from.getUTCFullYear(), from.getUTCMonth() + interval, 1, h, m))
    const last = daysInMonthUTC(tmp.getUTCFullYear(), tmp.getUTCMonth())
    return new Date(Date.UTC(
      tmp.getUTCFullYear(), tmp.getUTCMonth(), Math.min(originDay, last), h, m,
    ))
  }

  let next = step(new Date(Date.UTC(lastTime.getUTCFullYear(), lastTime.getUTCMonth(), 1)))
  let i = 0
  while (next.getTime() <= now.getTime()) {
    guard(++i, 'monthly')
    next = step(next)
  }
  return next
}

// ---------- ngày cuối tháng dương ----------

/** Port `calcNextSolarLastDayOfMonth` (calNextTime.go:90) */
export function nextSolarLastDayOfMonth(now: Date): Date {
  const h = now.getUTCHours()
  const m = now.getUTCMinutes()
  const s = now.getUTCSeconds()
  const ms = now.getUTCMilliseconds()

  const lastThis = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() + 1, 0, h, m, s, ms))
  if (now.getUTCDate() < lastThis.getUTCDate()) return lastThis
  return new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() + 2, 0, h, m, s, ms))
}

// ---------- interval_seconds ----------

/** Port `calcNextIntervalSeconds` (calNextTime.go:75) — dùng nhân chia, O(1) */
export function nextIntervalSeconds(origin: Date, now: Date, intervalSeconds: number): Date {
  if (intervalSeconds <= 0) throw new Error('interval_seconds phải > 0')
  const stepMs = intervalSeconds * 1000
  const k = Math.floor((now.getTime() - origin.getTime()) / stepMs) + 1
  return new Date(origin.getTime() + k * stepMs)
}

// ---------- âm lịch ----------

/** Port `calcNextLunarMonthly` (calNextTime.go:260) */
export function nextLunarMonthly(origin: Date, now: Date): Date {
  const lunarOrigin = solarToLunar(origin)
  if (!lunarOrigin) {
    throw new Error('origin_time nằm ngoài phạm vi bảng âm lịch')
  }
  const { h, m } = hhmmOf(origin)
  const nextSolar = findNextLunarMonthly(now, lunarOrigin.day)
  if (!nextSolar) {
    throw new Error('không tìm được ngày âm kế tiếp (vượt quá phạm vi bảng âm lịch)')
  }
  return new Date(Date.UTC(
    nextSolar.getUTCFullYear(), nextSolar.getUTCMonth(), nextSolar.getUTCDate(), h, m,
  ))
}

/** MỞ RỘNG: Go chưa có ở worker engine (chỉ có ở engine đã bị loại) */
export function nextLunarLastDayOfMonth(_origin: Date, now: Date): Date {
  const nextSolar = findNextLunarLastDayOfMonth(now)
  if (!nextSolar) {
    throw new Error('không tìm được ngày cuối tháng âm kế tiếp')
  }
  return nextSolar
}

// ---------- dispatcher ----------

function requireOrigin(r: Reminder): Date {
  if (!r.origin_time) throw new Error('origin_time là bắt buộc để tính lịch lặp')
  return new Date(r.origin_time)
}

/**
 * Tính thời điểm FRP kế tiếp.
 * Port `Tinhtoan_NextRecurringV2` (calNextTime.go:34).
 * Worker Go luôn truyền lastTime = origin_time → lịch không trôi.
 */
export function nextRecurring(r: Reminder, now: Date): Date {
  const p = r.recurrence_pattern
  if (!p) throw new Error('recurrence_pattern là bắt buộc với recurring reminder')

  const origin = requireOrigin(r)

  // Âm lịch: calendar_type = 'lunar' thắng pattern.type (đúng theo Go)
  if (r.calendar_type === 'lunar') return nextLunarMonthly(origin, now)

  switch (p.type) {
    case 'interval_seconds':
      return nextIntervalSeconds(origin, now, p.interval_seconds ?? 0)
    case 'daily':
      return nextDaily(origin, origin, now, p.interval)
    case 'weekly':
      return nextWeekly(origin, origin, now, p.interval)
    case 'monthly':
      return nextSolarMonthly(origin, origin, now, p.interval)
    case 'solar_last_day_of_month':
      return nextSolarLastDayOfMonth(now)
    case 'lunar_last_day_of_month':
      return nextLunarLastDayOfMonth(origin, now)
    default:
      throw new Error(`Không hỗ trợ recurrence type: ${(p as { type: string }).type}`)
  }
}

/**
 * Tính FRP kế tiếp **từ thời điểm user bấm Hoàn thành**.
 * Port `Tinhtoan_NextRecurringV2_fromapi` (calNextTime.go:11)
 * + MỞ RỘNG weekly và lunar (Go thiếu 2 nhánh này).
 */
export function nextRecurringFromComplete(r: Reminder, now: Date): Date {
  const p = r.recurrence_pattern
  if (!p) throw new Error('recurrence_pattern là bắt buộc với recurring reminder')

  const origin = r.origin_time ? new Date(r.origin_time) : now

  switch (p.type) {
    case 'daily':
      return nextDaily(origin, now, now, p.interval)
    case 'weekly': // MỞ RỘNG (Go thiếu)
      return nextWeekly(origin, now, now, p.interval)
    case 'monthly':
      return r.calendar_type === 'lunar'
        ? nextLunarMonthly(origin, now) // MỞ RỘNG (Go thiếu)
        : nextSolarMonthly(origin, now, now, p.interval)
    case 'solar_last_day_of_month':
      return nextSolarLastDayOfMonth(now)
    case 'lunar_last_day_of_month': // MỞ RỘNG (Go thiếu)
      return nextLunarLastDayOfMonth(origin, now)
    case 'interval_seconds':
      return nextIntervalSeconds(now, now, p.interval_seconds ?? 0)
    default:
      throw new Error(`Không hỗ trợ recurrence type: ${(p as { type: string }).type}`)
  }
}

// ---------- CRP ----------

/** Khoảng cách retry CRP, fallback 60 giây (như worker_onetime_v2.go:156-160) */
export function crpIntervalMs(r: Reminder): number {
  return (r.crp_interval_sec > 0 ? r.crp_interval_sec : 60) * 1000
}

/** Thời điểm retry CRP kế tiếp, hoặc null nếu hết lượt */
export function nextCrp(r: Reminder, now: Date): string | null {
  if (r.max_crp <= 0) return null
  if (r.crp_count >= r.max_crp) return null
  return new Date(now.getTime() + crpIntervalMs(r)).toISOString()
}

// ---------- next_action_at ----------

/**
 * Thời điểm gần nhất cần xử lý reminder.
 * Port `CalculateNextActionAt` (schedule_calculator.go:29).
 *
 * ⚠️ KHÔNG đưa `now` vào danh sách ứng viên (bản draft cũ của docs làm vậy →
 *    reminder bị xử lý lại mỗi phút vô hạn).
 * ⚠️ Bỏ qua `snooze_until` đã hết hạn (Go dùng short-circuit IsSnoozeUntilActive).
 */
export function nextActionAt(r: Reminder, now: Date): string | null {
  if (r.snooze_until && Date.parse(r.snooze_until) > now.getTime()) {
    return r.snooze_until
  }

  const candidates: number[] = []
  if (r.next_recurring) candidates.push(Date.parse(r.next_recurring))
  if (r.max_crp > 0 && r.next_crp) candidates.push(Date.parse(r.next_crp))
  if (candidates.length === 0) return null

  return new Date(Math.min(...candidates)).toISOString()
}
