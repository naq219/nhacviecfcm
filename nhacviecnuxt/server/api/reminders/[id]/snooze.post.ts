import { applySnooze } from '../../../utils/reminder-processor'
import { getReminder } from '../../../utils/reminders'

const MAX_SNOOZE_SEC = 7 * 24 * 3600

export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  const id = getRouterParam(event, 'id')!
  const { duration } = await readBody(event) as { duration?: number }

  const seconds = Number(duration)
  // Go không validate → gửi số âm sẽ set snooze_until trong quá khứ
  if (!Number.isFinite(seconds) || seconds <= 0 || seconds > MAX_SNOOZE_SEC) {
    throw createError({ statusCode: 400, message: 'duration phải từ 1 giây đến 7 ngày' })
  }

  const current = await getReminder(id, session.user.id)
  if (!current) throw createError({ statusCode: 404, message: 'Không tìm thấy reminder' })

  const db = useDb()
  const now = new Date()
  const changes = applySnooze(now, seconds)

  await db.execute({
    sql: `UPDATE reminders SET snooze_until = ?, next_action_at = ?, updated_at = ?
          WHERE id = ? AND user_id = ?`,
    args: [changes.snooze_until!, changes.next_action_at!, now.toISOString(), id, session.user.id],
  })

  return await getReminder(id, session.user.id)
})
