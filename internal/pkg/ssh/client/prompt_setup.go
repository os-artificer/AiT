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
)

// beginRemotePromptSetup arms pause/suppress state shared by login/load and nested SSH setup.
func (c *Client) beginRemotePromptSetup(dirAbsolute bool, initialCwd string) {
	if UseRemoteSystemPrompt() {
		return
	}

	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)
	c.setupInjectDone.Store(false)
	c.setupLeakGraceUntil.Store(0)
	c.setupKeystrokeGraceUntil.Store(0)
	c.remotePromptReadySeen.Store(false)
	c.remotePromptShown.Store(false)
	c.remotePromptDirAbs = dirAbsolute
	c.initialRemoteCwd = initialCwd
	c.setupFilter = setupOutputFilter{}
}

// runRemotePromptSetup finishes remote PS1 setup and makes the first prompt visible before unpause.
func (c *Client) runRemotePromptSetup(stdin io.Writer, user, host string, tmuxEnabled bool) {
	if UseRemoteSystemPrompt() {
		return
	}

	c.finishRemoteSetup(stdin, user, host, tmuxEnabled)
}
