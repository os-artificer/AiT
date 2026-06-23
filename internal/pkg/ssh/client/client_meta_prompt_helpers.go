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
	"time"
)

const metaRecoverWhitespaceWindow = 2 * time.Second

// resyncRemotePS1AfterMeta forces remote precmd and PS1 redraw after local-only meta
// output. Unlike syncRemotePromptBeforeUnpause, it does not skip when scroll already
// holds an AiT prompt line — meta does not run commands on the remote shell, so zsh
// readline must be re-aligned before the next keystroke.
func (c *Client) resyncRemotePS1AfterMeta(stdin io.Writer, user, host string) {
	if stdin == nil || c.localShell.Active() || !c.remotePromptShown.Load() {
		return
	}

	c.triggerRemotePrecmd(stdin)
	c.refreshRemotePrompt(stdin)
	refreshAt := time.Now()
	c.waitRemoteOutputQuietSince(refreshAt, remoteOutputQuietPeriod, remoteAITPromptRefreshWait)
	c.waitForRemoteAITPrompt(user, host, remoteAITPromptRefreshWait)
}

// syncRemoteShellAfterMeta wakes the remote shell and injects a prompt redraw while
// meta output is still local-only. Used after !ait meta commands complete.
func (c *Client) syncRemoteShellAfterMeta(stdin io.Writer) {
	if stdin == nil || c.localShell.Active() {
		return
	}

	c.triggerRemotePrecmd(stdin)
	c.refreshRemotePrompt(stdin)
	refreshAt := time.Now()
	c.waitRemoteOutputQuietSince(refreshAt, remoteOutputQuietPeriod, remoteAITPromptRefreshWait)
}

// emitMetaPromptStdout writes the active remote prompt to stdout without a trailing newline.
// Remote redraw echo is dropped as a duplicate once remotePromptShown is set.
func (c *Client) emitMetaPromptStdout(user, host string) bool {
	if c == nil {
		return false
	}
	if UseRemoteSystemPrompt() {
		return c.emitBufferedNativePrompt()
	}
	if c.scroll != nil && c.emitBufferedRemotePrompt(user, host) {
		return true
	}
	if (c.setupSucceeded() || c.remotePromptReadySeen.Load()) && c.paintClientRemotePrompt(user, host) {
		return true
	}
	if c.shouldStripNativePrompt() {
		return false
	}
	return c.emitBufferedNativePrompt()
}

// finishMetaInteraction ends local meta output and repaints the remote shell prompt.
// When the prompt was already shown, remote PS1 is synced while stdout stays paused
// (same as login unpause) so the first keystroke does not redraw oh-my-zsh native prompts.
func (c *Client) finishMetaInteraction(stdin io.Writer) {
	user, host := c.promptUser(), c.promptHost()

	recoverUntil := time.Now().Add(metaRecoverWhitespaceWindow)
	c.metaRecoverWhitespaceUntil.Store(recoverUntil.UnixNano())
	c.setupKeystrokeGraceUntil.Store(recoverUntil.UnixNano())
	c.setupLeakGraceUntil.Store(recoverUntil.UnixNano())

	// Native prompt mode: prefer scrollback re-emit (no remote poke). When scrollback
	// lacks a recognizable prompt (host mismatch, bash PS1 variant, fresh login),
	// fall back to a silent remote redraw captured while stdout stays paused.
	if UseRemoteSystemPrompt() {
		c.finishNativePromptAfterMeta(stdin)
		return
	}

	if c.remotePromptShown.Load() {
		if stdin != nil && !c.localShell.Active() {
			wasPaused := c.pauseRemoteOut.Load()
			if !wasPaused {
				c.pauseRemoteOut.Store(true)
			}
			c.resyncRemotePS1AfterMeta(stdin, user, host)
			if !wasPaused {
				c.pauseRemoteOut.Store(false)
			}
		}
		c.repaintInteractivePrompt(user, host)
		return
	}

	c.syncRemoteShellAfterMeta(stdin)
	c.waitForRemoteAITPrompt(user, host, remoteAITPromptRefreshWait)
	c.repaintInteractivePrompt(user, host)
}

// finishNativePromptAfterMeta re-shows the remote shell prompt after local-only meta output.
// It tries scrollback re-emit first, then a silent remote redraw when no prompt is found.
func (c *Client) finishNativePromptAfterMeta(stdin io.Writer) bool {
	if c.emitBufferedNativePrompt() {
		return true
	}
	if stdin == nil || c.localShell.Active() {
		return false
	}

	beforeLen := 0
	if c.scroll != nil {
		beforeLen = len(c.scroll.String())
	}
	c.syncNativeRemoteShellAfterMeta(stdin)
	if c.emitBufferedNativePrompt() {
		return true
	}
	return c.emitScrollDeltaPrompt(beforeLen)
}
