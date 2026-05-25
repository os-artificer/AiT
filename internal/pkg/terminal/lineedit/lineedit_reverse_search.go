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

package lineedit

import (
	"bufio"
	"fmt"
	"io"

	"github.com/os-artificer/AiT/internal/pkg/terminal/history"
)

const (
	ctrlG = 7
	ctrlR = 18
)

func runReverseSearch(br *bufio.Reader, out io.Writer, prompt string, hist *history.LineHistory, st *editState) lineEditResult {
	if hist == nil || hist.Len() == 0 {
		return lineEditResult{}
	}

	savedRunes := append([]rune(nil), st.runes...)
	savedCursor := st.cursor

	query := ""
	matchIdx := hist.Len()
	matchLine := ""
	found := false

	matchIdx, matchLine, found = hist.ReverseSearch(query, matchIdx)
	writeReverseSearch(out, query, matchLine, found)

	for {
		r, _, err := br.ReadRune()
		if err != nil {
			return lineEditResult{err: err, done: true}
		}

		switch r {
		case ctrlR:
			if !found {
				matchIdx, matchLine, found = hist.ReverseSearch(query, hist.Len())
			} else {
				matchIdx, matchLine, found = hist.ReverseSearch(query, matchIdx)
			}
			writeReverseSearch(out, query, matchLine, found)
		case ctrlG:
			return abortReverseSearch(out, prompt, st, savedRunes, savedCursor)
		case keyESC:
			key, err := parseEscapeSeq(br)
			if err != nil {
				return lineEditResult{err: err, done: true}
			}
			if key == escapeNone {
				return abortReverseSearch(out, prompt, st, savedRunes, savedCursor)
			}
			if key == escapeRight && found {
				return acceptReverseSearchEdit(out, prompt, st, matchLine)
			}
			if key == escapeEnd && found {
				return acceptReverseSearchEdit(out, prompt, st, matchLine)
			}
		case ctrlC:
			return abortReverseSearch(out, prompt, st, savedRunes, savedCursor)
		case ctrlE:
			if found {
				return acceptReverseSearchEdit(out, prompt, st, matchLine)
			}
		case '\r', '\n':
			if found {
				st.setLine(matchLine)
				_, _ = io.WriteString(out, "\r\n")
				return lineEditResult{line: matchLine, done: true}
			}
			return abortReverseSearch(out, prompt, st, savedRunes, savedCursor)
		case keyBS, keyDEL:
			if len(query) > 0 {
				query = query[:len(query)-1]
				matchIdx, matchLine, found = hist.ReverseSearch(query, hist.Len())
				writeReverseSearch(out, query, matchLine, found)
			}
		default:
			if !isLineInputRune(r) || r == '\t' {
				continue
			}
			query += string(r)
			matchIdx, matchLine, found = hist.ReverseSearch(query, hist.Len())
			writeReverseSearch(out, query, matchLine, found)
		}
		_ = matchIdx
	}
}

func writeReverseSearch(out io.Writer, query, match string, found bool) {
	display := match
	if !found {
		display = ""
	}
	_, _ = fmt.Fprintf(out, "\r\x1b[K(reverse-i-search)`%s': %s", query, display)
}

func abortReverseSearch(out io.Writer, prompt string, st *editState, saved []rune, cursor int) lineEditResult {
	st.runes = saved
	st.cursor = cursor
	st.histIdx = -1
	redrawLine(out, prompt, st.runes, st.cursor)
	return lineEditResult{}
}

func acceptReverseSearchEdit(out io.Writer, prompt string, st *editState, match string) lineEditResult {
	st.setLine(match)
	redrawLine(out, prompt, st.runes, st.cursor)
	return lineEditResult{}
}
