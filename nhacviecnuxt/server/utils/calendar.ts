import { LUNAR_TO_SOLAR, SOLAR_TO_LUNAR } from './lunar-table'

/**
 * Âm lịch Việt Nam (múi giờ +07).
 *
 * Dữ liệu lấy từ bảng sinh bởi `go run ./cmd/gen_lunar_table` (thuật toán
 * Jean Meeus / Hồ Ngọc Đức trong `internal/services/lunar_calendar.go`).
 *
 * ⚠️ Bảng CÓ HẠN (xem header lunar-table.ts). Hàm dưới trả null khi ra ngoài
 * phạm vi — caller phải xử lý (xem `schedule.ts`).
 */

export interface LunarDate {
  day: number
  month: number
  year: number
  isLeapMonth: boolean
}

const VN_OFFSET_MS = 7 * 60 * 60 * 1000

/** "YYYY-MM-DD" của một thời điểm theo múi giờ VN (+07) */
function vnDateKey(date: Date): string {
  return new Date(date.getTime() + VN_OFFSET_MS).toISOString().slice(0, 10)
}

/** "YYYY-MM-DD" theo UTC */
function utcDateKey(date: Date): string {
  return date.toISOString().slice(0, 10)
}

/** Dương → Âm. Trả null nếu ngày nằm ngoài phạm vi bảng. */
export function solarToLunar(date: Date): LunarDate | null {
  const t = SOLAR_TO_LUNAR[vnDateKey(date)]
  if (!t) return null
  return { day: t[0], month: t[1], year: t[2], isLeapMonth: t[3] === 1 }
}

/** Âm → Dương. Trả null nếu ngày âm không tồn tại (vd 30 tháng thiếu). */
export function lunarToSolar(
  year: number,
  month: number,
  day: number,
  isLeapMonth = false,
): Date | null {
  const key = `${year}-${month}-${day}${isLeapMonth ? '-true' : ''}`
  const solar = LUNAR_TO_SOLAR[key]
  if (!solar) return null
  return new Date(`${solar}T00:00:00.000Z`)
}

/** Ngày đầu tiên trong phạm vi bảng (để báo lỗi rõ ràng). */
export function lunarTableRange(): { from: string; to: string } | null {
  const keys = Object.keys(SOLAR_TO_LUNAR).sort()
  if (keys.length === 0) return null
  return { from: keys[0]!, to: keys[keys.length - 1]! }
}

/**
 * Số ngày của tháng âm lịch (29 hoặc 30). Trả 0 nếu không xác định được.
 */
export function daysInLunarMonth(year: number, month: number, isLeapMonth = false): number {
  if (lunarToSolar(year, month, 30, isLeapMonth)) return 30
  if (lunarToSolar(year, month, 29, isLeapMonth)) return 29
  return 0
}

/**
 * Tháng âm lịch kế tiếp sau (year, month).
 */
export function nextLunarMonth(year: number, month: number): { year: number; month: number } {
  return month >= 12 ? { year: year + 1, month: 1 } : { year, month: month + 1 }
}

/**
 * Ngày dương của "ngày âm `lunarDay` trong tháng âm kế tiếp" tính từ `now`.
 * Nếu tháng âm đó không có ngày `lunarDay` (tháng thiếu) → kẹp về ngày cuối tháng.
 * Trả null nếu vượt quá phạm vi bảng.
 */
export function findNextLunarMonthly(now: Date, lunarDay: number): Date | null {
  const current = solarToLunar(now)
  if (!current) return null

  let { year, month } = nextLunarMonth(current.year, current.month)

  // Thử tối đa 3 tháng âm kế tiếp (đủ để vượt qua tháng thiếu ngày)
  for (let i = 0; i < 3; i++) {
    const dim = daysInLunarMonth(year, month)
    if (dim > 0) {
      const day = Math.min(lunarDay, dim)
      const solar = lunarToSolar(year, month, day)
      if (solar) return solar
    }
    const next = nextLunarMonth(year, month)
    year = next.year
    month = next.month
  }
  return null
}

/**
 * Ngày dương của "ngày cuối tháng âm kế tiếp" tính từ `now`.
 * Dùng cho `lunar_last_day_of_month`.
 */
export function findNextLunarLastDayOfMonth(now: Date): Date | null {
  const current = solarToLunar(now)
  if (!current) return null

  let { year, month } = nextLunarMonth(current.year, current.month)

  for (let i = 0; i < 3; i++) {
    const dim = daysInLunarMonth(year, month)
    if (dim > 0) {
      const solar = lunarToSolar(year, month, dim)
      if (solar) return solar
    }
    const next = nextLunarMonth(year, month)
    year = next.year
    month = next.month
  }
  return null
}

export { utcDateKey, vnDateKey }
