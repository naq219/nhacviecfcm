import { setWorkerEnabled } from '../utils/system-status'

export default defineEventHandler(async (event) => {
  // Route quản trị — Go đang để hở, bản này BẮT BUỘC auth
  await requireUserSession(event)
  const { worker_enabled, last_error } = await readBody(event) as {
    worker_enabled?: boolean
    last_error?: string
  }

  await setWorkerEnabled(worker_enabled === true, last_error ?? '')
  return { ok: true }
})
