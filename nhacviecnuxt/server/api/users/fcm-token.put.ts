import { updateFcmToken } from '../../utils/auth'

/**
 * ⚠️ Route CŨ — app Android đang dùng. GIỮ LẠI để không broke app hiện tại.
 * App mới nên dùng `POST /api/devices` (hỗ trợ nhiều thiết bị).
 *
 * Route này vẫn ghi vào users.fcm_token VÀ upsert luôn vào bảng devices,
 * nên dữ liệu cũ tự đồng bộ mà không cần migrate.
 */
export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  const { fcm_token } = await readBody(event) as { fcm_token?: string }

  if (!fcm_token || typeof fcm_token !== 'string') {
    throw createError({ statusCode: 400, message: 'fcm_token là bắt buộc' })
  }

  await updateFcmToken(session.user.id, fcm_token)
  return { ok: true }
})
