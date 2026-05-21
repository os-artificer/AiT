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

func run(sess *ssh.Session, cmd string) (string, error) {
	out, err := sess.CombinedOutput(cmd)
	return strings.TrimSpace(string(out)), err
}

// HasSession reports whether tmux session name exists on the remote host.
// conn must be an active SSH client. name is the tmux session name (see SessionName).
// Returns (true, nil) when present, (false, nil) when absent, or NewSession/run errors.
func HasSession(conn *ssh.Client, name string) (bool, error) {
	sess, err := conn.NewSession()
	if err != nil {
		return false, err
	}
	defer sess.Close()

	_, err = run(sess, fmt.Sprintf("tmux has-session -t %q 2>/dev/null", name))
	return err == nil, nil
}

// NewSession creates a detached tmux session on the remote host.
// conn must be connected; name must be a valid tmux session identifier.
// Returns NewSession or remote tmux command errors (e.g. tmux not installed).
func NewSession(conn *ssh.Client, name string) error {
	sess, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	_, err = run(sess, fmt.Sprintf("tmux new-session -d -s %q", name))
	return err
}

// Detach detaches clients from tmux session name (best effort; ignores missing clients).
// Returns NewSession or remote tmux command errors.
func Detach(conn *ssh.Client, name string) error {
	sess, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	_, err = run(sess, fmt.Sprintf("tmux detach-client -s %q 2>/dev/null || true", name))
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
	sess, err := conn.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()

	return run(sess, fmt.Sprintf("tmux display-message -p -t %q '#{pane_current_path}'", name))
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
