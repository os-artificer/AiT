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
	"bytes"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/prompt"
)

// setupLeakFragments are literal substrings that identify AiT-injected setup
// echo (PS1 export, history toggles, pwd probe markers, capture delimiters).
// Lines containing any of them are dropped from stdout/scrollback filters.
var setupLeakFragments = []string{
	"stty -echo",
	"stty echo",
	"set +o history",
	"set -o history",
	"_AIT_HIST",
	"HISTFILE=/dev/null",
	".__AITPWD",
	"AITPWDEND",
	`basename "${PWD`,
	"${PWD/#$HOME",
	`"${PWD:-/}"`,
	`printf '%s' "${d:-/}"`,
	`&& printf ":%s"`,
	"git rev-parse --abbrev-ref HEAD",
	"AIT_CAPTURE_B_",
	"AIT_CAPTURE_E_",
	"export HISTCONTROL=ignorespace:erasedups",
	"setopt HIST_IGNORE_SPACE",
	"PROMPT_SUBST",
	"printf '\\033[1A",
	`[ -z "$ZSH_VERSION" ] && true`,
	`print -Pn "$PROMPT"`,
	"ait_prompt_command()",
	"export PROMPT_COMMAND=ait_prompt_command",
	`print -Pn "$PROMPT"`,
	"ait_precmd()",
	"precmd_functions=(${precmd_functions:#ait_precmd} ait_precmd)",
	"zle reset-prompt",
	"dev/null",
	"HISTFILE=",
	"unset _AIT_HIST",
	"precmd_functions",
	prompt.RemotePromptReadySentinel,
}

// nativePromptSuffixes are trailing terminators accepted by isNativePromptLike.
var nativePromptSuffixes = []string{"->", ">", "$", "#", "%", "→", "➜", "❯"}

const nativePromptMinLen = 8

type setupOutputFilter struct {
	pending []byte
}

func nextSetupFilterLine(pending []byte) (line string, rest []byte, ok bool) {
	idx := bytes.IndexByte(pending, '\n')
	if idx < 0 {
		idx = bytes.IndexByte(pending, '\r')
	}
	if idx < 0 {
		return "", pending, false
	}

	line = strings.TrimRight(string(pending[:idx]), "\r")
	rest = pending[idx+1:]
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	}
	return line, rest, true
}

func (f *setupOutputFilter) filter(p []byte, user, host string) []byte {
	if len(p) == 0 {
		return nil
	}
	f.pending = append(f.pending, p...)
	var out bytes.Buffer
	for {
		line, rest, ok := nextSetupFilterLine(f.pending)
		if !ok {
			break
		}
		f.pending = rest
		if !isSetupLeakLine(line, user, host) {
			out.WriteString(line)
			out.WriteByte('\n')
		}
	}

	// A real prompt line arrives without a trailing newline; release it
	// immediately so waitForRemoteAITPrompt converges without extra delay.
	if pending := string(f.pending); pending != "" {
		if prompt.IsRemotePromptLine(user, host, pending) {
			out.WriteString(pending)
			f.pending = nil
		}
	}
	return out.Bytes()
}

func (f *setupOutputFilter) flush(user, host string) []byte {
	if len(f.pending) == 0 {
		return nil
	}
	line := strings.TrimRight(string(f.pending), "\r\n")
	f.pending = nil
	if isSetupLeakLine(line, user, host) {
		return nil
	}
	return []byte(line)
}

func filterSetupLeakBytes(p []byte, user, host string) []byte {
	var f setupOutputFilter
	out := f.filter(p, user, host)
	if tail := f.flush(user, host); len(tail) > 0 {
		out = append(out, tail...)
	}
	return out
}

type setupStdoutStripper struct {
	pending []byte
}

// nextSetupStdoutLine splits one complete line from pending, returning raw bytes (content + original separator).
func nextSetupStdoutLine(pending []byte) (line string, kept []byte, rest []byte, ok bool) {
	idxNL := bytes.IndexByte(pending, '\n')
	idxCR := bytes.IndexByte(pending, '\r')
	if idxNL < 0 && idxCR < 0 {
		return "", nil, pending, false
	}

	var end int
	if idxNL >= 0 && (idxCR < 0 || idxNL <= idxCR) {
		end = idxNL + 1
		if idxCR >= 0 && idxCR == idxNL-1 {
			line = string(pending[:idxCR])
		} else {
			line = string(pending[:idxNL])
		}
	} else {
		line = string(pending[:idxCR])
		end = idxCR + 1
		if idxNL >= 0 && idxNL == idxCR+1 {
			end = idxNL + 1
		}
	}

	kept = pending[:end]
	rest = pending[end:]
	return line, kept, rest, true
}

func (s *setupStdoutStripper) strip(user, host string) []byte {
	if len(s.pending) == 0 {
		return nil
	}

	var out bytes.Buffer
	for {
		line, kept, rest, ok := nextSetupStdoutLine(s.pending)
		if !ok {
			break
		}
		s.pending = rest
		if !isSetupLeakLine(line, user, host) {
			out.Write(kept)
		}
	}
	return out.Bytes()
}

