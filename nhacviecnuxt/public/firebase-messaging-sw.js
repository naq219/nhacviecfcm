/* eslint-disable no-undef */
// Service Worker cho Firebase Cloud Messaging (Web Push).
// File nằm trong public/ nên được serve đúng ở gốc: /firebase-messaging-sw.js
//
// Lưu ý:
//  - KHÔNG dùng import module ở đây (trình duyệt cũ không hỗ trợ SW module).
//  - Phải là file tĩnh ở gốc domain, không được nằm trong /_nuxt/.

importScripts('https://www.gstatic.com/firebasejs/10.14.1/firebase-app-compat.js')
importScripts('https://www.gstatic.com/firebasejs/10.14.1/firebase-messaging-compat.js')

// Cấu hình được truyền qua URL khi đăng ký SW (?apiKey=...&...).
// Cách này tránh phải import config vào file tĩnh.
const params = new URLSearchParams(self.location.search)

firebase.initializeApp({
  apiKey: params.get('apiKey'),
  authDomain: params.get('authDomain'),
  projectId: params.get('projectId'),
  messagingSenderId: params.get('messagingSenderId'),
  appId: params.get('appId'),
})

const messaging = firebase.messaging()

// Thông báo khi tab đang Ở TRANG SAU (background) hoặc tab đã đóng.
messaging.onBackgroundMessage((payload) => {
  const notification = payload.notification || {}
  const title = notification.title || 'SenNote'
  const options = {
    body: notification.body || '',
    icon: notification.icon || '/icon-192.png',
    badge: '/icon-192.png',
    data: {
      ...(payload.data || {}),
      // link do server gửi qua webpush.fcmOptions.link
      url: payload.fcmOptions?.link || (payload.data && payload.data.url) || '/',
    },
  }
  self.registration.showNotification(title, options)
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
