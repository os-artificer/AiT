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
	"path"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/prompt"
)

const clientPromptDefaultDir = "~"

// formatClientPromptDir formats cwd for the client-painted remote prompt display line.
func formatClientPromptDir(cwd string, dirAbsolute bool) string {
	cwd = strings.TrimSpace(cwd)
	if cwd == "" {
		return clientPromptDefaultDir
	}
	if !dirAbsolute {
		base := path.Base(strings.TrimRight(cwd, "/"))
		if base == "" || base == "." {
			return clientPromptDefaultDir
		}
		return base
	}
	return cwd
}

// setupSucceeded reports whether remote PS1 setup completed (sentinel seen).
func (c *Client) setupSucceeded() bool {
	if c == nil {
		return false
	}
	if c.remotePromptReadySeen.Load() {
		return true
	}
	if c.scroll == nil {
		return false
	}
	return containsPromptReadySentinel(c.scroll.String())
}

// paintClientRemotePrompt writes an AiT remote prompt to interactive stdout without
// waiting for remote PTY echo. Returns true when the line was written.
func (c *Client) paintClientRemotePrompt(user, host string) bool {
	if c == nil || user == "" || host == "" {
		return false
	}

	dir := formatClientPromptDir(c.initialRemoteCwd, c.remotePromptDirAbs)
	line := prompt.RemoteFormat(user, host, dir, "")

	out := formatActivePromptStdout(line)
	if _, err := c.writeInteractiveStdout([]byte(out)); err != nil {
		return false
	}
	return true
}
