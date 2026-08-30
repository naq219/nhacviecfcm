import type { InValue } from '@libsql/client'
import type { Reminder } from '#shared/types'
import { isFcmConfigured, sendFcm } from './fcm'
import { crpIntervalMs, nextActionAt, nextRecurring, nextRecurringFromComplete } from './schedule'
import { getReminderRaw } from './reminders'
import { deactivateToken, listActiveTokens } from './devices'
import { disableWorker } from './system-status'

/**
 * Xử lý reminder đến hạn — port 11 nhánh worker của Go.
 * Đặc tả: docs/04b-tinh-lich-frp-crp.md#7-11-nhánh-worker
 */

// ---------- helper ----------

/** iso có phải thời điểm đã tới / quá khứ không; null → false */
function isPast(iso: string | null | undefined, now: Date): boolean {
  if (!iso) return false
  return Date.parse(iso) <= now.getTime()
}

/** user đã bấm Hoàn thành SAU lần gửi gần nhất chưa? (điều kiện nhánh C2/C4) */
function completedAfterLastSent(r: Reminder): boolean {
  if (!r.last_completed_at) return false
  if (!r.last_sent_at) return true
  return Date.parse(r.last_completed_at) > Date.parse(r.last_sent_at)
}

/** Ghi 1 lô thay đổi và TỰ TÍNH LẠI next_action_at */
async function patch(id: string, changes: Partial<Reminder>): Promise<void> {
  const db = useDb()
  const now = new Date()
  const merged = { ...(await getReminderRaw(id)), ...changes } as Reminder

  merged.next_action_at = nextActionAt(merged, now)
  merged.updated_at = now.toISOString()

  const keys = Array.from(new Set([...Object.keys(changes), 'next_action_at', 'updated_at']))
  const setSql = keys.map(k => `${k} = ?`).join(', ')
  const values = keys.map(k => ((merged as unknown as Record<string, unknown>)[k] ?? null) as InValue)

  await db.execute({
    sql: `UPDATE reminders SET ${setSql} WHERE id = ?`,
    args: [...values, id],
  })
}

/** Áp dụng thay đổi sau khi gửi FCM THÀNH CÔNG (11 nhánh A1–C5) */
async function applySent(r: Reminder, now: Date): Promise<void> {
  const nowIso = now.toISOString()
  const crpNext = new Date(now.getTime() + crpIntervalMs(r)).toISOString()

  // ---- Nhóm C: recurring + crp_until_complete ----
  if (r.type === 'recurring' && r.repeat_strategy === 'crp_until_complete') {
    if (
      r.max_crp > 0
      && r.is_sended_one_time === 1
      && r.crp_count < r.max_crp
      && isPast(r.next_crp, now)
      && !completedAfterLastSent(r)
    ) {
      // C5: retry CRP — không bao giờ tự completed
      return patch(r.id, { crp_count: r.crp_count + 1, last_sent_at: nowIso, next_crp: crpNext })
    }
    if (r.max_crp > 0 && !r.is_sended_one_time) {
      // C3: lần đầu, có CRP
      return patch(r.id, {
        is_sended_one_time: 1, last_sent_at: nowIso, crp_count: 0, next_crp: crpNext,
      })
    }
    if (r.max_crp > 0 && completedAfterLastSent(r)) {
      // C4: chu kỳ mới sau khi complete, có CRP
      return patch(r.id, { last_sent_at: nowIso, crp_count: 0, next_crp: crpNext })
    }
    if (!r.is_sended_one_time) {
      // C1: lần đầu, không CRP — next_recurring GIỮ NGUYÊN
      return patch(r.id, { is_sended_one_time: 1, last_sent_at: nowIso })
    }
    // C2: sau khi complete, không CRP
    return patch(r.id, { last_sent_at: nowIso })
  }

  // ---- Nhóm B: recurring + repeat_strategy = none ----
  if (r.type === 'recurring') {
    if (r.max_crp > 0 && isPast(r.next_recurring, now)) {
      // B2: sang chu kỳ FRP mới + reset CRP
      return patch(r.id, {
        last_sent_at: nowIso,
        next_recurring: nextRecurring(r, now).toISOString(),
        crp_count: 0,
        next_crp: crpNext,
      })
    }
    if (r.max_crp > 0 && r.crp_count < r.max_crp && isPast(r.next_crp, now)) {
      // B3: retry CRP trong cùng chu kỳ
      return patch(r.id, { crp_count: r.crp_count + 1, last_sent_at: nowIso, next_crp: crpNext })
    }
    // B1: không CRP
    return patch(r.id, {
      last_sent_at: nowIso,
      next_recurring: nextRecurring(r, now).toISOString(),
    })
  }

  // ---- Nhóm A: one_time ----
  if (r.max_crp > 0 && !r.is_sended_one_time) {
    // A2: gửi lần đầu, CHƯA completed (còn lượt retry)
    return patch(r.id, { is_sended_one_time: 1, last_sent_at: nowIso, next_crp: crpNext })
  }
  if (r.max_crp > 0 && r.is_sended_one_time) {
    // A3: retry CRP
    const crpCount = r.crp_count + 1
    const done = crpCount >= r.max_crp
    // ⚠️ Không truyền key nào có giá trị undefined — spread sẽ ghi đè thành NULL
    return patch(r.id, {
      crp_count: crpCount,
      last_sent_at: nowIso,
      next_crp: done ? null : crpNext,
      ...(done
        ? { status: 'completed' as const, last_completed_at: nowIso }
        : { next_action_at: crpNext }),
    })
  }
  // A1: one_time không CRP → completed luôn
  return patch(r.id, {
    is_sended_one_time: 1,
    status: 'completed',
    last_sent_at: nowIso,
    last_completed_at: nowIso,
  })
}

