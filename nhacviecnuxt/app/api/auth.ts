export interface SessionUser {
  id: string
  email: string
}

export const apiRegister = (email: string, password: string) =>
  $fetch<{ user: SessionUser }>('/api/auth/register', { method: 'POST', body: { email, password } })

export const apiLogin = (email: string, password: string) =>
  $fetch<{ user: SessionUser }>('/api/auth/login', { method: 'POST', body: { email, password } })

export const apiLogout = () => $fetch<{ ok: boolean }>('/api/auth/logout', { method: 'POST' })
