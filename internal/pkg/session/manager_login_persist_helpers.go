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
	"time"

	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/store"
	"github.com/os-artificer/AiT/internal/pkg/tmux"
)

type loginPersistPlan struct {
	persist  bool
	id       string
	existing *store.Session
}

// planLoginPersistence decides whether login should upsert a session row and which id to use.
// With --save, --alias, or --tmux, resolvePersistSessionID matches the full connection key (username, host, port, alias).
// Without flags, persist only when an empty-alias row for the connection already exists.
func (m *Manager) planLoginPersistence(opts LoginOpts, newID string) (loginPersistPlan, error) {
	if shouldPersistLogin(opts) {
		id, err := m.resolvePersistSessionID(opts, newID)
		if err != nil {
			return loginPersistPlan{}, err
		}
		plan := loginPersistPlan{persist: true, id: id}
		if existing, err := m.Store.GetSession(id); err == nil {
			plan.existing = existing
		}
		return plan, nil
	}

	existing, err := m.Store.FindSessionByKey(opts.User, opts.Host, normalizePort(opts.Port), "")
	if err != nil {
		if errors.Is(err, store.ErrSessionNotFound) {
			return loginPersistPlan{persist: false, id: newID}, nil
		}
		return loginPersistPlan{}, err
	}

	return loginPersistPlan{
		persist:  true,
		id:       existing.ID,
		existing: existing,
	}, nil
}

// buildSessionForLogin constructs a store.Session for upsert, preserving alias and dial metadata when updating.
// TmuxEnabled always follows opts.Tmux for this login so bookmark vs tmux mode can be switched per attempt.
func buildSessionForLogin(opts LoginOpts, sshOpts sshclient.Options, plan loginPersistPlan, expires int64) *store.Session {
	authMethod := sshclient.DetectAuthMethod(sshOpts)
	alias := opts.Alias
	tmuxName := tmux.SessionName(plan.id)
	tmuxEnabled := opts.Tmux
	loginDest := opts.LoginDestination
	proxyJump := formatProxyJump(opts.ProxyJump)
	proxyCmd := opts.ProxyCommand
	if plan.existing != nil {
		if alias == "" {
			alias = plan.existing.Alias
		}
		if plan.existing.TmuxSession != "" {
			tmuxName = plan.existing.TmuxSession
		}
		if loginDest == "" {
			loginDest = plan.existing.LoginDestination
		}
		if proxyJump == "" {
			proxyJump = plan.existing.ProxyJump
		}
		if proxyCmd == "" {
			proxyCmd = plan.existing.ProxyCommand
		}
	}

	return &store.Session{
		ID: plan.id, Alias: alias, Host: opts.Host, Port: normalizePort(opts.Port),
		Username: opts.User, IdentityPath: opts.IdentityPath,
		AuthMethod: authMethod, AuthExpiresAt: expires,
		TmuxEnabled: tmuxEnabled, TmuxSession: tmuxName,
		LoginDestination: loginDest, ProxyJump: proxyJump, ProxyCommand: proxyCmd,
		Status: store.StatusSuspended,
	}
}

func loginAuthExpiry(m *Manager) int64 {
	ttl, _ := m.Config.AuthTTL()
	return time.Now().Add(ttl).Unix()
}
