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
	"strings"
	"time"
)

const (
	nestedPromptMaxWait      = 30 * time.Second
	nestedMOTDReadyFallback  = 2 * time.Second
	nestedPromptPollInterval = remoteOutputQuietPollStep
)

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
	if c.nestedState != nestedPromptIdle {
		c.nestedMu.Unlock()
		return
	}

	c.pendingNestedUser = user
	c.pendingNestedHost = host
	c.nestedRemoteHost = host
	c.nestedState = nestedPromptAwaitingShell
	c.nestedAwaitingSince = time.Now()
	c.nestedMOTDSeen = false
	c.nestedMOTDAt = time.Time{}
	c.nestedShellReadySeen = false
	c.nestedPasswordPending = false
	c.nestedMu.Unlock()

	// Poll readiness so idle shells (MOTD + prompt, then silence) still get
	// chain/host updates without waiting for another PTY chunk (NP1).
	go c.pollNestedPromptReady()
}

// pollNestedPromptReady waits until nestedReadyToInject, then applies chain/host
// (and optional AiT PS1). Exits when inject runs or the await is cancelled.
func (c *Client) pollNestedPromptReady() {
	if c == nil {
		return
	}

	ticker := time.NewTicker(nestedPromptPollInterval)
	defer ticker.Stop()
	deadline := time.Now().Add(nestedPromptMaxWait)

	for time.Now().Before(deadline) {
		<-ticker.C
		if c.tryNestedPromptInject() {
			return
		}
	}
	_ = c.tryNestedPromptInject()
}

func (c *Client) maybeInjectNestedPrompt(chunk []byte) {
	if c == nil || len(chunk) == 0 {
		return
	}

	c.nestedMu.Lock()
	if c.nestedState != nestedPromptAwaitingShell {
		c.nestedMu.Unlock()
		return
	}

	c.observeNestedChunkLocked(chunk)
	c.nestedMu.Unlock()

	_ = c.tryNestedPromptInject()
}

// tryNestedPromptInject applies nested identity/chain when ready. Returns true when
// the await is finished (injected, skipped, or no longer awaiting).
func (c *Client) tryNestedPromptInject() bool {
	if c == nil {
		return true
	}

	c.nestedMu.Lock()
	if c.nestedState != nestedPromptAwaitingShell {
		c.nestedMu.Unlock()
		return true
	}

	c.observeNestedScrollLocked()
	if !c.nestedReadyToInject() {
		c.nestedMu.Unlock()
		return false
	}

	user := c.pendingNestedUser
	host := c.pendingNestedHost
	stdin := c.interactiveStdin
	c.activePromptUser = user
	c.activePromptHost = host
	c.SetChainMode(true)
	c.nestedState = nestedPromptInjecting
	skipPS1 := UseRemoteSystemPrompt() || stdin == nil || user == "" || host == ""
	if skipPS1 {
		c.nestedState = nestedPromptIdle
		c.nestedMu.Unlock()
		return true
	}
	c.nestedMu.Unlock()

	c.runNestedPS1Inject(stdin, user, host)

	c.nestedMu.Lock()
	if c.nestedState == nestedPromptInjecting {
		c.nestedState = nestedPromptIdle
	}
	c.nestedMu.Unlock()
	return true
}

func (c *Client) observeNestedChunkLocked(chunk []byte) {
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

	if scrollChunkLooksLikeShellPrompt(chunk) {
		c.nestedShellReadySeen = true
		c.nestedPasswordPending = false
	}
}

func (c *Client) observeNestedScrollLocked() {
	if c.scroll == nil {
		return
	}
	chunk := []byte(c.scroll.String())
	if scrollChunkLooksLikeMOTD(chunk) {
		if !c.nestedMOTDSeen {
			c.nestedMOTDAt = time.Now()
		}
		c.nestedMOTDSeen = true
	}
	if scrollChunkLooksLikeShellPrompt(chunk) {
		c.nestedShellReadySeen = true
		c.nestedPasswordPending = false
	}
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
			// MOTD without a detected prompt: fall through after a short wait (NP3).
			if c.nestedMOTDAt.IsZero() || time.Since(c.nestedMOTDAt) < nestedMOTDReadyFallback {
				return false
			}
		} else if !c.nestedMOTDAt.IsZero() && time.Since(c.nestedMOTDAt) < shellReadyDelay {
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

// runNestedPS1Inject installs AiT remote PS1 on the nested hop. Caller must not hold nestedMu.
func (c *Client) runNestedPS1Inject(stdin io.Writer, user, host string) {
	if stdin == nil || user == "" || host == "" {
		return
	}

	c.beginRemotePromptSetup(c.remotePromptDirAbs, "")

	cmds := RemotePromptInitialCommands(user, host, c.remotePromptDirAbs)
	if len(cmds) == 0 {
		return
	}

	if err := c.injectInitialCommands(stdin, cmds); err != nil {
		c.remoteSetupSuppress.Store(false)
		c.pauseRemoteOut.Store(false)
		return
	}

	c.runRemotePromptSetup(stdin, user, host, false)
}

// onNestedSSHSessionClosed restores dial prompt identity after a manual nested ssh exits.
// Returns true when the outer-shell prompt should be repainted locally.
func (c *Client) onNestedSSHSessionClosed(closedHost string) bool {
	if c == nil || closedHost == "" {
		return false
	}

	c.nestedMu.Lock()
	defer c.nestedMu.Unlock()

	nested := c.nestedRemoteHost
	if nested == "" {
		nested = c.pendingNestedHost
	}
	if nested != "" && !hostsEquivalent(nested, closedHost) && !hostsEquivalent(c.activePromptHost, closedHost) {
		return false
	}

	c.activePromptUser = c.opts.User
	c.activePromptHost = c.opts.Host
	c.pendingNestedUser = ""
	c.pendingNestedHost = ""
	c.nestedRemoteHost = ""
	c.nestedState = nestedPromptIdle
	c.nestedPasswordPending = false
	c.nestedMOTDSeen = false
	c.nestedShellReadySeen = false
	c.nestedAwaitingSince = time.Time{}
	c.nestedMOTDAt = time.Time{}
	if c.ChainMode() {
		c.SetChainMode(false)
	}
	return true
}
