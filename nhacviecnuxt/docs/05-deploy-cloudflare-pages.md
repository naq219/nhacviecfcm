# 05 — Deploy lên Cloudflare Pages (sennote.pages.dev)

> Mục tiêu: toàn bộ app (API + frontend + cron) online tại **https://sennote.pages.dev**, mỗi lần push là tự deploy.
> Yêu cầu: đã xong doc 03 + 04, app chạy ngon ở local.

---

## 0. Hiểu nhanh

- **Cloudflare Pages** = host miễn phí. Với Nuxt fullstack, app chạy dưới dạng **Pages Functions** (chính là Cloudflare Workers) — vừa render web, vừa chạy API, vừa nhận cron.
- **URL** = tên project: đặt tên project `sennote` → `sennote.pages.dev`.
- **Cron** (từ doc 04) được Nitro tự sinh khi build → Cloudflare tự nhận, không cần cấu hình thêm.
- **Không còn** vấn đề CORS/mixed-content như mô hình "frontend gọi backend riêng" — API cùng origin.

## 1. Chuẩn bị

1. Tài khoản Cloudflare (miễn phí): https://dash.cloudflare.com/sign-up
2. (Cách A) tài khoản GitHub + code đã push lên repo.
3. Build thử ở local cho chắc:

```powershell
cd E:\PROJECT\nhacviecnuxt
npm run build
```

✅ Chạy xong không lỗi và sinh ra thư mục `dist/` là OK (nhờ `nitro.preset: 'cloudflare-pages'` ở `nuxt.config.ts`).

## 2. Cách A — Deploy qua GitHub (khuyến nghị)

### 2.1 Push code lên GitHub

```powershell
cd E:\PROJECT\nhacviecnuxt
git init
git add .
git commit -m "khoi tao sen note fullstack nuxt"
git remote add origin https://github.com/<ban>/nhacviecnuxt.git
git push -u origin main
```

> ⚠️ Kiểm tra trước khi push: `.env` KHÔNG được commit (`.gitignore` của Nuxt đã lo `.env`).

> ⚠️ **Nếu bạn để dự án Nuxt NẰM TRONG repo Go** (`E:\PROJECT\nhacviecfcm\nhacviecnuxt`),
> `git init` sẽ tạo **nested repo**. Hãy chọn 1 trong 2:
> - **Khuyến nghị:** chuyển dự án ra `E:\PROJECT\nhacviecnuxt` (xem doc 02 bước 0.5), rồi `git init` bình thường.
> - Hoặc giữ nguyên và thêm `nhacviecnuxt/` vào `.gitignore` của repo `nhacviecfcm` để repo Go bỏ qua nó.
>
> Kiểm tra nhanh bạn có đang trong repo khác không: `git rev-parse --show-toplevel`

### 2.2 Tạo project trên Cloudflare

1. https://dash.cloudflare.com → **Workers & Pages** → **Create** → tab **Pages** → **Connect to Git**.
2. Chọn GitHub → repo `nhacviecnuxt` → **Begin setup**.
3. Điền form:
   - **Project name:** `sennote` ← quyết định URL `sennote.pages.dev`
   - **Production branch:** `main`
   - **Framework preset:** `Nuxt`
   - **Build command:** `npm run build`
   - **Build output directory:** `dist`
4. Mở **Environment variables**, thêm các biến (giá trị y hệt `.env` local):

   | Tên | Nội dung |
   |-----|----------|
   | `NUXT_SESSION_PASSWORD` | chuỗi ≥ 32 ký tự |
   | `NUXT_TURSO_DATABASE_URL` | `libsql://...` |
   | `NUXT_TURSO_AUTH_TOKEN` | token Turso |
   | `NUXT_FCM_CLIENT_EMAIL` | `client_email` service account |
   | `NUXT_FCM_PRIVATE_KEY` | `private_key` (phải giữ nguyên chuỗi có `\n`) |
   | `NUXT_FCM_PROJECT_ID` | `project_id` |
   | `NODE_VERSION` | `20` |

   > ℹ️ Tên biến bắt buộc có prefix `NUXT_` (đúng y hệt `.env` local) — Nuxt chỉ nạp env có prefix này vào `runtimeConfig`.

5. **Save and Deploy** → chờ 1–3 phút.
6. ✅ Mở https://sennote.pages.dev → đăng ký tài khoản → tạo reminder → hoạt động bình thường = THÀNH CÔNG.

> ℹ️ Các biến FCM/`NUXT_SESSION_PASSWORD` nên đánh dấu **Secret** (mục **Secret variables**) để không bị hiện trong log — cách set giống nhau, chỉ chọn "Encrypt" khi nhập.

### 2.3 Deploy sau này

Không cần làm gì: `git push` lên `main` → Cloudflare tự build + deploy. Xem log ở tab **Deployments**.

## 3. Cách B — Deploy thủ công bằng wrangler

```powershell
cd E:\PROJECT\nhacviecnuxt
npm run build
npx wrangler login                                   # đăng nhập Cloudflare (1 lần)
npx wrangler pages deploy dist --project-name=sennote
```

> ⚠️ **Cách B build ở máy bạn → vẫn phải set env trên Cloudflare.** Đã kiểm chứng: Nitro
> **không** nướng env vào `dist/` (đọc lúc runtime qua `globalThis.__env__`). Vì vậy thiếu env
> trên Cloudflare là app chạy nhưng mọi API trả 500.

