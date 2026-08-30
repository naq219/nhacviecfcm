<script setup lang="ts">
import { apiLogin, apiRegister } from '~/api/auth'

const { fetch: fetchSession } = useUserSession()

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit(action: 'login' | 'register') {
  error.value = ''
  loading.value = true
  try {
    if (action === 'login') await apiLogin(email.value, password.value)
    else await apiRegister(email.value, password.value)
    await fetchSession()
    await navigateTo('/')
  }
  catch (e: any) {
    error.value = e?.data?.message ?? 'Có lỗi xảy ra'
  }
  finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="wrap">
    <h1>SenNote</h1>

    <form @submit.prevent="submit('login')">
      <input v-model="email" type="email" placeholder="Email" required autocomplete="email">
      <input
        v-model="password"
        type="password"
        placeholder="Mật khẩu (ít nhất 8 ký tự)"
        required
        minlength="8"
        autocomplete="current-password"
      >
      <button type="submit" :disabled="loading">
        Đăng nhập
      </button>
      <button type="button" :disabled="loading" @click="submit('register')">
        Đăng ký
      </button>
    </form>

    <p v-if="error" class="err">
      {{ error }}
    </p>
  </main>
</template>

<style scoped>
.wrap { max-width: 320px; margin: 4rem auto; font-family: system-ui, sans-serif; }
form { display: flex; flex-direction: column; gap: 8px; }
input { padding: 8px; }
button { padding: 8px; cursor: pointer; }
.err { color: #c0392b; }
</style>
