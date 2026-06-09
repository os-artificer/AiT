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
	"time"
)

const nestedPromptMaxWait = 30 * time.Second

// initActivePrompt sets active prompt identity to the dial target (single-hop default).
func (c *Client) initActivePrompt() {
	c.activePromptUser = c.opts.User
	c.activePromptHost = c.opts.Host
}

func (c *Client) promptUser() string {
	if c == nil {
		return ""
	}
	return c.activePromptUser
}

func (c *Client) promptHost() string {
	if c == nil {
		return ""
	}
	return c.activePromptHost
}

func (c *Client) promptMarker() string {
	return remoteSetupPromptMarker(c.promptUser(), c.promptHost())
}

// noteInteractiveInput records bytes forwarded to the remote shell and detects submitted lines.
func (c *Client) noteInteractiveInput(fwd []byte) {
	if c == nil || len(fwd) == 0 || c.localShell.Active() {
		return
	}

	for _, b := range fwd {
		switch b {
		case '\n', '\r':
			line := strings.TrimSpace(string(c.forwardLineBuf))
			c.forwardLineBuf = c.forwardLineBuf[:0]
			c.onInteractiveLineSubmitted(line)
		case 0x15:
			c.forwardLineBuf = c.forwardLineBuf[:0]
		case 0x7f, 0x08:
			if len(c.forwardLineBuf) > 0 {
				c.forwardLineBuf = c.forwardLineBuf[:len(c.forwardLineBuf)-1]
			}
		default:
			if b >= 0x20 || b == '\t' {
				c.forwardLineBuf = append(c.forwardLineBuf, b)
			}
		}
	}
}

func (c *Client) onInteractiveLineSubmitted(line string) {
	if line == "" {
		return
	}

	user, host, ok := parseSSHCommandLine(line, c.promptUser())
	if !ok {
		return
	}

	c.nestedMu.Lock()
	defer c.nestedMu.Unlock()

	if c.nestedState != nestedPromptIdle {
		return
	}

	c.pendingNestedUser = user
	c.pendingNestedHost = host
	c.nestedState = nestedPromptAwaitingShell
	c.nestedAwaitingSince = time.Now()
	c.nestedMOTDSeen = false
	c.nestedMOTDAt = time.Time{}
	c.nestedShellReadySeen = false
	c.nestedPasswordPending = false
}

func (c *Client) maybeInjectNestedPrompt(chunk []byte) {
	if c == nil || len(chunk) == 0 {
		return
	}

	c.nestedMu.Lock()
	defer c.nestedMu.Unlock()

	if c.nestedState != nestedPromptAwaitingShell {
		return
	}

	if scrollChunkLooksLikePasswordPrompt(chunk) {
		c.nestedPasswordPending = true
		return
	}

	if scrollChunkLooksLikeMOTD(chunk) {
		if !c.nestedMOTDSeen {
			c.nestedMOTDAt = time.Now()
		}
		c.nestedMOTDSeen = true
		c.nestedPasswordPending = false
	}

	if c.nestedMOTDSeen && scrollChunkLooksLikeShellPrompt(chunk) {
		c.nestedShellReadySeen = true
	}

	if !c.nestedReadyToInject() {
		return
	}

	c.nestedState = nestedPromptInjecting
	c.runNestedPromptInjectLocked()
	c.nestedState = nestedPromptIdle
}

func (c *Client) nestedReadyToInject() bool {
	if c.nestedPasswordPending {
		last := c.lastRemoteOut.Load()
		if last == 0 {
			return false
		}
		if time.Since(remoteOutAt(last)) < remoteOutputQuietPeriod {
			return false
		}
	}

	if c.nestedMOTDSeen {
		if !c.nestedShellReadySeen {
			return false
		}
		if !c.nestedMOTDAt.IsZero() && time.Since(c.nestedMOTDAt) < shellReadyDelay {
			return false
		}
	} else if time.Since(c.nestedAwaitingSince) < nestedPromptMaxWait {
		return false
	}

	last := c.lastRemoteOut.Load()
	if last == 0 {
		return false
	}
	return time.Since(remoteOutAt(last)) >= remoteOutputQuietPeriod
}

func (c *Client) runNestedPromptInjectLocked() {
	stdin := c.interactiveStdin
	if stdin == nil {
		return
	}

	user := c.pendingNestedUser
	host := c.pendingNestedHost
	if user == "" || host == "" {
		return
	}

	c.activePromptUser = user
	c.activePromptHost = host
	c.SetChainMode(true)

	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)
	c.setupInjectDone.Store(false)
	c.setupFilter = setupOutputFilter{}

	cmds := RemotePromptInitialCommands(user, host, c.remotePromptDirAbs)
	if len(cmds) == 0 {
		return
	}

	if err := c.injectInitialCommands(stdin, cmds); err != nil {
		return
	}

	c.finishRemoteSetup(stdin, user, host, false)

	marker := remoteSetupPromptMarker(user, host)
	if strings.Contains(c.scroll.String(), marker) {
		return
	}

	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)
	c.setupInjectDone.Store(false)
	c.setupFilter = setupOutputFilter{}

	if err := c.injectInitialCommands(stdin, cmds); err != nil {
		c.remoteSetupSuppress.Store(false)
		c.pauseRemoteOut.Store(false)
		return
	}

	c.finishRemoteSetup(stdin, user, host, false)
}
