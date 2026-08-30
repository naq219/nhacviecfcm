/**
 * Gửi push FCM qua HTTP v1 API bằng Web Crypto.
 * `firebase-admin` KHÔNG chạy trên Cloudflare Workers (cần Node API).
 */

const TOKEN_URI = 'https://oauth2.googleapis.com/token'

export interface FcmResult {
  ok: boolean
  /** token chết / không hợp lệ → có thể dọn (không phải lỗi hệ thống) */
  unregistered: boolean
  message: string
}

function pemToArrayBuffer(pem: string): ArrayBuffer {
  const b64 = pem
    .replace(/-----BEGIN PRIVATE KEY-----/, '')
    .replace(/-----END PRIVATE KEY-----/, '')
    .replace(/\\n/g, '') // còn literal "\n" do .env không unescape
    .replace(/\s/g, '') // xoá khoảng trắng / xuống dòng thật
  const bin = atob(b64)
  const buf = new Uint8Array(bin.length)
  for (let i = 0; i < bin.length; i++) buf[i] = bin.charCodeAt(i)
  return buf.buffer
}

function b64url(bytes: Uint8Array): string {
  let s = ''
  for (const b of bytes) s += String.fromCharCode(b)
  return btoa(s).replace(/=+$/, '').replace(/\+/g, '-').replace(/\//g, '_')
}

function encObj(o: object): string {
  return btoa(JSON.stringify(o)).replace(/=+$/, '').replace(/\+/g, '-').replace(/\//g, '_')
}

/**
 * FCM đã được cấu hình chưa?
 *
 * Quan trọng: thiếu cấu hình KHÔNG PHẢI lỗi hệ thống. Nếu không check, cron sẽ
 * gọi FCM → 401 → kill-switch tắt worker vĩnh viễn, và khi khai báo FCM sau này
 * worker vẫn bị tắt (phải bật tay).
 */
export function isFcmConfigured(): boolean {
  const cfg = useRuntimeConfig()
  return Boolean(cfg.fcmProjectId && cfg.fcmClientEmail && cfg.fcmPrivateKey)
}

/** Lấy access token OAuth2 bằng JWT (RS256) ký bằng Web Crypto */
async function getAccessToken(): Promise<string> {
  const cfg = useRuntimeConfig()
  const now = Math.floor(Date.now() / 1000)

  const header = { alg: 'RS256', typ: 'JWT' }
  const claims = {
    iss: cfg.fcmClientEmail,
    scope: 'https://www.googleapis.com/auth/firebase.messaging',
    aud: TOKEN_URI,
    iat: now,
    exp: now + 3600,
  }
  const unsigned = `${encObj(header)}.${encObj(claims)}`

  const key = await crypto.subtle.importKey(
    'pkcs8',
    pemToArrayBuffer(cfg.fcmPrivateKey),
    { name: 'RSASSA-PKCS1-v1_5', hash: 'SHA-256' },
    false,
    ['sign'],
  )
  const sig = await crypto.subtle.sign(
    'RSASSA-PKCS1-v1_5',
    key,
    new TextEncoder().encode(unsigned),
  )
  const assertion = `${unsigned}.${b64url(new Uint8Array(sig))}`

  const res = await fetch(TOKEN_URI, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'urn:ietf:params:oauth:grant-type:jwt-bearer',
      assertion,
    }),
  })
  const data = await res.json() as { access_token?: string }
  return data.access_token as string
}

/**
 * Gửi 1 push tới 1 device token (dùng chung cho Android / iOS / web).
 *
 * - `notification`: Android + iOS + web đều hiển thị được.
 * - `webpush.fcmOptions.link`: CHỈ web dùng (bấm vào mở link).
 *   Android/iOS bỏ qua trường này nên gửi luôn cũng không sao.
 */
export async function sendFcm(
  token: string,
  title: string,
  body?: string | null,
  link?: string,
): Promise<FcmResult> {
  const cfg = useRuntimeConfig()
  const accessToken = await getAccessToken()

  const message: Record<string, unknown> = {
    token,
    notification: { title, body: body ?? '' },
  }
  if (link) {
    message.webpush = { fcmOptions: { link } }
  }

  const res = await fetch(
    `https://fcm.googleapis.com/v1/projects/${cfg.fcmProjectId}/messages:send`,
    {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ message }),
    },
  )

  if (res.ok) return { ok: true, unregistered: false, message: '' }

  // Phân biệt token chết (dọn được) và lỗi hệ thống (tắt worker)
  const text = await res.text()
  const unregistered = /UNREGISTERED|NOT_FOUND|INVALID_ARGUMENT/.test(text)
  return { ok: false, unregistered, message: text.slice(0, 500) }
}
