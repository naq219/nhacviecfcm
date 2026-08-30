import { verifyUser } from '../../utils/auth'

export default defineEventHandler(async (event) => {
  const { email, password } = await readBody(event) as { email?: string, password?: string }

  const user = await verifyUser(String(email ?? '').trim().toLowerCase(), String(password ?? ''))
  if (!user) {
    throw createError({ statusCode: 401, message: 'Sai email hoặc mật khẩu' })
  }

  await setUserSession(event, { user })
  return { user }
})
