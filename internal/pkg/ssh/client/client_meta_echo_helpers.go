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

// metaLineEchoPrefix returns the visible prompt prefix for meta-line redraw echo.
func (c *Client) metaLineEchoPrefix() string {
	if c.localShell.Active() {
		return prompt.Local
	}

	user, host := c.promptUser(), c.promptHost()
	if UseRemoteSystemPrompt() {
		scroll := c.scroll.String()
		activeHost := c.promptHost()
		line := lastNativePromptLineForHost(scroll, activeHost)
		if line == "" {
			line = lastNativePromptLineFallback(scroll, activeHost)
		}
		if line != "" {
			if cleaned := extractNativePromptForEmit(line); cleaned != "" {
				return cleaned
			}
		}
	}

	if line := lastRemotePromptLine(c.scroll.String(), user, host); line != "" {
		return strings.TrimRight(line, "\r\n")
	}

	cwd := c.remoteCwd
	if cwd == "" {
		cwd = "/"
	}
	return prompt.RemoteFormat(user, host, cwd, "")
}
