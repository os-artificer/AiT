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
	"time"

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
	CreatedAt     int64
	UpdatedAt     int64
	Status        string
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
	return nil
}

// Close closes the underlying database connection.
// Returns the error from sql.DB.Close, if any.
func (s *Store) Close() error {
	return s.db.Close()
}

// UpsertSession inserts a new session or updates an existing row by id.
// sess must be non-nil with a non-empty ID; CreatedAt is set on first insert and UpdatedAt on every call.
// Returns a database error from Exec, if any.
func (s *Store) UpsertSession(sess *Session) error {
	now := time.Now().Unix()
	if sess.CreatedAt == 0 {
		sess.CreatedAt = now
	}
	sess.UpdatedAt = now

	tmux := 0
	if sess.TmuxEnabled {
		tmux = 1
	}

	_, err := s.db.Exec(`
		INSERT INTO sessions (
			id, alias, host, port, username, identity_path, auth_method,
			auth_expires_at, tmux_enabled, tmux_session, created_at, updated_at, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			alias=excluded.alias, host=excluded.host, port=excluded.port,
			username=excluded.username, identity_path=excluded.identity_path,
			auth_method=excluded.auth_method, auth_expires_at=excluded.auth_expires_at,
			tmux_enabled=excluded.tmux_enabled, tmux_session=excluded.tmux_session,
			updated_at=excluded.updated_at, status=excluded.status
	`, sess.ID, nullStr(sess.Alias), sess.Host, sess.Port, sess.Username,
		nullStr(sess.IdentityPath), sess.AuthMethod, sess.AuthExpiresAt, tmux,
		nullStr(sess.TmuxSession), sess.CreatedAt, sess.UpdatedAt, sess.Status)
	return err
}

// GetSession loads a session by UUID id or unique alias. Id lookup takes precedence over alias.
// Returns the session, ErrSessionNotFound if neither matches, ErrSessionAmbiguous if id and alias
// refer to different rows, or a database error.
func (s *Store) GetSession(idOrAlias string) (*Session, error) {
	byID, errID := s.getSessionWhere(`id = ?`, idOrAlias)
	if errID == nil {
		byAlias, errAlias := s.getSessionWhere(`alias = ?`, idOrAlias)
		if errAlias == nil && byAlias.ID != byID.ID {
			return nil, ErrSessionAmbiguous
		}
		return byID, nil
	}

	if !errors.Is(errID, ErrSessionNotFound) {
		return nil, errID
	}
	return s.getSessionWhere(`alias = ?`, idOrAlias)
}

func (s *Store) getSessionWhere(clause string, arg string) (*Session, error) {
	row := s.db.QueryRow(`
		SELECT id, alias, host, port, username, identity_path, auth_method,
			auth_expires_at, tmux_enabled, tmux_session, created_at, updated_at, status
		FROM sessions WHERE `+clause, arg)
	return scanSession(row)
}

// ListSessions returns all saved sessions, newest updated_at first.
// Returns an empty slice when none exist, or a database error from Query/Scan.
func (s *Store) ListSessions() ([]*Session, error) {
	rows, err := s.db.Query(`
		SELECT id, alias, host, port, username, identity_path, auth_method,
			auth_expires_at, tmux_enabled, tmux_session, created_at, updated_at, status
		FROM sessions ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Session
	for rows.Next() {
		var sess Session
		var alias, identity, tmuxName sql.NullString
		var tmuxEnabled int
		if err := rows.Scan(
			&sess.ID, &alias, &sess.Host, &sess.Port, &sess.Username,
			&identity, &sess.AuthMethod, &sess.AuthExpiresAt, &tmuxEnabled,
			&tmuxName, &sess.CreatedAt, &sess.UpdatedAt, &sess.Status,
		); err != nil {
			return nil, err
		}

		sess.Alias = alias.String
		sess.IdentityPath = identity.String
		sess.TmuxSession = tmuxName.String
		sess.TmuxEnabled = tmuxEnabled == 1
		out = append(out, &sess)
	}
	return out, rows.Err()
}

// DeleteSession removes a session and its session_state row in one transaction.
// idOrAlias is resolved via GetSession. Returns ErrSessionNotFound, GetSession errors,
// or a database/transaction error.
func (s *Store) DeleteSession(idOrAlias string) error {
	sess, err := s.GetSession(idOrAlias)
	if err != nil {
		return err
	}
	id := sess.ID

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM session_state WHERE session_id = ?`, id); err != nil {
		return err
	}
	res, err := tx.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrSessionNotFound
	}

	return tx.Commit()
}

