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

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/prompt"
)

// lastRemotePromptLine returns the last scrollback line that looks like an AiT remote prompt.
func lastRemotePromptLine(scroll string, user, host string) string {
	if scroll == "" {
		return ""
	}

	normalized := strings.ReplaceAll(scroll, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	lines := strings.Split(normalized, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if !prompt.IsRemotePromptLine(user, host, line) {
			continue
		}
		return line
	}

	return ""
}

// emitBufferedRemotePrompt writes the buffered remote prompt from scrollback to stdout.
// Returns true when a prompt line was found and written.
func (c *Client) emitBufferedRemotePrompt(user, host string) bool {
	line := lastRemotePromptLine(c.scroll.String(), user, host)
	if line == "" {
		return false
	}

	_, _ = c.writeInteractiveStdout([]byte("\r" + line))
	return true
}
