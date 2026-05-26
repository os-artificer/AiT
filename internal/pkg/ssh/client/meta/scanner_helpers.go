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

	s.buf = s.buf[:len(s.buf)-1]
	if len(s.buf) == 0 {
		s.state = stateIdle
	}

	return []byte{0x08, ' ', 0x08}, true
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
