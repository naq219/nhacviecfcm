import { createClient } from '@libsql/client'

const db = createClient({
  url: process.env.NUXT_TURSO_DATABASE_URL,
  authToken: process.env.NUXT_TURSO_AUTH_TOKEN,
})

const rs = await db.execute(`
  SELECT
    (SELECT COUNT(*) FROM users)     AS users,
    (SELECT COUNT(*) FROM reminders) AS reminders,
    (SELECT COUNT(*) FROM devices)   AS devices,
    (SELECT worker_enabled FROM system_status WHERE mid = 1) AS worker_enabled
`)
console.log('thong ke:', JSON.stringify(rs.rows[0]))

// Phải xoá reminder trước (có FK trỏ sang users)
const ids = await db.execute(
  `SELECT id FROM users
    WHERE email LIKE 'smoke%'
       OR email LIKE 'dbg%'
       OR email LIKE 'fcmcheck%'
       OR email LIKE '%@test.local'`,
)
const idList = ids.rows.map(r => String(r.id))

if (idList.length === 0) {
  console.log('khong co user test nao')
}
else {
  // SQLite không có IN (?) với mảng → xoá từng user
  for (const id of idList) {
    await db.execute('DELETE FROM reminders WHERE user_id = ?', [id])
    await db.execute('DELETE FROM devices WHERE user_id = ?', [id])
    await db.execute('DELETE FROM users WHERE id = ?', [id])
  }
  console.log('da xoa user test:', idList.length)
}

const after = await db.execute(`
  SELECT
    (SELECT COUNT(*) FROM users)     AS users,
    (SELECT COUNT(*) FROM reminders) AS reminders,
    (SELECT COUNT(*) FROM devices)   AS devices,
    (SELECT worker_enabled FROM system_status WHERE mid = 1) AS worker_enabled
`)
console.log('sau don dep:', JSON.stringify(after.rows[0]))
