/**
 * Băm mật khẩu bằng Web Crypto (PBKDF2-SHA256).
 *
 * ⚠️ Tại sao không dùng `hashPassword` / `verifyPassword` của nuxt-auth-utils?
 * Hai hàm đó dùng `@adonisjs/hash` + driver scrypt, phụ thuộc `node:crypto`
 * (`scrypt` + `timingSafeEqual`). Trên Cloudflare Workers:
 *   - `hashPassword` (tạo hash) chạy được,
 *   - NHƯNG `verifyPassword` luôn trả false → **không đăng nhập được**.
 * Đã kiểm chứng: hash lưu trên production đem verify ở Node local → true,
 * verify trên Workers → false.
 *
 * Web Crypto (`crypto.subtle`) là native trên Workers nên không cần bật
 * compatibility flag `nodejs_compat`.
 */

const ITERATIONS = 100_000
const KEY_LENGTH_BYTES = 32
const SALT_LENGTH_BYTES = 16
const SCHEME = 'pbkdf2-sha256'

function toBase64(bytes: Uint8Array): string {
  let s = ''
  for (const b of bytes) s += String.fromCharCode(b)
  return btoa(s)
}

function fromBase64(value: string): Uint8Array {
  const bin = atob(value)
  const out = new Uint8Array(bin.length)
  for (let i = 0; i < bin.length; i++) out[i] = bin.charCodeAt(i)
  return out
}

async function derive(password: string, salt: Uint8Array, iterations: number): Promise<Uint8Array> {
  const key = await crypto.subtle.importKey(
    'raw',
    new TextEncoder().encode(password),
    'PBKDF2',
    false,
    ['deriveBits'],
  )
  const bits = await crypto.subtle.deriveBits(
    { name: 'PBKDF2', hash: 'SHA-256', salt: salt as BufferSource, iterations },
    key,
    KEY_LENGTH_BYTES * 8,
  )
  return new Uint8Array(bits)
}

/** So sánh 2 mảng byte trong thời gian không đổi (tránh timing attack) */
function constantTimeEqual(a: Uint8Array, b: Uint8Array): boolean {
  if (a.length !== b.length) return false
  let diff = 0
  for (let i = 0; i < a.length; i++) diff |= a[i]! ^ b[i]!
  return diff === 0
}

/**
 * Tạo hash. Định dạng:
 *   pbkdf2-sha256$<iterations>$<salt base64>$<hash base64>
 */
export async function makePasswordHash(password: string): Promise<string> {
  const salt = crypto.getRandomValues(new Uint8Array(SALT_LENGTH_BYTES))
  const hash = await derive(password, salt, ITERATIONS)
  return `${SCHEME}$${ITERATIONS}$${toBase64(salt)}$${toBase64(hash)}`
}

/**
 * Kiểm tra mật khẩu.
 * Trả false (không ném lỗi) nếu hash sai định dạng, để route tự trả 401.
 */
export async function verifyPasswordHash(stored: string, password: string): Promise<boolean> {
  const parts = stored.split('$')
  if (parts.length !== 4) return false
  const [scheme, iterRaw, saltRaw, hashRaw] = parts as [string, string, string, string]
  if (scheme !== SCHEME) return false

  const iterations = Number(iterRaw)
  if (!Number.isInteger(iterations) || iterations <= 0) return false

  try {
    const salt = fromBase64(saltRaw)
    const expected = fromBase64(hashRaw)
    const actual = await derive(password, salt, iterations)
    return constantTimeEqual(actual, expected)
  }
  catch {
    return false
  }
}
