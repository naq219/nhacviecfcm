import type { RecurrencePattern } from '#shared/types'

/** Mô tả lịch lặp bằng tiếng Việt để hiển thị */
export function describeRecurrence(p: RecurrencePattern | null | undefined): string {
  if (!p) return 'không lặp'

  switch (p.type) {
    case 'daily':
      return `mỗi ${p.interval && p.interval > 1 ? p.interval : ''} ngày`.replace('  ', ' ').trim()
    case 'weekly':
      return `mỗi ${p.interval && p.interval > 1 ? p.interval : ''} tuần`.replace('  ', ' ').trim()
    case 'monthly':
      return `mỗi ${p.interval && p.interval > 1 ? p.interval : ''} tháng`.replace('  ', ' ').trim()
    case 'interval_seconds': {
      const s = p.interval_seconds ?? 0
      if (s % 86400 === 0) return `mỗi ${s / 86400} ngày`
      if (s % 3600 === 0) return `mỗi ${s / 3600} giờ`
      if (s % 60 === 0) return `mỗi ${s / 60} phút`
      return `mỗi ${s} giây`
    }
    case 'solar_last_day_of_month':
      return 'ngày cuối tháng dương'
    case 'lunar_last_day_of_month':
      return 'ngày cuối tháng âm'
    default:
      return (p as { type: string }).type
  }
}
