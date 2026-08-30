import { createUser } from '../../utils/auth'

export default defineEventHandler(async (event) => {
  const { email, password } = await readBody(event) as { email?: string, password?: string }

  if (!email || !password || String(password).length < 8) {
    throw createError({ statusCode: 400, message: 'Email và mật khẩu (ít nhất 8 ký tự) là bắt buộc' })
  }

  const user = await createUser(String(email).trim().toLowerCase(), String(password))
  if (!user) {
    throw createError({ statusCode: 409, message: 'Email đã tồn tại' })
  }

  await setUserSession(event, { user })
  return { user }
})
