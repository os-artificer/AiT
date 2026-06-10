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
	"context"

	"github.com/os-artificer/AiT/internal/pkg/creds"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/ssh/openssh"
	"github.com/os-artificer/AiT/internal/pkg/store"
	"github.com/os-artificer/AiT/internal/pkg/tmux"

	"github.com/google/uuid"
)

// Login dials the remote host and runs an interactive shell or a one-shot remote command.
// A session profile is persisted when opts.Save, opts.Alias, or opts.Tmux is set, or when the same
// username+host+port already exists in the store (existing rows are updated and TTL refreshed).
// ctx cancels dial and interactive I/O. opts must include Host; RemoteCommand and Tmux cannot be combined.
// Returns dial/auth errors, store errors, errhint for invalid flag combinations, or sshclient.RemoteExitError for one-shot exits.
func (m *Manager) Login(ctx context.Context, opts LoginOpts) error {
	opts.Port = normalizePort(opts.Port)
	if len(opts.RemoteCommand) > 0 && opts.Tmux {
		return openssh.TmuxWithRemoteCommandErr()
	}
	if opts.Tmux && !stdinIsTerminal() {
		return TmuxRequiresTTYErr()
	}

	sshOpts := sshclient.Options{
		Host: opts.Host, Port: opts.Port, User: opts.User,
		LoginDestination: opts.LoginDestination,
		IdentityPath:     opts.IdentityPath, IdentityExplicit: opts.IdentityExplicit,
		Password: opts.Password, UseAgent: true,
		ProxyJump: opts.ProxyJump, ProxyCommand: opts.ProxyCommand,
		KnownHostsPath: opts.KnownHostsPath, StrictHostKeyChecking: opts.StrictHostKeyChecking,
		ConnectTimeout: opts.ConnectTimeout, BatchMode: opts.BatchMode,
	}

	client, err := m.dialLogin(sshOpts, opts.Password, opts.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	if len(opts.RemoteCommand) > 0 {
		return m.loginRunOnce(ctx, client, opts)
	}
	return m.loginInteractive(ctx, client, opts, sshOpts)
}

func (m *Manager) loginRunOnce(ctx context.Context, client *sshclient.Client, opts LoginOpts) error {
	plan, err := m.planLoginPersistence(opts, uuid.New().String())
	if err != nil {
		return err
	}
	if plan.persist {
		if err := m.persistLoginSession(client, opts, plan); err != nil {
			return err
		}
	}

	cmdline := sshclient.BuildRemoteCommandLine(opts.RemoteCommand)
	return client.RunRemoteCommand(ctx, cmdline)
}

func (m *Manager) loginInteractive(ctx context.Context, client *sshclient.Client, opts LoginOpts, sshOpts sshclient.Options) error {
	newID := uuid.New().String()
	plan, err := m.planLoginPersistence(opts, newID)
	if err != nil {
		return err
	}

	sess := buildSessionForLogin(opts, sshOpts, plan, loginAuthExpiry(m))
	newRow := plan.existing == nil

	if plan.persist {
		if err := m.persistInteractiveLogin(sess, opts); err != nil {
			return err
		}
	}

	if sess.TmuxEnabled {
		if err := tmux.Ensure(client.Conn(), sess.TmuxSession); err != nil {
			m.cleanupOnTmuxFailure(plan.persist, newRow, sess.ID)
			return tmuxEnsureErr(sess, err)
		}
	}

	initialCmds := m.remotePromptInitialCommands(sess.Username, sess.Host)
	err = m.runInteractive(ctx, client, sess, initialCmds, plan.persist, "", "")
	if err != nil && plan.persist {
		if newRow {
			m.deleteLoginSession(sess.ID)
		} else {
			m.clearSessionOnFailure(sess.ID, true)
		}
	}
	return err
}

// resolveLoginSessionID returns an existing session id for opts.Alias when present; otherwise newID.
// When connection parameters changed, clears stale scrollback for that id.
func (m *Manager) resolveLoginSessionID(opts LoginOpts, newID string) string {
	existing, err := m.Store.GetSession(opts.Alias)
	if err != nil {
		return newID
	}

	id := existing.ID
	if existing.Host != opts.Host ||
		normalizePort(existing.Port) != opts.Port ||
		(opts.User != "" && existing.Username != opts.User) {
		if st, _ := m.Store.GetSessionState(id); st != nil && st.ScrollbackPath != "" {
			m.removeScrollbackFile(st.ScrollbackPath)
		}
		_ = m.Store.DeleteSessionState(id)
	}
	return id
}

// persistInteractiveLogin upserts sess and records meta/keyring for a saved login.
func (m *Manager) persistInteractiveLogin(sess *store.Session, opts LoginOpts) error {
	if err := m.Store.UpsertSession(sess); err != nil {
		return err
	}

	if opts.Password != "" {
		_ = creds.SetSessionUnlock(sess.ID, opts.Password)
	}
	return m.Store.SetMeta(store.MetaCurrentSession, sess.ID)
}

func (m *Manager) persistLoginSession(client *sshclient.Client, opts LoginOpts, plan loginPersistPlan) error {
	_ = client
	if !plan.persist {
		return nil
	}

	sshOpts := sshclient.Options{
		Host: opts.Host, Port: opts.Port, User: opts.User,
		IdentityPath: opts.IdentityPath, IdentityExplicit: opts.IdentityExplicit,
		Password: opts.Password,
	}
	sess := buildSessionForLogin(opts, sshOpts, plan, loginAuthExpiry(m))
	if err := m.Store.UpsertSession(sess); err != nil {
		return err
	}

	if opts.Password != "" {
		_ = creds.SetSessionUnlock(sess.ID, opts.Password)
	}
	return m.Store.SetMeta(store.MetaCurrentSession, sess.ID)
}
