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
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/pathsafe"
	"github.com/os-artificer/AiT/internal/pkg/platform"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/ssh/client/meta"
	"github.com/os-artificer/AiT/internal/pkg/store"
	"github.com/os-artificer/AiT/internal/pkg/terminal"

	"golang.org/x/term"
)

// shouldPersistLogin reports whether login should upsert a session row (--save, --alias, or --tmux).
func shouldPersistLogin(opts LoginOpts) bool {
	return opts.Save || opts.Alias != "" || opts.Tmux
}

func normalizePort(port int) int {
	if port == 0 {
		return defaultSSHPort
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

// retryDialAfterPasswordPrompt retries DialChain once after an interactive password prompt.
// firstErr is returned unchanged when prompting is skipped or the user provides no password.
func retryDialAfterPasswordPrompt(sshOpts sshclient.Options, firstErr error, cliPassword, hostLabel string) (*sshclient.Client, error) {
	if firstErr == nil || !shouldPromptSSHPassword(sshOpts, cliPassword) || !isSSHAuthFailure(firstErr) {
		return nil, firstErr
	}

	user := sshOpts.User
	if user == "" {
		user = defaultSSHPromptUser
	}

	p, perr := promptPassword(fmt.Sprintf(promptPasswordLoginFmt, user, hostLabel))
	if perr != nil || p == "" {
		return nil, firstErr
	}

	sshOpts.Password = p
	return sshclient.DialChain(sshOpts)
}

func isSSHAuthFailure(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, sshHostKeyErrKnownHosts) ||
		strings.Contains(msg, sshHostKeyErrHostKey) ||
		strings.Contains(msg, sshHostKeyErrHostKeyCompact) ||
		strings.Contains(msg, sshHostKeyErrKeyUnknown) ||
		strings.Contains(msg, sshHostKeyErrVerificationFailed) ||
		strings.Contains(msg, sshHostKeyErrPossibleMITM) {
		return false
	}

	return strings.Contains(msg, sshAuthErrUnableToAuthenticate) ||
		strings.Contains(msg, sshAuthErrPermissionDenied) ||
		strings.Contains(msg, sshAuthErrNoSupportedAuth) ||
		strings.Contains(msg, sshAuthErrNoAuthMethod)
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
		Host:             sess.Host,
		LoginDestination: sessionLoginDestination(sess),
		Port:             normalizePort(sess.Port),
		User:             sess.Username,
		IdentityPath:     sess.IdentityPath,
		Password:         password,
		UseAgent:         true,
	}
}

func sessionLoginDestination(sess *store.Session) string {
	return sessionDialDestination(sess)
}

func sessionHostLabel(sess *store.Session) string {
	if sess.Username != "" {
		return sess.Username + "@" + sess.Host
	}
	return sess.Host
}

// remotePromptInitialCommands returns PS1 setup lines from session config (login/load shared path).
func (m *Manager) remotePromptInitialCommands(user, host string) []string {
	dirAbsolute := false
	if m != nil && m.Config != nil {
		dirAbsolute = m.Config.RemotePromptPathAbsolute()
	}
	return sshclient.RemotePromptInitialCommands(user, host, dirAbsolute)
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

// HasActiveSession reports whether a session is currently marked active (connected).
// It reads MetaCurrentSession from the store, fetches the session row, and checks status == Active.
// Returns true only when both the meta points to an existing session and that session is active.
func (m *Manager) HasActiveSession() bool {
	if m == nil || m.Store == nil {
		return false
	}
	current, err := m.Store.GetMeta(store.MetaCurrentSession)
	if err != nil || current == "" {
		return false
	}
	sess, err := m.Store.GetSession(current)
	if err != nil {
		return false
	}
	return sess.Status == store.StatusActive
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
		return errhint.WithTryMsg(
			fmt.Sprintf(errCurrentSessionActiveFmt, cur.AliasOrID()),
			hintActiveSessionSwitchExplain,
			meta.LineSuspend, trySwitchForce, trySessionHelp,
		)
	}

	fmt.Fprintln(os.Stderr, warnSwitchWithoutSuspend)
	_ = m.Store.UpdateStatus(cur.ID, store.StatusSuspended)
	return nil
}

// deleteLoginSession removes a just-created login row and clears current-session meta.
func (m *Manager) deleteLoginSession(id string) {
	_ = m.Store.DeleteSession(id)
	cur, _ := m.Store.GetMeta(store.MetaCurrentSession)
	if cur == id {
		_ = m.Store.SetMeta(store.MetaCurrentSession, "")
	}
}

// cleanupOnTmuxFailure clears or deletes persisted session state when tmux setup fails before attach.
// newRow deletes the row when login created it; existing rows are marked suspended.
func (m *Manager) cleanupOnTmuxFailure(persist, newRow bool, sessID string) {
	if !persist {
		return
	}
	if newRow {
		m.deleteLoginSession(sessID)
		return
	}
	m.clearSessionOnFailure(sessID, true)
}
