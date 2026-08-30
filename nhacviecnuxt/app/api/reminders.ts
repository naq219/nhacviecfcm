import type { Reminder, ReminderPayload } from '#shared/types'

// Mọi request đi qua lớp này — component/composable KHÔNG gọi $fetch trực tiếp.

export const apiListReminders = (status?: string) =>
  $fetch<Reminder[]>('/api/reminders', { query: status ? { status } : {} })

export const apiGetReminder = (id: string) => $fetch<Reminder>(`/api/reminders/${id}`)

export const apiCreateReminder = (body: ReminderPayload) =>
  $fetch<Reminder>('/api/reminders', { method: 'POST', body })

export const apiUpdateReminder = (id: string, body: Partial<ReminderPayload>) =>
  $fetch<Reminder>(`/api/reminders/${id}`, { method: 'PUT', body })

export const apiDeleteReminder = (id: string) =>
  $fetch<{ ok: boolean }>(`/api/reminders/${id}`, { method: 'DELETE' })

export const apiSnoozeReminder = (id: string, duration: number) =>
  $fetch<Reminder>(`/api/reminders/${id}/snooze`, { method: 'POST', body: { duration } })

export const apiCompleteReminder = (id: string) =>
  $fetch<Reminder>(`/api/reminders/${id}/complete`, { method: 'POST' })
