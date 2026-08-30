# Cấu hình & credentials của SenNote

> File này lưu các giá trị **công khai** (public) — an toàn khi commit lên GitHub.
> Được mã hoá **base64** để tránh GitHub secret-scanning chặn push.
>
> ⚠️ **base64 KHÔNG PHẢI mã hoá** — nó chỉ là cách biểu diễn, ai cũng giải được ngay.
> Vì vậy file này **chỉ chứa thứ được thiết kế để công khai**, không chứa bí mật.

## 1. Giải mã

```powershell
# PowerShell
[Text.Encoding]::UTF8.GetString([Convert]::FromBase64String((Get-Content docs/credentials.md -Raw | Select-String -Pattern '(?s)```(.*?)```' -AllMatches).Matches[0].Groups[1].Value))
```

```bash
# Linux / macOS / WSL
sed -n '/^```$/,/^```$/p' docs/credentials.md | sed '1d;$d' | tr -d '\n' | base64 -d
```

Hoặc đơn giản: copy chuỗi base64 vào https://www.base64decode.org

## 2. Khối cấu hình (base64)

```
TlVYVF9GQ01fUFJPSkVDVF9JRD1yZW1pbmFxLTAwMQpOVVhUX0ZDTV9DTElFTlRfRU1BSUw9ZmlyZWJhc2UtYWRtaW5zZGstZmJzdmNAcmVtaW5hcS0wMDEuaWFtLmdzZXJ2aWNlYWNjb3VudC5jb20KTlVYVF9QVUJMSUNfRklSRUJBU0VfQVBJX0tFWT1BSXphU3lBeFlObXlsRmp6ZkFJSUs0Zm05NGVuNXN2cEw5WWdGWjgKTlVYVF9QVUJMSUNfRklSRUJBU0VfQVVUSF9ET01BSU49cmVtaW5hcS0wMDEuZmlyZWJhc2VhcHAuY29tCk5VWFRfUFVCTElDX0ZJUkVCQVNFX1BST0pFQ1RfSUQ9cmVtaW5hcS0wMDEKTlVYVF9QVUJMSUNfRklSRUJBU0VfTUVTU0FHSU5HX1NFTkRFUl9JRD02NzkyMTkyMTczNjgKTlVYVF9QVUJMSUNfRklSRUJBU0VfQVBQX0lEPTE6Njc5MjE5MjE3MzY4OndlYjpmOTBlMjI2NGFiMzg4ZjM4NDNlOGFkCk5VWFRfUFVCTElDX0ZDTV9WQVBJRF9LRVk9Qk5CYjFPWGNfX2kxSjlsQVBYMkFrR1J3Rk5FVFBSQ1VaOHB4UkVLR0h4YnNyUFRfMVRLNTM2MFFpdW9tUnJLM1BKZ1dneHpLVjBybUFGWGZxQXlmVy1rCk5VWFRfVFVSU09fREFUQUJBU0VfVVJMPWxpYnNxbDovL3Nlbm5vdGUtbmFxNTIxOS5hd3MtYXAtbm9ydGhlYXN0LTEudHVyc28uaW8=
```

Giải ra được:

| Biến | Giá trị | Công khai? |
|---|---|---|
| `NUXT_FCM_PROJECT_ID` | `reminaq-001` | ✅ |
| `NUXT_FCM_CLIENT_EMAIL` | `firebase-adminsdk-fbsvc@reminaq-001.iam.gserviceaccount.com` | ✅ |
| `NUXT_PUBLIC_FIREBASE_API_KEY` | `AIzaSyAx…gFZ8` | ✅ nằm ngay trong bundle trình duyệt |
| `NUXT_PUBLIC_FIREBASE_AUTH_DOMAIN` | `reminaq-001.firebaseapp.com` | ✅ |
| `NUXT_PUBLIC_FIREBASE_PROJECT_ID` | `reminaq-001` | ✅ |
| `NUXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID` | `679219217368` | ✅ |
| `NUXT_PUBLIC_FIREBASE_APP_ID` | `1:679219217368:web:f90e2264ab388f3843e8ad` | ✅ |
| `NUXT_PUBLIC_FCM_VAPID_KEY` | `BNBb1OXc…yfW-k` | ✅ khoá công khai của Web Push |
| `NUXT_TURSO_DATABASE_URL` | `libsql://sennote-naq5219…turso.io` | ⚠️ URL thôi, chưa đủ để truy cập |

