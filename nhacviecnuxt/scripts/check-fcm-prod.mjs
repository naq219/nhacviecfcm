/** Kiểm tra FCM đã sống trên production chưa (cần đăng nhập) */
/**
 * Dùng:  node scripts/check-fcm-prod.mjs <base> [email] [password]
 * Nếu không truyền email/password → tự đăng ký tài khoản tạm rồi kiểm tra
 * (nhớ chạy scripts/cleanup-test-users.mjs để dọn sau).
 */
const BASE = process.argv[2] || 'https://sennote.pages.dev'
let email = process.argv[3]
let password = process.argv[4]

let cookie = ''

async function call(method, path, body) {
  const res = await fetch(`${BASE}${path}`, {
    method,
    headers: {
      ...(body ? { 'Content-Type': 'application/json' } : {}),
      ...(cookie ? { Cookie: cookie } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
  })
  const sc = res.headers.get('set-cookie')
  if (sc) cookie = sc.split(';')[0]
  const text = await res.text()
  let json
  try { json = JSON.parse(text) } catch { /* bo qua */ }
  return { status: res.status, json, text }
}

let login
if (email && password) {
  login = await call('POST', '/api/auth/login', { email, password })
}
else {
  email = `fcmcheck${Math.floor(Math.random() * 1e6)}@test.local`
  password = 'Test12345678'
  login = await call('POST', '/api/auth/register', { email, password })
  console.log('Da tao tai khoan tam:', email)
}

if (login.status !== 200) {
  console.log('Dang nhap that bai:', login.status, login.text?.slice(0, 200))
  process.exitCode = 1
}
else {
  console.log('Dang nhap OK:', email)
  await runCheck()
}

async function runCheck() {

  const st = await call('GET', '/api/system_status')
  console.log('\nGET /api/system_status ->', st.status)
  console.log(JSON.stringify(st.json, null, 2))

  const fcm = st.json?.fcm
  console.log('\n--- ket luan ---')
  if (!fcm) {
    console.log('Khong co thong tin fcm')
    process.exitCode = 1
    return
  }
  console.log('FCM da cau hinh      :', fcm.configured ? 'CO' : 'CHUA')
  console.log('Lay access token duoc:', fcm.accessToken ? 'CO' : 'KHONG')
  if (fcm.error) console.log('Loi                  :', fcm.error)
  process.exitCode = (fcm.configured && fcm.accessToken) ? 0 : 1
}
