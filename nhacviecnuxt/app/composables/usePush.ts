import { getApps, getApp, initializeApp } from 'firebase/app'
import { getMessaging, getToken, onMessage, isSupported } from 'firebase/messaging'
import type { DevicePlatform } from '#shared/types'

/**
 * Đăng ký nhận thông báo đẩy trên trình duyệt (Web Push qua FCM).
 *
 * Chỉ chạy ở client. Gọi `enable()` sau khi user đã đăng nhập.
 *
 * Yêu cầu:
 *  - HTTPS (localhost được coi là an toàn, dùng được khi dev)
 *  - `public/firebase-messaging-sw.js`
 *  - `NUXT_PUBLIC_FIREBASE_*` + `NUXT_PUBLIC_FCM_VAPID_KEY`
 */

type State = 'unsupported' | 'denied' | 'default' | 'granted' | 'ready' | 'loading'

export function usePush() {
  const state = useState<State>('push-state', () => 'default')
  const token = useState<string>('push-token', () => '')
  const error = useState<string>('push-error', () => '')
  const supported = useState<boolean>('push-supported', () => false)

  /** Kiểm tra trình duyệt có hỗ trợ Web Push không */
  async function check() {
    state.value = 'loading'
    error.value = ''

    const ok = await isSupported().catch(() => false)
    supported.value = ok
    if (!ok) {
      state.value = 'unsupported'
      error.value = 'Trình duyệt này không hỗ trợ thông báo đẩy'
      return
    }

    if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
      state.value = 'unsupported'
      error.value = 'Thiếu Service Worker hoặc Push API'
      return
    }

    state.value = (Notification.permission as State) || 'default'
  }

  /** Xin quyền + lấy token + đăng ký vào server */
  async function enable(): Promise<boolean> {
    error.value = ''
    state.value = 'loading'

    try {
      if (!(await isSupported())) {
        state.value = 'unsupported'
        error.value = 'Trình duyệt không hỗ trợ thông báo đẩy'
        return false
      }

      const permission = await Notification.requestPermission()
      state.value = permission as State
      if (permission !== 'granted') {
        error.value = 'Bạn đã từ chối quyền nhận thông báo'
        return false
      }

      const cfg = useRuntimeConfig().public
      if (!cfg.fcmVapidKey || !cfg.firebaseApiKey) {
        error.value = 'Thiếu NUXT_PUBLIC_FCM_VAPID_KEY / NUXT_PUBLIC_FIREBASE_API_KEY'
        return false
      }

      const app = getApps().length ? getApp() : initializeApp({
        apiKey: cfg.firebaseApiKey,
        authDomain: cfg.firebaseAuthDomain,
        projectId: cfg.firebaseProjectId,
        messagingSenderId: cfg.firebaseMessagingSenderId,
        appId: cfg.firebaseAppId,
      })

      // Truyền config qua query để service worker đọc được (file tĩnh không import được config)
      const swUrl = `/firebase-messaging-sw.js?apiKey=${encodeURIComponent(cfg.firebaseApiKey)}`
        + `&authDomain=${encodeURIComponent(cfg.firebaseAuthDomain)}`
        + `&projectId=${encodeURIComponent(cfg.firebaseProjectId)}`
        + `&messagingSenderId=${encodeURIComponent(cfg.firebaseMessagingSenderId)}`
        + `&appId=${encodeURIComponent(cfg.firebaseAppId)}`

      const registration = await navigator.serviceWorker.register(swUrl)
      await navigator.serviceWorker.ready

      const messaging = getMessaging(app)
      const fcmToken = await getToken(messaging, {
        vapidKey: cfg.fcmVapidKey,
        serviceWorkerRegistration: registration,
      })

      token.value = fcmToken

      // Gửi token về server theo chuẩn đa thiết bị
      await $fetch('/api/devices', {
        method: 'POST',
        body: { token: fcmToken, platform: 'web' as DevicePlatform },
      })

      // Thông báo khi tab ĐANG MỞ (foreground)
      onMessage(messaging, (payload) => {
        const n = payload.notification
        if (n?.title) {
          // eslint-disable-next-line no-new
          new Notification(n.title, { body: n.body ?? '', icon: '/icon-192.png' })
        }
      })

      state.value = 'ready'
      return true
    }
    catch (e: any) {
      error.value = e?.data?.message ?? e?.message ?? 'Không bật được thông báo'
      return false
    }
  }

  /** Gỡ đăng ký thiết bị này */
  async function disable(): Promise<boolean> {
    if (!token.value) return true
    try {
      await $fetch('/api/devices', { method: 'DELETE', body: { token: token.value } })
      token.value = ''
      state.value = 'default'
      return true
    }
    catch (e: any) {
      error.value = e?.data?.message ?? 'Không tắt được thông báo'
      return false
    }
  }

  return { state, token, error, supported, check, enable, disable }
}
