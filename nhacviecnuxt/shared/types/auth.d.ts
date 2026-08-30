// Khai báo type cho session user của nuxt-auth-utils
// để `session.user.id` / `session.user.email` không báo lỗi TypeScript.
declare module '#auth-utils' {
  interface User {
    id: string
    email: string
  }

  interface UserSession {
    user: User
  }
}

export {}
