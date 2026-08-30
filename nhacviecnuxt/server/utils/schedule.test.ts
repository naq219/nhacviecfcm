import { describe, expect, it } from 'vitest'
import {
  nextActionAt,
  nextCrp,
  nextDaily,
  nextIntervalSeconds,
  nextRecurring,
  nextRecurringFromComplete,
  nextSolarLastDayOfMonth,
  nextSolarMonthly,
  nextWeekly,
  normalizeInterval,
} from './schedule'
import { solarToLunar } from './calendar'
import type { Reminder } from '#shared/types'

/**
 * Bộ giá trị chuẩn lấy trực tiếp từ `internal/services/calNextTime_test.go` (Go).
 * Chạy: npm run test
 * Thêm case mới → chạy lại `go test ./internal/services/ -run TestCalcNext -v` bên Go để lấy expected.
 */

/** Tạo Date UTC — tương đương createTime() trong test Go */
function t(year: number, month: number, day: number, hour = 0, minute = 0): Date {
  return new Date(Date.UTC(year, month - 1, day, hour, minute, 0, 0))
}

function iso(d: Date): string {
  return d.toISOString()
}

function reminder(over: Partial<Reminder>): Reminder {
  return {
    id: 'r1',
    user_id: 'u1',
    title: 'test',
    type: 'recurring',
    status: 'active',
    repeat_strategy: 'none',
    calendar_type: 'solar',
    origin_time: null,
    max_crp: 0,
    crp_count: 0,
    crp_interval_sec: 0,
    is_sended_one_time: 0,
    created_at: '2025-01-01T00:00:00.000Z',
    updated_at: '2025-01-01T00:00:00.000Z',
    ...over,
  }
}

describe('normalizeInterval — chống vòng lặp vô hạn', () => {
  it('interval 0 / âm / undefined → 1', () => {
    expect(normalizeInterval(0)).toBe(1)
    expect(normalizeInterval(-5)).toBe(1)
    expect(normalizeInterval(undefined)).toBe(1)
    expect(normalizeInterval(3)).toBe(3)
  })

  // Go đang TREO ở case này (TestCalcNextDaily_ZeroInterval)
  it('daily với interval = 0 không bị treo', () => {
    const next = nextDaily(t(2025, 1, 1, 8), t(2025, 1, 5, 8), t(2025, 1, 6, 7), 0)
    expect(iso(next)).toBe(iso(t(2025, 1, 6, 8)))
  })
})

describe('daily — port calcNextDailyTime', () => {
  it('SimpleDaily', () => {
    expect(iso(nextDaily(t(2025, 1, 1, 8), t(2025, 1, 5, 8), t(2025, 1, 6, 7), 1)))
      .toBe(iso(t(2025, 1, 6, 8)))
  })
  it('EveryTwoDays', () => {
    expect(iso(nextDaily(t(2025, 1, 1, 9, 30), t(2025, 1, 5, 9, 30), t(2025, 1, 6, 10), 2)))
      .toBe(iso(t(2025, 1, 7, 9, 30)))
  })
  it('CrossMonth', () => {
    expect(iso(nextDaily(t(2025, 1, 1, 10), t(2025, 1, 31, 10), t(2025, 2, 1, 9), 1)))
      .toBe(iso(t(2025, 2, 1, 10)))
  })
})

describe('weekly — port calcNextWeekly (thứ lấy từ origin_time)', () => {
  it('SameWeek: origin T6, đang T2 → T6 kế tiếp', () => {
    expect(iso(nextWeekly(t(2025, 1, 3, 10), t(2025, 1, 6, 10), t(2025, 1, 6, 11), 1)))
      .toBe(iso(t(2025, 1, 10, 10)))
  })
  it('EveryTwoWeeks', () => {
    expect(iso(nextWeekly(t(2025, 1, 6, 14), t(2025, 1, 6, 14), t(2025, 1, 7, 10), 2)))
      .toBe(iso(t(2025, 1, 20, 14)))
  })
  it('Sunday', () => {
    expect(iso(nextWeekly(t(2025, 1, 5, 9), t(2025, 1, 11, 9), t(2025, 1, 11, 10), 1)))
      .toBe(iso(t(2025, 1, 12, 9)))
  })
})

