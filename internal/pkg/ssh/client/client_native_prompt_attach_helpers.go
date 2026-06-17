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

// finishNativePromptAfterAttach re-emits the native shell prompt after load/switch attach.
// Grace filtering may hold the prompt in setupGraceStripper.pending without writing scrollback;
// this path waits, then emits from scrollback or flushes grace pending.
func (c *Client) finishNativePromptAfterAttach() {
	if c == nil || c.remotePromptShown.Load() {
		return
	}

	user, host := c.promptUser(), c.promptHost()
	deadline := time.Now().Add(remoteAITPromptMaxWait)

	for time.Now().Before(deadline) {
		if c.remotePromptShown.Load() {
			return
		}
		if c.tryEmitNativePromptIfNeeded(user, host) {
			return
		}

		until := c.setupLeakGraceUntil.Load()
		if until != 0 && time.Now().UnixNano() >= until {
			break
		}

		time.Sleep(remoteOutputQuietPollStep)
	}

	if c.remotePromptShown.Load() {
		return
	}

	c.tryEmitNativePromptIfNeeded(user, host)
}

// tryEmitNativePromptIfNeeded emits a native prompt from grace pending or scrollback.
// Returns true when the prompt was written and remotePromptShown was set.
func (c *Client) tryEmitNativePromptIfNeeded(user, host string) bool {
	if c == nil || c.remotePromptShown.Load() || !UseRemoteSystemPrompt() {
		return false
	}

	if c.emitNativePromptFromGracePending(user, host) {
		c.remotePromptShown.Store(true)
		return true
	}
	if c.emitBufferedNativePrompt() {
		c.remotePromptShown.Store(true)
		return true
	}
	return false
}

// chunkContainsNativePrompt reports whether p includes a complete native shell prompt line.
func chunkContainsNativePrompt(p []byte, user, host string) bool {
	if len(p) == 0 {
		return false
	}
	normalized := strings.ReplaceAll(string(p), "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	for _, line := range strings.Split(normalized, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		if isSetupLeakLine(line, user, host) {
			continue
		}
		if isNativePromptLike(line) || isLikelyShellPromptLine(line) {
			return true
		}
	}
	trimmed := strings.TrimRight(string(p), "\r\n")
	if trimmed != "" && !isSetupLeakLine(trimmed, user, host) {
		if isNativePromptLike(trimmed) || isLikelyShellPromptLine(trimmed) {
			return true
		}
	}
	return false
}

// emitNativePromptFromGracePending writes a native prompt held in grace pending to stdout.
func (c *Client) emitNativePromptFromGracePending(user, host string) bool {
	raw := c.setupGraceStripper.flushPendingNativePrompt(user, host)
	if len(raw) == 0 {
		return false
	}

	cleaned := extractNativePromptForEmit(string(raw))
	if cleaned == "" {
		return false
	}

	out := formatActivePromptStdout(cleaned)
	_, err := c.writeInteractiveStdout([]byte(out))
	return err == nil
}
