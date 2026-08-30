// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: ['nuxt-auth-utils'],

  app: {
    head: {
      title: 'SenNote',
      htmlAttrs: { lang: 'vi' },
    },
  },

  runtimeConfig: {
    // Chỉ đọc được ở SERVER (server/utils, server/api, server/tasks)
    // Env phải có prefix NUXT_: tursoDatabaseUrl <- NUXT_TURSO_DATABASE_URL
    tursoDatabaseUrl: '',
    tursoAuthToken: '',
    fcmClientEmail: '',
    fcmPrivateKey: '',
    fcmProjectId: '',
    /** Secret để Worker cron gọi POST /api/cron/reminders-check */
    cronSecret: '',

    public: {
      // ⚠️ public = GỬI XUỐNG TRÌNH DUYỆT. Chỉ để các giá trị Firebase
      // vốn đã an toàn khi công khai (apiKey/config của web app).
      // KHÔNG bao giờ bỏ private key hay service account vào đây.
      firebaseApiKey: '',
      firebaseAuthDomain: '',
      firebaseProjectId: '',
      firebaseMessagingSenderId: '',
      firebaseAppId: '',
      // VAPID key dùng cho Web Push (Firebase Console → Cloud Messaging → Web push certificates)
      fcmVapidKey: '',
    },
  },

  nitro: {
    // ⚠️ CHỈ đặt preset khi build, KHÔNG đặt khi chạy dev.
    // Lý do: preset cloudflare-pages khiến Nitro dev dùng "cloudflare-dev emulation",
    // cần wrangler và KHÔNG mở endpoint /_nitro/tasks/... → test cron tay bị 404.
    // Khi build (NODE_ENV=production) vẫn ra đúng dist/ chuẩn Cloudflare Pages.
    ...(process.env.NODE_ENV === 'production' ? { preset: 'cloudflare-pages' } : {}),

    experimental: {
      tasks: true, // bắt buộc để có scheduled tasks (cron)
    },
    // Cron: mỗi phút chạy task "reminders:check".
    // PHẢI nằm TRONG khối nitro: {} — để top-level sẽ bị bỏ qua.
    scheduledTasks: {
      '* * * * *': 'reminders:check',
    },
  },
})
