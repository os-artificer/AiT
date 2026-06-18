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

// containsPromptReadySentinel reports whether s contains the setup ready handshake marker.
func containsPromptReadySentinel(s string) bool {
	return strings.Contains(s, prompt.RemotePromptReadySentinel)
}

// containsRemotePromptLine reports whether s contains a real AiT remote prompt line.
// Unlike a marker substring check, setup echo that merely contains the prompt
// prefix (e.g. an echoed "export PS1='...'") does not match.
func containsRemotePromptLine(s, user, host string) bool {
	return lastRemotePromptLine(s, user, host) != ""
}

// lastNativePromptLine returns the last scrollback line that looks like a native shell prompt.
func lastNativePromptLine(scroll string) string {
	return lastNativePromptLineForHost(scroll, "")
}

// lastNativePromptLineForHost returns the last native prompt whose user@host matches activeHost.
// When activeHost is empty, any native prompt line may match.
func lastNativePromptLineForHost(scroll, activeHost string) string {
	if scroll == "" {
		return ""
	}
	if line := findLastNativePromptRaw(scroll, activeHost); line != "" {
		return line
	}
	return ""
}

// emitBufferedRemotePromptIfReady emits a buffered AiT prompt or client paint while pauseRemoteOut holds
// setup output in scrollback. It does not trigger precmd or native fallback.
// Returns true when a prompt line was written to stdout.
func (c *Client) emitBufferedRemotePromptIfReady(stdin io.Writer, user, host string) bool {
	if !c.pauseRemoteOut.Load() {
		return false
	}

	if containsRemotePromptLine(c.scroll.String(), user, host) {
		if c.emitBufferedRemotePrompt(user, host) {
			c.remotePromptShown.Store(true)
			return true
		}
	}

	if c.setupSucceeded() {
		if c.paintClientRemotePrompt(user, host) {
			c.remotePromptShown.Store(true)
			return true
		}
	}

	c.refreshRemotePrompt(stdin)
	c.waitForRemoteAITPrompt(user, host, remoteAITPromptRefreshWait)
	if c.emitBufferedRemotePrompt(user, host) {
		c.remotePromptShown.Store(true)
		return true
	}
	if c.setupSucceeded() && c.paintClientRemotePrompt(user, host) {
		c.remotePromptShown.Store(true)
		return true
	}
	return false
}

// formatActivePromptStdout formats a prompt for stdout with the cursor left on the same line.
// Clears to end of line so typed command echo (e.g. "exit") cannot linger after repaint.
func formatActivePromptStdout(line string) string {
	line = strings.TrimRight(line, "\r\n")
	return "\r" + line + promptClearToEOL
}

// lastNativePromptLineFallback returns the last native prompt when dial-host filtering
// misses the genuine shell PS1 (e.g. root@devhost:~# while dial target is remote.example.com).
// It skips stale inner-shell prompts left in scrollback after nested ssh closes.
func lastNativePromptLineFallback(scroll, activeHost string) string {
	if staleInnerNativePromptAfterNestedClose(scroll, activeHost) {
		return ""
	}
	for _, raw := range iterNativePromptCandidates(scroll) {
		if cleaned := extractNativePromptForEmit(raw); cleaned != "" {
			return raw
		}
	}
	return ""
}

// staleInnerNativePromptAfterNestedClose reports scrollback that still ends with an inner
// native prompt after a nested ssh session closed back to the outer dial target.
func staleInnerNativePromptAfterNestedClose(scroll, activeHost string) bool {
	if scroll == "" {
		return false
	}
	normalized := strings.ReplaceAll(scroll, "\r\n", "\n")
	if !strings.Contains(normalized, "Connection to") || !strings.Contains(normalized, "closed") {
		return false
	}
	line := lastNativePromptLine(scroll)
	if line == "" {
		return false
	}
	return !nativePromptMatchesActiveHost(line, activeHost)
}

// emitBufferedNativePrompt writes the last native shell prompt from scrollback to stdout.
// Returns true when a native prompt line was found and written.
func (c *Client) emitBufferedNativePrompt() bool {
	if c == nil || c.scroll == nil {
		return false
	}

	activeHost := c.promptHost()
	scroll := c.scroll.String()
	line := lastNativePromptLineForHost(scroll, activeHost)
	if line == "" {
		line = lastNativePromptLineFallback(scroll, activeHost)
	}
	if line == "" {
		return false
	}

	cleaned := extractNativePromptForEmit(line)
	if cleaned == "" {
		return false
	}

	out := formatActivePromptStdout(cleaned)
	if _, err := c.writeInteractiveStdout([]byte(out)); err != nil {
		return false
	}
	return true
}

// lastPromptCandidateLine returns the last prompt-like line in chunk using relaxed heuristics.
func lastPromptCandidateLine(chunk string) string {
	for _, raw := range iterNativePromptCandidates(chunk) {
		if cleaned := extractNativePromptForEmit(raw); cleaned != "" {
			return raw
		}
	}
	return ""
}

// emitScrollDeltaPrompt writes a prompt line captured in scrollback since beforeLen to stdout.
func (c *Client) emitScrollDeltaPrompt(beforeLen int) bool {
	if c == nil || c.scroll == nil {
		return false
	}
	scroll := c.scroll.String()
	if len(scroll) <= beforeLen {
		return false
	}
	line := lastPromptCandidateLine(scroll[beforeLen:])
	if line == "" {
		return false
	}
	cleaned := extractNativePromptForEmit(line)
	if cleaned == "" {
		return false
	}
	out := formatActivePromptStdout(cleaned)
	_, err := c.writeInteractiveStdout([]byte(out))
	return err == nil
}

// emitBufferedRemotePrompt writes the buffered remote prompt from scrollback to stdout.
// Returns true when a prompt line was found and written.
func (c *Client) emitBufferedRemotePrompt(user, host string) bool {
	line := lastRemotePromptLine(c.scroll.String(), user, host)
	if line == "" {
		return false
	}

	out := formatActivePromptStdout(line)
	if _, err := c.writeInteractiveStdout([]byte(out)); err != nil {
		return false
	}
	return true
}
