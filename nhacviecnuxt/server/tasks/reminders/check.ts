import type { Reminder } from '#shared/types'
import { processReminder } from '../../utils/reminder-processor'
import { isWorkerEnabled } from '../../utils/system-status'

/** Số reminder xử lý tối đa mỗi lần chạy (Cloudflare có giới hạn thời gian) */
const BATCH_LIMIT = 100

export default defineTask({
  meta: {
    name: 'reminders:check',
    description: 'Quét reminder đến hạn và gửi FCM',
  },
  async run() {
    // 0. Kill-switch
    if (!await isWorkerEnabled()) {
      return { result: 'worker đang tắt' }
    }

    const db = useDb()
    const nowIso = new Date().toISOString()

    // 1. Lấy lô reminder đến hạn
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

    return { result: `Đã xử lý ${sent} reminder, ${failed} lỗi` }
  },
})
