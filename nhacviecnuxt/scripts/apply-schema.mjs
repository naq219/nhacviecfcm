/**
 * Apply db/schema.sql vào Turso.
 * Dùng thay cho Turso CLI (không cài được trên Windows nếu không có WSL).
 *
 * Chạy:  node scripts/apply-schema.mjs
 * (đọc NUXT_TURSO_DATABASE_URL / NUXT_TURSO_AUTH_TOKEN từ .env, nạp bằng --env-file)
 */
import { readFileSync } from 'node:fs'
import { createClient } from '@libsql/client'

const url = process.env.NUXT_TURSO_DATABASE_URL
const authToken = process.env.NUXT_TURSO_AUTH_TOKEN

if (!url || !authToken) {
  console.error('Thieu NUXT_TURSO_DATABASE_URL hoac NUXT_TURSO_AUTH_TOKEN')
  process.exit(1)
}

const db = createClient({ url, authToken })
const sql = readFileSync(new URL('../db/schema.sql', import.meta.url), 'utf8')

// executeMultiple chạy được nhiều câu lệnh trong 1 lần gọi
await db.executeMultiple(sql)

const tables = await db.execute(
  "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name",
)
console.log('Cac bang hien co:', tables.rows.map(r => r.name).join(', '))

const idx = await db.execute(
  "SELECT name FROM sqlite_master WHERE type='index' AND name LIKE 'idx_%' ORDER BY name",
)
console.log('Cac index:', idx.rows.map(r => r.name).join(', '))

const st = await db.execute('SELECT mid, worker_enabled FROM system_status WHERE mid = 1')
console.log('system_status:', JSON.stringify(st.rows[0] ?? null))

console.log('OK')