// ---------- entry ----------

/**
 * Link mở khi user bấm vào thông báo trên web (Android/iOS bỏ qua trường này).
 * Hiện chưa có trang chi tiết reminder → trỏ về trang chủ.
 * Khi thêm `app/pages/reminders/[id].vue` thì đổi thành `/reminders/${id}`.
 */
function webLink(): string {
  return '/'
}

/** Xử lý 1 reminder đến hạn: gửi FCM tới MỌI thiết bị + cập nhật trạng thái */
export async function processReminder(r: Reminder): Promise<void> {
  const now = new Date()

  // 1. Công tắc tổng: user tắt hẳn thông báo → bỏ qua
  const masterOn = await isMasterSwitchOn(r.user_id)
  if (!masterOn) {
    console.warn(`[reminder] bỏ qua ${r.id}: user đã tắt thông báo`)
    return
  }

  // 2. FCM chưa được cấu hình → bỏ qua, KHÔNG tắt worker
  //    (thiếu credentials không phải lỗi hệ thống)
  if (!isFcmConfigured()) {
    console.warn('[reminder] FCM chưa cấu hình — bỏ qua gửi thông báo')
    return
  }

  // 3. Lấy TẤT CẢ thiết bị đang hoạt động của user
  const targets = await listActiveTokens(r.user_id)
  if (targets.length === 0) {
    console.warn(`[reminder] bỏ qua ${r.id}: user chưa đăng ký thiết bị nào`)
    return
  }

  // 4. Gửi từng thiết bị. Chỉ cần 1 thiết bị nhận được là tính thành công.
  let anyOk = false

  for (const target of targets) {
    const link = target.platform === 'web' ? webLink() : undefined
    const res = await sendFcm(target.token, r.title, r.description, link)

    if (res.ok) {
      anyOk = true
      continue
    }

    if (res.unregistered) {
      // 5. Token chết → CHỈ tắt đúng token đó (bản cũ tắt cả user)
      await deactivateToken(target.token)
      console.warn(`[reminder] token hết hạn (${target.platform}), đã tắt: ${target.token.slice(0, 12)}…`)
      continue
    }

    // 6. Lỗi hệ thống FCM → kill-switch (không phải lỗi do user)
    await disableWorker(`FCM send failed for reminder ${r.id}: ${res.message}`)
    return
  }

  if (!anyOk) {
    console.warn(`[reminder] ${r.id}: không thiết bị nào nhận được (tất cả token đều chết)`)
    return
  }

  // 7. Gửi thành công ít nhất 1 thiết bị
  await applySent(r, now)
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

/**
 * Bấm Hoàn thành — port `OnUserComplete` (reminder_service.go:176).
 * ⚠️ recurring GIỮ status = 'active' (set completed là làm chết reminder lặp).
 */
export function applyComplete(r: Reminder, now: Date): Partial<Reminder> {
  const nowIso = now.toISOString()

  if (r.type === 'one_time') {
    return {
      status: 'completed',
      last_completed_at: nowIso,
      last_crp_completed_at: nowIso,
      crp_count: 0,
    }
  }

  const next = nextRecurringFromComplete(r, now).toISOString()
  return {
    last_completed_at: nowIso,
    last_crp_completed_at: nowIso,
    crp_count: 9999, // sentinel: đã xong CRP, chờ FRP mới
    next_recurring: next,
    next_crp: next,
    next_action_at: next,
    // status GIỮ NGUYÊN
  }
}

/** Snooze — port `SnoozeReminder` (reminder_service.go:160) */
export function applySnooze(now: Date, seconds: number): Partial<Reminder> {
  const until = new Date(now.getTime() + seconds * 1000).toISOString()
  return { snooze_until: until, next_action_at: until }
}
