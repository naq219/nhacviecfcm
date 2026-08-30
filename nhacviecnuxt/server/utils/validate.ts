import type { ReminderPayload } from '#shared/types'

const MAX_TITLE = 200
const MAX_DESC = 2000

/**
 * Validate payload tạo reminder — port `validateReminderForCreate`
 * (reminder_handler.go:306-364) + bổ sung giới hạn độ dài (Go không có).
 */
export function validateReminderPayload(data: Partial<ReminderPayload>): string | null {
  if (!data.title || !data.title.trim()) return 'title là bắt buộc'
  if (data.title.length > MAX_TITLE) return `title không quá ${MAX_TITLE} ký tự`
  if (data.description && data.description.length > MAX_DESC) {
    return `description không quá ${MAX_DESC} ký tự`
  }

  if (data.type !== 'one_time' && data.type !== 'recurring') {
    return 'type phải là one_time hoặc recurring'
  }

  if (data.calendar_type && data.calendar_type !== 'solar' && data.calendar_type !== 'lunar') {
    return 'calendar_type phải là solar hoặc lunar'
  }

  if (data.repeat_strategy
    && data.repeat_strategy !== 'none'
    && data.repeat_strategy !== 'crp_until_complete') {
    return 'repeat_strategy phải là none hoặc crp_until_complete'
  }

  const maxCrp = data.max_crp ?? 0
  const crpInterval = data.crp_interval_sec ?? 0
  if (maxCrp < 0) return 'max_crp không được âm'
  if (maxCrp > 0 && crpInterval <= 0) {
    return 'crp_interval_sec phải > 0 khi max_crp > 0'
  }

  if (data.type === 'recurring') {
    if (!data.recurrence_pattern) return 'recurrence_pattern là bắt buộc với recurring'
    const p = data.recurrence_pattern
    if (p.type === 'interval_seconds' && (p.interval_seconds ?? 0) <= 0) {
      return 'interval_seconds phải > 0'
    }
    if (!data.next_action_at && !data.origin_time) {
      return 'cần next_action_at hoặc origin_time với recurring'
    }
  }
  else if (data.recurrence_pattern) {
    return 'one_time không được có recurrence_pattern'
  }

  if (data.next_action_at && Number.isNaN(Date.parse(data.next_action_at))) {
    return 'next_action_at không đúng định dạng ISO 8601'
  }
  if (data.origin_time && Number.isNaN(Date.parse(data.origin_time))) {
    return 'origin_time không đúng định dạng ISO 8601'
  }

  return null
}
