import { getSystemStatus } from '../utils/system-status'

export default defineEventHandler(async (event) => {
  // Route quản trị — Go đang để hở, bản này BẮT BUỘC auth
  await requireUserSession(event)
  return getSystemStatus()
})
