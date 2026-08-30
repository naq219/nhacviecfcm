import type { InValue } from '@libsql/client'
import type { RecurrencePattern, Reminder, ReminderPayload, ReminderStatus } from '#shared/types'
import { nextActionAt, nextRecurring } from './schedule'

/**
 * Reminder service — business logic + SQL.
 * Mọi hàm đều nhận `userId` và WHERE kèm `user_id` để chống IDOR.
 */

type Row = Record<string, unknown>

function num(v: unknown, fallback = 0): number {
  const n = Number(v)
  return Number.isFinite(n) ? n : fallback
}

function str(v: unknown): string | null {
  return typeof v === 'string' && v.length > 0 ? v : null
}

function zeroOne(v: unknown): 0 | 1 {
  return num(v) === 1 ? 1 : 0
}

export function mapRow(row: Row): Reminder {
  const pattern = str(row.recurrence_pattern)
  return {
    id: String(row.id),
    user_id: String(row.user_id),
    title: String(row.title ?? ''),
    description: str(row.description),
    tag: str(row.tag),
    type: row.type === 'recurring' ? 'recurring' : 'one_time',
    status: (str(row.status) ?? 'active') as ReminderStatus,
    recurrence_pattern: pattern ? (JSON.parse(pattern) as RecurrencePattern) : null,
    repeat_strategy: row.repeat_strategy === 'crp_until_complete' ? 'crp_until_complete' : 'none',
    calendar_type: row.calendar_type === 'lunar' ? 'lunar' : 'solar',
    origin_time: str(row.origin_time),
    next_recurring: str(row.next_recurring),
    next_crp: str(row.next_crp),
    max_crp: num(row.max_crp),
    crp_count: num(row.crp_count),
    crp_interval_sec: num(row.crp_interval_sec),
    is_sended_one_time: zeroOne(row.is_sended_one_time),
    next_action_at: str(row.next_action_at),
    last_sent_at: str(row.last_sent_at),
    last_completed_at: str(row.last_completed_at),
    last_crp_completed_at: str(row.last_crp_completed_at),
    snooze_until: str(row.snooze_until),
    created_at: String(row.created_at ?? ''),
    updated_at: String(row.updated_at ?? ''),
  }
}

/** Đọc 1 reminder thô (dùng nội bộ, không kèm user_id) */
export async function getReminderRaw(id: string): Promise<Reminder | null> {
  const db = useDb()
  const rs = await db.execute({ sql: 'SELECT * FROM reminders WHERE id = ?', args: [id] })
  return rs.rows[0] ? mapRow(rs.rows[0] as Row) : null
}

/** Danh sách reminder của 1 user. `status` = null → lấy tất cả. */
export async function listReminders(userId: string, status?: ReminderStatus): Promise<Reminder[]> {
  const db = useDb()
  const rs = await db.execute({
    sql: `SELECT * FROM reminders
          WHERE user_id = ? AND (? IS NULL OR status = ?)
          ORDER BY next_action_at IS NULL, next_action_at ASC`,
    args: [userId, status ?? null, status ?? null],
  })
  return rs.rows.map(r => mapRow(r as Row))
}

/** Chi tiết 1 reminder (chỉ khi thuộc về userId) */
export async function getReminder(id: string, userId: string): Promise<Reminder | null> {
  const db = useDb()
  const rs = await db.execute({
    sql: 'SELECT * FROM reminders WHERE id = ? AND user_id = ?',
    args: [id, userId],
  })
  return rs.rows[0] ? mapRow(rs.rows[0] as Row) : null
}

/** Tạo reminder mới */
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

  // origin_time: mỏ neo tính lịch. Thiếu → suy từ next_action_at, rồi tới now
  const originTime = data.origin_time ?? nextActionAt ?? nowIso

  const row = {
    id,
    user_id: userId,
    title: data.title,
    description: data.description ?? null,
    tag: data.tag ?? null,
    type: data.type,
    status: 'active',
    recurrence_pattern: data.recurrence_pattern ? JSON.stringify(data.recurrence_pattern) : null,
    repeat_strategy: data.repeat_strategy ?? 'none',
    calendar_type: data.calendar_type ?? 'solar',
    origin_time: originTime,
    max_crp: data.max_crp ?? 0,
    crp_interval_sec: data.crp_interval_sec ?? 0,
    is_sended_one_time: 0,
    // next_recurring CHỈ dành cho recurring (Go từ chối one_time có next_recurring)
    next_recurring: data.type === 'recurring' ? nextActionAt : null,
    next_crp: nextActionAt,
    next_action_at: nextActionAt,
  }

  await db.execute({
    sql: `INSERT INTO reminders
      (id, user_id, title, description, tag, type, status, recurrence_pattern,
       repeat_strategy, calendar_type, origin_time, max_crp, crp_interval_sec,
       is_sended_one_time, next_recurring, next_crp, next_action_at, created_at, updated_at)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
    args: [
      row.id, row.user_id, row.title, row.description, row.tag, row.type, row.status,
      row.recurrence_pattern, row.repeat_strategy, row.calendar_type, row.origin_time,
      row.max_crp, row.crp_interval_sec, row.is_sended_one_time,
      row.next_recurring, row.next_crp, row.next_action_at, nowIso, nowIso,
    ],
  })

  return mapRow({ ...row, created_at: nowIso, updated_at: nowIso })
}

/**
 * Cập nhật reminder — merge "có gửi thì lấy", giống Go (reminder_service.go:103-129).
 * KHÔNG ghi đè bằng `?? default` (client quên gửi description sẽ bị xoá mất).
 */
export async function updateReminder(
  id: string,
  userId: string,
  data: Partial<ReminderPayload>,
): Promise<Reminder | null> {
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

  const now = new Date()
  const patternChanged =
    data.recurrence_pattern !== undefined
    || data.origin_time !== undefined
    || data.calendar_type !== undefined
    || data.next_action_at !== undefined

  if (data.next_action_at !== undefined) merged.next_action_at = data.next_action_at
  if (merged.type === 'recurring' && patternChanged) {
    merged.next_recurring = nextRecurring(merged, now).toISOString()
  }
  if (patternChanged) merged.next_action_at = nextActionAt(merged, now)

  merged.updated_at = now.toISOString()

  const args: InValue[] = [
    merged.title, merged.description ?? null, merged.tag ?? null, merged.type,
    merged.recurrence_pattern ? JSON.stringify(merged.recurrence_pattern) : null,
    merged.repeat_strategy, merged.calendar_type, merged.origin_time ?? null,
    merged.max_crp, merged.crp_interval_sec,
    merged.next_recurring ?? null, merged.next_action_at ?? null,
    merged.updated_at, id, userId,
  ]

  await db.execute({
    sql: `UPDATE reminders SET title = ?, description = ?, tag = ?, type = ?,
            recurrence_pattern = ?, repeat_strategy = ?, calendar_type = ?, origin_time = ?,
            max_crp = ?, crp_interval_sec = ?, next_recurring = ?, next_action_at = ?, updated_at = ?
          WHERE id = ? AND user_id = ?`,
    args,
  })

  return getReminder(id, userId)
}

/** Xoá reminder */
export async function deleteReminder(id: string, userId: string): Promise<boolean> {
  const db = useDb()
  const rs = await db.execute({
    sql: 'DELETE FROM reminders WHERE id = ? AND user_id = ?',
    args: [id, userId],
  })
  return rs.rowsAffected > 0
}
