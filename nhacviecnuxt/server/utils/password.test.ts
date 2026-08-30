import { describe, expect, it } from 'vitest'
import { makePasswordHash, verifyPasswordHash } from './password'

/**
 * ⚠️ Module này thay thế hashPassword/verifyPassword của nuxt-auth-utils
 * vì scrypt của node:crypto không verify được trên Cloudflare Workers.
 */

describe('password (Web Crypto PBKDF2)', () => {
  it('tạo hash đúng định dạng pbkdf2-sha256$iter$salt$hash', async () => {
    const h = await makePasswordHash('Test12345678')
    const parts = h.split('$')
    expect(parts).toHaveLength(4)
    expect(parts[0]).toBe('pbkdf2-sha256')
    expect(Number(parts[1])).toBeGreaterThan(0)
  })

  it('verify đúng với mật khẩu đúng', async () => {
    const h = await makePasswordHash('Test12345678')
    expect(await verifyPasswordHash(h, 'Test12345678')).toBe(true)
  })

  it('verify sai với mật khẩu sai', async () => {
    const h = await makePasswordHash('Test12345678')
    expect(await verifyPasswordHash(h, 'SaiMatKhau')).toBe(false)
  })

  it('2 lần hash cùng mật khẩu cho ra hash KHÁC nhau (salt ngẫu nhiên)', async () => {
    const a = await makePasswordHash('Test12345678')
    const b = await makePasswordHash('Test12345678')
    expect(a).not.toBe(b)
    // nhưng cả hai đều verify được
    expect(await verifyPasswordHash(a, 'Test12345678')).toBe(true)
    expect(await verifyPasswordHash(b, 'Test12345678')).toBe(true)
  })

  it('hash cũ của nuxt-auth-utils ($scrypt$) bị từ chối (không ném lỗi)', async () => {
    const scryptHash = '$scrypt$n=16384,r=8,p=1$DC3c9SEJbiHAeAYEqRvIvg$TtTgv2I+F0MS8'
    expect(await verifyPasswordHash(scryptHash, 'Test12345678')).toBe(false)
  })

  it('hash hỏng / rỗng trả false thay vì ném lỗi', async () => {
    expect(await verifyPasswordHash('', 'x')).toBe(false)
    expect(await verifyPasswordHash('bat ky chuoi rac', 'x')).toBe(false)
    expect(await verifyPasswordHash('pbkdf2-sha256$abc$x$y', 'x')).toBe(false)
  })

  it('mật khẩu có dấu tiếng Việt + unicode vẫn hoạt động', async () => {
    const pw = 'MậtKhẩuViệt123@#'
    const h = await makePasswordHash(pw)
    expect(await verifyPasswordHash(h, pw)).toBe(true)
    expect(await verifyPasswordHash(h, 'MậtKhẩuViệt123@')).toBe(false)
  })
})
