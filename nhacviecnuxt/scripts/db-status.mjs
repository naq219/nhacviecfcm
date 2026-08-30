/**
 * Xem nhanh trạng thái hệ thống trên production.
 *
 *   node --env-file=.env scripts/db-status.mjs
 *
 * Hiển thị: thiết bị đã đăng ký, reminder đang chạy, kill-switch worker.
 */
import { createClient } from '@libsql/client'

const db = createClient({
  url: process.env.NUXT_TURSO_DATABASE_URL,
  authToken: process.env.NUXT_TURSO_AUTH_TOKEN,
})

const pad = (v, n) => String(v ?? '').padEnd(n)

const devices = await db.execute(`
  SELECT u.email, d.platform, d.is_active, d.created_at
  FROM devices d JOIN users u ON u.id = d.user_id
  ORDER BY d.created_at
`)
console.log('=== THIET BI (devices) ===')
if (devices.rows.length === 0) console.log('  (chua co thiet bi nao)')
for (const r of devices.rows) {
  console.log(`  ${pad(r.email, 22)}| ${pad(r.platform, 8)}| active=${r.is_active} | ${r.created_at}`)
}

const reminders = await db.execute(`
  SELECT u.email, r.title, r.status, r.type, r.is_sended_one_time, r.last_sent_at, r.next_action_at
  FROM reminders r JOIN users u ON u.id = r.user_id
  ORDER BY r.created_at DESC
`)
console.log('\n=== REMINDERS ===')
for (const r of reminders.rows) {
  console.log(
    `  ${pad(r.title, 10)}| ${pad(r.email, 22)}| ${pad(r.status, 10)}| ${pad(r.type, 10)}`
    + `| sent=${r.is_sended_one_time} | next=${r.next_action_at ?? '-'} | last=${r.last_sent_at ?? '-'}`,
  )
}

const st = await db.execute('SELECT worker_enabled, last_error FROM system_status WHERE mid = 1')
console.log('\n=== WORKER (kill-switch) ===')
console.log(`  worker_enabled = ${st.rows[0]?.worker_enabled}   last_error = ${st.rows[0]?.last_error || '(rong)'}`)
