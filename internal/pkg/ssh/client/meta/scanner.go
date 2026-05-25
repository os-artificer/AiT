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

// maxLine caps the bytes captured for a single meta line so a runaway
// paste cannot keep the scanner in meta mode indefinitely.
const maxLine = 4096

// maxLook is the longest possible opener length; bounds the holding buf.
const maxLook = 12 // len("!ait local ")

type scannerState int

const (
	stateIdle scannerState = iota
	stateHolding
	stateMeta
)

// Scanner is the byte-level state machine that detects !ait / !! meta
// lines. It is intentionally free of side effects so it can be unit-tested
// in isolation from the SSH session.
type Scanner struct {
	state       scannerState
	buf         []byte
	atLineStart bool
	aiAvailable bool
}

// New returns a Scanner. aiAvailable controls whether !! and !ait prefixes
// fall back to a "loose" capture mode when the AI client is missing
// (otherwise the bash !! history expansion would leak to the remote shell).
func New(aiAvailable bool) *Scanner {
	return &Scanner{atLineStart: true, aiAvailable: aiAvailable}
}

// Feed consumes one input byte for meta-line detection (!ait, !!).
// forward is sent to the remote shell; echo is shown locally; when hasLine
// is true, line is the completed meta command (with trailing CR/LF stripped).
func (s *Scanner) Feed(b byte) (forward []byte, echo []byte, line string, hasLine bool) {
	return s.feedOne(b, nil, nil)
}

func (s *Scanner) feedOne(b byte, fwd, ech []byte) ([]byte, []byte, string, bool) {
	if echOut, handled := s.applyLineEdit(b); handled {
		return fwd, append(ech, echOut...), "", false
	}

	switch s.state {

	case stateMeta:
		if len(s.buf) >= maxLine {
			fwd = append(fwd, s.buf...)
			s.buf = s.buf[:0]
			s.state = stateIdle
			s.atLineStart = false
			return s.feedOne(b, fwd, ech)
		}
		s.buf = append(s.buf, b)
		ech = append(ech, b)
		if b == '\n' || b == '\r' {
			line := strings.TrimRight(string(s.buf), "\r\n")
			s.buf = s.buf[:0]
			s.state = stateIdle
			s.atLineStart = true
			return fwd, ech, line, true
		}
		return fwd, ech, "", false

	case stateHolding:
		candidate := string(s.buf) + string([]byte{b})
		if f, e, line, hasLine, handled := s.feedHoldingLoose(candidate, b, fwd, ech); handled {
			if hasLine {
				return f, e, line, true
			}
			return f, e, "", false
		}
		if f, e, line, hasLine, handled := s.feedHoldingStrict(candidate, b, fwd, ech); handled {
			if hasLine {
				return f, e, line, true
			}
			return f, e, "", false
		}
		if prefixCandidate(candidate) && len(s.buf)+1 < maxLook {
			s.buf = append(s.buf, b)
			return fwd, append(ech, b), "", false
		}
		fwd = append(fwd, s.buf...)
		s.buf = s.buf[:0]
		s.state = stateIdle
		return s.feedOne(b, fwd, ech)

	default: // stateIdle
		// Always intercept '!' for !ait / !! meta. Requiring atLineStart breaks
		// when the remote prompt has no trailing newline before user input.
		if b == '!' {
			s.buf = append(s.buf[:0], b)
			s.state = stateHolding
			return fwd, append(ech, b), "", false
		}
		fwd = append(fwd, b)
		s.noteForwarded(b)
		return fwd, ech, "", false
	}
}

// looseBangCapture is fail-closed mode when AI is unavailable: capture !! or !ait lines.
func (s *Scanner) looseBangCapture(candidate string, b byte) bool {
	if s.aiAvailable {
		return false
	}

	if strings.HasPrefix(candidate, "!!") {
		return true
	}

	return strings.HasPrefix(candidate, "!ait")
}

func (s *Scanner) applyLineEdit(b byte) (echo []byte, handled bool) {
	switch b {
	case 0x15: // Ctrl+U — kill line; remote should clear too.
		s.buf = s.buf[:0]
		s.state = stateIdle
		s.atLineStart = true
		return nil, false
	case 0x7f, 0x08: // Backspace
		return s.applyMetaBackspace()

	case '\r':
		s.atLineStart = true
	}
	return nil, false
}

func (s *Scanner) noteForwarded(b byte) {
	switch b {
	case '\n', '\r':
		s.atLineStart = true
	case 0x15:
		s.atLineStart = true
	case 0x7f, 0x08:
		s.atLineStart = false
	default:
		s.atLineStart = false
	}
}
