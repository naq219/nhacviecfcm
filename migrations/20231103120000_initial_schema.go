package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// up queries
		queries := []string{
			`
-- remiaq Initial Database Schema

-- Table: users
CREATE TABLE IF NOT EXISTS musers (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    fcm_token TEXT,
    is_fcm_active BOOLEAN DEFAULT TRUE,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_musers_email ON musers(email);
CREATE INDEX IF NOT EXISTS idx_musers_fcm_active ON musers(is_fcm_active);

-- Table: reminders
CREATE TABLE IF NOT EXISTS reminders (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    type TEXT NOT NULL CHECK(type IN ('one_time', 'recurring')),
    calendar_type TEXT DEFAULT 'solar' CHECK(calendar_type IN ('solar', 'lunar')),
    next_trigger_at DATETIME NOT NULL,
    trigger_time_of_day TEXT,
    recurrence_pattern TEXT,
    repeat_strategy TEXT DEFAULT 'none' CHECK(repeat_strategy IN ('none', 'retry_until_complete')),
    retry_interval_sec INTEGER,
    max_retries INTEGER DEFAULT 0,
    retry_count INTEGER DEFAULT 0,
    status TEXT DEFAULT 'active' CHECK(status IN ('active', 'completed', 'paused')),
    snooze_until DATETIME,
    last_completed_at DATETIME NULL,
    last_sent_at DATETIME,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES musers(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_reminders_next_trigger ON reminders(next_trigger_at);
CREATE INDEX IF NOT EXISTS idx_reminders_user_id ON reminders(user_id);
CREATE INDEX IF NOT EXISTS idx_reminders_status ON reminders(status);
CREATE INDEX IF NOT EXISTS idx_reminders_user_status ON reminders(user_id, status);
CREATE INDEX IF NOT EXISTS idx_reminders_status_trigger ON reminders(status, next_trigger_at);

-- Table: system_status (singleton table)
CREATE TABLE IF NOT EXISTS system_status (
    mid INTEGER PRIMARY KEY CHECK (mid = 1),
    worker_enabled BOOLEAN DEFAULT TRUE,
    last_error TEXT,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert default system status
INSERT OR IGNORE INTO system_status (mid, worker_enabled) VALUES (1, TRUE);
			`,
		}

		for _, query := range queries {
			if _, err := app.DB().NewQuery(query).Execute(); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		// down queries (optional)
		return nil
	})
}