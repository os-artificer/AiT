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

import "strings"

// metaPrefixCandidate reports whether s is a strict prefix of any opener.
func metaPrefixCandidate(s string) bool {
	for _, p := range metaPrefixes {
		if len(s) < len(p) && strings.HasPrefix(p, s) {
			return true
		}
	}
	return false
}

// metaPrefixMatch returns the matched opener if s equals one exactly.
func metaPrefixMatch(s string) (string, bool) {
	for _, p := range metaPrefixes {
		if s == p {
			return p, true
		}
	}
	return "", false
}

// metaLineEnd reports whether b terminates a meta line.
func metaLineEnd(b byte) bool {
	return b == '\n' || b == '\r'
}

func (s *metaScanner) applyMetaBackspace() ([]byte, bool) {
	if s.state != scanHolding && s.state != scanMeta {
		s.atLineStart = false
		return nil, false
	}
	if len(s.buf) == 0 {
		return nil, true
	}

	s.buf = s.buf[:len(s.buf)-1]
	if len(s.buf) == 0 {
		s.state = scanIdle
	}

	return []byte{0x08, ' ', 0x08}, true
}

func metaCompleteLine(candidate string, b byte) (line string, done bool) {
	if !metaLineEnd(b) {
		return "", false
	}
	return strings.TrimRight(candidate, "\r\n"), true
}

func (s *metaScanner) feedHoldingLoose(candidate string, b byte, fwd, ech []byte) (outFwd, outEch []byte, line string, hasLine, handled bool) {
	if s.aiAvailable || !s.looseBangCapture(candidate, b) {
		return fwd, ech, "", false, false
	}
	if line, ok := metaCompleteLine(candidate, b); ok {
		s.buf = s.buf[:0]
		s.state = scanIdle
		s.atLineStart = true
		return fwd, ech, line, true, true
	}
	s.buf = append(s.buf, b)
	s.state = scanMeta
	return fwd, append(ech, b), "", false, true
}

func (s *metaScanner) feedHoldingStrict(candidate string, b byte, fwd, ech []byte) (outFwd, outEch []byte, line string, hasLine, handled bool) {
	if _, ok := metaPrefixMatch(candidate); !ok {
		return fwd, ech, "", false, false
	}
	if line, ok := metaCompleteLine(candidate, b); ok {
		s.buf = s.buf[:0]
		s.state = scanIdle
		s.atLineStart = true
		return fwd, ech, line, true, true
	}
	s.buf = append(s.buf, b)
	s.state = scanMeta
	return fwd, append(ech, b), "", false, true
}