describe('monthly — port calcNextSolarMonthly', () => {
  it('SimpleMonthly', () => {
    expect(iso(nextSolarMonthly(t(2025, 1, 15, 10, 30), t(2025, 1, 15, 10, 30), t(2025, 2, 1, 12), 1)))
      .toBe(iso(t(2025, 2, 15, 10, 30)))
  })
  it('Day31ToFebruary → kẹp về 28', () => {
    expect(iso(nextSolarMonthly(t(2025, 1, 31, 8), t(2025, 1, 31, 8), t(2025, 2, 1, 9), 1)))
      .toBe(iso(t(2025, 2, 28, 8)))
  })
  it('EveryTwoMonths', () => {
    expect(iso(nextSolarMonthly(t(2025, 1, 10, 15), t(2025, 1, 10, 15), t(2025, 2, 15, 10), 2)))
      .toBe(iso(t(2025, 3, 10, 15)))
  })
  it('LeapYear: 29/1 → 29/2 năm nhuận', () => {
    expect(iso(nextSolarMonthly(t(2024, 1, 29, 12), t(2024, 1, 29, 12), t(2024, 2, 15, 10), 1)))
      .toBe(iso(t(2024, 2, 29, 12)))
  })
})

describe('interval_seconds — port calcNextIntervalSeconds', () => {
  it('OneHour', () => {
    expect(iso(nextIntervalSeconds(t(2025, 1, 1, 10), t(2025, 1, 1, 10, 30), 3600)))
      .toBe(iso(t(2025, 1, 1, 11)))
  })
  // Go chỉ assert "> now" và chia hết cho 600. After() là so sánh NGHIÊM → 12:10.
  it('SystemDowntime 2 tiếng → catch-up, kết quả > now (12:10)', () => {
    const next = nextIntervalSeconds(t(2025, 1, 1, 10), t(2025, 1, 1, 12), 600)
    expect(iso(next)).toBe(iso(t(2025, 1, 1, 12, 10)))
  })
  it('interval_seconds <= 0 → ném lỗi', () => {
    expect(() => nextIntervalSeconds(t(2025, 1, 1), t(2025, 1, 2), 0)).toThrow()
  })
})

describe('solar_last_day_of_month — port calcNextSolarLastDayOfMonth', () => {
  it('SameMonth', () => {
    expect(iso(nextSolarLastDayOfMonth(t(2025, 1, 15, 14, 30)))).toBe(iso(t(2025, 1, 31, 14, 30)))
  })
  it('AlreadyLastDay → tháng sau', () => {
    expect(iso(nextSolarLastDayOfMonth(t(2025, 1, 31, 10)))).toBe(iso(t(2025, 2, 28, 10)))
  })
  it('February', () => {
    expect(iso(nextSolarLastDayOfMonth(t(2025, 2, 15, 9)))).toBe(iso(t(2025, 2, 28, 9)))
  })
  it('LeapYearFebruary', () => {
    expect(iso(nextSolarLastDayOfMonth(t(2024, 2, 15, 11)))).toBe(iso(t(2024, 2, 29, 11)))
  })
})

describe('nextRecurring — tích hợp (port Tinhtoan_NextRecurringV2)', () => {
  it('Daily', () => {
    const r = reminder({
      origin_time: iso(t(2025, 1, 1, 8)),
      recurrence_pattern: { type: 'daily', interval: 1 },
    })
    expect(iso(nextRecurring(r, t(2025, 1, 5, 9)))).toBe(iso(t(2025, 1, 6, 8)))
  })
  it('Weekly: origin T4, now CN → T4 kế tiếp', () => {
    const r = reminder({
      origin_time: iso(t(2025, 1, 1, 10)),
      recurrence_pattern: { type: 'weekly', interval: 1 },
    })
    expect(iso(nextRecurring(r, t(2025, 1, 5, 12)))).toBe(iso(t(2025, 1, 8, 10)))
  })
  it('thiếu recurrence_pattern → ném lỗi', () => {
    expect(() => nextRecurring(reminder({ origin_time: iso(t(2025, 1, 1)) }), t(2025, 1, 5))).toThrow()
  })
  it('thiếu origin_time → ném lỗi', () => {
    expect(() => nextRecurring(reminder({ recurrence_pattern: { type: 'daily' } }), t(2025, 1, 5))).toThrow()
  })
})