// DeleteSessionState removes the suspend snapshot for session id, if present.
// id is the session UUID, not an alias. Returns a database error from Exec, if any.
func (s *Store) DeleteSessionState(id string) error {
	_, err := s.db.Exec(`DELETE FROM session_state WHERE session_id = ?`, id)
	return err
}

// DeleteAllSessions removes every session and session_state row in one transaction.
// Returns a database or transaction error; does not return ErrSessionNotFound.
func (s *Store) DeleteAllSessions() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM session_state`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM sessions`); err != nil {
		return err
	}

	return tx.Commit()
}

// SaveSessionState upserts the suspend snapshot for st.SessionID.
// st must be non-nil with a non-empty SessionID. Returns a database error from Exec, if any.
func (s *Store) SaveSessionState(st *SessionState) error {
	_, err := s.db.Exec(`
		INSERT INTO session_state (session_id, cols, rows, remote_cwd, scrollback_path, suspended_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(session_id) DO UPDATE SET
			cols=excluded.cols, rows=excluded.rows, remote_cwd=excluded.remote_cwd,
			scrollback_path=excluded.scrollback_path, suspended_at=excluded.suspended_at
	`, st.SessionID, st.Cols, st.Rows, nullStr(st.RemoteCwd), nullStr(st.ScrollbackPath), st.SuspendedAt)
	return err
}

// GetSessionState loads the suspend snapshot for sessionID.
// Returns (nil, nil) when no snapshot exists, the state, or a database error other than no rows.
func (s *Store) GetSessionState(sessionID string) (*SessionState, error) {
	row := s.db.QueryRow(`
		SELECT session_id, cols, rows, remote_cwd, scrollback_path, suspended_at
		FROM session_state WHERE session_id = ?
	`, sessionID)
	var st SessionState
	var cwd, path sql.NullString
	var cols, rows sql.NullInt64
	err := row.Scan(&st.SessionID, &cols, &rows, &cwd, &path, &st.SuspendedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if cols.Valid {
		st.Cols = int(cols.Int64)
	}
	if rows.Valid {
		st.Rows = int(rows.Int64)
	}
	st.RemoteCwd = cwd.String
	st.ScrollbackPath = path.String
	return &st, nil
}

// SetMeta sets or replaces an application metadata key (e.g. MetaCurrentSession).
// Returns a database error from Exec, if any.
func (s *Store) SetMeta(key, value string) error {
	_, err := s.db.Exec(`
		INSERT INTO app_meta (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value
	`, key, value)
	return err
}

// GetMeta returns the value for key, or "" when the key is absent.
// Returns a database error from QueryRow other than sql.ErrNoRows.
func (s *Store) GetMeta(key string) (string, error) {
	var v string
	err := s.db.QueryRow(`SELECT value FROM app_meta WHERE key = ?`, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return v, err
}

// UpdateStatus sets status and updated_at for session id.
// id is the session UUID. Returns a database error from Exec, if any.
func (s *Store) UpdateStatus(id, status string) error {
	_, err := s.db.Exec(`UPDATE sessions SET status=?, updated_at=? WHERE id=?`,
		status, time.Now().Unix(), id)
	return err
}

// RefreshAuthExpiry sets auth_expires_at and updated_at for session id.
// expiresAt is Unix seconds. Returns a database error from Exec, if any.
func (s *Store) RefreshAuthExpiry(id string, expiresAt int64) error {
	_, err := s.db.Exec(`UPDATE sessions SET auth_expires_at=?, updated_at=? WHERE id=?`,
		expiresAt, time.Now().Unix(), id)
	return err
}
