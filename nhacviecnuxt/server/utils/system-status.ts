import type { SystemStatus } from '#shared/types'
import { getAccessToken, isFcmConfigured } from './fcm'

/**
 * Kill-switch cho cron — port từ `system_status` của Go.
 *
 * Quy tắc (internal/worker/common.go):
 *  - Lỗi do user (không có token / tắt FCM) → CHỈ log, giữ worker_enabled.
 *  - Lỗi FCM hệ thống → ghi last_error rồi TẮT worker.
 */

export async function isWorkerEnabled(): Promise<boolean> {
  const db = useDb()
  const rs = await db.execute({
    sql: 'SELECT worker_enabled FROM system_status WHERE mid = 1',
  })
  return Number(rs.rows[0]?.worker_enabled) === 1
}

export async function getSystemStatus(): Promise<SystemStatus> {
  const db = useDb()
  const rs = await db.execute({
    sql: 'SELECT mid, worker_enabled, last_error, updated_at FROM system_status WHERE mid = 1',
  })
  const row = rs.rows[0] as Record<string, unknown> | undefined
  return {
    mid: 1,
    worker_enabled: Number(row?.worker_enabled) === 1 ? 1 : 0,
    last_error: typeof row?.last_error === 'string' ? row.last_error : null,
    updated_at: String(row?.updated_at ?? ''),
  }
}

/** Tắt worker + ghi lý do (dùng khi FCM lỗi hệ thống) */
export async function disableWorker(errorMessage: string): Promise<void> {
  const db = useDb()
  await db.execute({
    sql: 'UPDATE system_status SET worker_enabled = 0, last_error = ?, updated_at = ? WHERE mid = 1',
    args: [errorMessage.slice(0, 500), new Date().toISOString()],
  })
}

/**
 * Kiểm tra FCM có thực sự dùng được trong runtime hiện tại không.
 * Quan trọng vì Web Crypto trên Workers có thể khác Node — cần xác minh
 * tận nơi chứ không chỉ test ở local.
 */
export async function checkFcmHealth(): Promise<{
  configured: boolean
  accessToken: boolean
  error?: string
}> {
  if (!isFcmConfigured()) {
    return { configured: false, accessToken: false }
  }
  try {
    const token = await getAccessToken()
    return { configured: true, accessToken: Boolean(token) }
  }
  catch (e) {
    return { configured: true, accessToken: false, error: String(e).slice(0, 200) }
  }
}

/** Bật/tắt worker thủ công (route quản trị) */
export async function setWorkerEnabled(enabled: boolean, errorMessage = ''): Promise<void> {
  const db = useDb()
  await db.execute({
    sql: 'UPDATE system_status SET worker_enabled = ?, last_error = ?, updated_at = ? WHERE mid = 1',
    args: [enabled ? 1 : 0, errorMessage, new Date().toISOString()],
  })
}
