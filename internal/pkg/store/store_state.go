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
)

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

// DeleteSessionState removes the suspend snapshot for session id, if present.
// id is the session UUID, not an alias. Returns a database error from Exec, if any.
func (s *Store) DeleteSessionState(id string) error {
	_, err := s.db.Exec(`DELETE FROM session_state WHERE session_id = ?`, id)
	return err
}
