<script setup lang="ts">
import { usePush } from '~/composables/usePush'

const props = defineProps<{ compact?: boolean }>()

const { state, error, check, enable, disable } = usePush()

onMounted(() => check())

const label = computed(() => {
  switch (state.value) {
    case 'unsupported': return 'Trình duyệt không hỗ trợ'
    case 'denied': return 'Đã chặn quyền thông báo'
    case 'ready': return 'Đang nhận thông báo'
    case 'loading': return 'Đang xử lý...'
    default: return 'Bật thông báo'
  }
})

const canEnable = computed(() => state.value === 'default' || state.value === 'loading')
const canDisable = computed(() => state.value === 'ready')
</script>

<template>
  <div class="push">
    <button
      v-if="canEnable"
      type="button"
      :disabled="state === 'loading' || state === 'unsupported' || state === 'denied'"
      @click="enable"
    >
      {{ label }}
    </button>
    <button v-else-if="canDisable" type="button" @click="disable">
      {{ label }} (bấm để tắt)
    </button>
    <span v-else class="muted">{{ label }}</span>

    <p v-if="error" class="err">
      {{ error }}
    </p>
    <p v-if="state === 'denied' && !props.compact" class="hint">
      Mở khoá: bấm icon khoá trên thanh địa chỉ → cho phép Thông báo → tải lại trang.
    </p>
  </div>
</template>

<style scoped>
.push { display: inline-flex; flex-direction: column; gap: 4px; align-items: flex-start; }
button { padding: 6px 12px; cursor: pointer; }
.muted { color: #888; font-size: 13px; }
.err { color: #c0392b; font-size: 12px; margin: 0; }
.hint { color: #777; font-size: 12px; margin: 0; max-width: 260px; }
</style>
