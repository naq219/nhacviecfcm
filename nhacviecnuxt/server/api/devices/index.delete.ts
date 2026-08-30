import { unregisterDevice } from '../../utils/devices'

/**
 * Gỡ đăng ký thiết bị (logout / tắt thông báo trên thiết bị đó).
 * Body: { token } — hoặc gọi không body để gỡ bằng query.
 */
export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  const body = await readBody(event).catch(() => ({})) as { token?: string }
  const token = body.token ?? (getQuery(event).token as string | undefined)

  if (!token) {
    throw createError({ statusCode: 400, message: 'token là bắt buộc' })
  }

  const ok = await unregisterDevice(session.user.id, token)
  if (!ok) throw createError({ statusCode: 404, message: 'Không tìm thấy thiết bị' })
  return { ok: true }
})
