import { getApps, getApp, initializeApp } from 'firebase/app'
import { getMessaging, getToken, onMessage, isSupported } from 'firebase/messaging'
import type { Messaging } from 'firebase/messaging'
import type { DevicePlatform } from '#shared/types'

/**
 * Đăng ký nhận thông báo đẩy trên trình duyệt (Web Push qua FCM).
 *
 * ⚠️ Hai bẫy đã gặp và cách xử lý trong bản này:
 *
 * 1) **Trình duyệt đã cấp quyền từ trước** → `Notification.permission === 'granted'`
 *    nhưng device CHƯA được đăng ký cho tài khoản hiện tại (mỗi tài khoản có
 *    thiết bị riêng). Nếu chỉ hiện nút khi permission là 'default' thì user
 *    mãi mãi không đăng ký được.
 *    → `check()` tự động đăng ký luôn khi permission đã 'granted'.
 *
 * 2) **Thông báo lúc tab đang mở bị mất**: FCM đẩy về trang (foreground) thay vì
 *    Service Worker, nên cần listener `onMessage`. Listener này chỉ tồn tại trong
 *    phiên hiện tại → mỗi lần tải lại trang phải gắn lại.
 *    → `register()` luôn gắn lại `onMessage`.
 */

type State = 'unsupported' | 'denied' | 'default' | 'granted' | 'ready' | 'loading'

const OFF_FLAG = 'sennote:push-off'

export function usePush() {
  const state = useState<State>('push-state', () => 'default')
  const token = useState<string>('push-token', () => '')
  const error = useState<string>('push-error', () => '')
  const supported = useState<boolean>('push-supported', () => false)
  const busy = ref(false)

  let messaging: Messaging | null = null

  function firebaseConfig() {
    const c = useRuntimeConfig().public
    return {
      apiKey: c.firebaseApiKey,
      authDomain: c.firebaseAuthDomain,
      projectId: c.firebaseProjectId,
      messagingSenderId: c.firebaseMessagingSenderId,
      appId: c.firebaseAppId,
      vapidKey: c.fcmVapidKey,
    }
  }

  function isOff(): boolean {
    if (!import.meta.client) return false
    return localStorage.getItem(OFF_FLAG) === '1'
  }

  /** Nếu đã có quyền → tự đăng ký luôn (fix việc nút bị ẩn) */
  async function check() {
    if (busy.value) return
    error.value = ''

    const ok = await isSupported().catch(() => false)
    supported.value = ok
    if (!ok) {
      state.value = 'unsupported'
      error.value = 'Trình duyệt này không hỗ trợ thông báo đẩy'
      return
    }

    const permission = Notification.permission
    if (permission === 'denied') {
      state.value = 'denied'
      return
    }

    if (permission === 'granted') {
      state.value = 'granted'
      if (!isOff()) await register()
      return
    }

    state.value = 'default'
  }

  /** Xin quyền rồi đăng ký (dùng khi user bấm nút) */
  async function enable(): Promise<boolean> {
    error.value = ''
    busy.value = true
    try {
      if (!(await isSupported().catch(() => false))) {
        state.value = 'unsupported'
        error.value = 'Trình duyệt không hỗ trợ thông báo đẩy'
        return false
      }

      const permission = await Notification.requestPermission()
      if (permission !== 'granted') {
        state.value = permission === 'denied' ? 'denied' : 'default'
        error.value = permission === 'denied'
          ? 'Bạn đã chặn quyền thông báo — mở lại ở icon khoá trên thanh địa chỉ'
          : 'Bạn chưa cho phép nhận thông báo'
        return false
      }

      if (import.meta.client) localStorage.removeItem(OFF_FLAG)
      return await register()
    }
    catch (e: any) {
      error.value = e?.data?.message ?? e?.message ?? 'Không bật được thông báo'
      return false
    }
    finally {
      busy.value = false
    }
  }

  /**
   * Lấy token FCM + đăng ký vào server + gắn lại listener foreground.
   * Gọi lại nhiều lần an toàn (getToken trả token đã缓存, server upsert).
   */
  async function register(): Promise<boolean> {
    const cfg = firebaseConfig()
    if (!cfg.vapidKey || !cfg.apiKey) {
      error.value = 'Thiếu NUXT_PUBLIC_FCM_VAPID_KEY / NUXT_PUBLIC_FIREBASE_API_KEY'
      return false
    }

    busy.value = true
    try {
      const app = getApps().length ? getApp() : initializeApp({
        apiKey: cfg.apiKey,
        authDomain: cfg.authDomain,
        projectId: cfg.projectId,
        messagingSenderId: cfg.messagingSenderId,
        appId: cfg.appId,
      })

      // Truyền config qua query để Service Worker (file tĩnh) đọc được
      const swUrl = `/firebase-messaging-sw.js?apiKey=${encodeURIComponent(cfg.apiKey)}`
        + `&authDomain=${encodeURIComponent(cfg.authDomain)}`
        + `&projectId=${encodeURIComponent(cfg.projectId)}`
        + `&messagingSenderId=${encodeURIComponent(cfg.messagingSenderId)}`
        + `&appId=${encodeURIComponent(cfg.appId)}`

      const registration = await navigator.serviceWorker.register(swUrl)
      await navigator.serviceWorker.ready

      messaging = getMessaging(app)
      const fcmToken = await getToken(messaging, {
        vapidKey: cfg.vapidKey,
        serviceWorkerRegistration: registration,
      })
      token.value = fcmToken

      // Gắn lại listener foreground — MẤT sau mỗi lần reload nên phải gắn ở đây.
      //
      // ⚠️ Chỉ hiện ở đây khi tab đang mở. Khi tab đóng/không focus,
      // FCM gọi onBackgroundMessage trong Service Worker.
      // Hai đường LOẠI TRỪ NHAU — không được xử lý ở cả 2 nơi kẻo hiện 2 lần.
      // `tag` trùng với SW để nếu có trùng lặp thì thay thế thay vì xếp chồng.
      onMessage(messaging, (payload) => {
        const n = payload.notification
        if (!n?.title) return
        try {
          navigator.serviceWorker.ready.then(reg => reg.showNotification(n.title!, {
            body: n.body ?? '',
            icon: '/icon-192.png',
            badge: '/icon-192.png',
            tag: 'sennote',
            data: { url: payload.fcmOptions?.link || '/' },
          }))
        }
        catch {
          // bỏ qua nếu không tạo được (vd chưa có SW)
        }
      })

      await $fetch('/api/devices', {
        method: 'POST',
        body: { token: fcmToken, platform: 'web' as DevicePlatform },
      })

      state.value = 'ready'
      return true
    }
    catch (e: any) {
      error.value = e?.data?.message ?? e?.message ?? 'Không đăng ký được thiết bị'
      return false
    }
    finally {
      busy.value = false
    }
  }

  /** Gỡ thiết bị này khỏi tài khoản hiện tại */
  async function disable(): Promise<boolean> {
    if (!token.value) return true
    busy.value = true
    try {
      await $fetch('/api/devices', { method: 'DELETE', body: { token: token.value } })
      if (import.meta.client) localStorage.setItem(OFF_FLAG, '1')
      token.value = ''
      state.value = 'granted' // vẫn còn quyền, chỉ là đã gỡ thiết bị
      return true
    }
    catch (e: any) {
      error.value = e?.data?.message ?? 'Không tắt được thông báo'
      return false
    }
    finally {
      busy.value = false
    }
  }

  return { state, token, error, supported, busy, check, enable, disable, register }
}
