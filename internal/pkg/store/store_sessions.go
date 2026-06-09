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
	"time"
)

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
			auth_expires_at, tmux_enabled, tmux_session, login_destination, proxy_jump, proxy_command,
			created_at, updated_at, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			alias=excluded.alias, host=excluded.host, port=excluded.port,
			username=excluded.username, identity_path=excluded.identity_path,
			auth_method=excluded.auth_method, auth_expires_at=excluded.auth_expires_at,
			tmux_enabled=excluded.tmux_enabled, tmux_session=excluded.tmux_session,
			login_destination=excluded.login_destination, proxy_jump=excluded.proxy_jump,
			proxy_command=excluded.proxy_command,
			updated_at=excluded.updated_at, status=excluded.status
	`, sess.ID, nullStr(sess.Alias), sess.Host, sess.Port, sess.Username,
		nullStr(sess.IdentityPath), sess.AuthMethod, sess.AuthExpiresAt, tmux,
		nullStr(sess.TmuxSession), nullStr(sess.LoginDestination), nullStr(sess.ProxyJump),
		nullStr(sess.ProxyCommand), sess.CreatedAt, sess.UpdatedAt, sess.Status)
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

// FindSessionByKey returns the most recently updated session for username, host, port, and alias.
// port should already be normalized (for example 22 when unset). Empty alias matches rows with
// NULL or empty alias. Returns ErrSessionNotFound when no row matches, or a database error.
func (s *Store) FindSessionByKey(username, host string, port int, alias string) (*Session, error) {
	if alias == "" {
		row := s.db.QueryRow(`
			SELECT `+sessionsSelectCols+`
			FROM sessions
			WHERE username = ? AND host = ? AND port = ?
			  AND (alias IS NULL OR alias = '')
			ORDER BY updated_at DESC
			LIMIT 1`, username, host, port)
		return scanSession(row)
	}

	row := s.db.QueryRow(`
		SELECT `+sessionsSelectCols+`
		FROM sessions
		WHERE username = ? AND host = ? AND port = ? AND alias = ?
		ORDER BY updated_at DESC
		LIMIT 1`, username, host, port, alias)
	return scanSession(row)
}

func (s *Store) getSessionWhere(clause string, arg string) (*Session, error) {
	row := s.db.QueryRow(`
		SELECT `+sessionsSelectCols+`
		FROM sessions WHERE `+clause, arg)
	return scanSession(row)
}

// ListSessions returns all saved sessions, newest updated_at first.
// Returns an empty slice when none exist, or a database error from Query/Scan.
func (s *Store) ListSessions() ([]*Session, error) {
	rows, err := s.db.Query(`
		SELECT ` + sessionsSelectCols + `
		FROM sessions ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Session
	for rows.Next() {
		var sess Session
		var alias, identity, tmuxName, loginDest, proxyJump, proxyCmd sql.NullString
		var tmuxEnabled int
		if err := rows.Scan(
			&sess.ID, &alias, &sess.Host, &sess.Port, &sess.Username,
			&identity, &sess.AuthMethod, &sess.AuthExpiresAt, &tmuxEnabled,
			&tmuxName, &loginDest, &proxyJump, &proxyCmd,
			&sess.CreatedAt, &sess.UpdatedAt, &sess.Status,
		); err != nil {
			return nil, err
		}

		sess.Alias = alias.String
		sess.IdentityPath = identity.String
		sess.TmuxSession = tmuxName.String
		sess.TmuxEnabled = tmuxEnabled == 1
		sess.LoginDestination = loginDest.String
		sess.ProxyJump = proxyJump.String
		sess.ProxyCommand = proxyCmd.String
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
