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
)

// resolvePersistSessionID returns an existing session id when opts match a saved alias or connection
// key (username, host, port, alias); otherwise newID. Alias lookup runs first when opts.Alias is set.
// Returns store.ErrSessionNotFound as (newID, nil). Other store errors are returned unchanged.
func (m *Manager) resolvePersistSessionID(opts LoginOpts, newID string) (string, error) {
	if opts.Alias != "" {
		if _, err := m.Store.GetSession(opts.Alias); err == nil {
			return m.resolveLoginSessionID(opts, newID), nil
		}
	}

	existing, err := m.Store.FindSessionByKey(opts.User, opts.Host, normalizePort(opts.Port), opts.Alias)
	if err != nil {
		if errors.Is(err, store.ErrSessionNotFound) {
			return newID, nil
		}
		return "", err
	}

	return existing.ID, nil
}
