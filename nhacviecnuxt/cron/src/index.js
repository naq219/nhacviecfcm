/**
 * Worker cron cho SenNote.
 *
 * ⚠️ Tại sao cần Worker này?
 * Cloudflare **Pages không hỗ trợ Cron Triggers** — handler `scheduled()` của Nitro
 * vẫn được sinh ra trong bundle nhưng không bao giờ được gọi. Vì vậy lịch nhắc
 * không tự chạy. Worker này có cron thật (1 phút) và gọi về endpoint HTTP của Pages.
 *
 * Cấu hình: xem wrangler.toml cạnh file này.
 * Secret CRON_SECRET phải trùng với NUXT_CRON_SECRET trên Pages.
 */
export default {
  async scheduled(_controller, env, ctx) {
    const url = env.TARGET_URL
    const secret = env.CRON_SECRET

    if (!url || !secret) {
      console.error('Thieu TARGET_URL hoac CRON_SECRET')
      return
    }

    try {
      const res = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'x-cron-secret': secret,
        },
        // cron chạy mỗi phút → không để request treo quá 25 giây
        signal: AbortSignal.timeout(25_000),
      })
      const text = await res.text()
      console.log(`[cron] ${res.status} ${text.slice(0, 300)}`)
    }
    catch (err) {
      console.error('[cron] loi:', String(err))
    }
  },
}
