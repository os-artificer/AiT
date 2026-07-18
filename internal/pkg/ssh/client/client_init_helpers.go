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

	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

const (
	remoteOutputQuietPeriod    = 250 * time.Millisecond
	remoteOutputQuietMaxWait   = 2 * time.Second
	remoteOutputQuietPollStep  = 25 * time.Millisecond
	remoteAITPromptMaxWait     = 5 * time.Second
	remoteAITPromptRefreshWait = 500 * time.Millisecond
	shellQuietMaxWait          = 5 * time.Second
	shellReadyMaxWait          = 10 * time.Second
	tmuxAttachQuietPeriod      = 450 * time.Millisecond
	tmuxAttachQuietMaxWait     = 3 * time.Second

	// setupLeakGraceMax caps the post-setup stdout leak-filter window. The
	// filter normally ends as soon as the first real AiT prompt is shown;
	// this cap is a safety valve for hosts where the prompt never installs
	// (e.g. zsh themes resetting PROMPT in precmd) while echo dribbles in
	// slowly over many seconds.
	setupLeakGraceMax = 30 * time.Second
)

func (c *Client) touchRemoteOut() {
	c.lastRemoteOut.Store(time.Now().UnixNano())
}

func remoteOutAt(last int64) time.Time {
	return time.Unix(0, last)
}

// waitRemoteOutputQuietAfter blocks until remote stdout/stderr has been quiet for quiet
// after injectAt, or maxWait elapses. Ignores lastRemoteOut timestamps before injectAt.
// When no post-inject output arrives, returns once quiet has elapsed since injectAt
// (SuppressEcho / test stubs often produce no PTY bytes after the write).
func (c *Client) waitRemoteOutputQuietAfter(injectAt time.Time, quiet, maxWait time.Duration) {
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		last := c.lastRemoteOut.Load()
		if last == 0 || remoteOutAt(last).Before(injectAt) {
			if time.Since(injectAt) >= quiet {
				return
			}
			time.Sleep(remoteOutputQuietPollStep)
			continue
		}
		if time.Since(remoteOutAt(last)) >= quiet {
			return
		}
		time.Sleep(remoteOutputQuietPollStep)
	}
}

// waitRemoteOutputQuietSince returns when no remote output newer than since has occurred for quiet,
// or maxWait elapses. When nothing arrives after since, returns after quiet from since.
func (c *Client) waitRemoteOutputQuietSince(since time.Time, quiet, maxWait time.Duration) {
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		last := c.lastRemoteOut.Load()
		if last != 0 && !remoteOutAt(last).Before(since) {
			if time.Since(remoteOutAt(last)) >= quiet {
				return
			}
		} else if time.Since(since) >= quiet {
			return
		}
		time.Sleep(remoteOutputQuietPollStep)
	}
}

// waitForShellReadyBeforeInject blocks until the remote login shell looks ready for PS1 setup.
// For zsh/oh-my-zsh hosts it waits until a native or AiT prompt line appears in scrollback and
// output has been quiet; bash hosts fall back to quiet-since-attach when no prompt line is seen.
func (c *Client) waitForShellReadyBeforeInject(attachAt time.Time, user, host string, maxWait time.Duration) {
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		scroll := c.scroll.String()
		if containsRemotePromptLine(scroll, user, host) || lastNativePromptLine(scroll) != "" {
			last := c.lastRemoteOut.Load()
			if last != 0 && time.Since(remoteOutAt(last)) >= remoteOutputQuietPeriod {
				return
			}
		} else {
			last := c.lastRemoteOut.Load()
			if last != 0 && !remoteOutAt(last).Before(attachAt) {
				if time.Since(remoteOutAt(last)) >= remoteOutputQuietPeriod {
					return
				}
			} else if time.Since(attachAt) >= remoteOutputQuietPeriod {
				return
			}
		}
		time.Sleep(remoteOutputQuietPollStep)
	}
}

