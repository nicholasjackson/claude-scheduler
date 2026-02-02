package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store wraps the SQLite database connection.
type Store struct {
	db *sql.DB
}

// Open creates or opens the SQLite database at the given path.
// It creates the parent directory if needed and initializes the schema.
func Open(dbPath string) (*Store, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating db directory: %w", err)
	}

	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("setting WAL mode: %w", err)
	}

	// Allow up to 5 seconds of retry when another goroutine holds the write lock.
	if _, err := sqlDB.Exec("PRAGMA busy_timeout=5000"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("setting busy timeout: %w", err)
	}

	s := &Store{db: sqlDB}
	if err := s.init(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("initializing schema: %w", err)
	}

	return s, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) init() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS jobs (
			id             TEXT PRIMARY KEY,
			name           TEXT NOT NULL,
			start_date     TEXT NOT NULL DEFAULT '',
			interval_value INTEGER NOT NULL DEFAULT 0,
			interval_unit  TEXT NOT NULL DEFAULT 'hours',
			next_run       TEXT NOT NULL DEFAULT '',
			last_run       TEXT NOT NULL DEFAULT '',
			status         TEXT NOT NULL DEFAULT 'pending',
			output         TEXT NOT NULL DEFAULT '',
			prompt         TEXT NOT NULL DEFAULT '',
			active         INTEGER NOT NULL DEFAULT 1
		)
	`)
	if err != nil {
		return err
	}

	// Migrate existing databases that lack the new columns.
	s.db.Exec("ALTER TABLE jobs ADD COLUMN prompt TEXT NOT NULL DEFAULT ''")
	s.db.Exec("ALTER TABLE jobs ADD COLUMN active INTEGER NOT NULL DEFAULT 1")
	s.db.Exec("ALTER TABLE jobs ADD COLUMN start_date TEXT NOT NULL DEFAULT ''")
	s.db.Exec("ALTER TABLE jobs ADD COLUMN interval_value INTEGER NOT NULL DEFAULT 0")
	s.db.Exec("ALTER TABLE jobs ADD COLUMN interval_unit TEXT NOT NULL DEFAULT 'hours'")

	// Drop legacy cron column if it exists.
	s.db.Exec("ALTER TABLE jobs DROP COLUMN cron")

	// Job runs history table.
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS job_runs (
			id         TEXT PRIMARY KEY,
			job_id     TEXT NOT NULL,
			started_at TEXT NOT NULL,
			ended_at   TEXT NOT NULL DEFAULT '',
			status     TEXT NOT NULL DEFAULT 'running',
			output     TEXT NOT NULL DEFAULT '',
			FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_job_runs_job_id ON job_runs(job_id)`)
	if err != nil {
		return err
	}

	// MCP servers table.
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS mcp_servers (
			id      TEXT PRIMARY KEY,
			name    TEXT NOT NULL UNIQUE,
			type    TEXT NOT NULL DEFAULT 'http',
			url     TEXT NOT NULL DEFAULT '',
			command TEXT NOT NULL DEFAULT '',
			args    TEXT NOT NULL DEFAULT '[]',
			env     TEXT NOT NULL DEFAULT '{}',
			headers TEXT NOT NULL DEFAULT '{}'
		)
	`)
	if err != nil {
		return err
	}

	// Join table for per-job MCP server selection.
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS job_mcp_servers (
			job_id        TEXT NOT NULL,
			mcp_server_id TEXT NOT NULL,
			PRIMARY KEY (job_id, mcp_server_id),
			FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE,
			FOREIGN KEY (mcp_server_id) REFERENCES mcp_servers(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Enable foreign key enforcement (SQLite has it off by default).
	_, err = s.db.Exec("PRAGMA foreign_keys = ON")
	return err
}