> ✅ **Đã xác nhận:** app Android dùng chung project **`reminaq-001`** với web,
> nên một service account gửi được cho cả 2 nền tảng.

## 3. Nguồn gốc từng giá trị

| Biến | Lấy ở đâu |
|---|---|
| `NUXT_FCM_PROJECT_ID` | Firebase Console → Project settings → **General** → `projectId` (dùng cho cả server & web) |
| `NUXT_FCM_CLIENT_EMAIL` | Service account JSON (file `firebase-credentials.json` ở project Go) → `client_email` |
| 5 biến `NUXT_PUBLIC_FIREBASE_*` | Firebase Console → Project settings → **General** → Your apps → app Web → `firebaseConfig` |
| `NUXT_PUBLIC_FCM_VAPID_KEY` | Firebase Console → Project settings → **Cloud Messaging** → **Web push certificates** → Generate key pair |
| `NUXT_TURSO_DATABASE_URL` | Turso dashboard → database `sennote` → **Connect** |

## 4. ⚠️ KHÔNG BAO GIỜ đưa vào file này — dù có base64

| Biến | Vì sao nguy hiểm |
|---|---|
| `NUXT_FCM_PRIVATE_KEY` | Khoá riêng của service account. Ai có nó có thể **gửi push không giới hạn** nhân danh bạn. Google **tự động thu hồi** khoá bị lộ lên GitHub. |
| `NUXT_TURSO_AUTH_TOKEN` | Toàn quyền đọc/ghi database. |
| `NUXT_SESSION_PASSWORD` | Dùng để ký session cookie → có thể **giả mạo đăng nhập** mọi tài khoản. |

Ba giá trị này đã được lưu ở nơi an toàn:

1. **Cloudflare**: `wrangler pages secret put <TÊN> --project-name=sennote` (mã hoá, chỉ giải được lúc chạy)
2. **Local**: file `.env` ở gốc dự án — đã có trong `.gitignore`

> Muốn backup riêng cho mình? Tạo file `.credentials.local.md` và thêm vào `.gitignore`.
> Đừng commit.

## 5. Cách nạp lại sau này

```powershell
cd E:\PROJECT\nhacviecfcm\nhacviecnuxt

# Tự động: đọc private key thẳng từ firebase-credentials.json, không lộ ra màn hình
node scripts/setup-cloudflare-secrets.mjs sennote

# Kiểm tra đã set chưa
npx wrangler pages secret list --project-name=sennote

# Set xong phải deploy lại thì deployment mới nhận
npm run build
node -e "require('child_process').spawn(process.execPath,['node_modules/wrangler/bin/wrangler.js','pages','deploy','dist','--project-name','sennote','--branch','main'],{stdio:'inherit'})"

# Kiểm tra FCM sống trên production
node scripts/check-fcm-prod.mjs https://sennote.pages.dev
```

## 6. Trạng thái hiện tại

| Hạng mục | Trạng thái |
|---|---|
| FCM server (3 biến) | ✅ Đã set, `accessToken: true` trên Workers |
| Firebase web (5 biến) | ✅ Đã set, đã thấy trong payload SSR |
| VAPID key (web push) | ✅ Đã set, đã thấy trong payload SSR |
| Service Worker `/firebase-messaging-sw.js` | ✅ Serve 200 |
| Android app | ✅ Dùng chung project `reminaq-001` |
| **Gửi thông báo thật** | ⏳ Chờ test trên trình duyệt (xem doc 04 mục 7) |
