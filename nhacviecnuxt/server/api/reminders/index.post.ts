import type { ReminderPayload } from '#shared/types'
import { createReminder } from '../../utils/reminders'
import { validateReminderPayload } from '../../utils/validate'

export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  const body = await readBody(event) as Partial<ReminderPayload>

  const error = validateReminderPayload(body)
  if (error) throw createError({ statusCode: 400, message: error })

  return createReminder(session.user.id, body as ReminderPayload)
})
