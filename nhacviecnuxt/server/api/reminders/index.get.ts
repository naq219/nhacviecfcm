import type { ReminderStatus } from '#shared/types'
import { listReminders } from '../../utils/reminders'

export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  const status = getQuery(event).status as ReminderStatus | undefined
  return listReminders(session.user.id, status)
})
