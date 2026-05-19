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

package sshclient

import "strings"

// metaPrefixes is the exhaustive set of strict meta openers. The scanner
// only commits to meta mode once a candidate buffer matches one exactly.
var metaPrefixes = []string{"\\ait ", "\\ait\n", "!! ", "!!\n"}

// metaMaxLine caps the bytes captured for a single meta line so a runaway
// paste cannot keep the scanner in meta mode indefinitely.
const metaMaxLine = 4096

// metaMaxLook is the longest possible opener length; bounds the holding buf.
const metaMaxLook = 6 // len("\\ait ")

type scannerState int

const (
	scanIdle scannerState = iota
	scanHolding
	scanMeta
)

// metaScanner is the byte-level state machine that detects \ait / !! meta
// lines. It is intentionally free of side effects so it can be unit-tested
// in isolation from the SSH session.
type metaScanner struct {
	state       scannerState
	buf         []byte
	atLineStart bool
}

func newMetaScanner() *metaScanner {
	return &metaScanner{atLineStart: true}
}

// Feed consumes one byte. It returns the bytes that must be forwarded to the
// remote shell stdin (in order), and a completed meta line if the byte
// closed one. The returned forward slice is only valid until the next call.
func (s *metaScanner) Feed(b byte) (forward []byte, line string, hasLine bool) {
	return s.feedOne(b, nil)
}

func (s *metaScanner) feedOne(b byte, fwd []byte) ([]byte, string, bool) {
	switch s.state {
	case scanMeta:
		if len(s.buf) >= metaMaxLine {
			// Runaway line: bail out, forward what we'd buffered as raw bytes.
			fwd = append(fwd, s.buf...)
			s.buf = s.buf[:0]
			s.state = scanIdle
			s.atLineStart = false
			return s.feedOne(b, fwd)
		}
		s.buf = append(s.buf, b)
		if b == '\n' || b == '\r' {
			line := strings.TrimRight(string(s.buf), "\r\n")
			s.buf = s.buf[:0]
			s.state = scanIdle
			s.atLineStart = true
			return fwd, line, true
		}
		return fwd, "", false

	case scanHolding:
		candidate := string(s.buf) + string([]byte{b})
		if _, ok := metaPrefixMatch(candidate); ok {
			// Openers that end in newline (e.g. "!!\n") are themselves
			// the entire meta line, so dispatch immediately without entering
			// scanMeta state.
			if b == '\n' || b == '\r' {
				line := strings.TrimRight(candidate, "\r\n")
				s.buf = s.buf[:0]
				s.state = scanIdle
				s.atLineStart = true
				return fwd, line, true
			}
			s.buf = append(s.buf, b)
			s.state = scanMeta
			return fwd, "", false
		}
		if metaPrefixCandidate(candidate) && len(s.buf)+1 < metaMaxLook {
			s.buf = append(s.buf, b)
			return fwd, "", false
		}
		// Not a meta opener — flush held bytes and re-process current byte.
		fwd = append(fwd, s.buf...)
		s.buf = s.buf[:0]
		s.state = scanIdle
		return s.feedOne(b, fwd)

	default: // scanIdle
		if s.atLineStart && (b == '\\' || b == '!') {
			s.buf = append(s.buf[:0], b)
			s.state = scanHolding
			return fwd, "", false
		}
		fwd = append(fwd, b)
		s.atLineStart = b == '\n' || b == '\r'
		return fwd, "", false
	}
}

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
