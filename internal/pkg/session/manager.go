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
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	"github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
	"github.com/os-artificer/AiT/internal/pkg/pathsafe"
	"github.com/os-artificer/AiT/internal/pkg/platform"
	"github.com/os-artificer/AiT/internal/pkg/sshclient"
	"github.com/os-artificer/AiT/internal/pkg/store"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
	"github.com/os-artificer/AiT/internal/pkg/tmux"
	"golang.org/x/term"
)

// LoginOpts for new connection.
type LoginOpts struct {
	Alias        string
	Host         string
	Port         int
	User         string
	IdentityPath string
	// IdentityExplicit is true when the user passed -i on the CLI.
	IdentityExplicit bool
	Password         string
	Tmux             bool

	RemoteCommand []string

	ProxyJump             []string
	ProxyCommand          string
	KnownHostsPath        string
	StrictHostKeyChecking string
	ConnectTimeout        time.Duration
	BatchMode             bool
}

// Manager orchestrates session lifecycle.
type Manager struct {
	Store  *store.Store
	Config *config.Config
	AI     *ai.Client
	// MetaCLI runs ait subcommands from in-session !ait lines (set by internal/ait).
	MetaCLI func(args []string) error
}

func normalizePort(port int) int {
	if port == 0 {
		return 22
	}
	return port
}

