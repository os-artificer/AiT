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

// SessionName builds the deterministic tmux session name for a session ID.
func SessionName(sessionID string) string {
	return "ait-" + sessionID
}

func run(sess *ssh.Session, cmd string) (string, error) {
	out, err := sess.CombinedOutput(cmd)
	return strings.TrimSpace(string(out)), err
}

// HasSession checks if tmux session exists on remote.
func HasSession(conn *ssh.Client, name string) (bool, error) {
	sess, err := conn.NewSession()
	if err != nil {
		return false, err
	}
	defer sess.Close()
	_, err = run(sess, fmt.Sprintf("tmux has-session -t %q 2>/dev/null", name))
	return err == nil, nil
}

// NewSession creates a detached tmux session.
func NewSession(conn *ssh.Client, name string) error {
	sess, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()
	_, err = run(sess, fmt.Sprintf("tmux new-session -d -s %q", name))
	return err
}

// Detach detaches clients from session (best effort).
func Detach(conn *ssh.Client, name string) error {
	sess, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()
	_, err = run(sess, fmt.Sprintf("tmux detach-client -s %q 2>/dev/null || true", name))
	return err
}

// AttachCommand returns shell command to attach to session.
func AttachCommand(name string) string {
	return fmt.Sprintf("exec tmux attach-session -t %q", name)
}

// PanePath returns the current working directory of the tmux pane.
func PanePath(conn *ssh.Client, name string) (string, error) {
	sess, err := conn.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()
	return run(sess, fmt.Sprintf("tmux display-message -p -t %q '#{pane_current_path}'", name))
}

// Ensure creates session if missing.
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
