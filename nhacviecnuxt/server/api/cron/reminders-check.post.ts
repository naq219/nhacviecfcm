import { runReminderCheck } from '../../utils/run-reminder-check'

/**
 * Endpoint để Worker cron gọi vào mỗi phút.
 *
 * ⚠️ Lý do tồn tại: **Cloudflare Pages không hỗ trợ Cron Triggers**,
 * nên Nitro task `reminders:check` không bao giờ tự chạy trên Pages.
 * Worker `sennote-cron` (có cron) sẽ gọi endpoint này.
 *
 * Bảo vệ bằng secret trong header `x-cron-secret`:
 *  - Không dùng session vì cron không có cookie.
 *  - Phải khai báo `NUXT_CRON_SECRET` (Pages) và `CRON_SECRET` (Worker) GIỐNG NHAU.
 */
export default defineEventHandler(async (event) => {
  const cfg = useRuntimeConfig()
  const expected = cfg.cronSecret

  if (!expected) {
    throw createError({
      statusCode: 503,
      message: 'Chưa cấu hình NUXT_CRON_SECRET',
    })
  }

  const provided = getHeader(event, 'x-cron-secret')
  if (provided !== expected) {
    throw createError({ statusCode: 401, message: 'Sai cron secret' })
  }

  return runReminderCheck()
})
