export interface User {
  id: string
  email: string
  /** @deprecated Chỉ để tương thích app Android cũ — dùng bảng `devices` */
  fcm_token?: string | null
  /** Công tắc TỔNG: user muốn tắt hẳn thông báo trên mọi thiết bị */
  is_fcm_active?: 0 | 1
}

export type DevicePlatform = 'android' | 'ios' | 'web'

export interface Device {
  id: string
  user_id: string
  token: string
  platform: DevicePlatform
  is_active: 0 | 1
  created_at: string
  updated_at: string
}

/** Dữ liệu public trả về cho client (KHÔNG gồm password_hash) */
export interface PublicUser {
  id: string
  email: string
}

export type ReminderType = 'one_time' | 'recurring'
export type ReminderStatus = 'active' | 'completed' | 'paused'
export type RepeatStrategy = 'none' | 'crp_until_complete'
export type CalendarType = 'solar' | 'lunar'

export type RecurrenceType =
  | 'daily'
  | 'weekly'
  | 'monthly'
  | 'interval_seconds'
  | 'solar_last_day_of_month'
  | 'lunar_last_day_of_month'

export interface RecurrencePattern {
  type: RecurrenceType
  /** ngày / tuần / tháng tuỳ type; <= 0 hoặc thiếu → 1 */
  interval?: number
  /** chỉ dùng cho interval_seconds; bắt buộc > 0 */
  interval_seconds?: number
  /** chỉ để tương thích dữ liệu cũ; worker lấy thứ từ origin_time */
  day_of_week?: number
  /** chỉ để tương thích dữ liệu cũ; worker lấy ngày từ origin_time */
  day_of_month?: number
  calendar_type?: CalendarType
  /** chỉ để tương thích dữ liệu cũ; worker lấy giờ từ origin_time */
  trigger_time_of_day?: string
}

export interface Reminder {
  id: string
  user_id: string
  title: string
  description?: string | null
  tag?: string | null
  type: ReminderType
  status: ReminderStatus
  recurrence_pattern?: RecurrencePattern | null
  repeat_strategy: RepeatStrategy
  calendar_type: CalendarType
  /** Mỏ neo tính lịch lặp (ISO 8601 UTC) */
  origin_time: string | null
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

/**
 * Payload tạo reminder: `title` + `type` bắt buộc, còn lại tuỳ chọn
 * (server tự suy `origin_time` từ `next_action_at` nếu không gửi).
 */
export type ReminderPayload = Pick<Reminder, 'title' | 'type'> & Partial<Pick<Reminder,
  | 'description' | 'tag'
  | 'recurrence_pattern' | 'repeat_strategy' | 'calendar_type'
  | 'origin_time' | 'next_action_at'
  | 'max_crp' | 'crp_interval_sec'
>> & {
  /** giây — tiện test cron (port từ Go `for_test`) */
  for_test?: number
}

export interface SystemStatus {
  mid: 1
  worker_enabled: 0 | 1
  last_error: string | null
  updated_at: string
}
