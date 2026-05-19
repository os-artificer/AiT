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
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/platform"
	_ "modernc.org/sqlite"
)

const (
	StatusActive    = "active"
	StatusSuspended = "suspended"

	AuthMethodKey      = "key"
	AuthMethodPassword = "password"
	AuthMethodAgent    = "agent"

	MetaCurrentSession = "current_session_id"
)

// ErrSessionNotFound is returned when a session lookup misses.
var ErrSessionNotFound = errors.New("session not found")

// ErrSessionAmbiguous is returned when id and alias both match different rows.
var ErrSessionAmbiguous = errors.New("session id/alias lookup is ambiguous")

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

// Open opens or creates the database at the default path.
func Open() (*Store, error) {
	path, err := platform.DBPath()
	if err != nil {
		return nil, err
	}
	return OpenPath(path)
}

func buildDSN(path string) string {
	if path == ":memory:" {
		return "file::memory:?mode=memory&_pragma=foreign_keys(1)"
	}
	u := url.URL{
		Scheme:   "file",
		Path:     filepath.ToSlash(path),
		RawQuery: "_pragma=foreign_keys(1)",
	}
	return u.String()
}

// OpenPath opens SQLite at path (":memory:" for tests).
func OpenPath(path string) (*Store, error) {
	if path != ":memory:" {
		dir := filepath.Dir(path)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0o700); err != nil {
				return nil, err
			}
		}
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
	if path != ":memory:" {
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

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// UpsertSession inserts or updates a session.
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

// GetSession loads a session by id or alias (id match takes precedence).
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

func scanSession(row *sql.Row) (*Session, error) {
	var sess Session
	var alias, identity, tmuxName sql.NullString
	var tmuxEnabled int
	err := row.Scan(
		&sess.ID, &alias, &sess.Host, &sess.Port, &sess.Username,
		&identity, &sess.AuthMethod, &sess.AuthExpiresAt, &tmuxEnabled,
		&tmuxName, &sess.CreatedAt, &sess.UpdatedAt, &sess.Status,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	sess.Alias = alias.String
	sess.IdentityPath = identity.String
	sess.TmuxSession = tmuxName.String
	sess.TmuxEnabled = tmuxEnabled == 1
	return &sess, nil
}

// ListSessions returns all sessions ordered by updated_at desc.
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

// DeleteSession removes one session (and its state) in a single transaction.
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

// DeleteSessionState removes only the session_state row for id, if any.
func (s *Store) DeleteSessionState(id string) error {
	_, err := s.db.Exec(`DELETE FROM session_state WHERE session_id = ?`, id)
	return err
}

// DeleteAllSessions removes all sessions and state in a single transaction.
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

// SaveSessionState upserts suspend snapshot.
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

// GetSessionState loads suspend snapshot.
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

// SetMeta sets an app_meta key.
func (s *Store) SetMeta(key, value string) error {
	_, err := s.db.Exec(`
		INSERT INTO app_meta (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value
	`, key, value)
	return err
}

// GetMeta gets an app_meta key.
func (s *Store) GetMeta(key string) (string, error) {
	var v string
	err := s.db.QueryRow(`SELECT value FROM app_meta WHERE key = ?`, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return v, err
}

// UpdateStatus updates session status.
func (s *Store) UpdateStatus(id, status string) error {
	_, err := s.db.Exec(`UPDATE sessions SET status=?, updated_at=? WHERE id=?`,
		status, time.Now().Unix(), id)
	return err
}

// RefreshAuthExpiry updates auth_expires_at.
func (s *Store) RefreshAuthExpiry(id string, expiresAt int64) error {
	_, err := s.db.Exec(`UPDATE sessions SET auth_expires_at=?, updated_at=? WHERE id=?`,
		expiresAt, time.Now().Unix(), id)
	return err
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
