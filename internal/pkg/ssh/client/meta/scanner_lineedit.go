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
	"unicode"

	"github.com/os-artificer/AiT/internal/pkg/terminal/lineedit"
)

const (
	ctrlA = 1
	ctrlE = 5
	ctrlK = 11
	ctrlL = 12
	ctrlU = 21
	ctrlW = 23
)

type metaEscapeKey int

const (
	metaEscNone metaEscapeKey = iota
	metaEscLeft
	metaEscRight
	metaEscHome
	metaEscEnd
	metaEscDelete
)

// enterIdle resets holding/meta edit flags when returning to stateIdle.
func (s *Scanner) enterIdle() {
	s.state = stateIdle
	s.deferLocalEcho = false
	s.holdingRedraw = false
}

// SyncFrom updates per-client context used for local passthrough and line redraw.
// The SSH client calls this before every Feed byte so echoPrefix can track the live prompt.
// localShellActive enables !! passthrough to the local PTY instead of meta capture.
func (s *Scanner) SyncFrom(localShellActive bool, echoPrefix string) {
	s.localShellActive = localShellActive
	s.echoPrefix = echoPrefix
}

func (s *Scanner) runes() []rune {
	return []rune(string(s.buf))
}

func (s *Scanner) setRunes(r []rune) {
	s.buf = []byte(string(r))
	if s.cursor > len(r) {
		s.cursor = len(r)
	}
}

func (s *Scanner) redrawEcho() []byte {
	return lineedit.AppendRedrawLine(nil, s.echoPrefix, s.runes(), s.cursor)
}

func (s *Scanner) clearHoldingLineEcho() []byte {
	return lineedit.AppendRedrawLine(nil, s.echoPrefix, nil, 0)
}

func (s *Scanner) noteHoldingRedraw(ech []byte) []byte {
	if s.localShellActive {
		return ech
	}
	s.holdingRedraw = true
	return append(ech, s.redrawEcho()...)
}

func (s *Scanner) flushHoldingToForward(ech, fwd []byte) ([]byte, []byte) {
	if s.holdingRedraw {
		ech = s.clearHoldingLineEcho()
		s.holdingRedraw = false
	}
	fwd = append(fwd, s.buf...)
	return ech, fwd
}

func (s *Scanner) redrawEchoAfterClear() []byte {
	return lineedit.AppendClearAndRedrawLine(nil, s.echoPrefix, s.runes(), s.cursor)
}

func (s *Scanner) inEditState() bool {
	return s.state == stateHolding || s.state == stateMeta
}

func (s *Scanner) tryLocalBangPassthrough(candidate string, b byte, fwd, ech []byte) ([]byte, []byte, bool) {
	if !s.localShellActive || !strings.HasPrefix(candidate, "!!") {
		return fwd, ech, false
	}
	if isBangAitTypoPrefix(candidate) {
		return fwd, ech, false
	}
	// Hold at exactly "!!" until the next byte disambiguates !!ait meta typo vs shell !!.
	if candidate == "!!" {
		s.buf = append(s.buf, b)
		s.cursor = len(s.runes())
		if s.shouldDeferHoldingEcho(b) {
			return fwd, ech, true
		}
		return fwd, s.holdingAppendEcho(ech, b), true
	}

	fwd = append(fwd, s.buf...)
	fwd = append(fwd, b)
	s.buf = s.buf[:0]
	s.cursor = 0
	s.enterIdle()
	s.atLineStart = false
	return fwd, ech, true
}

func (s *Scanner) holdingAppendEcho(ech []byte, b byte) []byte {
	if s.localShellActive && s.deferLocalEcho {
		s.deferLocalEcho = false
		return append(ech, s.buf...)
	}
	return append(ech, b)
}

func (s *Scanner) shouldDeferHoldingEcho(b byte) bool {
	return s.deferLocalEcho && b == '!'
}

// echoEnteringStateMeta transitions to stateMeta after buf already contains b.
func (s *Scanner) echoEnteringStateMeta(ech []byte, b byte) []byte {
	s.state = stateMeta
	s.deferLocalEcho = false
	s.holdingRedraw = false
	if s.localShellActive {
		return s.holdingAppendEcho(ech, b)
	}
	return append(ech, s.redrawEcho()...)
}

