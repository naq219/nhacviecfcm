import { runReminderCheck } from '../../utils/run-reminder-check'

export default defineTask({
  meta: {
    name: 'reminders:check',
    description: 'Quét reminder đến hạn và gửi FCM',
  },
  async run() {
    // ⚠️ Cloudflare Pages KHÔNG gọi task này (không hỗ trợ Cron Triggers).
    // Trên production, lịch được kích qua POST /api/cron/reminders-check
    // do Worker "sennote-cron" gọi mỗi phút. Xem docs/05 mục 4.
    return runReminderCheck()
  },
})
