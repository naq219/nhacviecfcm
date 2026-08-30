<script setup lang="ts">
import { usePush } from '~/composables/usePush'

const { state, error, busy, supported, check, enable, disable } = usePush()

onMounted(() => check())

/**
 * ⚠️ Quan trọng: KHÔNG ẩn nút khi `state === 'granted'`.
 * Trình duyệt có thể đã cấp quyền từ trước nhưng thiết bị chưa được đăng ký
 * cho tài khoản đang đăng nhập (mỗi tài khoản có thiết bị riêng).
 * Khi đó vẫn phải cho user bấm để đăng ký.
 */
const label = computed(() => {
  switch (state.value) {
    case 'unsupported': return 'Trình duyệt không hỗ trợ'
    case 'denied': return 'Đã chặn quyền thông báo'
    case 'loading': return 'Đang xử lý…'
    case 'ready': return 'Đang nhận thông báo'
    case 'granted': return 'Bật thông báo cho tài khoản này'
    default: return 'Bật thông báo'
  }
})

const isReady = computed(() => state.value === 'ready')
const blocked = computed(() => state.value === 'unsupported' || state.value === 'denied')
</script>

<template>
  <div class="push">
    <button
      v-if="!isReady && !blocked"
      type="button"
      :disabled="busy"
      @click="enable"
    >
      {{ busy ? 'Đang xử lý…' : label }}
    </button>

    <template v-else-if="isReady">
      <button type="button" class="on" :disabled="busy" @click="disable">
        {{ label }} — bấm để tắt
      </button>
    </template>

    <span v-else class="muted">{{ label }}</span>

    <p v-if="error" class="err">
      {{ error }}
    </p>

    <p v-if="state === 'denied'" class="hint">
      Mở khoá: bấm icon khoá trên thanh địa chỉ → cho phép Thông báo → tải lại trang.
    </p>
    <p v-else-if="state === 'granted' && !supported" class="hint">
      Trình duyệt chưa cấp quyền ở mức Service Worker.
    </p>
  </div>
</template>

<style scoped>
.push { display: inline-flex; flex-direction: column; gap: 4px; align-items: flex-start; }
button { padding: 6px 12px; cursor: pointer; }
button.on { border-color: #2e9e5b; }
.muted { color: #888; font-size: 13px; }
.err { color: #c0392b; font-size: 12px; margin: 0; }
.hint { color: #777; font-size: 12px; margin: 0; max-width: 280px; }
</style>
