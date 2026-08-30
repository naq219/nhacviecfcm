/**
 * Đẩy toàn bộ env lên Cloudflare Pages dưới dạng secret.
 *
 *   node scripts/setup-cloudflare-secrets.mjs [project]
 *
 * - 3 biến server FCM được đọc TRỰC TIẾP từ firebase-credentials.json
 *   → private key KHÔNG bao giờ hiện ra ở log / chat.
 * - Các biến NUXT_PUBLIC_* là cấu hình Firebase web, vốn đã public
 *   (nằm ngay trong bundle trình duyệt), nên an toàn khi lưu.
 */
import { readFileSync, existsSync } from 'node:fs'
import { spawn } from 'node:child_process'
import { fileURLToPath } from 'node:url'

const PROJECT = process.argv[2] || 'sennote'
const CRED = 'E:/PROJECT/nhacviecfcm/firebase-credentials.json'

// Gọi trực tiếp entry JS của wrangler bằng node → không cần shell,
// không phụ thuộc .cmd hay PATH (tránh lỗi "not recognized" trên Windows)
const WRANGLER_JS = fileURLToPath(new URL('../node_modules/wrangler/bin/wrangler.js', import.meta.url))

if (!existsSync(CRED)) {
  console.error('Khong tim thay', CRED)
  process.exit(1)
}

const cred = JSON.parse(readFileSync(CRED, 'utf8'))

/** Cấu hình Firebase web (public — nằm trong bundle trình duyệt) */
const FIREBASE_WEB = {
  NUXT_PUBLIC_FIREBASE_API_KEY: 'AIzaSyAxYNmylFjzfAIIK4fm94en5svpL9YgFZ8',
  NUXT_PUBLIC_FIREBASE_AUTH_DOMAIN: 'reminaq-001.firebaseapp.com',
  NUXT_PUBLIC_FIREBASE_PROJECT_ID: 'reminaq-001',
  NUXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID: '679219217368',
  NUXT_PUBLIC_FIREBASE_APP_ID: '1:679219217368:web:f90e2264ab388f3843e8ad',
  // VAPID: chờ cung cấp, bỏ qua nếu rỗng
  NUXT_PUBLIC_FCM_VAPID_KEY: process.env.VAPID_KEY || '',
}

const secrets = {
  // Server FCM — đọc từ file, không hardcode
  NUXT_FCM_PROJECT_ID: cred.project_id,
  NUXT_FCM_CLIENT_EMAIL: cred.client_email,
  NUXT_FCM_PRIVATE_KEY: cred.private_key,
  ...FIREBASE_WEB,
}

function putSecret(name, value) {
  return new Promise((resolve) => {
    const child = spawn(
      process.execPath,
      [WRANGLER_JS, 'pages', 'secret', 'put', name, '--project-name', PROJECT],
      { stdio: ['pipe', 'pipe', 'pipe'] },
    )
    let out = ''
    child.stdout.on('data', d => (out += d.toString()))
    child.stderr.on('data', d => (out += d.toString()))
    child.on('close', (code) => {
      const ok = code === 0
      console.log(`${ok ? 'OK  ' : 'FAIL'} ${name}`)
      if (!ok) console.log('     ' + out.replace(/\s+/g, ' ').slice(0, 220))
      resolve(ok)
    })
    child.stdin.write(value)
    child.stdin.end()
  })
}

let fail = 0
for (const [name, value] of Object.entries(secrets)) {
  if (!value) { console.log(`SKIP ${name} (rong)`); continue }
  const ok = await putSecret(name, value)
  if (!ok) fail++
}

console.log(fail === 0 ? '\nXong: tat ca secret da duoc set' : `\nCo ${fail} secret that bai`)
process.exit(fail === 0 ? 0 : 1)
