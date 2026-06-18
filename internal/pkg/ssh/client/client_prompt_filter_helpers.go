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
	"regexp"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/prompt"
)

var ansiPromptEscapeRE = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)
var ansiOSCEscapeBELRE = regexp.MustCompile(`\x1b\][^\x07]*\x07`)
var ansiOSCEscapeSTRE = regexp.MustCompile(`\x1b\][^\x1b]*\x1b\\`)
var trailingRightPromptRE = regexp.MustCompile(`\s+\[[^\]]+\]\s*$`)

// stripANSIForPromptMatch removes CSI color/style sequences before prompt heuristics.
func stripANSIForPromptMatch(s string) string {
	s = ansiPromptEscapeRE.ReplaceAllString(s, "")
	s = ansiOSCEscapeBELRE.ReplaceAllString(s, "")
	s = ansiOSCEscapeSTRE.ReplaceAllString(s, "")
	return s
}

func trimRightPromptSegments(s string) string {
	s = strings.TrimSpace(s)
	for {
		next := trailingRightPromptRE.ReplaceAllString(s, "")
		if next == s {
			return s
		}
		s = strings.TrimSpace(next)
	}
}

// aitPromptManaged reports whether AiT PS1 setup completed and should own the prompt.
func (c *Client) aitPromptManaged() bool {
	if c == nil {
		return false
	}
	return c.setupSucceeded()
}

// shouldStripNativePrompt reports whether native shell prompt lines should be removed
// from PTY stdout. True after setup sentinel or once an AiT prompt was painted locally.
func (c *Client) shouldStripNativePrompt() bool {
	if UseRemoteSystemPrompt() {
		return false
	}

	if c == nil {
		return false
	}
	return c.aitPromptManaged() || c.remotePromptShown.Load()
}

// stripUnwantedPromptLines removes prompt-echo lines from a PTY chunk without dropping
// unrelated output on other lines. managed drops native shell prompts; dropEcho also
// removes duplicate AiT prompt lines already shown locally. The bool is true when any
// prompt line was removed (caller should repaint the AiT prompt locally).
func stripUnwantedPromptLines(p []byte, user, host string, managed, dropEcho bool) ([]byte, bool) {
	if len(p) == 0 {
		return p, false
	}

	normalized := strings.ReplaceAll(string(p), "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	lines := strings.Split(normalized, "\n")

	var promptRemoved bool
	var kept []string
	for _, line := range lines {
		plain := stripANSIForPromptMatch(line)
		if containsClearTokens(line) {
			kept = append(kept, line)
			continue
		}
		if managed && isNativePromptLike(plain) {
			promptRemoved = true
			continue
		}
		if dropEcho && isDuplicatePromptLine(line, user, host) {
			promptRemoved = true
			continue
		}
		kept = append(kept, line)
	}
	if !promptRemoved {
		return p, false
	}

	for len(kept) > 0 && strings.TrimSpace(kept[len(kept)-1]) == "" {
		kept = kept[:len(kept)-1]
	}
	if len(kept) == 0 {
		return nil, true
	}

	out := strings.Join(kept, "\n")
	if strings.HasSuffix(string(p), "\r\n") && !strings.HasSuffix(out, "\r\n") {
		out += "\r\n"
	} else if strings.HasSuffix(string(p), "\n") && !strings.HasSuffix(out, "\n") && !strings.HasSuffix(out, "\r\n") {
		out += "\n"
	}
	return []byte(out), true
}

// removeManagedNativePromptLines drops native shell prompt lines from PTY output when
// AiT setup owns the prompt. Returns the filtered bytes and whether any line was removed.
func removeManagedNativePromptLines(p []byte) ([]byte, bool) {
	out, nativeRemoved := stripUnwantedPromptLines(p, "", "", true, false)
	return out, nativeRemoved
}

// repaintInteractivePrompt writes the current AiT prompt to stdout without a trailing newline.
func (c *Client) repaintInteractivePrompt(user, host string) {
	if c == nil || user == "" || host == "" {
		return
	}
	_ = c.emitMetaPromptStdout(user, host)
}

// repaintAfterNestedSSHClose restores the outer-shell prompt after a nested ssh exits.
func (c *Client) repaintAfterNestedSSHClose(user, host string) {
	if c == nil || user == "" || host == "" {
		return
	}

	if UseRemoteSystemPrompt() {
		if c.emitBufferedNativePrompt() {
			c.remotePromptShown.Store(true)
		}
		return
	}

	if stdin := c.interactiveStdin; stdin != nil && !c.localShell.Active() {
		c.triggerRemotePrecmd(stdin)
		c.refreshRemotePrompt(stdin)
	}

	if c.emitMetaPromptStdout(user, host) {
		c.remotePromptShown.Store(true)
	}
}

// filterRemoteOutputChunkRepaint applies setup suppression, managed-native stripping, and
// duplicate-prompt dropping. The bool is true when a managed native prompt was stripped
// and the AiT prompt should be repainted locally.
func (c *Client) filterRemoteOutputChunkRepaint(chunk []byte, user, host string) ([]byte, bool) {
	repaint := false

	if c.remoteSetupSuppress.Load() {
		chunk = c.setupFilter.filter(chunk, user, host)
		if c.setupInjectDone.Load() && containsRemotePromptLine(string(chunk), user, host) {
			c.remoteSetupSuppress.Store(false)
		}
	}
	if len(chunk) == 0 {
		return nil, false
	}

	if c.shouldStripNativePrompt() {
		var promptStripped bool
		chunk, promptStripped = stripUnwantedPromptLines(
			chunk, user, host, c.shouldStripNativePrompt(), c.remotePromptShown.Load(),
		)
		if promptStripped {
			repaint = true
		}
		if len(chunk) == 0 {
			return nil, repaint
		}
	}
	return chunk, repaint
}

func containsClearTokens(s string) bool {
	return strings.Contains(s, "\x0c") || strings.Contains(s, "\x1b[2J") || strings.Contains(s, "\x1b[H")
}

// isDuplicatePromptLine reports whether a single line is a duplicate prompt echo.
func isDuplicatePromptLine(line, user, host string) bool {
	line = strings.TrimSpace(stripANSIForPromptMatch(line))
	if line == "" {
		return false
	}
	if prompt.IsRemotePromptLine(user, host, line) {
		return true
	}
	return isNativePromptLike(line)
}
