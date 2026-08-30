<script setup lang="ts">
import type { ReminderPayload } from '#shared/types'
import { useReminders } from '~/composables/useReminders'

definePageMeta({ middleware: 'auth' })

const { user, clear } = useUserSession()
const { reminders, loading, error, load, create, remove, snooze, complete } = useReminders()

const filter = ref('')

onMounted(() => load())

async function onCreate(payload: ReminderPayload) {
  await create(payload)
}

async function onChange(value: string) {
  await load(value || undefined)
}
</script>

<template>
  <main class="wrap">
    <header class="head">
      <h1>Reminder của {{ user?.email }}</h1>
      <div class="head-actions">
        <PushToggle />
        <button @click="clear()">
          Đăng xuất
        </button>
      </div>
    </header>

    <ReminderForm @create="onCreate" />

    <div class="row">
      <label>Lọc theo trạng thái
        <select v-model="filter" @change="onChange(filter)">
          <option value="">
            Tất cả
          </option>
          <option value="active">Đang chạy</option>
          <option value="paused">Tạm dừng</option>
          <option value="completed">Hoàn thành</option>
        </select>
      </label>
    </div>

    <p v-if="error" class="err">
      {{ error }}
    </p>

    <ReminderList
      :reminders="reminders"
      :loading="loading"
      @remove="remove"
      @snooze="id => snooze(id, 600)"
      @complete="complete"
    />
  </main>
</template>

<style scoped>
.wrap { max-width: 720px; margin: 2rem auto; padding: 0 12px; font-family: system-ui, sans-serif; }
.head { display: flex; justify-content: space-between; align-items: center; gap: 12px; flex-wrap: wrap; }
.head h1 { font-size: 20px; }
.head-actions { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.row { margin: 8px 0; }
label { display: inline-flex; gap: 6px; align-items: center; font-size: 13px; color: #555; }
.err { color: #c0392b; }
button { padding: 6px 12px; cursor: pointer; }
</style>
