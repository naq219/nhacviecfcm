-- SenNote — schema Turso (libSQL)
-- Mọi cột datetime lưu ISO 8601 UTC (vd 2026-08-28T07:00:00.000Z)
-- để so sánh trực tiếp bằng chuỗi với cron chạy theo giờ UTC.

-- Bảng người dùng
CREATE TABLE IF NOT EXISTS users (
  id            TEXT PRIMARY KEY,          -- crypto.randomUUID()
  email         TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,             -- hashPassword() của nuxt-auth-utils
  fcm_token     TEXT,                      -- token FCM của mobile (1 user 1 token, ghi đè)
  is_fcm_active INTEGER NOT NULL DEFAULT 1,
  created_at    TEXT NOT NULL,             -- ISO 8601 (UTC)
  updated_at    TEXT NOT NULL
);

-- Thiết bị nhận thông báo (1 user nhiều thiết bị: Android + web + iOS)
-- Lý do tách: trước đây users.fcm_token chỉ chứa 1 token, thiết bị đăng ký sau
-- sẽ GHI ĐÈ thiết bị trước (mở web là mất thông báo trên Android).
CREATE TABLE IF NOT EXISTS devices (
  id         TEXT PRIMARY KEY,
  user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token      TEXT NOT NULL,
  platform   TEXT NOT NULL DEFAULT 'android',  -- 'android' | 'ios' | 'web'
  is_active  INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_devices_token ON devices(token);
CREATE INDEX IF NOT EXISTS idx_devices_user ON devices(user_id, is_active);

-- Bảng nhắc nhở
CREATE TABLE IF NOT EXISTS reminders (
  id                    TEXT PRIMARY KEY,
  user_id               TEXT NOT NULL REFERENCES users(id),
  title                 TEXT NOT NULL,
  description           TEXT,
  tag                   TEXT,
  type                  TEXT NOT NULL,            -- 'one_time' | 'recurring'
  status                TEXT NOT NULL DEFAULT 'active', -- 'active'|'completed'|'paused'
  recurrence_pattern    TEXT,                     -- JSON (xem docs/03)
  repeat_strategy       TEXT NOT NULL DEFAULT 'none',   -- 'none'|'crp_until_complete'
  calendar_type         TEXT NOT NULL DEFAULT 'solar',  -- 'solar'|'lunar'
  origin_time           TEXT,                     -- MỎ NEO TÍNH LỊCH LẶP
  next_recurring        TEXT,                     -- ISO 8601 UTC
  next_crp              TEXT,
  max_crp               INTEGER NOT NULL DEFAULT 0,
  crp_count             INTEGER NOT NULL DEFAULT 0,
  crp_interval_sec      INTEGER NOT NULL DEFAULT 0,
  is_sended_one_time    INTEGER NOT NULL DEFAULT 0,
  next_action_at        TEXT,                     -- thời điểm gần nhất cần xử lý
  last_sent_at          TEXT,
  last_completed_at     TEXT,
  last_crp_completed_at TEXT,
  snooze_until          TEXT,
  created_at            TEXT NOT NULL,
  updated_at            TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_reminders_next_action ON reminders(next_action_at, status);
CREATE INDEX IF NOT EXISTS idx_reminders_user ON reminders(user_id, status);

-- Kill-switch cho cron (port từ system_status của Go)
CREATE TABLE IF NOT EXISTS system_status (
  mid            INTEGER PRIMARY KEY CHECK (mid = 1),
  worker_enabled INTEGER NOT NULL DEFAULT 0,
  last_error     TEXT,
  created_at     TEXT NOT NULL,
  updated_at     TEXT NOT NULL
);
INSERT OR IGNORE INTO system_status (mid, worker_enabled, last_error, created_at, updated_at)
VALUES (1, 1, '', datetime('now'), datetime('now'));