// flushPendingPrompt emits a complete remote prompt held in pending without a trailing newline.
func (s *setupStdoutStripper) flushPendingPrompt(user, host string) []byte {
	if len(s.pending) == 0 {
		return nil
	}
	line := strings.TrimRight(string(s.pending), "\r\n")
	if isSetupLeakLine(line, user, host) {
		return nil
	}
	if !prompt.IsRemotePromptLine(user, host, line) {
		return nil
	}
	out := append([]byte(nil), s.pending...)
	s.pending = nil
	return out
}

// pendingNativePromptCandidate returns the last native prompt line held in pending,
// including when MOTD or other earlier lines share the buffer.
func (s *setupStdoutStripper) pendingNativePromptCandidate(user, host string) string {
	if len(s.pending) == 0 {
		return ""
	}
	line := lastPromptCandidateLine(string(s.pending))
	if line == "" {
		return ""
	}
	if isSetupLeakLine(line, user, host) {
		return ""
	}
	if extractNativePromptForEmit(line) == "" {
		return ""
	}
	return line
}

// flushPendingNativePrompt emits pending bytes when they form a complete native
// shell prompt (e.g. "root@host [~] →" without a trailing newline), so the user
// is never left without a prompt during the grace window. Single characters and
// other wrapped setup-echo fragments never match and stay held.
func (s *setupStdoutStripper) flushPendingNativePrompt(user, host string) []byte {
	line := s.pendingNativePromptCandidate(user, host)
	if line == "" {
		return nil
	}
	s.pending = nil
	return []byte(line)
}

// filterChunk appends p and returns stdout-safe bytes for complete non-leak lines.
// Pending bytes flush only when they form a complete AiT or native shell prompt;
// wrapped setup-echo fragments (including single characters) stay held until the
// line completes or the grace window ends.
func (s *setupStdoutStripper) filterChunk(p []byte, user, host string, allowNative bool) []byte {
	if len(p) == 0 {
		return nil
	}
	s.pending = append(s.pending, p...)
	out := s.strip(user, host)

	if promptBytes := s.flushPendingPrompt(user, host); len(promptBytes) > 0 {
		return append(out, promptBytes...)
	}
	if allowNative {
		if nativeBytes := s.flushPendingNativePrompt(user, host); len(nativeBytes) > 0 {
			return append(out, nativeBytes...)
		}
	}
	return out
}

// peekPendingNativePrompt reports whether pending holds a complete native shell prompt.
func (s *setupStdoutStripper) peekPendingNativePrompt(user, host string) bool {
	return s.pendingNativePromptCandidate(user, host) != ""
}

func (s *setupStdoutStripper) reset() {
	s.pending = nil
}

// endGrace drops incomplete setup-echo fragments when the grace window ends.
// A complete native shell prompt in pending is flushed so load/switch never
// leaves the user without a visible prompt.
func (s *setupStdoutStripper) endGrace(user, host string) []byte {
	if native := s.flushPendingNativePrompt(user, host); len(native) > 0 {
		return native
	}
	s.pending = nil
	return nil
}

func (s *setupStdoutStripper) flushStdout(user, host string) []byte {
	if len(s.pending) == 0 {
		return nil
	}
	line := strings.TrimRight(string(s.pending), "\r\n")
	tail := append([]byte(nil), s.pending...)
	s.pending = nil
	if isSetupLeakLine(line, user, host) {
		return nil
	}
	return tail
}

// stripSetupLeakStdout removes setup leak lines from stdout-bound PTY output while preserving original line endings.
func stripSetupLeakStdout(p []byte, user, host string) []byte {
	var s setupStdoutStripper
	s.pending = append(s.pending, p...)
	out := s.strip(user, host)
	if tail := s.flushStdout(user, host); len(tail) > 0 {
		out = append(out, tail...)
	}
	return out
}

// StripSetupLeakLines removes AiT setup/probe echo lines (PS1 export, stty
// toggles, pwd probe markers and wrapped fragments) from recorded scrollback
// data while preserving the original line endings of the remaining lines.
// user and host scope the prompt-specific patterns; data is returned unchanged
// when it contains no leak lines.
func StripSetupLeakLines(data, user, host string) string {
	if data == "" {
		return ""
	}
	return string(stripSetupLeakStdout([]byte(data), user, host))
}

// SanitizeReplayScrollback strips setup leaks, prompt lines, and collapses excessive blank lines for load/switch seeding.
func SanitizeReplayScrollback(data, user, host string) string {
	if data == "" {
		return ""
	}
	data = StripSetupLeakLines(data, user, host)
	data = stripReplayPromptLines(data, user, host)
	return collapseBlankRuns(data)
}

