import type { Reminder } from '#shared/types'
import { processReminder } from './reminder-processor'
import { isWorkerEnabled } from './system-status'

/** Số reminder xử lý tối đa mỗi lần chạy (Cloudflare có giới hạn thời gian) */
const BATCH_LIMIT = 100

export interface RunCheckResult {
  result: string
  due: number
  sent: number
  failed: number
}

/**
 * Một lần chạy "quét reminder đến hạn và gửi FCM".
 *
 * Được gọi bởi 2 nơi:
 *  - `server/tasks/reminders/check.ts` (Nitro task) — chỉ hoạt động nền tảng có cron
 *  - `server/api/cron/reminders-check.post.ts` — do Worker cron gọi vào
 *
 * ⚠️ Lý do phải có đường HTTP: **Cloudflare Pages KHÔNG hỗ trợ Cron Triggers**.
 * Handler `scheduled()` của Nitro vẫn được sinh ra nhưng không bao giờ được gọi.
 */
export async function runReminderCheck(): Promise<RunCheckResult> {
  if (!await isWorkerEnabled()) {
    return { result: 'worker đang tắt', due: 0, sent: 0, failed: 0 }
  }

  const db = useDb()
  const nowIso = new Date().toISOString()

  const rs = await db.execute({
    sql: `SELECT * FROM reminders
           WHERE status = 'active'
             AND next_action_at IS NOT NULL
             AND next_action_at <= ?
             AND (snooze_until IS NULL OR snooze_until <= ?)
           ORDER BY next_action_at ASC
           LIMIT ?`,
    args: [nowIso, nowIso, BATCH_LIMIT],
  })

  let sent = 0
  let failed = 0

  for (const row of rs.rows) {
    try {
      await processReminder(row as unknown as Reminder)
      sent++
    }
    catch (e) {
      failed++
      console.error(`[reminder] lỗi ${(row as { id?: string }).id}:`, e)
    }
  }

  return {
    result: `Đã xử lý ${sent} reminder, ${failed} lỗi`,
    due: rs.rows.length,
    sent,
    failed,
  }
}
