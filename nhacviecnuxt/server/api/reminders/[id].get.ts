import { getReminder } from '../../utils/reminders'

export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  const id = getRouterParam(event, 'id')!

  const r = await getReminder(id, session.user.id)
  if (!r) throw createError({ statusCode: 404, message: 'Không tìm thấy reminder' })
  return r
})