// stripReplayPromptLines removes AiT and native shell prompt lines from saved scrollback seed data.
func stripReplayPromptLines(data, user, host string) string {
	if data == "" {
		return ""
	}

	normalized := strings.ReplaceAll(data, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	lines := strings.Split(normalized, "\n")

	var out []string
	for _, line := range lines {
		if prompt.IsRemotePromptLine(user, host, line) || isNativePromptLike(line) {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// collapseBlankRuns folds consecutive empty lines to at most one blank line.
func collapseBlankRuns(data string) string {
	if data == "" {
		return ""
	}

	normalized := strings.ReplaceAll(data, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	lines := strings.Split(normalized, "\n")

	var out []string
	blankRun := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			blankRun++
			if blankRun <= 1 {
				out = append(out, line)
			}
			continue
		}
		blankRun = 0
		out = append(out, line)
	}

	return strings.Join(out, "\n")
}

func isSetupLeakLine(line, user, host string) bool {
	if strings.TrimSpace(line) == "" {
		return false
	}
	if isScrollNoiseLine(line) {
		return true
	}
	if isSetupLeakFragment(line) {
		return true
	}
	prefix := prompt.RemotePrefix(user, host)
	if strings.Contains(line, prefix) && !prompt.IsRemotePromptLine(user, host, line) {
		return true
	}
	if strings.Contains(line, "export PS1='") && strings.Contains(line, prefix) {
		return true
	}
	if strings.Contains(line, "export PROMPT='") && strings.Contains(line, prefix) {
		return true
	}
	return false
}

func isSetupLeakFragment(s string) bool {
	for _, frag := range setupLeakFragments {
		if strings.Contains(s, frag) {
			return true
		}
	}
	return false
}

// isNativePromptLike reports whether line looks like a complete native shell
// prompt (e.g. "root@remote [~] →" or "user@host:~$"). It requires a
// user@host marker, a minimum length, and a trailing prompt terminator so
// wrapped setup-echo fragments and single characters never match.
func isNativePromptLike(line string) bool {
	line = strings.TrimSpace(trimRightPromptSegments(stripANSIForPromptMatch(line)))
	if len(line) < nativePromptMinLen || !strings.Contains(line, "@") {
		return false
	}
	if strings.HasPrefix(line, "AiT ") {
		return false
	}
	for _, suffix := range nativePromptSuffixes {
		if strings.HasSuffix(line, suffix) {
			return true
		}
	}
	return false
}

// isLikelyShellPromptLine is a relaxed native-prompt heuristic for bash PS1 variants
// that isNativePromptLike rejects (short prompts, bracketed styles, ANSI-heavy lines).
func isLikelyShellPromptLine(line string) bool {
	if isNativePromptLike(line) {
		return true
	}
	plain := strings.TrimSpace(stripANSIForPromptMatch(line))
	if len(plain) < 5 || strings.HasPrefix(plain, "AiT ") || !strings.Contains(plain, "@") {
		return false
	}
	for _, suffix := range nativePromptSuffixes {
		if strings.HasSuffix(plain, suffix) {
			return true
		}
	}
	runes := []rune(plain)
	switch runes[len(runes)-1] {
	case '#', '$', '>', '%':
		return true
	default:
		return false
	}
}

// nativePromptHostInLine extracts the host label from a native prompt like "root@host [~] ->".
func nativePromptHostInLine(line string) string {
	plain := strings.TrimSpace(stripANSIForPromptMatch(line))
	at := strings.Index(plain, "@")
	if at < 0 {
		return ""
	}
	rest := plain[at+1:]
	end := len(rest)
	for i, r := range rest {
		if r == ' ' || r == '[' || r == ':' {
			end = i
			break
		}
	}
	return strings.TrimSpace(rest[:end])
}

// nativePromptMatchesActiveHost reports whether line's user@host matches activeHost.
// When activeHost is empty, any native prompt line matches.
func nativePromptMatchesActiveHost(line, activeHost string) bool {
	if activeHost == "" {
		return true
	}
	h := nativePromptHostInLine(line)
	if h == "" {
		return false
	}
	return hostsEquivalent(h, activeHost)
}

// isScrollNoiseLine reports whether line is a setup-echo fragment too short to
// carry meaning: ZLE line redraw leaves lone characters ("#", "S", "a") on
// their own lines when echoing wrapped setup commands.
func isScrollNoiseLine(line string) bool {
	return len([]rune(strings.TrimSpace(line))) == 1
}

// filterRemoteOutputChunk applies setup suppression and duplicate-prompt dropping before scroll/stdout.
// Suppression ends only on a real AiT prompt line (not setup echo containing the marker text).
// Returns nil when the chunk should not be forwarded.
func (c *Client) filterRemoteOutputChunk(chunk []byte, user, host string) []byte {
	out, _ := c.filterRemoteOutputChunkRepaint(chunk, user, host)
	return out
}

// isWhitespaceOnlyPTYOutput reports whether chunk is only terminal whitespace
// (newlines/spaces) with no visible shell content.
func isWhitespaceOnlyPTYOutput(chunk []byte) bool {
	if len(chunk) == 0 {
		return false
	}
	for _, b := range chunk {
		switch b {
		case '\r', '\n', ' ', '\t':
			continue
		default:
			return false
		}
	}
	return true
}

func isDuplicatePromptOutput(p []byte, user, host string) bool {
	if len(p) == 0 {
		return false
	}
	normalized := strings.ReplaceAll(string(p), "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	lines := strings.Split(normalized, "\n")
	for _, line := range lines {
		if isDuplicatePromptLine(line, user, host) {
			return true
		}
	}
	return false
}
