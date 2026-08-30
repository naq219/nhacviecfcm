import type { Device, DevicePlatform } from '#shared/types'
import { listDevices, registerDevice, unregisterDevice } from '../../utils/devices'

const PLATFORMS: DevicePlatform[] = ['android', 'ios', 'web']

/** Đăng ký 1 thiết bị nhận thông báo (web / Android mới / iOS) */
export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  const { token, platform } = await readBody(event) as {
    token?: string
    platform?: string
  }

  if (!token || typeof token !== 'string') {
    throw createError({ statusCode: 400, message: 'token là bắt buộc' })
  }

  const plat = (platform ?? 'web') as DevicePlatform
  if (!PLATFORMS.includes(plat)) {
    throw createError({ statusCode: 400, message: 'platform phải là android | ios | web' })
  }

  await registerDevice(session.user.id, token, plat)
  return { ok: true, platform: plat }
})
