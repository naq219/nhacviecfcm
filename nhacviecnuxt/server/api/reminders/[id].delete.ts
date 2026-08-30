import { deleteReminder } from '../../utils/reminders'

export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  const id = getRouterParam(event, 'id')!

  const ok = await deleteReminder(id, session.user.id)
  if (!ok) throw createError({ statusCode: 404, message: 'Không tìm thấy reminder' })
  return { ok: true }
})
