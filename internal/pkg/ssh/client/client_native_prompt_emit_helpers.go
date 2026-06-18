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
	"regexp"
	"strings"
	"time"
)

var (
	zshRightPromptTailRE = regexp.MustCompile(`\x1b\[\d+C\[\d{1,2}:\d{2}:\d{2}\]\x1b\[\d+D.*$`)
	ansiBracketedPasteRE = regexp.MustCompile(`\x1b\[\?[0-9]+[hl]`)
)

// nativeRemotePromptRedrawCmd refreshes the remote prompt without print -Pn, which leaves
// stray zsh fragments on stdout in native-prompt mode.
const nativeRemotePromptRedrawCmd = `[ -z "$ZSH_VERSION" ] && true` +
	`; [ -n "$ZSH_VERSION" ] && { zle 2>/dev/null && zle reset-prompt 2>/dev/null; }`

// findLastNativePromptRaw returns the last scrollback prompt fragment for activeHost.
func findLastNativePromptRaw(scroll, activeHost string) string {
	for _, raw := range iterNativePromptCandidates(scroll) {
		if !nativePromptMatchesActiveHost(raw, activeHost) {
			continue
		}
		if cleaned := extractNativePromptForEmit(raw); cleaned != "" {
			return raw
		}
	}
	return ""
}

// iterNativePromptCandidates yields scrollback fragments newest-first, splitting on newlines
// and carriage returns so zsh line overwrites remain searchable.
func iterNativePromptCandidates(scroll string) []string {
	if scroll == "" {
		return nil
	}

	normalized := strings.ReplaceAll(scroll, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	var out []string
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\r")
		for j := len(parts) - 1; j >= 0; j-- {
			part := strings.TrimSpace(parts[j])
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

// extractNativePromptForEmit strips terminal mode, OSC, and zsh right-prompt noise from a
// scrollback prompt fragment before writing it to stdout.
func extractNativePromptForEmit(raw string) string {
	if raw == "" {
		return ""
	}

	if idx := strings.LastIndex(raw, "\r"); idx >= 0 {
		raw = raw[idx+1:]
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	if strings.HasPrefix(strings.TrimSpace(stripANSIForPromptMatch(raw)), "AiT ") {
		return ""
	}

	raw = stripScrollbackControlPrefix(raw)
	if raw == "" {
		return ""
	}

	if strings.HasPrefix(strings.TrimSpace(stripANSIForPromptMatch(raw)), "AiT ") {
		return ""
	}

	raw = ansiBracketedPasteRE.ReplaceAllString(raw, "")
	raw = ansiOSCEscapeBELRE.ReplaceAllString(raw, "")
	raw = ansiOSCEscapeSTRE.ReplaceAllString(raw, "")

	if at := strings.LastIndex(raw, "@"); at >= 0 {
		start := nativePromptUserStart(raw, at)
		raw = raw[start:]
	}

	raw = zshRightPromptTailRE.ReplaceAllString(raw, "")
	for _, cut := range []string{"\x1b[?2004", "\x1b[?1h", "\x1b[?1l", "\x1b="} {
		if idx := strings.Index(raw, cut); idx >= 0 {
			raw = raw[:idx]
		}
	}
	raw = strings.TrimRight(raw, " \t")
	raw = strings.TrimSuffix(raw, "\x1b[K")
	raw = strings.TrimRight(raw, " \t")
	if !isLikelyShellPromptLine(raw) {
		return ""
	}
	return raw
}

// nativePromptUserStart returns the index of the first byte of the username in a prompt fragment.
func nativePromptUserStart(raw string, at int) int {
	if at <= 0 {
		return 0
	}

	start := at
	for start > 0 {
		prev := raw[start-1]
		if prev == '\x1b' || prev == '[' || prev == ']' || prev == ';' {
			break
		}
		if (prev >= 'a' && prev <= 'z') || (prev >= 'A' && prev <= 'Z') ||
			(prev >= '0' && prev <= '9') || prev == '_' || prev == '-' {
			start--
			continue
		}
		break
	}

	// Drop CSI/SGR tails split across scrollback chunks (e.g. "33mroot@...", "\x1b[01;33m").
	start = skipSplitSGRTail(raw, start, at)

	return includeANSIColorPrefix(raw, start)
}

// skipSplitSGRTail advances past CSI parameters and an SGR "m" terminator before the username.
func skipSplitSGRTail(raw string, start, at int) int {
	for start < at && start < len(raw) {
		ch := raw[start]
		if ch >= '0' && ch <= '9' || ch == ';' || ch == '?' {
			start++
			continue
		}
		if ch == 'm' && hasSGRTailBefore(raw, start) {
			start++
			continue
		}
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			break
		}
		start++
	}
	return start
}

// hasSGRTailBefore reports whether index mIndex points at an SGR "m" after CSI parameters.
func hasSGRTailBefore(raw string, mIndex int) bool {
	if mIndex <= 0 || mIndex >= len(raw) || raw[mIndex] != 'm' {
		return false
	}
	j := mIndex - 1
	for j >= 0 && (raw[j] >= '0' && raw[j] <= '9' || raw[j] == ';' || raw[j] == '?') {
		j--
	}
	return j < mIndex-1
}

// includeANSIColorPrefix extends start backward to include one SGR sequence immediately before username.
func includeANSIColorPrefix(raw string, start int) int {
	if start <= 0 || raw[start-1] != 'm' {
		return start
	}
	for j := start - 2; j >= 0; j-- {
		if raw[j] == '\x1b' {
			return j
		}
	}
	return start
}

// stripScrollbackControlPrefix removes clear-screen CSI and split-chunk CSI tails before prompt emit.
func stripScrollbackControlPrefix(raw string) string {
	for {
		before := raw
		raw = strings.TrimLeft(raw, "\b \t")
		for _, prefix := range []string{"\x1b[2J", "\x1b[H", "\x1b[J", "\x0c"} {
			raw = strings.TrimPrefix(raw, prefix)
		}
		raw = trimPartialCSIPrefix(raw)
		if raw == before {
			break
		}
		raw = strings.TrimSpace(raw)
	}
	return raw
}

// trimPartialCSIPrefix drops CSI tails left by scrollback chunk splits (e.g. "2Jroot@...").
// Requires numeric/parameter bytes before the final letter so "AiT ..." is not mangled.
func trimPartialCSIPrefix(raw string) string {
	paramEnd := 0
	for paramEnd < len(raw) && (raw[paramEnd] >= '0' && raw[paramEnd] <= '9' ||
		raw[paramEnd] == ';' || raw[paramEnd] == '?') {
		paramEnd++
	}
	if paramEnd == 0 {
		return raw
	}

	i := paramEnd
	for i < len(raw) && raw[i] >= 'A' && raw[i] <= 'Z' {
		i++
	}
	if i <= paramEnd {
		return raw
	}
	if i < len(raw) {
		next := raw[i]
		if (next >= 'a' && next <= 'z') || next == '_' {
			return raw[i:]
		}
	}
	return raw
}

// syncNativeRemoteShellAfterMeta wakes the remote shell with a native-safe redraw inject.
func (c *Client) syncNativeRemoteShellAfterMeta(stdin io.Writer) {
	if stdin == nil || c.localShell.Active() {
		return
	}

	c.triggerRemotePrecmd(stdin)
	_ = c.Inject(stdin, nativeRemotePromptRedrawCmd, InjectOpts{SuppressEcho: true})
	refreshAt := time.Now()
	c.waitRemoteOutputQuietSince(refreshAt, remoteOutputQuietPeriod, remoteAITPromptRefreshWait)
}
