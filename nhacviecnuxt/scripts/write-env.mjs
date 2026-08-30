/**
 * Sinh file .env từ firebase-credentials.json + cấu hình Firebase web.
 * Private key được đọc trực tiếp từ file → không hardcode, không lộ ra log.
 *
 *   node scripts/write-env.mjs
 */
import { readFileSync, existsSync, writeFileSync } from 'node:fs'

const CRED = 'E:/PROJECT/nhacviecfcm/firebase-credentials.json'
const OUT = new URL('../.env', import.meta.url)
const SESSION_PW = process.env.SESSION_PASSWORD || ''

if (!existsSync(CRED)) { console.error('Thieu', CRED); process.exit(1) }
const cred = JSON.parse(readFileSync(CRED, 'utf8'))

const lines = [
  `NUXT_SESSION_PASSWORD=${SESSION_PW}`,
  `NUXT_TURSO_DATABASE_URL=${process.env.TURSO_URL || ''}`,
  `NUXT_TURSO_AUTH_TOKEN=${process.env.TURSO_TOKEN || ''}`,
  '',
  '# FCM server (doc tu firebase-credentials.json)',
  `NUXT_FCM_CLIENT_EMAIL=${cred.client_email}`,
  `NUXT_FCM_PRIVATE_KEY=${JSON.stringify(cred.private_key)}`,
  `NUXT_FCM_PROJECT_ID=${cred.project_id}`,
  '',
  '# Firebase web (public)',
  'NUXT_PUBLIC_FIREBASE_API_KEY=AIzaSyAxYNmylFjzfAIIK4fm94en5svpL9YgFZ8',
  'NUXT_PUBLIC_FIREBASE_AUTH_DOMAIN=reminaq-001.firebaseapp.com',
  'NUXT_PUBLIC_FIREBASE_PROJECT_ID=reminaq-001',
  'NUXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID=679219217368',
  'NUXT_PUBLIC_FIREBASE_APP_ID=1:679219217368:web:f90e2264ab388f3843e8ad',
  `NUXT_PUBLIC_FCM_VAPID_KEY=${process.env.VAPID_KEY || ''}`,
  '',
]

writeFileSync(OUT, lines.join('\n'), 'utf8')
console.log('Da ghi .env')
console.log('  VAPID:', process.env.VAPID_KEY ? 'co' : 'CHUA CO (web push se khong hoat dong)')