// triggerRemotePrecmd sends a history-ignored blank line to force zsh/bash precmd prompt redraw.
// It does not use Inject so post-inject CaptureMaxWait is avoided.
func (c *Client) triggerRemotePrecmd(stdin io.Writer) {
	if stdin == nil {
		return
	}

	_, _ = io.WriteString(stdin, " \n")
	c.waitRemoteOutputQuietAfter(time.Now(), remoteOutputQuietPeriod, remoteOutputQuietMaxWait)
}

// waitForRemoteAITPrompt polls scrollback until a real AiT remote prompt line appears or maxWait elapses.
// Setup echo containing the marker text does not satisfy the wait.
func (c *Client) waitForRemoteAITPrompt(user, host string, maxWait time.Duration) {
	if c == nil || c.scroll == nil {
		return
	}

	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		if containsRemotePromptLine(c.scroll.String(), user, host) {
			return
		}
		time.Sleep(remoteOutputQuietPollStep)
	}
}

// waitForRemotePromptReady blocks until setup emits the ready sentinel or a real AiT prompt line.
func (c *Client) waitForRemotePromptReady(user, host string, maxWait time.Duration) {
	if c == nil || c.scroll == nil {
		return
	}

	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		if c.remotePromptReadySeen.Load() || containsPromptReadySentinel(c.scroll.String()) {
			c.remotePromptReadySeen.Store(true)
			return
		}
		if containsRemotePromptLine(c.scroll.String(), user, host) {
			return
		}
		time.Sleep(remoteOutputQuietPollStep)
	}
}

// ensureFirstPromptVisible emits a prompt to stdout before setup unpause. Returns whether a line was written.
func (c *Client) ensureFirstPromptVisible(stdin io.Writer, user, host string) bool {
	if c.remotePromptShown.Load() {
		return true
	}

	if c.emitBufferedRemotePrompt(user, host) {
		c.remotePromptShown.Store(true)
		return true
	}

	if c.setupSucceeded() && c.paintClientRemotePrompt(user, host) {
		c.remotePromptShown.Store(true)
		return true
	}

	if c.pauseRemoteOut.Load() && c.emitBufferedRemotePromptIfReady(stdin, user, host) {
		return true
	}

	if c.remotePromptReadySeen.Load() || containsPromptReadySentinel(c.scroll.String()) {
		c.remotePromptReadySeen.Store(true)
		if c.emitBufferedRemotePrompt(user, host) {
			c.remotePromptShown.Store(true)
			return true
		}
		if c.paintClientRemotePrompt(user, host) {
			c.remotePromptShown.Store(true)
			return true
		}
		return c.remotePromptShown.Load()
	}

	c.triggerRemotePrecmd(stdin)
	refreshAt := time.Now()
	c.refreshRemotePrompt(stdin)
	c.waitRemoteOutputQuietSince(refreshAt, remoteOutputQuietPeriod, remoteAITPromptRefreshWait)
	c.waitForRemotePromptReady(user, host, remoteAITPromptRefreshWait)

	if c.emitBufferedRemotePrompt(user, host) {
		c.remotePromptShown.Store(true)
		return true
	}
	if c.setupSucceeded() && c.paintClientRemotePrompt(user, host) {
		c.remotePromptShown.Store(true)
		return true
	}

	if !c.setupSucceeded() && c.emitBufferedNativePrompt() {
		c.remotePromptShown.Store(true)
		return true
	}

	return c.remotePromptShown.Load()
}

// syncRemotePromptBeforeUnpause wakes the remote shell while PTY output is still paused so
// precmd runs and the real prompt is in scrollback before stdout unpause. Client-only paint
// does not put zsh/bash into the readline-ready state on oh-my-zsh hosts.
func (c *Client) syncRemotePromptBeforeUnpause(stdin io.Writer, user, host string) {
	if stdin == nil || !c.remotePromptShown.Load() || !c.setupSucceeded() {
		return
	}

	if containsRemotePromptLine(c.scroll.String(), user, host) {
		if c.emitBufferedRemotePrompt(user, host) {
			return
		}
	}

	c.triggerRemotePrecmd(stdin)
	c.refreshRemotePrompt(stdin)
	refreshAt := time.Now()
	c.waitRemoteOutputQuietSince(refreshAt, remoteOutputQuietPeriod, remoteAITPromptRefreshWait)
	c.waitForRemoteAITPrompt(user, host, remoteAITPromptRefreshWait)
	_ = c.emitBufferedRemotePrompt(user, host)
}

