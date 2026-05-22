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
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/os-artificer/AiT/internal/pkg/creds"
	"github.com/os-artificer/AiT/internal/pkg/pathsafe"
	"github.com/os-artificer/AiT/internal/pkg/platform"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/store"
	"github.com/os-artificer/AiT/internal/pkg/terminal"

	"golang.org/x/term"
)

func normalizePort(port int) int {
	if port == 0 {
		return 22
	}
	return port
}

func shouldPromptSSHPassword(sshOpts sshclient.Options, cliPassword string) bool {
	if sshOpts.BatchMode || cliPassword != "" {
		return false
	}
	if sshOpts.IdentityExplicit {
		return false
	}
	return true
}

func isSSHAuthFailure(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "knownhosts") ||
		strings.Contains(msg, "host key") ||
		strings.Contains(msg, "hostkey") ||
		strings.Contains(msg, "key is unknown") ||
		strings.Contains(msg, "host key verification failed") ||
		strings.Contains(msg, "possible mitm") {
		return false
	}

	return strings.Contains(msg, "unable to authenticate") ||
		strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "no supported authentication") ||
		strings.Contains(msg, "no authentication method")
}

func promptPassword(msg string) (string, error) {
	fmt.Fprint(os.Stderr, msg)
	b, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func quotePath(p string) string {
	return pathsafe.QuoteShellSingle(p)
}

func sshOptionsFromSession(sess *store.Session, password string) sshclient.Options {
	return sshclient.Options{
		Host:         sess.Host,
		Port:         normalizePort(sess.Port),
		User:         sess.Username,
		IdentityPath: sess.IdentityPath,
		Password:     password,
		UseAgent:     true,
	}
}

func sessionHostLabel(sess *store.Session) string {
	if sess.Username != "" {
		return sess.Username + "@" + sess.Host
	}
	return sess.Host
}

func (m *Manager) purgeSessionArtifacts(sessID string) {
	st, _ := m.Store.GetSessionState(sessID)
	if st != nil && st.ScrollbackPath != "" {
		m.removeScrollbackFile(st.ScrollbackPath)
	}
	_ = creds.DeleteSessionUnlock(sessID)
}

func (m *Manager) replayScrollbackIfPresent(st *store.SessionState) {
	if st == nil || st.ScrollbackPath == "" {
		return
	}

	dir, err := platform.ScrollbackDir()
	if err != nil {
		return
	}

	path, err := terminal.ResolveScrollbackPath(dir, st.ScrollbackPath)
	if err != nil {
		return
	}

	data, err := terminal.LoadFromFile(path)
	if err != nil {
		return
	}

	terminal.Replay(data)
}

func (m *Manager) reconcileCurrentOnSwitch(current string, force bool) error {
	if current == "" {
		return nil
	}

	cur, gerr := m.Store.GetSession(current)
	if errors.Is(gerr, store.ErrSessionNotFound) {
		// Stale meta pointing at a deleted session — drop it so the next
		// Switch/Load doesn't re-run this lookup.
		_ = m.Store.SetMeta(store.MetaCurrentSession, "")
		return nil
	}

	if gerr != nil || cur.Status != store.StatusActive {
		return nil
	}

	return m.warnOrBlockActiveSwitch(cur, force)
}

func (m *Manager) warnOrBlockActiveSwitch(cur *store.Session, force bool) error {
	if !force {
		return fmt.Errorf(
			"current session %s is still active; suspend it first inside SSH with !ait suspend, or use --force",
			cur.AliasOrID(),
		)
	}

	fmt.Fprintln(os.Stderr, "warning: switching without suspend; marking current session suspended")
	_ = m.Store.UpdateStatus(cur.ID, store.StatusSuspended)
	return nil
}