// Login connects and saves session (or runs a one-shot remote command).
func (m *Manager) Login(ctx context.Context, opts LoginOpts) error {
	opts.Port = normalizePort(opts.Port)
	if len(opts.RemoteCommand) > 0 && opts.Tmux {
		return fmt.Errorf("cannot use --tmux with a remote command")
	}

	sshOpts := sshclient.Options{
		Host: opts.Host, Port: opts.Port, User: opts.User,
		IdentityPath: opts.IdentityPath, IdentityExplicit: opts.IdentityExplicit,
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

func (m *Manager) dialLogin(sshOpts sshclient.Options, password, host string) (*sshclient.Client, error) {
	client, err := sshclient.DialChain(sshOpts)
	if err != nil && shouldPromptSSHPassword(sshOpts, password) && isSSHAuthFailure(err) {
		user := sshOpts.User
		if user == "" {
			user = "user"
		}
		if p, perr := promptPassword(fmt.Sprintf("Password for %s@%s: ", user, host)); perr == nil && p != "" {
			sshOpts.Password = p
			client, err = sshclient.DialChain(sshOpts)
		}
	}
	return client, err
}

func (m *Manager) loginRunOnce(ctx context.Context, client *sshclient.Client, opts LoginOpts) error {
	if opts.Alias != "" {
		if err := m.persistLoginSession(client, opts); err != nil {
			return err
		}
	}
	cmdline := sshclient.BuildRemoteCommandLine(opts.RemoteCommand)
	return client.RunRemoteCommand(ctx, cmdline)
}

func (m *Manager) loginInteractive(ctx context.Context, client *sshclient.Client, opts LoginOpts, sshOpts sshclient.Options) error {
	id := uuid.New().String()
	var reuseChangedConn bool
	if opts.Alias != "" {
		if existing, err := m.Store.GetSession(opts.Alias); err == nil {
			id = existing.ID
			if existing.Host != opts.Host ||
				normalizePort(existing.Port) != opts.Port ||
				(opts.User != "" && existing.Username != opts.User) {
				reuseChangedConn = true
			}
		}
	}
	if reuseChangedConn {
		if st, _ := m.Store.GetSessionState(id); st != nil && st.ScrollbackPath != "" {
			m.removeScrollbackFile(st.ScrollbackPath)
		}
		_ = m.Store.DeleteSessionState(id)
	}

	authMethod := sshclient.DetectAuthMethod(sshOpts)
	ttl, _ := m.Config.AuthTTL()
	expires := time.Now().Add(ttl).Unix()
	tmuxName := tmux.SessionName(id)
	sess := &store.Session{
		ID: id, Alias: opts.Alias, Host: opts.Host, Port: opts.Port,
		Username: opts.User, IdentityPath: opts.IdentityPath,
		AuthMethod: authMethod, AuthExpiresAt: expires,
		TmuxEnabled: opts.Tmux, TmuxSession: tmuxName,
		Status: store.StatusSuspended,
	}
	if err := m.Store.UpsertSession(sess); err != nil {
		return err
	}
	if opts.Password != "" {
		_ = creds.SetSessionUnlock(id, opts.Password)
	}
	_ = m.Store.SetMeta(store.MetaCurrentSession, id)

	if opts.Tmux {
		if err := tmux.Ensure(client.Conn(), tmuxName); err != nil {
			m.clearSessionOnFailure(id)
			return fmt.Errorf("tmux: %w (is tmux installed on remote?)", err)
		}
	}

	initialCmds := sshclient.RemotePromptInitialCommands(sess.Username, sess.Host)
	err := m.runInteractive(ctx, client, sess, initialCmds)
	if err != nil {
		m.clearSessionOnFailure(id)
	}
	return err
}

func (m *Manager) persistLoginSession(client *sshclient.Client, opts LoginOpts) error {
	_ = client
	id := uuid.New().String()
	if opts.Alias != "" {
		if existing, err := m.Store.GetSession(opts.Alias); err == nil {
			id = existing.ID
		}
	}
	ttl, _ := m.Config.AuthTTL()
	expires := time.Now().Add(ttl).Unix()
	sess := &store.Session{
		ID: id, Alias: opts.Alias, Host: opts.Host, Port: normalizePort(opts.Port),
		Username: opts.User, IdentityPath: opts.IdentityPath,
		AuthExpiresAt: expires, Status: store.StatusSuspended,
	}
	return m.Store.UpsertSession(sess)
}

// Load restores a saved session.
func (m *Manager) Load(ctx context.Context, idOrAlias string, useTmux bool, password string) error {
	sess, err := m.Store.GetSession(idOrAlias)
	if err != nil {
		return err
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

	client, err := sshclient.Dial(sshclient.Options{
		Host: sess.Host, Port: normalizePort(sess.Port), User: sess.Username,
		IdentityPath: sess.IdentityPath, Password: pass, UseAgent: true,
	})
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
	if st != nil && st.ScrollbackPath != "" {
		if dir, derr := platform.ScrollbackDir(); derr == nil {
			if path, perr := terminal.ResolveScrollbackPath(dir, st.ScrollbackPath); perr == nil {
				if data, lerr := terminal.LoadFromFile(path); lerr == nil {
					terminal.Replay(data)
				}
			}
		}
	}

	var initialCmds []string
	initialCmds = append(sshclient.RemotePromptInitialCommands(sess.Username, sess.Host), initialCmds...)
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
			return err
		}
	}

	_ = m.Store.SetMeta(store.MetaCurrentSession, sess.ID)
	return m.runInteractive(ctx, client, sess, initialCmds)
}

// SSHOptionsForSession builds dial options for a saved session (e.g. sync).
func (m *Manager) SSHOptionsForSession(idOrAlias, passwordOverride string) (sshclient.Options, error) {
	sess, err := m.Store.GetSession(idOrAlias)
	if err != nil {
		return sshclient.Options{}, err
	}
	pass, _, err := m.resolvePassword(sess, passwordOverride)
	if err != nil {
		return sshclient.Options{}, err
	}
	return sshclient.Options{
		Host: sess.Host, Port: normalizePort(sess.Port), User: sess.Username,
		IdentityPath: sess.IdentityPath, Password: pass, UseAgent: true,
	}, nil
}

func (m *Manager) clearSessionOnFailure(id string) {
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
		p, err := promptPassword(fmt.Sprintf("Password for %s@%s (auth expired): ", sess.Username, sess.Host))
		if err != nil {
			return "", false, err
		}
		return p, true, nil
	}
	if sess.IdentityPath != "" {
		p, err := promptPassword(fmt.Sprintf("Passphrase for key %s (auth expired): ", sess.IdentityPath))
		if err != nil {
			return "", false, err
		}
		return p, true, nil
	}
	return "", false, nil
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

func (m *Manager) runInteractive(ctx context.Context, client *sshclient.Client, sess *store.Session, initialCmds []string) error {
	defer client.Close()
	tmuxName := sess.TmuxSession
	if tmuxName == "" {
		tmuxName = tmux.SessionName(sess.ID)
	}
	if err := m.Store.UpdateStatus(sess.ID, store.StatusActive); err != nil {
		return err
	}
	err := client.RunInteractive(ctx, sshclient.InteractiveOpts{
		SessionID:       sess.ID,
		TmuxEnabled:     sess.TmuxEnabled,
		TmuxName:        tmuxName,
		InitialCommands: initialCmds,
		AI:              m.AI,
		OnMetaCLI:       m.MetaCLI,
		OnSuspend: func(scroll *terminal.Buffer, cwd string, cols, rows int) error {
			return m.suspend(sess.ID, scroll, cwd, cols, rows)
		},
	})
	if err != nil {
		m.clearSessionOnFailure(sess.ID)
	}
	return err
}

func (m *Manager) suspend(id string, scroll *terminal.Buffer, cwd string, cols, rows int) error {
	dir, err := platform.ScrollbackDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, id+".log")
	if err := scroll.SaveToFile(path); err != nil {
		return err
	}
	st := &store.SessionState{
		SessionID: id, Cols: cols, Rows: rows,
		RemoteCwd: cwd, ScrollbackPath: terminal.ScrollbackFileName(id), SuspendedAt: time.Now().Unix(),
	}
	if err := m.Store.SaveSessionState(st); err != nil {
		return err
	}
	return m.Store.UpdateStatus(id, store.StatusSuspended)
}

// SuspendCurrent explains how to suspend when invoked from the local shell.
func (m *Manager) SuspendCurrent() error {
	id, _ := m.Store.GetMeta(store.MetaCurrentSession)
	if id == "" {
		return fmt.Errorf("no current session; use !ait suspend inside SSH or ait load first")
	}
	fmt.Println("Suspend from the local shell is not supported.")
	fmt.Println("Reconnect to the session and run !ait suspend inside SSH.")
	return nil
}

