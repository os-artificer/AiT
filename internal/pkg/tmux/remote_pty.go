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
)

// PTYInjector is a function that sends a command through the PTY stdin pipe
// with history and echo suppression, returning captured output.
// When capture is true, the output between delimiters is captured from scrollback.
type PTYInjector func(cmd string, capture bool) (string, error)

// HasSessionViaPTY reports whether tmux session name exists on the remote host,
// using the PTY pipe for command injection.
// inject sends commands through the PTY with silent injection.
// Returns (true, nil) when present, (false, nil) when absent, or injection errors.
func HasSessionViaPTY(inject PTYInjector, name string) (bool, error) {
	output, err := inject(fmt.Sprintf("tmux has-session -t %q 2>/dev/null; echo \"AIT_EXIT_$?\"", name), true)
	if err != nil {
		return false, err
	}

	return strings.Contains(output, "AIT_EXIT_0"), nil
}

// NewSessionViaPTY creates a detached tmux session on the remote host,
// using the PTY pipe for command injection.
// inject sends commands through the PTY with silent injection.
// Returns injection errors (e.g. tmux not installed).
func NewSessionViaPTY(inject PTYInjector, name string) error {
	_, err := inject(fmt.Sprintf("tmux new-session -d -s %q", name), false)
	return err
}

// DetachViaPTY detaches clients from tmux session name (best effort),
// using the PTY pipe for command injection.
// inject sends commands through the PTY with silent injection.
func DetachViaPTY(inject PTYInjector, name string) error {
	_, err := inject(fmt.Sprintf("tmux detach-client -s %q 2>/dev/null || true", name), false)
	return err
}

// PanePathViaPTY returns the current working directory of tmux session name's pane,
// using the PTY pipe for command injection.
// inject sends commands through the PTY with silent injection.
// Returns the path string, or injection errors.
func PanePathViaPTY(inject PTYInjector, name string) (string, error) {
	output, err := inject(fmt.Sprintf("tmux display-message -p -t %q '#{pane_current_path}'", name), true)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

// EnsureViaPTY creates tmux session name on the remote host when it does not
// already exist, using the PTY pipe for command injection.
// inject sends commands through the PTY with silent injection.
// Returns HasSessionViaPTY or NewSessionViaPTY errors.
func EnsureViaPTY(inject PTYInjector, name string) error {
	ok, err := HasSessionViaPTY(inject, name)
	if err != nil {
		return err
	}

	if ok {
		return nil
	}
	return NewSessionViaPTY(inject, name)
}
