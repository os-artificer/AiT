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

package session

import (
	"errors"

	"github.com/os-artificer/AiT/internal/pkg/store"
	"github.com/os-artificer/AiT/internal/pkg/tmux"
)

// ensureSessionPersisted upserts the current interactive session when not yet in SQLite.
// Returns the session id to use for suspend (may differ after D1 dedupe merge).
func (m *Manager) ensureSessionPersisted(sess *store.Session) (string, error) {
	if m == nil || m.Store == nil || sess == nil {
		return "", errors.New("internal: ensureSessionPersisted: nil manager or session")
	}

	if _, err := m.Store.GetSession(sess.ID); err == nil {
		_ = m.Store.SetMeta(store.MetaCurrentSession, sess.ID)
		return sess.ID, nil
	} else if !errors.Is(err, store.ErrSessionNotFound) {
		return "", err
	}

	id := sess.ID
	var existing *store.Session

	found, err := m.Store.FindSessionByKey(sess.Username, sess.Host, normalizePort(sess.Port), "")
	if err == nil {
		id = found.ID
		existing = found
	} else if !errors.Is(err, store.ErrSessionNotFound) {
		return "", err
	}

	toSave := mergeSessionForAutoPersist(sess, existing, id)
	toSave.Status = store.StatusActive
	toSave.AuthExpiresAt = loginAuthExpiry(m)

	if err := m.Store.UpsertSession(toSave); err != nil {
		return "", err
	}

	_ = m.Store.SetMeta(store.MetaCurrentSession, id)
	return id, nil
}

func mergeSessionForAutoPersist(current, existing *store.Session, id string) *store.Session {
	out := *current
	out.ID = id

	if existing != nil {
		if out.Alias == "" {
			out.Alias = existing.Alias
		}
		if existing.TmuxSession != "" {
			out.TmuxSession = existing.TmuxSession
		}
		if out.LoginDestination == "" {
			out.LoginDestination = existing.LoginDestination
		}
		if out.ProxyJump == "" {
			out.ProxyJump = existing.ProxyJump
		}
		if out.ProxyCommand == "" {
			out.ProxyCommand = existing.ProxyCommand
		}
	}

	if out.TmuxSession == "" {
		out.TmuxSession = tmux.SessionName(id)
	}

	return &out
}
