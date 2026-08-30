const BASE = process.argv[2] || 'https://sennote.pages.dev'
let cookie = ''

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
  const sc = res.headers.get('set-cookie')
  if (sc) { cookie = sc.split(';')[0]; console.log(`   [set-cookie] ${cookie.slice(0, 45)}…`) }
  const text = await res.text()
  let json; try { json = JSON.parse(text) } catch {}
  return { status: res.status, json, text, raw: sc }
}

const email = `dbg${Math.floor(Math.random() * 1e6)}@test.local`
const pw = 'Test12345678'
console.log('BASE', BASE, '\nemail', email)

console.log('\n[A] register')
let r = await call('POST', '/api/auth/register', { email, password: pw })
console.log('   ->', r.status, JSON.stringify(r.json))

console.log('\n[B] login NGAY (chua logout)')
r = await call('POST', '/api/auth/login', { email, password: pw })
console.log('   ->', r.status, JSON.stringify(r.json))

console.log('\n[C] register lai cung email (phai 409 neu user ton tai)')
r = await call('POST', '/api/auth/register', { email, password: pw })
console.log('   ->', r.status, JSON.stringify(r.json))

console.log('\n[D] logout')
r = await call('POST', '/api/auth/logout')
console.log('   ->', r.status, JSON.stringify(r.json))
console.log('   cookie sau logout:', JSON.stringify(cookie))

console.log('\n[E] login SAU logout (cai bi 401 o e2e)')
r = await call('POST', '/api/auth/login', { email, password: pw })
console.log('   ->', r.status, JSON.stringify(r.json))

console.log('\n[F] login lai voi cookie da xoa sach')
cookie = ''
r = await call('POST', '/api/auth/login', { email, password: pw })
console.log('   ->', r.status, JSON.stringify(r.json))
