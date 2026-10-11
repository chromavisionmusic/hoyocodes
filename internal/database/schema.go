package database

import (
	"context"
	"fmt"
)

type Migration struct {
	Version     int
	Description string
	SQL         string
}

var migrations = []Migration{
	{
		Version:     1,
		Description: "Create initial tables for subscriptions, codes, and sent notifications",
		SQL: `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    description TEXT NOT NULL,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    guild_id TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    game TEXT NOT NULL,
    role_id TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (guild_id, game)
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_game ON subscriptions (game);
CREATE INDEX IF NOT EXISTS idx_subscriptions_guild ON subscriptions (guild_id);

CREATE TABLE IF NOT EXISTS codes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL,
    game TEXT NOT NULL,
    rewards TEXT,
    first_seen_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (code, game)
);

CREATE INDEX IF NOT EXISTS idx_codes_game ON codes (game);

CREATE TABLE IF NOT EXISTS sent_notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    guild_id TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    code TEXT NOT NULL,
    game TEXT NOT NULL,
    sent_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (guild_id, channel_id, code, game)
);

CREATE INDEX IF NOT EXISTS idx_sent_notifications_code ON sent_notifications (code, game);
`,
	},
}

// Migrate applies all pending schema migrations.
func (db *DB) Migrate(ctx context.Context) error {
	initMigrationsTable := `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    description TEXT NOT NULL,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
	if _, err := db.ExecContext(ctx, initMigrationsTable); err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	for _, m := range migrations {
		var exists bool
		err := db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)", m.Version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("checking migration version %d: %w", m.Version, err)
		}

		if exists {
			continue
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("beginning tx for migration %d: %w", m.Version, err)
		}

		if _, err := tx.ExecContext(ctx, m.SQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("executing migration %d (%s): %w", m.Version, m.Description, err)
		}

		insertQuery := "INSERT INTO schema_migrations (version, description) VALUES (?, ?)"
		if _, err := tx.ExecContext(ctx, insertQuery, m.Version, m.Description); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("recording migration %d: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %d: %w", m.Version, err)
		}
	}

	return nil
}
