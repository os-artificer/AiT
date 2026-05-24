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

package terminal

import (
	"bufio"
	"io"
	"unicode"
)

const (
	ctrlA = 1
	ctrlE = 5
	ctrlK = 11
	ctrlU = 21
	ctrlW = 23
)

type escapeKey int

const (
	escapeNone escapeKey = iota
	escapeUp
	escapeDown
	escapeLeft
	escapeRight
	escapeHome
	escapeEnd
)

type editState struct {
	runes   []rune
	cursor  int
	draft   []rune
	histIdx int
}

func newEditState() editState {
	return editState{histIdx: -1}
}

func (s *editState) setLine(line string) {
	s.runes = []rune(line)
	s.cursor = len(s.runes)
	s.histIdx = -1
}

func (s *editState) insertRune(out io.Writer, prompt string, r rune) {
	s.histIdx = -1
	atMidLine := s.cursor < len(s.runes)
	s.runes = append(s.runes[:s.cursor], append([]rune{r}, s.runes[s.cursor:]...)...)
	s.cursor++
	if atMidLine {
		redrawLine(out, prompt, s.runes, s.cursor)
		return
	}
	echoRune(out, r)
}

func (s *editState) deleteBeforeCursor(out io.Writer, prompt string) {
	if s.cursor <= 0 {
		return
	}
	s.cursor = eraseRuneBeforeCursor(out, s.runes, s.cursor)
	s.runes = append(s.runes[:s.cursor], s.runes[s.cursor+1:]...)
	redrawLine(out, prompt, s.runes, s.cursor)
}

func (s *editState) moveCursor(out io.Writer, prompt string, pos int) {
	if pos < 0 {
		pos = 0
	}
	if pos > len(s.runes) {
		pos = len(s.runes)
	}
	s.cursor = pos
	redrawLine(out, prompt, s.runes, s.cursor)
}

func (s *editState) killLine(out io.Writer, prompt string) {
	s.runes = s.runes[:0]
	s.cursor = 0
	redrawLine(out, prompt, s.runes, s.cursor)
}

func (s *editState) killToEnd(out io.Writer, prompt string) {
	s.runes = s.runes[:s.cursor]
	redrawLine(out, prompt, s.runes, s.cursor)
}

func (s *editState) killWordBefore(out io.Writer, prompt string) {
	if s.cursor == 0 {
		return
	}
	end := s.cursor
	start := end
	for start > 0 && unicode.IsSpace(s.runes[start-1]) {
		start--
	}
	for start > 0 && !unicode.IsSpace(s.runes[start-1]) {
		start--
	}
	s.runes = append(s.runes[:start], s.runes[end:]...)
	s.cursor = start
	redrawLine(out, prompt, s.runes, s.cursor)
}

func (s *editState) historyUp(out io.Writer, prompt string, hist *LineHistory) {
	if hist == nil || hist.Len() == 0 {
		return
	}
	if s.histIdx == -1 {
		s.draft = append([]rune(nil), s.runes...)
		s.histIdx = hist.Len() - 1
	} else if s.histIdx > 0 {
		s.histIdx--
	}
	s.runes = []rune(hist.At(s.histIdx))
	s.cursor = len(s.runes)
	redrawLine(out, prompt, s.runes, s.cursor)
}

func (s *editState) historyDown(out io.Writer, prompt string, hist *LineHistory) {
	if hist == nil || s.histIdx == -1 {
		return
	}
	if s.histIdx < hist.Len()-1 {
		s.histIdx++
		s.runes = []rune(hist.At(s.histIdx))
	} else {
		s.histIdx = -1
		s.runes = append([]rune(nil), s.draft...)
	}
	s.cursor = len(s.runes)
	redrawLine(out, prompt, s.runes, s.cursor)
}

func parseEscapeSeq(br *bufio.Reader) (escapeKey, error) {
	b, err := br.ReadByte()
	if err != nil {
		return escapeNone, err
	}
	if b != csiIntroducerByte {
		return escapeNone, nil
	}

	var digits []byte
	for {
		c, err := br.ReadByte()
		if err != nil {
			return escapeNone, err
		}
		if c >= '0' && c <= '9' {
			digits = append(digits, c)
			continue
		}
		if c == '~' {
			switch string(digits) {
			case "1":
				return escapeHome, nil
			case "4":
				return escapeEnd, nil
			}
			return escapeNone, nil
		}
		if c >= csiFinalByteMin && c <= csiFinalByteMax {
			switch c {
			case 'A':
				return escapeUp, nil
			case 'B':
				return escapeDown, nil
			case 'C':
				return escapeRight, nil
			case 'D':
				return escapeLeft, nil
			case 'H':
				return escapeHome, nil
			case 'F':
				return escapeEnd, nil
			}
			return escapeNone, nil
		}
	}
}

func handleEscapeKey(out io.Writer, prompt string, st *editState, hist *LineHistory, key escapeKey) {
	switch key {
	case escapeUp:
		st.historyUp(out, prompt, hist)
	case escapeDown:
		st.historyDown(out, prompt, hist)
	case escapeLeft:
		st.moveCursor(out, prompt, st.cursor-1)
	case escapeRight:
		st.moveCursor(out, prompt, st.cursor+1)
	case escapeHome:
		st.moveCursor(out, prompt, 0)
	case escapeEnd:
		st.moveCursor(out, prompt, len(st.runes))
	}
}