// Switch changes from current session to target. The target may be an ID or
// alias; it is resolved before any state change so that comparisons against
// the persisted current_session_id (always a UUID) are correct.
func (m *Manager) Switch(ctx context.Context, target string, useTmux bool, password string, force bool) error {
	tgt, err := m.Store.GetSession(target)
	if err != nil {
		return err
	}
	current, _ := m.Store.GetMeta(store.MetaCurrentSession)
	if current == tgt.ID {
		fmt.Println("already on session", tgt.AliasOrID())
		return nil
	}
	if current != "" {
		cur, gerr := m.Store.GetSession(current)
		switch {
		case errors.Is(gerr, store.ErrSessionNotFound):
			// Stale meta pointing at a deleted session — drop it so the next
			// Switch/Load doesn't re-run this lookup.
			_ = m.Store.SetMeta(store.MetaCurrentSession, "")
		case gerr == nil && cur.Status == store.StatusActive:
			if !force {
				return fmt.Errorf(
					"current session %s is still active; suspend it first inside SSH with !ait suspend, or use --force",
					cur.AliasOrID(),
				)
			}
			fmt.Fprintln(os.Stderr, "warning: switching without suspend; marking current session suspended")
			_ = m.Store.UpdateStatus(cur.ID, store.StatusSuspended)
		}
	}
	return m.Load(ctx, tgt.ID, useTmux || tgt.TmuxEnabled, password)
}

// PickSession interactively lists sessions.
func (m *Manager) PickSession() (string, error) {
	list, err := m.Store.ListSessions()
	if err != nil {
		return "", err
	}
	if len(list) == 0 {
		return "", fmt.Errorf("no saved sessions")
	}
	for i, s := range list {
		fmt.Printf("  %d) %s  %s@%s:%d  [%s]\n", i+1, s.AliasOrID(), s.Username, s.Host, s.Port, s.Status)
	}
	fmt.Print("Select session (number, id, or alias): ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return "", fmt.Errorf("invalid selection")
	}
	var n int
	if _, err := fmt.Sscanf(line, "%d", &n); err == nil {
		if n < 1 || n > len(list) {
			return "", fmt.Errorf("invalid selection")
		}
		return list[n-1].ID, nil
	}
	if sess, err := m.Store.GetSession(line); err == nil {
		return sess.ID, nil
	}
	return "", fmt.Errorf("invalid selection")
}

// Delete removes one session.
func (m *Manager) Delete(idOrAlias string, force bool) error {
	sess, err := m.Store.GetSession(idOrAlias)
	if err != nil {
		return err
	}
	if !force {
		fmt.Printf("Delete session %s (%s@%s)? [y/N] ", sess.AliasOrID(), sess.Username, sess.Host)
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(line)) != "y" {
			fmt.Println("cancelled")
			return nil
		}
	}
	st, _ := m.Store.GetSessionState(sess.ID)
	if st != nil && st.ScrollbackPath != "" {
		m.removeScrollbackFile(st.ScrollbackPath)
	}
	_ = creds.DeleteSessionUnlock(sess.ID)
	current, _ := m.Store.GetMeta(store.MetaCurrentSession)
	if current == sess.ID {
		_ = m.Store.SetMeta(store.MetaCurrentSession, "")
	}
	return m.Store.DeleteSession(sess.ID)
}

// DeleteAll removes all sessions.
func (m *Manager) DeleteAll(force bool) error {
	if !force {
		fmt.Print("Delete ALL sessions? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(line)) != "y" {
			fmt.Println("cancelled")
			return nil
		}
	}
	list, _ := m.Store.ListSessions()
	for _, s := range list {
		st, _ := m.Store.GetSessionState(s.ID)
		if st != nil && st.ScrollbackPath != "" {
			m.removeScrollbackFile(st.ScrollbackPath)
		}
		_ = creds.DeleteSessionUnlock(s.ID)
	}
	_ = m.Store.SetMeta(store.MetaCurrentSession, "")
	return m.Store.DeleteAllSessions()
}

// List prints sessions to w as a left-aligned table.
func (m *Manager) List(w io.Writer) error {
	list, err := m.Store.ListSessions()
	if err != nil {
		return err
	}
	if len(list) == 0 {
		fmt.Fprintln(w, "No saved sessions.")
		return nil
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tTARGET\tSTATUS\tTMUX\tREMAINING")
	now := time.Now().Unix()
	for _, s := range list {
		remain := time.Duration(s.AuthExpiresAt-now) * time.Second
		remainingLabel := remain.Round(time.Hour).String()
		if remain <= 0 {
			remainingLabel = "expired"
		}
		tmuxCol := "-"
		if s.TmuxEnabled {
			tmuxCol = "tmux"
		}
		fmt.Fprintf(tw, "%s\t%s@%s:%d\t%s\t%s\t%s\n",
			s.AliasOrID(), s.Username, s.Host, s.Port, s.Status, tmuxCol, remainingLabel)
	}
	return tw.Flush()
}

func quotePath(p string) string {
	return pathsafe.QuoteShellSingle(p)
}
