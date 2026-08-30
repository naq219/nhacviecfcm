import type { DevicePlatform } from '#shared/types'

/**
 * Quản lý nhiều thiết bị nhận thông báo của 1 user.
 *
 * ⚠️ Tương thích ngược: app Android cũ vẫn gọi `PUT /api/users/fcm-token`
 * và chỉ ghi vào `users.fcm_token`. Vì vậy `listActiveTokens()` gộp cả 2 nguồn:
 *  1. bảng `devices` (chuẩn mới)
 *  2. cột `users.fcm_token` (dữ liệu cũ + app chưa cập nhật)
 * và khử trùng theo token → không gửi 2 lần cho cùng 1 thiết bị.
 */

export interface ActiveToken {
  token: string
  platform: DevicePlatform
}

/** Đăng ký / làm mới 1 thiết bị (upsert theo token) */
export async function registerDevice(
  userId: string,
  token: string,
  platform: DevicePlatform,
): Promise<void> {
  const db = useDb()
  const now = new Date().toISOString()

  const exist = await db.execute({
    sql: 'SELECT id FROM devices WHERE token = ?',
    args: [token],
  })

  if (exist.rows.length > 0) {
    await db.execute({
      sql: `UPDATE devices SET user_id = ?, platform = ?, is_active = 1, updated_at = ?
            WHERE token = ?`,
      args: [userId, platform, now, token],
    })
    return
  }

  await db.execute({
    sql: `INSERT INTO devices (id, user_id, token, platform, is_active, created_at, updated_at)
          VALUES (?, ?, ?, ?, 1, ?, ?)`,
    args: [crypto.randomUUID(), userId, token, platform, now, now],
  })
}

/** Gỡ đăng ký (logout / user tắt thông báo trên thiết bị đó) */
export async function unregisterDevice(userId: string, token: string): Promise<boolean> {
  const db = useDb()
  const rs = await db.execute({
    sql: 'DELETE FROM devices WHERE user_id = ? AND token = ?',
    args: [userId, token],
  })
  return rs.rowsAffected > 0
}

/**
 * Tắt 1 token chết (FCM trả UNREGISTERED / NOT_FOUND).
 * Chỉ tắt đúng token đó — KHÔNG tắt cả user như bản cũ.
 */
export async function deactivateToken(token: string): Promise<void> {
  const db = useDb()
  const now = new Date().toISOString()

  await db.execute({
    sql: 'UPDATE devices SET is_active = 0, updated_at = ? WHERE token = ?',
    args: [now, token],
  })

  // Tương thích ngược: dọn luôn cột cũ nếu trùng
  await db.execute({
    sql: 'UPDATE users SET is_fcm_active = 0, updated_at = ? WHERE fcm_token = ?',
    args: [now, token],
  })
}

/** Danh sách các thiết bị của user */
export async function listDevices(userId: string): Promise<{ token: string, platform: DevicePlatform, is_active: 0 | 1 }[]> {
  const db = useDb()
  const rs = await db.execute({
    sql: `SELECT token, platform, is_active FROM devices
          WHERE user_id = ? ORDER BY created_at ASC`,
    args: [userId],
  })
  return rs.rows.map((r) => ({
    token: String(r.token),
    platform: String(r.platform ?? 'android') as DevicePlatform,
    is_active: Number(r.is_active) === 1 ? 1 : 0,
  }))
}

/**
 * Các token đang hoạt động để gửi thông báo (đã khử trùng).
 * Gồm cả token legacy trong `users.fcm_token`.
 */
export async function listActiveTokens(userId: string): Promise<ActiveToken[]> {
  const db = useDb()

  const rs = await db.execute({
    sql: 'SELECT token, platform FROM devices WHERE user_id = ? AND is_active = 1',
    args: [userId],
  })

  const result: ActiveToken[] = rs.rows.map(r => ({
    token: String(r.token),
    platform: String(r.platform ?? 'android') as DevicePlatform,
  }))

  // Legacy: users.fcm_token (app Android cũ chưa migrate sang /api/devices)
  const legacy = await db.execute({
    sql: `SELECT fcm_token FROM users
          WHERE id = ? AND is_fcm_active = 1
            AND fcm_token IS NOT NULL AND fcm_token <> ''`,
    args: [userId],
  })

  const seen = new Set(result.map(t => t.token))
  for (const row of legacy.rows) {
    const token = String(row.fcm_token)
    if (!seen.has(token)) {
      seen.add(token)
      result.push({ token, platform: 'android' })
    }
  }

  return result
}
