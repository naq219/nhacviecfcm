/* eslint-disable no-undef */
// Service Worker cho Firebase Cloud Messaging (Web Push).
// Nằm trong public/ → được serve ở gốc: /firebase-messaging-sw.js
//
// Lưu ý:
//  - KHÔNG dùng import module (trình duyệt cũ không hỗ trợ SW module).
//  - Phải là file tĩnh ở gốc domain, không được nằm trong /_nuxt/.
//  - Cấu hình truyền qua query string khi đăng ký SW (file tĩnh không đọc được
//    runtimeConfig, và cách này đảm bảo config CÓ SẴN lúc SW khởi động).
//
// ⚠️ ĐỪNG thêm listener `push` thủ công. Firebase messaging đã tự đăng ký
//    listener `push` của nó; thêm listener nữa => CÙNG 1 push bị xử lý 2 lần
//    => thông báo hiện 2 lần. Chỉ dùng onBackgroundMessage.

importScripts('https://www.gstatic.com/firebasejs/10.14.1/firebase-app-compat.js')
importScripts('https://www.gstatic.com/firebasejs/10.14.1/firebase-messaging-compat.js')

const params = new URLSearchParams(self.location.search)

firebase.initializeApp({
  apiKey: params.get('apiKey'),
  authDomain: params.get('authDomain'),
  projectId: params.get('projectId'),
  messagingSenderId: params.get('messagingSenderId'),
  appId: params.get('appId'),
})

const messaging = firebase.messaging()

/**
 * Hiển thị thông báo — DUY NHẤT MỘT NƠI gọi.
 *
 * `tag` cố định để nếu vì lý do gì bị gọi 2 lần, thông báo sau THAY THẾ
 * thông báo trước thay vì xếp chồng thành 2 cái.
 */
function showNotification(payload) {
  const n = (payload && payload.notification) || {}
  const data = (payload && payload.data) || {}
  const title = n.title || data.title || 'SenNote'

  const options = {
    body: n.body || data.body || '',
    icon: n.icon || '/icon-192.png',
    badge: '/icon-192.png',
    tag: 'sennote',
    renotify: true,
    data: {
      ...data,
      url: (payload && payload.fcmOptions && payload.fcmOptions.link) || data.url || '/',
    },
  }

  return self.registration.showNotification(title, options)
}

// Tab ĐÓNG / không focus → FCM gọi vào đây.
// Tab đang mở → FCM gửi thẳng về trang (xem onMessage trong usePush.ts).
// Hai đường này loại trừ nhau, không bao giờ cùng chạy cho 1 message.
messaging.onBackgroundMessage((payload) => {
  return showNotification(payload)
})

// Bấm vào thông báo → mở đúng trang
self.addEventListener('notificationclick', (event) => {
  event.notification.close()
  const url = (event.notification.data && event.notification.data.url) || '/'

  event.waitUntil(
    self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clientList) => {
      for (const client of clientList) {
        if (client.url === url && 'focus' in client) return client.focus()
      }
      if (self.clients.openWindow) return self.clients.openWindow(url)
    }),
  )
})

self.addEventListener('install', () => self.skipWaiting())
self.addEventListener('activate', () => self.clients.claim())
