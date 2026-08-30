/**
 * Kiểm tra service account FCM có dùng được không.
 *
 * Dùng CHÍNH XÁC cùng logic Web Crypto với server/utils/fcm.ts
 * (ký JWT RS256 → đổi lấy access token OAuth2).
 * Chạy local được vì Node 24 có sẵn globalThis.crypto.subtle.
 *
 *   node scripts/verify-fcm.mjs
 */
import { readFileSync, existsSync } from 'node:fs'

const CRED = 'E:/PROJECT/nhacviecfcm/firebase-credentials.json'
const TOKEN_URI = 'https://oauth2.googleapis.com/token'

if (!existsSync(CRED)) {
  console.error('Khong tim thay', CRED)
  process.exit(1)
}

const cred = JSON.parse(readFileSync(CRED, 'utf8'))
console.log('project_id  :', cred.project_id)
console.log('client_email:', cred.client_email)

// ---- Cùng logic với pemToArrayBuffer() trong server/utils/fcm.ts ----
function pemToArrayBuffer(pem) {
  const b64 = pem
    .replace(/-----BEGIN PRIVATE KEY-----/, '')
    .replace(/-----END PRIVATE KEY-----/, '')
    .replace(/\\n/g, '')
    .replace(/\s/g, '')
  const bin = atob(b64)
  const buf = new Uint8Array(bin.length)
  for (let i = 0; i < bin.length; i++) buf[i] = bin.charCodeAt(i)
  return buf.buffer
}

function b64url(bytes) {
  let s = ''
  for (const b of bytes) s += String.fromCharCode(b)
  return btoa(s).replace(/=+$/, '').replace(/\+/g, '-').replace(/\//g, '_')
}

function encObj(o) {
  return btoa(JSON.stringify(o)).replace(/=+$/, '').replace(/\+/g, '-').replace(/\//g, '_')
}

async function getAccessToken() {
  const now = Math.floor(Date.now() / 1000)
  const header = { alg: 'RS256', typ: 'JWT' }
  const claims = {
    iss: cred.client_email,
    scope: 'https://www.googleapis.com/auth/firebase.messaging',
    aud: TOKEN_URI,
    iat: now,
    exp: now + 3600,
  }
  const unsigned = `${encObj(header)}.${encObj(claims)}`

  const key = await crypto.subtle.importKey(
    'pkcs8',
    pemToArrayBuffer(cred.private_key),
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
  return { res, data: await res.json().catch(() => ({})) }
}

const { res, data } = await getAccessToken()

if (res.ok && data.access_token) {
  console.log('\nOK  lay access token thanh cong')
  console.log('    token length :', String(data.access_token).length)
  console.log('    token_type   :', data.token_type)
  console.log('    expires_in   :', data.expires_in)
}
else {
  console.log('\nFAIL  khong lay duoc access token')
  console.log('    HTTP     :', res.status)
  console.log('    response :', JSON.stringify(data).slice(0, 600))
  process.exit(1)
}
