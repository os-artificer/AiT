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

import "strings"

// ensureSessionsDialColumns adds login_destination, proxy_jump, and proxy_command for load re-resolve.
func (s *Store) ensureSessionsDialColumns() error {
	stmts := []string{
		`ALTER TABLE sessions ADD COLUMN login_destination TEXT`,
		`ALTER TABLE sessions ADD COLUMN proxy_jump TEXT`,
		`ALTER TABLE sessions ADD COLUMN proxy_command TEXT`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
			return err
		}
	}
	return nil
}

// dedupeConnectionKeys removes duplicate session rows that share the same
// (username, host, port, alias) connection key, keeping the row with the
// greatest updated_at (ties broken by id).
func (s *Store) dedupeConnectionKeys() error {
	rows, err := s.db.Query(`
		SELECT username, host, port, COALESCE(alias, ''), COUNT(*) AS n
		FROM sessions
		GROUP BY username, host, port, COALESCE(alias, '')
		HAVING n > 1`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type connKey struct {
		username, host, alias string
		port                  int
	}

	var keys []connKey
	for rows.Next() {
		var k connKey
		if err := rows.Scan(&k.username, &k.host, &k.port, &k.alias, new(int)); err != nil {
			return err
		}
		keys = append(keys, k)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, k := range keys {
		if err := s.deleteOlderConnectionDuplicates(k.username, k.host, k.port, k.alias); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) deleteOlderConnectionDuplicates(username, host string, port int, alias string) error {
	var keepID string
	err := s.db.QueryRow(`
		SELECT id FROM sessions
		WHERE username = ? AND host = ? AND port = ?
		  AND COALESCE(alias, '') = ?
		ORDER BY updated_at DESC, id DESC
		LIMIT 1`, username, host, port, alias).Scan(&keepID)
	if err != nil {
		return err
	}

	rows, err := s.db.Query(`
		SELECT id FROM sessions
		WHERE username = ? AND host = ? AND port = ?
		  AND COALESCE(alias, '') = ?
		  AND id != ?`, username, host, port, alias, keepID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var dropIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		dropIDs = append(dropIDs, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, id := range dropIDs {
		if err := s.DeleteSession(id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ensureConnectionKeyIndex() error {
	_, err := s.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_conn_key
		ON sessions(username, host, port, COALESCE(alias, ''))`)
	return err
}
