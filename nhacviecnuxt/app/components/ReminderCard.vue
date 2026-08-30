<script setup lang="ts">
import type { Reminder } from '#shared/types'

const props = defineProps<{ reminder: Reminder }>()

const emit = defineEmits<{
  remove: [id: string]
  snooze: [id: string]
  complete: [id: string]
}>()

const thu = describeRecurrence(props.reminder.recurrence_pattern)
</script>

<template>
  <li class="card">
    <div class="card-main">
      <div class="card-title">
        <strong>{{ props.reminder.title }}</strong>
        <span class="badge" :class="`badge-${props.reminder.status}`">{{ props.reminder.status }}</span>
        <span v-if="props.reminder.tag" class="badge badge-tag">{{ props.reminder.tag }}</span>
      </div>

      <p v-if="props.reminder.description" class="card-desc">
        {{ props.reminder.description }}
      </p>

      <dl class="card-meta">
        <div>
          <dt>Loại</dt>
          <dd>{{ props.reminder.type === 'recurring' ? `Lặp — ${thu}` : 'Một lần' }}</dd>
        </div>
        <div>
          <dt>Lịch</dt>
          <dd>{{ props.reminder.calendar_type === 'lunar' ? 'Âm' : 'Dương' }}</dd>
        </div>
        <div v-if="props.reminder.repeat_strategy === 'crp_until_complete'">
          <dt>Chiến lược</dt>
          <dd>Nhắc đến khi hoàn thành</dd>
        </div>
        <div v-if="props.reminder.max_crp > 0">
          <dt>Nhắc lại</dt>
          <dd>{{ props.reminder.crp_count }}/{{ props.reminder.max_crp }} lần</dd>
        </div>
        <div>
          <dt>Lần tới</dt>
          <dd>{{ formatVN(props.reminder.next_action_at) }} ({{ fromNow(props.reminder.next_action_at) }})</dd>
        </div>
        <div v-if="props.reminder.last_sent_at">
          <dt>Đã gửi</dt>
          <dd>{{ formatVN(props.reminder.last_sent_at) }}</dd>
        </div>
      </dl>
    </div>

    <div v-if="props.reminder.status === 'active'" class="card-actions">
      <button type="button" @click="emit('snooze', props.reminder.id)">
        Hoãn 10p
      </button>
      <button type="button" @click="emit('complete', props.reminder.id)">
        Hoàn thành
      </button>
      <button type="button" class="danger" @click="emit('remove', props.reminder.id)">
        Xoá
      </button>
    </div>
  </li>
</template>

<style scoped>
.card {
  border: 1px solid #e2e2e2;
  border-radius: 10px;
  padding: 12px 14px;
  margin-bottom: 10px;
  display: flex;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}
.card-main { flex: 1 1 260px; min-width: 0; }
.card-title { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.card-desc { margin: 6px 0 0; color: #555; }
.card-meta { display: flex; flex-wrap: wrap; gap: 4px 20px; margin: 10px 0 0; font-size: 13px; }
.card-meta dt { color: #888; font-size: 11px; text-transform: uppercase; }
.card-meta dd { margin: 0; }
.card-actions { display: flex; flex-direction: column; gap: 6px; }
.card-actions button { white-space: nowrap; padding: 5px 10px; cursor: pointer; }
.danger { color: #c0392b; }
.badge { font-size: 11px; padding: 2px 7px; border-radius: 999px; background: #eee; }
.badge-active { background: #d6f5e3; }
.badge-completed { background: #e6e6e6; }
.badge-paused { background: #fdecd2; }
.badge-tag { background: #e3ecff; }
</style>
