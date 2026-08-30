<script setup lang="ts">
import type { RecurrenceType, ReminderPayload } from '#shared/types'

const emit = defineEmits<{ create: [payload: ReminderPayload] }>()

const title = ref('')
const description = ref('')
const type = ref<'one_time' | 'recurring'>('one_time')
const calendarType = ref<'solar' | 'lunar'>('solar')
const recurrenceType = ref<RecurrenceType>('daily')
const interval = ref(1)
const intervalSeconds = ref(3600)
const repeatStrategy = ref<'none' | 'crp_until_complete'>('none')
const maxCrp = ref(0)
const crpIntervalSec = ref(300)
const nextActionAt = ref('')

// mặc định: 1 giờ tới, hiển thị theo giờ VN
const defaultWhen = () => {
  const d = new Date(Date.now() + 3600_000 + 7 * 3600_000)
  return d.toISOString().slice(0, 16)
}
onMounted(() => { if (!nextActionAt.value) nextActionAt.value = defaultWhen() })

function submit() {
  if (!title.value.trim()) return

  const payload: ReminderPayload = {
    title: title.value.trim(),
    description: description.value.trim() || null,
    type: type.value,
    calendar_type: calendarType.value,
    repeat_strategy: repeatStrategy.value,
    next_action_at: fromLocalInput(nextActionAt.value) || null,
    max_crp: maxCrp.value,
    crp_interval_sec: maxCrp.value > 0 ? crpIntervalSec.value : 0,
  }

  if (type.value === 'recurring') {
    payload.recurrence_pattern = recurrenceType.value === 'interval_seconds'
      ? { type: 'interval_seconds', interval_seconds: intervalSeconds.value }
      : { type: recurrenceType.value, interval: interval.value }
  }

  emit('create', payload)

  title.value = ''
  description.value = ''
}
</script>

<template>
  <form class="form" @submit.prevent="submit">
    <div class="row">
      <input v-model="title" placeholder="Tiêu đề nhắc nhở" required>
      <select v-model="type">
        <option value="one_time">
          Một lần
        </option>
        <option value="recurring">
          Lặp lại
        </option>
      </select>
    </div>

    <div class="row">
      <input v-model="description" placeholder="Mô tả (tuỳ chọn)">
      <label>Thời gian <input v-model="nextActionAt" type="datetime-local" required></label>
    </div>

    <template v-if="type === 'recurring'">
      <div class="row">
        <select v-model="recurrenceType">
          <option value="daily">
            Hàng ngày
          </option>
          <option value="weekly">
            Hàng tuần
          </option>
          <option value="monthly">
            Hàng tháng
          </option>
          <option value="interval_seconds">
            Mỗi N giây
          </option>
          <option value="solar_last_day_of_month">
            Cuối tháng dương
          </option>
          <option value="lunar_last_day_of_month">
            Cuối tháng âm
          </option>
        </select>

        <label v-if="recurrenceType === 'interval_seconds'">
          Số giây <input v-model.number="intervalSeconds" type="number" min="1">
        </label>
        <label v-else-if="recurrenceType !== 'solar_last_day_of_month' && recurrenceType !== 'lunar_last_day_of_month'">
          Chu kỳ <input v-model.number="interval" type="number" min="1">
        </label>

        <label>Lịch
          <select v-model="calendarType">
            <option value="solar">Dương</option>
            <option value="lunar">Âm</option>
          </select>
        </label>
      </div>

      <div class="row">
        <label>Chiến lược
          <select v-model="repeatStrategy">
            <option value="none">Tự sang chu kỳ mới</option>
            <option value="crp_until_complete">Nhắc đến khi hoàn thành</option>
          </select>
        </label>
      </div>
    </template>

    <div class="row">
      <label>Nhắc lại tối đa <input v-model.number="maxCrp" type="number" min="0"></label>
      <label v-if="maxCrp > 0">
        Mỗi (giây) <input v-model.number="crpIntervalSec" type="number" min="1">
      </label>
    </div>

    <button type="submit">
      Thêm reminder
    </button>
  </form>
</template>

<style scoped>
.form { display: flex; flex-direction: column; gap: 8px; margin: 12px 0 20px; }
.row { display: flex; gap: 8px; flex-wrap: wrap; align-items: center; }
input, select { padding: 6px 8px; }
label { display: inline-flex; gap: 4px; align-items: center; font-size: 13px; color: #555; }
button { align-self: flex-start; padding: 7px 14px; cursor: pointer; }
</style>
