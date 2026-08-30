/**
 * Smoke test end-to-end cho bản đã deploy.
 *
 *   node scripts/e2e-smoke.mjs https://sennote.pages.dev
 *
 * Kiểm tra: đăng ký → session → tạo one_time → tạo recurring → liệt kê
 *           → gọi API khi chưa đăng nhập (401) → validate lỗi (400)
 *           → đăng nhập lại → dọn dẹp.
 */
const BASE = process.argv[2] || 'https://sennote.pages.dev'

let cookie = ''
let pass = 0
let fail = 0

async function call(method, path, body) {
  const res = await fetch(`${BASE}${path}`, {
    method,
    headers: {
      ...(body ? { 'Content-Type': 'application/json' } : {}),
      ...(cookie ? { Cookie: cookie } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
    redirect: 'manual',
  })

  const setCookie = res.headers.get('set-cookie')
  if (setCookie) cookie = setCookie.split(';')[0]

  const text = await res.text()
  let json
  try { json = JSON.parse(text) } catch { /* không phải JSON */ }

  return { status: res.status, json, text }
}

function check(name, actual, expected) {
  const ok = actual === expected
  console.log(`  ${ok ? 'PASS' : 'FAIL'}  ${name} (${actual}, mong đợi ${expected})`)
  ok ? pass++ : fail++
}

function truthy(name, cond) {
  console.log(`  ${cond ? 'PASS' : 'FAIL'}  ${name}`)
  cond ? pass++ : fail++
}

const email = `smoke${Math.floor(Math.random() * 1e6)}@test.local`
const password = 'Test12345678'
const when = new Date(Date.now() + 60_000).toISOString()

console.log(`BASE = ${BASE}`)
console.log(`email = ${email}\n`)

console.log('[1] Đăng ký')
const r1 = await call('POST', '/api/auth/register', { email, password })
check('register 200', r1.status, 200)
truthy('trả về user.email đúng', r1.json?.user?.email === email)
truthy('có set session cookie', cookie.startsWith('nuxt-session='))

console.log('\n[2] Session hoạt động')
const r2 = await call('GET', '/api/auth/me')
check('me 200', r2.status, 200)
truthy('me trả đúng email', r2.json?.user?.email === email)

console.log('\n[3] Tạo reminder one_time')
const r3 = await call('POST', '/api/reminders', {
  title: 'Smoke one_time',
  type: 'one_time',
  next_action_at: when,
})
check('tạo one_time 200', r3.status, 200)
truthy('origin_time được tự suy', Boolean(r3.json?.origin_time))
truthy('next_action_at đúng', r3.json?.next_action_at === when)
truthy('one_time không có next_recurring', r3.json?.next_recurring === null)

console.log('\n[4] Tạo reminder recurring (daily)')
const r4 = await call('POST', '/api/reminders', {
  title: 'Smoke daily',
  type: 'recurring',
  next_action_at: when,
  recurrence_pattern: { type: 'daily', interval: 1 },
})
check('tạo recurring 200', r4.status, 200)
truthy('recurring có next_recurring', Boolean(r4.json?.next_recurring))

console.log('\n[5] Liệt kê reminder')
const r5 = await call('GET', '/api/reminders')
check('list 200', r5.status, 200)
truthy('có ít nhất 2 reminder', Array.isArray(r5.json) && r5.json.length >= 2)

console.log('\n[6] Validate — thiếu title phải 400')
const r6 = await call('POST', '/api/reminders', { type: 'one_time' })
check('thiếu title 400', r6.status, 400)

console.log('\n[7] Validate — max_crp>0 mà thiếu crp_interval_sec phải 400')
const r7 = await call('POST', '/api/reminders', {
  title: 'x', type: 'one_time', next_action_at: when, max_crp: 3,
})
check('max_crp thiếu interval 400', r7.status, 400)

console.log('\n[8] Chống IDOR — reminder của user khác phải 404')
const r8 = await call('GET', '/api/reminders/00000000-0000-0000-0000-000000000000')
check('reminder người khác 404', r8.status, 404)

console.log('\n[9] Đăng xuất')
const r9 = await call('POST', '/api/auth/logout')
check('logout 200', r9.status, 200)

console.log('\n[10] Sau logout — API phải 401')
const r10 = await call('GET', '/api/reminders')
check('chưa đăng nhập 401', r10.status, 401)

console.log('\n[11] Đăng nhập lại + dọn dẹp')
const r11 = await call('POST', '/api/auth/login', { email, password })
check('login 200', r11.status, 200)
for (const r of (await call('GET', '/api/reminders')).json ?? []) {
  await call('DELETE', `/api/reminders/${r.id}`)
}
const r12 = await call('GET', '/api/reminders')
truthy('đã xoá hết reminder', Array.isArray(r12.json) && r12.json.length === 0)

console.log(`\n===== ${pass} PASS / ${fail} FAIL =====`)
process.exit(fail === 0 ? 0 : 1)
