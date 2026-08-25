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

import (
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// maxLine caps the bytes captured for a single meta line so a runaway
// paste cannot keep the scanner in meta mode indefinitely.
const maxLine = 4096

// maxLook is the longest possible opener length; bounds the holding buf.
const maxLook = 14 // len("!!ait remote\n") typo + len("!ait remote\n")

type scannerState int

const (
	stateIdle scannerState = iota
	stateHolding
	stateMeta
)

type escParseState int

const (
	escNone escParseState = iota
	escReceived
	escCSI
)

const (
	keyESC            = 27
	csiIntroducerByte = '['
	csiFinalByteMin   = 0x40
	csiFinalByteMax   = 0x7e
)

// Scanner is the byte-level state machine that detects !ait / !! meta lines.
// It performs no I/O; display context (prompt prefix, local-shell mode) is
// injected via SyncFrom so tests can drive Feed in isolation.
type Scanner struct {
	state            scannerState
	buf              []byte
	cursor           int
	atLineStart      bool
	escState         escParseState
	escParams        []byte
	echoPrefix       string
	localShellActive bool
	deferLocalEcho   bool
	holdingRedraw    bool
}

// New returns a Scanner configured for meta-line detection.
// echoPrefix is prepended on carriage-return line redraws (remote/local prompt).
func New(echoPrefix string) *Scanner {
	return &Scanner{atLineStart: true, echoPrefix: echoPrefix}
}

// Feed consumes one input byte for meta-line detection (!ait, !!).
// forward is sent to the remote shell; echo is shown locally; when hasLine
// is true, line is the completed meta command (with trailing CR/LF stripped).
func (s *Scanner) Feed(b byte) (forward []byte, echo []byte, line string, hasLine bool) {
	return s.feedOne(b, nil, nil)
}

func (s *Scanner) feedOne(b byte, fwd, ech []byte) ([]byte, []byte, string, bool) {
	if s.inEditState() {
		if s.escState != escNone {
			if echOut, handled := s.consumeEscForEdit(b); handled {
				return fwd, append(ech, echOut...), "", false
			}
		}
		if b == keyESC {
			s.escState = escReceived
			return fwd, ech, "", false
		}
	}

	if echOut, handled := s.applyMetaLineEdit(b); handled {
		return fwd, append(ech, echOut...), "", false
	}

	switch s.state {

	case stateMeta:
		if len(s.buf) >= maxLine {
			fwd = append(fwd, s.buf...)
			s.buf = s.buf[:0]
			s.cursor = 0
			s.enterIdle()
			s.atLineStart = false
			return s.feedOne(b, fwd, ech)
		}
		if b == '\n' || b == '\r' {
			line := finalizeMetaLine(string(s.buf))
			s.buf = s.buf[:0]
			s.cursor = 0
			s.enterIdle()
			s.atLineStart = true
			ech = append(ech, b)
			return fwd, ech, line, true
		}
		ech = s.insertMetaByte(b, ech)
		return fwd, ech, "", false

	case stateHolding:
		candidate := string(s.buf) + string([]byte{b})
		if f, e, handled := s.tryLocalBangPassthrough(candidate, b, fwd, ech); handled {
			return f, e, "", false
		}
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
			s.cursor = len(s.runes())
			if s.shouldDeferHoldingEcho(b) {
				if s.localShellActive {
					return fwd, ech, "", false
				}
				return fwd, s.noteHoldingRedraw(ech), "", false
			}
			if s.localShellActive {
				return fwd, s.holdingAppendEcho(ech, b), "", false
			}
			return fwd, s.noteHoldingRedraw(ech), "", false
		}
		ech, fwd = s.flushHoldingToForward(ech, fwd)
		s.buf = s.buf[:0]
		s.cursor = 0
		s.enterIdle()
		return s.feedOne(b, fwd, ech)

	default: // stateIdle
		if b == '!' {
			s.buf = append(s.buf[:0], b)
			s.cursor = 1
			s.state = stateHolding
			s.deferLocalEcho = true
			return fwd, ech, "", false
		}
		if b == 0x15 { // Ctrl+U idle — forward to remote shell.
			s.atLineStart = true
			fwd = append(fwd, b)
			return fwd, ech, "", false
		}
		fwd = append(fwd, b)
		s.noteForwarded(b)
		return fwd, ech, "", false
	}
}

// looseBangCapture captures !ait meta lines (space/newline after prefix) while
// forwarding !! and non-meta !words (e.g. !aitssh) to the remote shell.
func (s *Scanner) looseBangCapture(candidate string, b byte) bool {
	if candidate == toolname.MetaPrefix {
		return true
	}
	return strings.HasPrefix(candidate, toolname.MetaPrefix+" ") ||
		strings.HasPrefix(candidate, toolname.MetaPrefix+"\n") ||
		strings.HasPrefix(candidate, toolname.MetaPrefix+"\r")
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
