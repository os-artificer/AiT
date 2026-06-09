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

package client

import (
	"io"

	"github.com/os-artificer/AiT/internal/pkg/tmux"
)

func (c *Client) ptyInjector(stdin io.Writer) tmux.PTYInjector {
	return func(cmd string, capture bool) (string, error) {
		return c.InjectCapture(stdin, cmd)
	}
}

// tmuxEnsure creates the remote tmux session when missing (direct SSH or chain PTY).
func (c *Client) tmuxEnsure(stdin io.Writer, name string) error {
	if c.ChainMode() {
		return tmux.EnsureViaPTY(c.ptyInjector(stdin), name)
	}
	return tmux.Ensure(c.conn, name)
}

// tmuxDetach detaches clients from the remote tmux session.
func (c *Client) tmuxDetach(stdin io.Writer, name string) error {
	if c.ChainMode() {
		return tmux.DetachViaPTY(c.ptyInjector(stdin), name)
	}
	return tmux.Detach(c.conn, name)
}

// tmuxPanePath returns the cwd of the tmux pane (direct SSH or chain PTY).
func (c *Client) tmuxPanePath(stdin io.Writer, name string) (string, error) {
	if c.ChainMode() {
		return tmux.PanePathViaPTY(c.ptyInjector(stdin), name)
	}
	return tmux.PanePath(c.conn, name)
}
