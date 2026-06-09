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
	"path/filepath"
)

func buildDSN(path string) string {
	if path == sqliteMemoryPath {
		return "file::memory:?mode=memory&_pragma=foreign_keys(1)"
	}

	u := url.URL{
		Scheme:   "file",
		Path:     filepath.ToSlash(path),
		RawQuery: "_pragma=foreign_keys(1)",
	}
	return u.String()
}

func scanSession(row *sql.Row) (*Session, error) {
	var sess Session
	var alias, identity, tmuxName, loginDest, proxyJump, proxyCmd sql.NullString
	var tmuxEnabled int
	err := row.Scan(
		&sess.ID, &alias, &sess.Host, &sess.Port, &sess.Username,
		&identity, &sess.AuthMethod, &sess.AuthExpiresAt, &tmuxEnabled,
		&tmuxName, &loginDest, &proxyJump, &proxyCmd,
		&sess.CreatedAt, &sess.UpdatedAt, &sess.Status,
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
	sess.LoginDestination = loginDest.String
	sess.ProxyJump = proxyJump.String
	sess.ProxyCommand = proxyCmd.String
	return &sess, nil
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
