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
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/tmux"
)

// partitionInitialCommands splits PS1 setup lines (pre-applied in detached tmux pane)
// from post-attach commands such as cwd restore.
func partitionInitialCommands(cmds []string, user, host string) (setup, after []string) {
	for _, line := range cmds {
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}
		if isSetupLeakLine(line, user, host) {
			setup = append(setup, line)
			continue
		}
		after = append(after, line)
	}
	return setup, after
}

// tmuxPreAttachSetup applies PS1 setup in a detached tmux pane before attach (direct SSH only).
func (c *Client) tmuxPreAttachSetup(name string, setup []string) error {
	if c.conn == nil || c.ChainMode() || len(setup) == 0 {
		return nil
	}

	for _, cmd := range setup {
		if err := tmux.SendLine(c.conn, name, cmd); err != nil {
			return err
		}
	}

	return tmux.ClearPane(c.conn, name)
}
