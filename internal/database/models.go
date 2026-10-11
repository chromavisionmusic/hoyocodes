package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Subscription struct {
	ID        int64     `json:"id"`
	GuildID   string    `json:"guild_id"`
	ChannelID string    `json:"channel_id"`
	Game      string    `json:"game"`
	RoleID    *string   `json:"role_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CodeRecord struct {
	ID          int64     `json:"id"`
	Code        string    `json:"code"`
	Game        string    `json:"game"`
	Rewards     string    `json:"rewards"`
	FirstSeenAt time.Time `json:"first_seen_at"`
}

// GetSubscriptionsByGame returns all subscriptions for a given game.
func (db *DB) GetSubscriptionsByGame(ctx context.Context, game string) ([]Subscription, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, guild_id, channel_id, game, role_id, created_at, updated_at
		FROM subscriptions
		WHERE game = ?
	`, game)
	if err != nil {
		return nil, fmt.Errorf("query subscriptions by game: %w", err)
	}
	defer rows.Close()

	var subs []Subscription
	for rows.Next() {
		var s Subscription
		var roleID sql.NullString
		if err := rows.Scan(&s.ID, &s.GuildID, &s.ChannelID, &s.Game, &roleID, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		if roleID.Valid {
			s.RoleID = &roleID.String
		}
		subs = append(subs, s)
	}
	return subs, rows.Err()
}

// UpsertSubscription adds or updates a subscription for a guild and game.
func (db *DB) UpsertSubscription(ctx context.Context, guildID, channelID, game string, roleID *string) error {
	query := `
		INSERT INTO subscriptions (guild_id, channel_id, game, role_id, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(guild_id, game) DO UPDATE SET
			channel_id = excluded.channel_id,
			role_id = excluded.role_id,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := db.ExecContext(ctx, query, guildID, channelID, game, roleID)
	if err != nil {
		return fmt.Errorf("upsert subscription: %w", err)
	}
	return nil
}

// DeleteSubscription removes a subscription for a guild and game.
func (db *DB) DeleteSubscription(ctx context.Context, guildID, game string) error {
	query := `DELETE FROM subscriptions WHERE guild_id = ? AND game = ?`
	_, err := db.ExecContext(ctx, query, guildID, game)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}
	return nil
}

// SaveCode records a new code if not already saved.
func (db *DB) SaveCode(ctx context.Context, code, game, rewards string) error {
	query := `
		INSERT INTO codes (code, game, rewards)
		VALUES (?, ?, ?)
		ON CONFLICT(code, game) DO UPDATE SET
			rewards = excluded.rewards
	`
	_, err := db.ExecContext(ctx, query, code, game, rewards)
	if err != nil {
		return fmt.Errorf("save code: %w", err)
	}
	return nil
}

// IsNotificationSent checks if a code was already sent to a guild channel.
func (db *DB) IsNotificationSent(ctx context.Context, guildID, channelID, code, game string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM sent_notifications
			WHERE guild_id = ? AND channel_id = ? AND code = ? AND game = ?
		)
	`
	err := db.QueryRowContext(ctx, query, guildID, channelID, code, game).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check sent notification: %w", err)
	}
	return exists, nil
}

// RecordNotificationSent records that a notification has been sent.
func (db *DB) RecordNotificationSent(ctx context.Context, guildID, channelID, code, game string) error {
	query := `
		INSERT INTO sent_notifications (guild_id, channel_id, code, game)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(guild_id, channel_id, code, game) DO NOTHING
	`
	_, err := db.ExecContext(ctx, query, guildID, channelID, code, game)
	if err != nil {
		return fmt.Errorf("record notification sent: %w", err)
	}
	return nil
}
