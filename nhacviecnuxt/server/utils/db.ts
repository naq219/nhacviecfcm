import { createClient } from '@libsql/client'

let _db: ReturnType<typeof createClient> | null = null

/**
 * Nơi DUY NHẤT tạo connection tới Turso.
 * Không gọi ở top-level module (Cloudflare chỉ đọc env trong vòng đời request/task).
 */
export function useDb() {
  if (_db) return _db

  const cfg = useRuntimeConfig()
  _db = createClient({
    url: cfg.tursoDatabaseUrl,
    authToken: cfg.tursoAuthToken,
  })
  return _db
}
