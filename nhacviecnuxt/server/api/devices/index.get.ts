/** Danh sách thiết bị đã đăng ký của user hiện tại */
export default defineEventHandler(async (event) => {
  const session = await requireUserSession(event)
  const rows = await listDevices(session.user.id)

  return rows.map(r => ({ ...r, token: maskToken(r.token) }) as Device)
})

function maskToken(token: string): string {
  return token.length <= 12 ? token : `${token.slice(0, 8)}…${token.slice(-4)}`
}
