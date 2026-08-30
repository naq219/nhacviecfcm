export default defineNuxtRouteMiddleware(async () => {
  const { loggedIn, ready, fetch } = useUserSession()

  // Chờ session nạp xong trước khi check — tránh bị đá về /login nhầm
  // (lúc SSR/mới vào trang, session chưa chắc đã sẵn sàng).
  if (!ready.value) await fetch()
  if (!loggedIn.value) return navigateTo('/login')
})