### 2.4 Cách B — những điểm bắt buộc phải chú ý

1. **Luôn thêm `--branch main`.** Nếu không, wrangler lấy tên branch git hiện tại
   (vd `worker-v4`) → deployment thành **Preview**, hệ quả:
   - `sennote.pages.dev` vẫn **404** (branch production chưa có deployment).
   - Preview **không nhận** production secrets → mọi API trả **500**.
   ```powershell
   npx wrangler pages deploy dist --project-name=sennote --branch main
   ```
2. **Set secrets bằng wrangler** (không cần vào dashboard):
   ```powershell
   npm i -D wrangler
   # gõ giá trị rồi Enter (hoặc pipe từ file)
   npx wrangler pages secret put NUXT_SESSION_PASSWORD      --project-name=sennote
   npx wrangler pages secret put NUXT_TURSO_AUTH_TOKEN      --project-name=sennote
   npx wrangler pages secret put NUXT_TURSO_DATABASE_URL    --project-name=sennote
   npx wrangler pages secret list --project-name=sennote
   ```
   ⚠️ Dùng `npx --yes wrangler@latest ...` sẽ bị npm nuốt mất argument → cài wrangler local.
3. **Set xong secret phải deploy lại** thì deployment mới nhận được.
4. **Kiểm tra env là Production hay Preview:**
   `npx wrangler pages deployment list --project-name=sennote`
5. **Test end-to-end sau deploy:**
   ```powershell
   node scripts/e2e-smoke.mjs https://sennote.pages.dev
   ```

> ℹ️ Lần đầu wrangler tự tạo project `sennote` (không cần `pages project create` nếu đã tạo).

## 4. Kiểm tra cron có chạy không

1. Vào Cloudflare → project `sennote` → **Settings** → **Functions** → **Cron Triggers** → thấy mục `* * * * *` (do Nitro tự sinh).
2. (Local) chạy thử task: mở `http://localhost:3000/_nitro/tasks/reminders:check` khi `npm run dev` đang chạy (xem doc 04 mục 7).
3. (Prod) tạo 1 reminder có `next_action_at` trong quá khứ, chờ 1–2 phút, xem thiết bị có nhận push không (cần mobile đã lưu FCM token).

## 5. Checklist sau deploy

- [ ] Mở https://sennote.pages.dev thấy trang login
- [ ] Đăng ký + đăng nhập + tạo/xoá reminder hoạt động
- [ ] F5 / vào trực tiếp `/login` không bị 404 (SSR nên không cần `_redirects`)
- [ ] F12 → Network: request tới `/api/...` trả 200, không có lỗi CORS
- [ ] Cron Triggers xuất hiện trong dashboard
- [ ] Đổi code → push → vài phút sau web online cập nhật

## 6. Lỗi thường gặp

| Triệu chứng | Nguyên nhân | Cách sửa |
|-------------|-------------|----------|
| Build trên Cloudflare fail, log lỗi Node | Node cũ | Thêm env `NODE_VERSION=20` rồi Retry |
| `sennote.pages.dev` báo 404 | Deploy thiếu `--branch main` → thành Preview | Deploy lại với `--branch main` |
| API trả 500 hàng loạt | Chưa set secrets trên Cloudflare (env đọc lúc runtime) | `wrangler pages secret put` rồi **deploy lại** |
| Đăng ký được nhưng **không đăng nhập được** | Dùng `hashPassword`/`verifyPassword` của nuxt-auth-utils (scrypt hỏng trên Workers) | Dùng `makePasswordHash`/`verifyPasswordHash` trong `server/utils/password.ts` |
| `wrangler pages secret put` báo lỗi npm | Dùng `npx --yes wrangler@latest` | Cài wrangler local (`npm i -D wrangler`) |
| Cron không chạy dù code đúng | Kill-switch `worker_enabled = 0` | `PUT /api/system_status { "worker_enabled": true }` |
| Login được local nhưng online lỗi session | Thiếu `NUXT_SESSION_PASSWORD` trên Cloudflare | Thêm đúng secret, deploy lại |
| `/api/ping` online lỗi `unauthorized` | Thiếu/ sai biến Turso | Kiểm tra `NUXT_TURSO_DATABASE_URL` + `NUXT_TURSO_AUTH_TOKEN` |
| Cron không thấy trong dashboard | `experimental.tasks` chưa bật, hoặc `scheduledTasks` đặt ngoài khối `nitro` | Kiểm tra `nuxt.config.ts`, build lại |
| Deploy xong vẫn thấy bản cũ | Cache trình duyệt | Ctrl+Shift+R (hard refresh) |
| Wrangler báo thiếu quyền | Chưa login | `npx wrangler login` lại |

## 7. Nâng cao (làm sau, không bắt buộc)

- **Custom domain:** project → **Custom domains** → trỏ domain riêng; `sennote.pages.dev` vẫn hoạt động song song.
- **Rollback:** tab **Deployments** → chọn deployment cũ → **⋯ → Rollback**.
- **Preview:** push nhánh khác `main` → Cloudflare tạo URL preview riêng để test trước khi merge.

---

🎉 Xong! SenNote (API + web + cron) đã online tại **https://sennote.pages.dev**.
