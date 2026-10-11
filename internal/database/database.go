package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps standard sql.DB with application-specific helpers.
type DB struct {
	*sql.DB
}

// Config holds SQLite configuration options.
type Config struct {
	Path            string
	BusyTimeout     time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DefaultConfig returns default SQLite configuration.
func DefaultConfig(path string) Config {
	return Config{
		Path:            path,
		BusyTimeout:     5 * time.Second,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 1 * time.Hour,
	}
}

// New creates and opens a new SQLite database connection with sensible defaults.
func New(path string) (*DB, error) {
	return NewWithConfig(DefaultConfig(path))
}

// NewWithConfig creates and opens a new SQLite database connection with the provided configuration.
func NewWithConfig(cfg Config) (*DB, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("database path cannot be empty")
	}

	// If not an in-memory database or URI, ensure the directory exists
	if cfg.Path != ":memory:" && !strings.HasPrefix(cfg.Path, "file:") {
		dir := filepath.Dir(cfg.Path)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("creating database directory %q: %w", dir, err)
			}
		}
	}

	dsn := cfg.Path
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite database: %w", err)
	}

	// Set connection pool settings
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	// Apply standard performance & safety pragmas
	busyTimeoutMs := int(cfg.BusyTimeout.Milliseconds())
	if busyTimeoutMs <= 0 {
		busyTimeoutMs = 5000
	}

	pragmas := []string{
		fmt.Sprintf("PRAGMA busy_timeout = %d;", busyTimeoutMs),
		"PRAGMA foreign_keys = ON;",
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA cache_size = -2000;", // 2MB cache
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("executing pragma %q: %w", pragma, err)
		}
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging sqlite database: %w", err)
	}

	return &DB{DB: db}, nil
}
