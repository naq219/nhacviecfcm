import type { PublicUser } from '#shared/types'
import { registerDevice } from './devices'
import { makePasswordHash, verifyPasswordHash } from './password'

/**
 * Auth service.
 *
 * ⚠️ KHÔNG dùng `hashPassword` / `verifyPassword` của nuxt-auth-utils —
 * chúng dựa trên scrypt của `node:crypto` và bị hỏng trên Cloudflare Workers
 * (verify luôn false). Dùng Web Crypto PBKDF2, xem `server/utils/password.ts`.
 */

interface UserRow {
  id: string
  email: string
  password_hash: string
}

/** Tạo user mới. Trả null nếu email đã tồn tại. */
export async function createUser(email: string, password: string): Promise<PublicUser | null> {
  const db = useDb()

  const exist = await db.execute({
    sql: 'SELECT id FROM users WHERE email = ?',
    args: [email],
  })
  if (exist.rows.length > 0) return null

  const id = crypto.randomUUID()
  const now = new Date().toISOString()

  await db.execute({
    sql: `INSERT INTO users (id, email, password_hash, is_fcm_active, created_at, updated_at)
          VALUES (?, ?, ?, 1, ?, ?)`,
    args: [id, email, await makePasswordHash(password), now, now],
  })

  return { id, email }
}

/** Kiểm tra email + password. Trả null nếu sai. */
export async function verifyUser(email: string, password: string): Promise<PublicUser | null> {
  const db = useDb()

  const rs = await db.execute({
    sql: 'SELECT id, email, password_hash FROM users WHERE email = ?',
    args: [email],
  })
  const row = rs.rows[0] as unknown as UserRow | undefined
  if (!row) return null

  const ok = await verifyPasswordHash(row.password_hash, password)
  if (!ok) return null

  return { id: row.id, email: row.email }
}

/**
 * Lưu token FCM của thiết bị (route CŨ của app Android).
 *
 * Vẫn ghi vào `users.fcm_token` để không broke app cũ, NHƯNG đồng thời
 * upsert vào bảng `devices` → dữ liệu tự đồng bộ với hệ thống đa thiết bị
 * mà không cần chạy migration.
 */
export async function updateFcmToken(userId: string, token: string): Promise<void> {
  const db = useDb()
  await db.execute({
    sql: 'UPDATE users SET fcm_token = ?, is_fcm_active = 1, updated_at = ? WHERE id = ?',
    args: [token, new Date().toISOString(), userId],
  })

  await registerDevice(userId, token, 'android')
}