func (s *Scanner) applyMetaLineEdit(b byte) (echo []byte, handled bool) {
	if !s.inEditState() {
		return nil, false
	}

	switch b {
	case ctrlL:
		return s.redrawEchoAfterClear(), true
	case ctrlA:
		s.cursor = 0
		return s.redrawEcho(), true
	case ctrlE:
		s.cursor = len(s.runes())
		return s.redrawEcho(), true
	case ctrlU:
		s.buf = s.buf[:0]
		s.cursor = 0
		s.enterIdle()
		s.atLineStart = true
		return s.redrawEcho(), true
	case ctrlK:
		rs := s.runes()
		s.setRunes(rs[:s.cursor])
		return s.redrawEcho(), true
	case ctrlW:
		return s.killWordBefore()
	case 0x7f, 0x08:
		return s.deleteBeforeCursor()
	}
	return nil, false
}

func (s *Scanner) killWordBefore() ([]byte, bool) {
	rs := s.runes()
	if s.cursor == 0 {
		return nil, true
	}

	end := s.cursor
	start := end
	for start > 0 && unicode.IsSpace(rs[start-1]) {
		start--
	}
	for start > 0 && !unicode.IsSpace(rs[start-1]) {
		start--
	}

	s.setRunes(append(rs[:start], rs[end:]...))
	s.cursor = start
	return s.redrawEcho(), true
}

func (s *Scanner) deleteBeforeCursor() ([]byte, bool) {
	if s.cursor <= 0 {
		return nil, true
	}

	rs := s.runes()
	s.setRunes(append(rs[:s.cursor-1], rs[s.cursor:]...))
	s.cursor--
	return s.redrawEcho(), true
}

func (s *Scanner) insertMetaByte(b byte, ech []byte) []byte {
	if isUTF8MultiByteStart(b) {
		n := s.stripTrailingPinyinAtCursor()
		for i := 0; i < n; i++ {
			ech = append(ech, 0x08, ' ', 0x08)
		}
	}

	rs := s.runes()
	atEnd := s.cursor == len(rs)
	if atEnd {
		s.buf = append(s.buf, b)
		s.cursor = len(s.runes())
		return append(ech, b)
	}

	r := rune(b)
	newRunes := append(rs[:s.cursor], append([]rune{r}, rs[s.cursor:]...)...)
	s.setRunes(newRunes)
	s.cursor++
	return append(ech, s.redrawEcho()...)
}

func (s *Scanner) consumeEscForEdit(b byte) (echo []byte, handled bool) {
	switch s.escState {
	case escReceived:
		if b == csiIntroducerByte {
			s.escState = escCSI
			s.escParams = s.escParams[:0]
			return nil, true
		}
		s.escState = escNone
		return nil, true
	case escCSI:
		if b >= csiFinalByteMin && b <= csiFinalByteMax {
			key := parseMetaCSIKey(string(s.escParams), b)
			s.escState = escNone
			s.escParams = s.escParams[:0]
			return s.applyMetaEscapeKey(key)
		}
		if b == '~' {
			key := parseMetaCSIKey(string(s.escParams), b)
			s.escState = escNone
			s.escParams = s.escParams[:0]
			return s.applyMetaEscapeKey(key)
		}
		if (b >= '0' && b <= '9') || b == ';' || b == ':' || b == '<' || b == '=' || b == '>' || b == '?' {
			s.escParams = append(s.escParams, b)
			return nil, true
		}
		s.escState = escNone
		s.escParams = s.escParams[:0]
		return nil, true
	}

	s.escState = escNone
	return nil, true
}

func parseMetaCSIKey(params string, final byte) metaEscapeKey {
	if final == '~' {
		switch params {
		case "1", "7":
			return metaEscHome
		case "3":
			return metaEscDelete
		case "4", "8":
			return metaEscEnd
		}
		return metaEscNone
	}

	switch final {
	case 'D':
		return metaEscLeft
	case 'C':
		return metaEscRight
	case 'H':
		return metaEscHome
	case 'F':
		return metaEscEnd
	case 'P':
		return metaEscDelete
	}
	return metaEscNone
}

func (s *Scanner) applyMetaEscapeKey(key metaEscapeKey) ([]byte, bool) {
	switch key {
	case metaEscLeft:
		if s.cursor > 0 {
			s.cursor--
		}
		return s.redrawEcho(), true
	case metaEscRight:
		if s.cursor < len(s.runes()) {
			s.cursor++
		}
		return s.redrawEcho(), true
	case metaEscHome:
		s.cursor = 0
		return s.redrawEcho(), true
	case metaEscEnd:
		s.cursor = len(s.runes())
		return s.redrawEcho(), true
	case metaEscDelete:
		return s.deleteBeforeCursor()
	default:
		return nil, true
	}
}

// isUTF8MultiByteStart reports whether b is the first byte of a multi-byte
// UTF-8 encoded character (CJK, emoji, etc.).
func isUTF8MultiByteStart(b byte) bool {
	return b >= 0xC2 && b <= 0xF4
}
