/**
 * Copyright 2026 AiT authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
**/

package store

import (
	"database/sql"
	"errors"
	"os"

	"github.com/os-artificer/AiT/internal/pkg/platform"

	_ "modernc.org/sqlite"
)

// ErrSessionNotFound is returned when a session lookup misses.
var ErrSessionNotFound = errors.New("session not found")

// ErrSessionAmbiguous is returned when id and alias both match different rows.
var ErrSessionAmbiguous = errors.New("session id/alias lookup is ambiguous")

const (
	StatusActive    = "active"
	StatusSuspended = "suspended"

	AuthMethodKey      = "key"
	AuthMethodPassword = "password"
	AuthMethodAgent    = "agent"

	MetaCurrentSession = "current_session_id"
)

// Session is a saved SSH session profile.
type Session struct {
	ID            string
	Alias         string
	Host          string
	Port          int
	Username      string
	IdentityPath  string
	AuthMethod    string
	AuthExpiresAt int64
	TmuxEnabled   bool
	TmuxSession   string
	// LoginDestination is the ssh config Host or CLI destination (e.g. devhost), not the AiT --alias.
	LoginDestination string
	// ProxyJump is comma-separated jump hosts from login -J or ssh -G (used on load re-resolve).
	ProxyJump string
	// ProxyCommand is the proxycommand from login -o or ssh -G when persisted.
	ProxyCommand string
	CreatedAt    int64
	UpdatedAt    int64
	Status       string
}

// SessionState is local terminal snapshot on suspend.
type SessionState struct {
	SessionID      string
	Cols           int
	Rows           int
	RemoteCwd      string
	ScrollbackPath string
	SuspendedAt    int64
}

// Store wraps SQLite persistence.
type Store struct {
	db *sql.DB
}

// Open opens or creates the SQLite session store at the platform default database path.
// Returns a Store with schema migrated, or an error from platform.DBPath, directory creation,
// sql.Open, or migrate.
func Open() (*Store, error) {
	path, err := platform.DBPath()
	if err != nil {
		return nil, err
	}
	return OpenPath(path)
}

// OpenPath opens SQLite at path. Use ":memory:" for in-process tests.
// path is the database file path; parent directories are created (except for
// :memory:) with mode 0700. A non-memory file is chmod 0600 after migrate.
// Returns a Store with schema applied.
// Errors: directory creation; sql.Open; migrate failure (db is closed).
func OpenPath(path string) (*Store, error) {
	if err := ensureDBParentDir(path); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", buildDSN(path))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}

	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	if path != sqliteMemoryPath {
		_ = os.Chmod(path, 0o600)
	}
	return s, nil
}

func (s *Store) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			alias TEXT UNIQUE,
			host TEXT NOT NULL,
			port INTEGER NOT NULL DEFAULT 22,
			username TEXT NOT NULL,
			identity_path TEXT,
			auth_method TEXT NOT NULL,
			auth_expires_at INTEGER NOT NULL,
			tmux_enabled INTEGER NOT NULL DEFAULT 0,
			tmux_session TEXT,
			login_destination TEXT,
			proxy_jump TEXT,
			proxy_command TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			status TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS session_state (
			session_id TEXT PRIMARY KEY REFERENCES sessions(id) ON DELETE CASCADE,
			cols INTEGER,
			rows INTEGER,
			remote_cwd TEXT,
			scrollback_path TEXT,
			suspended_at INTEGER
		)`,
		`CREATE TABLE IF NOT EXISTS app_meta (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return err
		}
	}

	if err := s.dedupeConnectionKeys(); err != nil {
		return err
	}
	if err := s.ensureSessionsDialColumns(); err != nil {
		return err
	}
	return s.ensureConnectionKeyIndex()
}

// Close closes the underlying database connection.
// Returns the error from sql.DB.Close, if any.
func (s *Store) Close() error {
	return s.db.Close()
}
