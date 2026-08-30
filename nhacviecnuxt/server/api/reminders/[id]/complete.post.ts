import { applyComplete } from '../../../utils/reminder-processor'
import { getReminder } from '../../../utils/reminders'

export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  const id = getRouterParam(event, 'id')!

  const current = await getReminder(id, session.user.id)
  if (!current) throw createError({ statusCode: 404, message: 'Không tìm thấy reminder' })

  const now = new Date()
  const changes = applyComplete(current, now)

  const db = useDb()
  await db.execute({
    sql: `UPDATE reminders SET
            status = COALESCE(?, status),
            last_completed_at = ?,
            last_crp_completed_at = ?,
            crp_count = ?,
            next_recurring = ?,
            next_crp = ?,
            next_action_at = ?,
            updated_at = ?
          WHERE id = ? AND user_id = ?`,
    args: [
      changes.status ?? null,
      changes.last_completed_at ?? null,
      changes.last_crp_completed_at ?? null,
      changes.crp_count ?? null,
      changes.next_recurring ?? null,
      changes.next_crp ?? null,
      changes.next_action_at ?? null,
      now.toISOString(),
      id,
      session.user.id,
    ],
  })

  return await getReminder(id, session.user.id)
})
