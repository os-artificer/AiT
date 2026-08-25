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
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/creds"
	"github.com/os-artificer/AiT/internal/pkg/pathsafe"
	"github.com/os-artificer/AiT/internal/pkg/platform"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/store"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
	"github.com/os-artificer/AiT/internal/pkg/tmux"
)

// Load reconnects to a saved session by id or alias, optionally enabling tmux and supplying a password override.
// ctx cancels dial and the interactive session. password, when non-empty, is cached in the keyring after a successful dial.
// Returns store.ErrSessionNotFound, resolve/dial errors, tmux.Ensure errors, or errors from the interactive loop.
func (m *Manager) Load(ctx context.Context, idOrAlias string, useTmux bool, password string) error {
	sess, err := m.Store.GetSession(idOrAlias)
	if err != nil {
		return wrapLoadNotFound(idOrAlias, err)
	}
	if useTmux {
		sess.TmuxEnabled = true
		if err := m.Store.UpsertSession(sess); err != nil {
			return err
		}
	}

	pass, prompted, err := m.resolvePassword(sess, password)
	if err != nil {
		return err
	}

	sshOpts, err := m.resolveSSHOptionsFromSession(sess, pass)
	if err != nil {
		return wrapLoadDialErr(sess, err)
	}
	client, err := m.dialLogin(sshOpts, pass, sessionHostLabel(sess))
	if err != nil {
		return err
	}

	// Only cache freshly-entered passwords (CLI flag override OR interactive
	// prompt) after a successful dial. Reusing a keyring-cached password
	// must not rewrite the same value. TTL is refreshed only when the user
	// supplied a new password so periodic load/switch cannot extend auth forever.
	if password != "" || prompted {
		_ = creds.SetSessionUnlock(sess.ID, pass)
		ttl, _ := m.Config.AuthTTL()
		_ = m.Store.RefreshAuthExpiry(sess.ID, time.Now().Add(ttl).Unix())
	}

	st, _ := m.Store.GetSessionState(sess.ID)
	replayScrollback := m.scrollbackDataForReplay(sess, st)
	initialRemoteCwd := ""
	if st != nil {
		initialRemoteCwd = st.RemoteCwd
	}

	var initialCmds []string
	initialCmds = append(m.remotePromptInitialCommands(sess.Username, sess.Host), initialCmds...)
	if st != nil && st.RemoteCwd != "" && pathsafe.RemotePath(st.RemoteCwd) && !sess.TmuxEnabled {
		initialCmds = append(initialCmds, "cd "+pathsafe.QuoteShellSingle(st.RemoteCwd))
	}

	if sess.TmuxEnabled {
		name := sess.TmuxSession
		if name == "" {
			name = tmux.SessionName(sess.ID)
		}
		if err := tmux.Ensure(client.Conn(), name); err != nil {
			client.Close()
			return tmuxEnsureErr(sess, err)
		}
	}

	_ = m.Store.SetMeta(store.MetaCurrentSession, sess.ID)
	return m.runInteractive(ctx, client, sess, initialCmds, true, replayScrollback, initialRemoteCwd)
}

// SSHOptionsForSession builds sshclient.Options for a saved session (e.g. aissh get, aissh put).
// idOrAlias is resolved via the store; passwordOverride bypasses the keyring when non-empty.
// Returns sshclient.Options and store lookup or password-resolution errors.
func (m *Manager) SSHOptionsForSession(idOrAlias, passwordOverride string) (sshclient.Options, error) {
	sess, err := m.Store.GetSession(idOrAlias)
	if err != nil {
		return sshclient.Options{}, err
	}

	pass, _, err := m.resolvePassword(sess, passwordOverride)
	if err != nil {
		return sshclient.Options{}, err
	}

	return m.resolveSSHOptionsFromSession(sess, pass)
}

// resolvePassword returns the password to use for the next dial and whether it
// came from a user-provided override (so the caller can decide when to persist
// it). It does NOT write to the keyring; callers must do that only after a
// successful dial.
func (m *Manager) resolvePassword(sess *store.Session, override string) (string, bool, error) {
	if override != "" {
		return override, true, nil
	}

	if time.Now().Unix() <= sess.AuthExpiresAt {
		if p, err := creds.GetSessionUnlock(sess.ID); err == nil && p != "" {
			return p, false, nil
		}
	}

	if sess.AuthMethod == store.AuthMethodPassword {
		p, err := promptPassword(fmt.Sprintf(promptPasswordExpiredFmt, sess.Username, sess.Host))
		if err != nil {
			return "", false, err
		}
		return p, true, nil
	}

	if sess.IdentityPath != "" {
		p, err := promptPassword(fmt.Sprintf(promptPassphraseExpiredFmt, sess.IdentityPath))
		if err != nil {
			return "", false, err
		}
		return p, true, nil
	}
	return "", false, nil
}

func (m *Manager) runInteractive(
	ctx context.Context,
	client *sshclient.Client,
	sess *store.Session,
	initialCmds []string,
	persisted bool,
	replayScrollback string,
	initialRemoteCwd string,
) error {
	defer client.Close()
	tmuxName := sess.TmuxSession
	if tmuxName == "" {
		tmuxName = tmux.SessionName(sess.ID)
	}
	if persisted {
		if err := m.suspendOtherActiveSessions(sess.ID); err != nil {
			return err
		}
		if err := m.Store.UpdateStatus(sess.ID, store.StatusActive); err != nil {
			return err
		}
	}

	err := client.RunInteractive(ctx, sshclient.InteractiveOpts{
		SessionID:               sess.ID,
		TmuxEnabled:             sess.TmuxEnabled,
		TmuxName:                tmuxName,
		InitialCommands:         initialCmds,
		InitialRemoteCwd:        initialRemoteCwd,
		ReplayScrollback:        replayScrollback,
		RemotePromptDirAbsolute: m.Config != nil && m.Config.RemotePromptPathAbsolute(),
		OnMetaCLI:               m.MetaCLI,
		AliasLookup:             m.Store,
		OnSuspend: func(scroll *terminal.Buffer, cwd string, cols, rows int) error {
			id, err := m.ensureSessionPersisted(sess)
			if err != nil {
				return err
			}
			return m.suspend(id, scroll, cwd, cols, rows)
		},
	})

	var chain *sshclient.ChainActionError
	if errors.As(err, &chain) {
		return m.runChainAction(ctx, chain)
	}

	if err != nil && persisted {
		m.clearSessionOnFailure(sess.ID, true)
	}
	return err
}

func (m *Manager) clearSessionOnFailure(id string, persisted bool) {
	if !persisted {
		return
	}

	_ = m.Store.UpdateStatus(id, store.StatusSuspended)
	cur, _ := m.Store.GetMeta(store.MetaCurrentSession)
	if cur == id {
		_ = m.Store.SetMeta(store.MetaCurrentSession, "")
	}
}

func (m *Manager) removeScrollbackFile(stored string) {
	dir, err := platform.ScrollbackDir()
	if err != nil {
		return
	}

	path, err := terminal.ResolveScrollbackPath(dir, stored)
	if err != nil {
		return
	}

	_ = os.Remove(path)
}
