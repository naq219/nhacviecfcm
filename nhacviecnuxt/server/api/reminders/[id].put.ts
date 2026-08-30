import type { ReminderPayload } from '#shared/types'
import { updateReminder } from '../../utils/reminders'
import { validateReminderPayload } from '../../utils/validate'

export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  const id = getRouterParam(event, 'id')!
  const body = await readBody(event) as Partial<ReminderPayload>

  // validate từng trường được gửi (payload có thể là partial)
  const error = validateReminderPayload({ ...body, title: body.title ?? 'x' })
  if (error) throw createError({ statusCode: 400, message: error })

  const r = await updateReminder(id, session.user.id, body)
  if (!r) throw createError({ statusCode: 404, message: 'Không tìm thấy reminder' })
  return r
})
