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

package tmux

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

// SessionName returns the deterministic remote tmux session name for an AiT session UUID.
// sessionID is the store session id, not an alias.
func SessionName(sessionID string) string {
	return "ait-" + sessionID
}

func runRemote(conn *ssh.Client, cmd string) (string, error) {
	sess, err := conn.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()

	out, err := sess.CombinedOutput(cmd)
	return strings.TrimSpace(string(out)), err
}

// HasSession reports whether tmux session name exists on the remote host.
// conn must be an active SSH client. name is the tmux session name (see SessionName).
// Returns (true, nil) when present, (false, nil) when absent, or NewSession/run errors.
func HasSession(conn *ssh.Client, name string) (bool, error) {
	_, err := runRemote(conn, fmt.Sprintf("tmux has-session -t %q 2>/dev/null", name))
	return err == nil, nil
}

// NewSession creates a detached tmux session on the remote host.
// conn must be connected; name must be a valid tmux session identifier.
// Returns NewSession or remote tmux command errors (e.g. tmux not installed).
func NewSession(conn *ssh.Client, name string) error {
	_, err := runRemote(conn, fmt.Sprintf("tmux new-session -d -s %q", name))
	return err
}

// Detach detaches clients from tmux session name (best effort; ignores missing clients).
// Returns NewSession or remote tmux command errors.
func Detach(conn *ssh.Client, name string) error {
	_, err := runRemote(conn, fmt.Sprintf("tmux detach-client -s %q 2>/dev/null || true", name))
	return err
}

// AttachCommand returns a shell command line that exec-replaces the process with tmux attach-session.
// name is the tmux session name (see SessionName).
func AttachCommand(name string) string {
	return fmt.Sprintf("exec tmux attach-session -t %q", name)
}

// PanePath returns the current working directory of tmux session name's pane.
// Returns the path string, or NewSession/tmux display-message errors.
func PanePath(conn *ssh.Client, name string) (string, error) {
	return runRemote(conn, fmt.Sprintf("tmux display-message -p -t %q '#{pane_current_path}'", name))
}

const (
	tmuxSendKeysEnter = "Enter"
	tmuxSendKeysClear = "C-l"
)

// shSingleQuote wraps s in single quotes for a POSIX sh argument (no $ expansion).
func shSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// BuildSendKeysLiteralCmd returns tmux send-keys -l for literal text into session name's pane.
// literal is passed without outer-shell $ expansion (single-quoted).
func BuildSendKeysLiteralCmd(name, literal string) string {
	return fmt.Sprintf("tmux send-keys -t %s -l %s", shSingleQuote(name), shSingleQuote(literal))
}

// BuildSendKeyCmd returns tmux send-keys for a special key (e.g. Enter, C-l).
func BuildSendKeyCmd(name, key string) string {
	return fmt.Sprintf("tmux send-keys -t %s %s", shSingleQuote(name), key)
}

// SendLine types line into session name's pane and submits with Enter.
// conn must be connected; line is passed literally to the pane shell.
// Returns NewSession or remote tmux send-keys errors.
func SendLine(conn *ssh.Client, name, line string) error {
	if _, err := runRemote(conn, BuildSendKeysLiteralCmd(name, line)); err != nil {
		return err
	}
	_, err := runRemote(conn, BuildSendKeyCmd(name, tmuxSendKeysEnter))
	return err
}

// ClearPane clears scrollback and redraws session name's pane (best effort before attach).
// Returns NewSession or remote tmux command errors.
func ClearPane(conn *ssh.Client, name string) error {
	if _, err := runRemote(conn, fmt.Sprintf("tmux clear-history -t %s 2>/dev/null || true", shSingleQuote(name))); err != nil {
		return err
	}
	_, err := runRemote(conn, BuildSendKeyCmd(name, tmuxSendKeysClear))
	return err
}

// Ensure creates tmux session name on the remote host when it does not already exist.
// Returns HasSession or NewSession errors.
func Ensure(conn *ssh.Client, name string) error {
	ok, err := HasSession(conn, name)
	if err != nil {
		return err
	}

	if ok {
		return nil
	}
	return NewSession(conn, name)
}
