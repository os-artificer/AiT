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

package meta

import "strings"

// AIPromptFromLine extracts the user prompt from a meta line that targets the
// remote AI REPL (!! ..., !ait ai ...). Returns ("", false) for other meta lines.
func AIPromptFromLine(line string) (prompt string, isAI bool) {
	switch {
	case strings.HasPrefix(line, "!!"):
		return strings.TrimSpace(strings.TrimPrefix(line, "!!")), true
	case strings.HasPrefix(line, PrefixAitAI):
		return strings.TrimSpace(strings.TrimPrefix(line, PrefixAitAI)), true
	default:
		return "", false
	}
}

// prefixCandidate reports whether s is a strict prefix of any opener.
func prefixCandidate(s string) bool {
	for _, p := range prefixes {
		if len(s) < len(p) && strings.HasPrefix(p, s) {
			return true
		}
	}
	return false
}

// prefixMatch returns the matched opener if s equals one exactly.
func prefixMatch(s string) (string, bool) {
	for _, p := range prefixes {
		if s == p {
			return p, true
		}
	}
	return "", false
}

// lineEnd reports whether b terminates a meta line.
func lineEnd(b byte) bool {
	return b == '\n' || b == '\r'
}

func (s *Scanner) applyMetaBackspace() ([]byte, bool) {
	if s.state != stateHolding && s.state != stateMeta {
		s.atLineStart = false
		return nil, false
	}
	if len(s.buf) == 0 {
		return nil, true
	}

	// Delete a complete UTF-8 character rather than a single byte.
	n := 1
	for n < len(s.buf) && s.buf[len(s.buf)-n]&0xC0 == 0x80 {
		n++
	}
	s.buf = s.buf[:len(s.buf)-n]
	if len(s.buf) == 0 {
		s.state = stateIdle
	}

	return []byte{0x08, ' ', 0x08}, true
}

// consumeEscInMeta drops escape/CSI sequences received while in stateMeta so
// that terminal control sequences (including IME cursor updates) do not pollute
// the meta command buffer.
func (s *Scanner) consumeEscInMeta(b byte, fwd, ech []byte) ([]byte, []byte, string, bool) {
	switch s.escState {
	case escReceived:
		if b == csiIntroducerByte {
			s.escState = escCSI
			return fwd, ech, "", false
		}
		// Simple escape sequence (ESC + one char) — consume and resume.
		s.escState = escNone
		return fwd, ech, "", false
	case escCSI:
		if b >= csiFinalByteMin && b <= csiFinalByteMax {
			// CSI sequence complete.
			s.escState = escNone
			return fwd, ech, "", false
		}
		if (b >= '0' && b <= '9') || b == ';' || b == ':' || b == '<' || b == '=' || b == '>' || b == '?' {
			// Intermediate/parameter bytes — keep consuming.
			return fwd, ech, "", false
		}
		// Unexpected byte — abort escape parsing.
		s.escState = escNone
		return fwd, ech, "", false
	}
	s.escState = escNone
	return fwd, ech, "", false
}

// isUTF8MultiByteStart reports whether b is the first byte of a multi-byte
// UTF-8 encoded character (CJK, emoji, etc.).
func isUTF8MultiByteStart(b byte) bool {
	return b >= 0xC2 && b <= 0xF4
}

// stripTrailingPinyin removes consecutive lowercase ASCII letters from the end
// of s.buf and returns how many were removed. This heuristic clears IME preedit
// (pinyin) text that precedes a confirmed Chinese character in raw TTY mode.
func (s *Scanner) stripTrailingPinyin() int {
	i := len(s.buf)
	for i > 0 {
		c := s.buf[i-1]
		if c >= 'a' && c <= 'z' {
			i--
		} else {
			break
		}
	}
	removed := len(s.buf) - i
	if removed > 0 {
		s.buf = s.buf[:i]
	}
	return removed
}

func completeLine(candidate string, b byte) (line string, done bool) {
	if !lineEnd(b) {
		return "", false
	}
	return strings.TrimRight(candidate, "\r\n"), true
}

func (s *Scanner) feedHoldingLoose(candidate string, b byte, fwd, ech []byte) (outFwd, outEch []byte, line string, hasLine, handled bool) {
	if s.aiAvailable || !s.looseBangCapture(candidate, b) {
		return fwd, ech, "", false, false
	}
	if line, ok := completeLine(candidate, b); ok {
		s.buf = s.buf[:0]
		s.state = stateIdle
		s.atLineStart = true
		return fwd, ech, line, true, true
	}
	s.buf = append(s.buf, b)
	s.state = stateMeta
	return fwd, append(ech, b), "", false, true
}

func (s *Scanner) feedHoldingStrict(candidate string, b byte, fwd, ech []byte) (outFwd, outEch []byte, line string, hasLine, handled bool) {
	if _, ok := prefixMatch(candidate); !ok {
		return fwd, ech, "", false, false
	}
	if line, ok := completeLine(candidate, b); ok {
		s.buf = s.buf[:0]
		s.state = stateIdle
		s.atLineStart = true
		return fwd, ech, line, true, true
	}
	s.buf = append(s.buf, b)
	s.state = stateMeta
	return fwd, append(ech, b), "", false, true
}
