package database_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/chromavisionmusic/hoyocodes/internal/database"
)

func TestDatabase_New_Memory(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create memory database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	// Verify foreign_keys pragma is enabled
	var foreignKeys int
	if err := db.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		t.Fatalf("failed to check foreign_keys: %v", err)
	}
	if foreignKeys != 1 {
		t.Errorf("expected foreign_keys to be 1, got %d", foreignKeys)
	}
}

func TestDatabase_New_FileCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "nested", "dir", "test.db")

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create file database in nested directory: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("failed to ping: %v", err)
	}
}

func TestDatabase_Migrate_And_Idempotency(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// First migration
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}

	// Second migration (idempotent run)
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("subsequent migration failed: %v", err)
	}

	// Verify schema_migrations table
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("failed to count migrations: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 migration recorded, got %d", count)
	}
}

func TestDatabase_Subscriptions(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	role := "123456789"
	err = db.UpsertSubscription(ctx, "guild_1", "chan_1", "genshin", &role)
	if err != nil {
		t.Fatalf("upsert subscription failed: %v", err)
	}

	subs, err := db.GetSubscriptionsByGame(ctx, "genshin")
	if err != nil {
		t.Fatalf("get subscriptions failed: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(subs))
	}
	if subs[0].GuildID != "guild_1" || subs[0].ChannelID != "chan_1" || *subs[0].RoleID != role {
		t.Errorf("unexpected subscription values: %+v", subs[0])
	}

	// Update channel for same guild & game
	newChan := "chan_2"
	err = db.UpsertSubscription(ctx, "guild_1", newChan, "genshin", nil)
	if err != nil {
		t.Fatalf("update subscription failed: %v", err)
	}

	subs, err = db.GetSubscriptionsByGame(ctx, "genshin")
	if err != nil {
		t.Fatalf("get subscriptions after update failed: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 subscription after update, got %d", len(subs))
	}
	if subs[0].ChannelID != newChan {
		t.Errorf("expected channel %s, got %s", newChan, subs[0].ChannelID)
	}
	if subs[0].RoleID != nil {
		t.Errorf("expected nil role_id, got %v", subs[0].RoleID)
	}

	// Delete subscription
	err = db.DeleteSubscription(ctx, "guild_1", "genshin")
	if err != nil {
		t.Fatalf("delete subscription failed: %v", err)
	}

	subs, err = db.GetSubscriptionsByGame(ctx, "genshin")
	if err != nil {
		t.Fatalf("get subscriptions after delete failed: %v", err)
	}
	if len(subs) != 0 {
		t.Errorf("expected 0 subscriptions after delete, got %d", len(subs))
	}
}

func TestDatabase_CodesAndNotifications(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Save code
	if err := db.SaveCode(ctx, "GENSHINGIFT", "genshin", "50 Primogems"); err != nil {
		t.Fatalf("save code failed: %v", err)
	}

	// Check notification before recording
	sent, err := db.IsNotificationSent(ctx, "guild_1", "chan_1", "GENSHINGIFT", "genshin")
	if err != nil {
		t.Fatalf("is notification sent failed: %v", err)
	}
	if sent {
		t.Errorf("expected notification not sent yet")
	}

	// Record notification
	if err := db.RecordNotificationSent(ctx, "guild_1", "chan_1", "GENSHINGIFT", "genshin"); err != nil {
		t.Fatalf("record notification sent failed: %v", err)
	}

	// Check notification after recording
	sent, err = db.IsNotificationSent(ctx, "guild_1", "chan_1", "GENSHINGIFT", "genshin")
	if err != nil {
		t.Fatalf("is notification sent failed: %v", err)
	}
	if !sent {
		t.Errorf("expected notification to be marked as sent")
	}
}
