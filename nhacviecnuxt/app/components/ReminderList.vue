<script setup lang="ts">
import type { Reminder } from '#shared/types'

defineProps<{
  reminders: Reminder[]
  loading: boolean
}>()

const emit = defineEmits<{
  remove: [id: string]
  snooze: [id: string]
  complete: [id: string]
}>()
</script>

<template>
  <section>
    <p v-if="loading">
      Đang tải...
    </p>
    <p v-else-if="!reminders.length">
      Chưa có reminder nào.
    </p>
    <ul v-else class="list">
      <ReminderCard
        v-for="r in reminders"
        :key="r.id"
        :reminder="r"
        @remove="emit('remove', $event)"
        @snooze="emit('snooze', $event)"
        @complete="emit('complete', $event)"
      />
    </ul>
  </section>
</template>

<style scoped>
.list { list-style: none; padding: 0; margin: 0; }
</style>
