import type { Reminder, ReminderPayload } from '#shared/types'
import {
  apiCompleteReminder,
  apiCreateReminder,
  apiDeleteReminder,
  apiListReminders,
  apiSnoozeReminder,
  apiUpdateReminder,
} from '~/api/reminders'

/** State + hành động cho danh sách reminder */
export function useReminders() {
  const reminders = useState<Reminder[]>('reminders', () => [])
  const loading = ref(false)
  const error = ref('')

  async function load(status?: string) {
    loading.value = true
    error.value = ''
    try {
      reminders.value = await apiListReminders(status)
    }
    catch (e: any) {
      error.value = e?.data?.message ?? 'Không tải được danh sách'
    }
    finally {
      loading.value = false
    }
  }

  async function create(data: ReminderPayload) {
    error.value = ''
    try {
      await apiCreateReminder(data)
      await load()
      return true
    }
    catch (e: any) {
      error.value = e?.data?.message ?? 'Tạo reminder thất bại'
      return false
    }
  }

  async function update(id: string, data: Partial<ReminderPayload>) {
    error.value = ''
    try {
      await apiUpdateReminder(id, data)
      await load()
      return true
    }
    catch (e: any) {
      error.value = e?.data?.message ?? 'Cập nhật thất bại'
      return false
    }
  }

  async function remove(id: string) {
    error.value = ''
    try {
      await apiDeleteReminder(id)
      await load()
      return true
    }
    catch (e: any) {
      error.value = e?.data?.message ?? 'Xoá thất bại'
      return false
    }
  }

  async function snooze(id: string, seconds: number) {
    error.value = ''
    try {
      await apiSnoozeReminder(id, seconds)
      await load()
    }
    catch (e: any) {
      error.value = e?.data?.message ?? 'Hoãn thất bại'
    }
  }

  async function complete(id: string) {
    error.value = ''
    try {
      await apiCompleteReminder(id)
      await load()
    }
    catch (e: any) {
      error.value = e?.data?.message ?? 'Hoàn thành thất bại'
    }
  }

  return { reminders, loading, error, load, create, update, remove, snooze, complete }
}