// finishRemoteSetup ends setup suppression, strips leaked setup text from scrollback, refreshes the remote PS1, and unpauses display.
// When tmuxEnabled is true, waits longer for attach redraw to finish before unpausing.
// Unpause is gated on remotePromptShown; client paint is used as a last resort when remote echo is absent.
func (c *Client) finishRemoteSetup(stdin io.Writer, _, _ string, tmuxEnabled bool) {
	user, host := c.promptUser(), c.promptHost()

	if tmuxEnabled {
		c.waitRemoteOutputQuietSince(time.Now(), tmuxAttachQuietPeriod, tmuxAttachQuietMaxWait)
	}

	if tail := c.setupFilter.flush(user, host); len(tail) > 0 {
		_, _ = c.scroll.Write(tail)
	}

	c.waitForRemotePromptReady(user, host, remoteOutputQuietMaxWait)
	if containsPromptReadySentinel(c.scroll.String()) {
		c.remotePromptReadySeen.Store(true)
	}
	quietSince := time.Now()
	if last := c.lastRemoteOut.Load(); last != 0 {
		quietSince = remoteOutAt(last)
	}
	c.waitRemoteOutputQuietSince(quietSince, remoteOutputQuietPeriod, remoteOutputQuietMaxWait)

	c.scroll.StripSubstrings(RemoteSetupStripFragments(user, host, c.remotePromptDirAbs)...)
	if tail := c.setupFilter.flush(user, host); len(tail) > 0 {
		_, _ = c.scroll.Write(tail)
	}

	if !c.ensureFirstPromptVisible(stdin, user, host) && !c.remotePromptShown.Load() {
		if c.paintClientRemotePrompt(user, host) {
			c.remotePromptShown.Store(true)
		}
	}

	if !c.remotePromptShown.Load() {
		c.waitForRemotePromptReady(user, host, remoteAITPromptMaxWait)
		_ = c.ensureFirstPromptVisible(stdin, user, host)
		if !c.remotePromptShown.Load() {
			if c.paintClientRemotePrompt(user, host) {
				c.remotePromptShown.Store(true)
			}
		}
	}

	if c.remotePromptShown.Load() {
		c.syncRemotePromptBeforeUnpause(stdin, user, host)
	}

	if c.setupSucceeded() && !containsRemotePromptLine(c.scroll.String(), user, host) {
		c.paintClientRemotePrompt(user, host)
	}

	c.remoteSetupSuppress.Store(false)
	graceEnd := time.Now().Add(setupLeakGraceMax)
	c.setupLeakGraceUntil.Store(graceEnd.UnixNano())
	c.setupKeystrokeGraceUntil.Store(graceEnd.UnixNano())
	c.setupGraceStripper.reset()

	if !c.remotePromptShown.Load() {
		if c.paintClientRemotePrompt(user, host) {
			c.remotePromptShown.Store(true)
		}
	}
	if c.remotePromptShown.Load() {
		c.remotePromptReadySeen.Store(true)
		c.pauseRemoteOut.Store(false)
	}
}

func drainInteractiveStdin() {
	terminal.DrainStdin()
}

// replayScrollbackToStdout seeds sanitized prior-session scrollback into memory only.
func (c *Client) replayScrollbackToStdout(data string) {
	c.seedReplayScrollback(data, c.promptUser(), c.promptHost())
}

// seedReplayScrollback sanitizes replay data and writes it into the in-memory scroll buffer only.
func (c *Client) seedReplayScrollback(data, user, host string) {
	if data == "" {
		return
	}

	data = SanitizeReplayScrollback(data, user, host)
	if data == "" {
		return
	}

	if !strings.HasSuffix(data, "\n") && !strings.HasSuffix(data, "\r\n") {
		data += "\n"
	}
	_, _ = c.scroll.Write([]byte(data))
}