describe('nextRecurringFromComplete — MỞ RỘNG so với Go', () => {
  const now = t(2027, 3, 10, 14)

  // Go: Tinhtoan_NextRecurringV2_fromapi THIẾU weekly và lunar
  it('weekly: Go đang thiếu nhánh này', () => {
    const r = reminder({
      origin_time: iso(t(2027, 3, 3, 9)),
      recurrence_pattern: { type: 'weekly', interval: 1 },
    })
    const next = nextRecurringFromComplete(r, now)
    expect(next.getTime()).toBeGreaterThan(now.getTime())
    expect(next.getUTCDay()).toBe(3) // cùng thứ với origin (T4)
  })

  it('lunar: Go đang thiếu nhánh này', () => {
    const r = reminder({
      calendar_type: 'lunar',
      origin_time: iso(t(2027, 3, 3, 9)),
      recurrence_pattern: { type: 'monthly', interval: 1 },
    })
    const next = nextRecurringFromComplete(r, now)
    expect(next.getTime()).toBeGreaterThan(now.getTime())
  })

  it('interval_seconds: mỏ neo chuyển sang now', () => {
    const r = reminder({
      origin_time: iso(t(2020, 1, 1)),
      recurrence_pattern: { type: 'interval_seconds', interval_seconds: 3600 },
    })
    expect(iso(nextRecurringFromComplete(r, now))).toBe(iso(t(2027, 3, 10, 15)))
  })
})

describe('âm lịch (bảng sinh từ Go)', () => {
  it('Tết 2027 (1/1 âm) → 2027-02-06', () => {
    const d = new Date('2027-02-06T00:00:00.000Z')
    const l = solarToLunar(d)
    expect(l).not.toBeNull()
    expect(`${l!.day}/${l!.month}/${l!.year}`).toBe('1/1/2027')
  })
  it('ngoài phạm vi bảng → null (không được âm thầm sai)', () => {
    expect(solarToLunar(new Date('1990-01-01T00:00:00.000Z'))).toBeNull()
  })
})

describe('nextActionAt', () => {
  const now = t(2025, 6, 1, 12)

  it('KHÔNG đưa now vào ứng viên (bug bản draft cũ: gửi lặp vô hạn)', () => {
    const r = reminder({ next_recurring: iso(t(2025, 6, 2, 8)) })
    expect(nextActionAt(r, now)).toBe(iso(t(2025, 6, 2, 8)))
  })

  it('snooze còn hiệu lực → ưu tiên tuyệt đối', () => {
    const r = reminder({
      snooze_until: iso(t(2025, 6, 1, 13)),
      next_recurring: iso(t(2025, 6, 2, 8)),
    })
    expect(nextActionAt(r, now)).toBe(iso(t(2025, 6, 1, 13)))
  })

  it('snooze ĐÃ HẾT HẠN → bỏ qua, không kẹt quá khứ', () => {
    const r = reminder({
      snooze_until: iso(t(2025, 5, 1, 13)),
      next_recurring: iso(t(2025, 6, 2, 8)),
    })
    expect(nextActionAt(r, now)).toBe(iso(t(2025, 6, 2, 8)))
  })

  it('min(next_recurring, next_crp) khi có CRP', () => {
    const r = reminder({
      max_crp: 3,
      next_recurring: iso(t(2025, 6, 5, 8)),
      next_crp: iso(t(2025, 6, 1, 18)),
    })
    expect(nextActionAt(r, now)).toBe(iso(t(2025, 6, 1, 18)))
  })

  it('bỏ qua next_crp khi max_crp = 0', () => {
    const r = reminder({
      max_crp: 0,
      next_recurring: iso(t(2025, 6, 5, 8)),
      next_crp: iso(t(2025, 6, 1, 18)),
    })
    expect(nextActionAt(r, now)).toBe(iso(t(2025, 6, 5, 8)))
  })

  it('không còn gì → null', () => {
    expect(nextActionAt(reminder({}), now)).toBeNull()
  })
})

describe('nextCrp', () => {
  const now = t(2025, 6, 1, 12)

  it('max_crp = 0 → null (không retry)', () => {
    expect(nextCrp(reminder({ max_crp: 0, crp_interval_sec: 60 }), now)).toBeNull()
  })
  it('đủ quota → null', () => {
    const r = reminder({ max_crp: 2, crp_count: 2, crp_interval_sec: 60 })
    expect(nextCrp(r, now)).toBeNull()
  })
  it('còn lượt → now + crp_interval_sec', () => {
    const r = reminder({ max_crp: 3, crp_count: 1, crp_interval_sec: 300 })
    expect(nextCrp(r, now)).toBe(iso(t(2025, 6, 1, 12, 5)))
  })
  it('crp_interval_sec <= 0 → fallback 60 giây', () => {
    const r = reminder({ max_crp: 3, crp_count: 0, crp_interval_sec: 0 })
    expect(nextCrp(r, now)).toBe(iso(t(2025, 6, 1, 12, 1)))
  })
})
