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

// metaPrefixes is the exhaustive set of strict meta openers. The scanner
// only commits to meta mode once a candidate buffer matches one exactly.
var metaPrefixes = []string{
	"!ait ai ",
	"!ait ai\n",
	"!ait local ",
	"!ait local\n",
	"!ait remote\n",
	"!ait ",
	"!! ",
	"!!\n",
}

// metaMaxLine caps the bytes captured for a single meta line so a runaway
// paste cannot keep the scanner in meta mode indefinitely.
const metaMaxLine = 4096

// metaMaxLook is the longest possible opener length; bounds the holding buf.
const metaMaxLook = 12 // len("!ait local ")

type scannerState int

const (
	scanIdle scannerState = iota
	scanHolding
	scanMeta
)

// metaScanner is the byte-level state machine that detects !ait / !! meta
// lines. It is intentionally free of side effects so it can be unit-tested
// in isolation from the SSH session.
type metaScanner struct {
	state       scannerState
	buf         []byte
	atLineStart bool
	aiAvailable bool
}

func newMetaScanner(aiAvailable bool) *metaScanner {
	return &metaScanner{atLineStart: true, aiAvailable: aiAvailable}
}

// Feed consumes one input byte for meta-line detection (!ait, !!).
// forward is sent to the remote shell; echo is shown locally; when hasLine is true, line is the completed meta command.
func (s *metaScanner) Feed(b byte) (forward []byte, echo []byte, line string, hasLine bool) {
	return s.feedOne(b, nil, nil)
}

func (s *metaScanner) feedOne(b byte, fwd, ech []byte) ([]byte, []byte, string, bool) {
	if echOut, handled := s.applyLineEdit(b); handled {
		return fwd, append(ech, echOut...), "", false
	}

	switch s.state {

	case scanMeta:
		if len(s.buf) >= metaMaxLine {
			fwd = append(fwd, s.buf...)
			s.buf = s.buf[:0]
			s.state = scanIdle
			s.atLineStart = false
			return s.feedOne(b, fwd, ech)
		}
		s.buf = append(s.buf, b)
		ech = append(ech, b)
		if b == '\n' || b == '\r' {
			line := strings.TrimRight(string(s.buf), "\r\n")
			s.buf = s.buf[:0]
			s.state = scanIdle
			s.atLineStart = true
			return fwd, ech, line, true
		}
		return fwd, ech, "", false

	case scanHolding:
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
		if metaPrefixCandidate(candidate) && len(s.buf)+1 < metaMaxLook {
			s.buf = append(s.buf, b)
			return fwd, append(ech, b), "", false
		}
		fwd = append(fwd, s.buf...)
		s.buf = s.buf[:0]
		s.state = scanIdle
		return s.feedOne(b, fwd, ech)

	default: // scanIdle
		// Always intercept '!' for !ait / !! meta. Requiring atLineStart breaks
		// when the remote prompt has no trailing newline before user input.
		if b == '!' {
			s.buf = append(s.buf[:0], b)
			s.state = scanHolding
			return fwd, append(ech, b), "", false
		}
		fwd = append(fwd, b)
		s.noteForwarded(b)
		return fwd, ech, "", false
	}
}

// looseBangCapture is fail-closed mode when AI is unavailable: capture !! or !ait lines.
func (s *metaScanner) looseBangCapture(candidate string, b byte) bool {
	if s.aiAvailable {
		return false
	}

	if strings.HasPrefix(candidate, "!!") {
		return true
	}

	return strings.HasPrefix(candidate, "!ait")
}

func (s *metaScanner) applyLineEdit(b byte) (echo []byte, handled bool) {
	switch b {
	case 0x15: // Ctrl+U — kill line; remote should clear too.
		s.buf = s.buf[:0]
		s.state = scanIdle
		s.atLineStart = true
		return nil, false
	case 0x7f, 0x08: // Backspace
		return s.applyMetaBackspace()

	case '\r':
		s.atLineStart = true
	}
	return nil, false
}

func (s *metaScanner) noteForwarded(b byte) {
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
