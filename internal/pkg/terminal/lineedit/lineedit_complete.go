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
	"io"
	"strings"
)

func applyCompleter(out io.Writer, prompt string, st *editState, completer func(string, int) (string, int, string)) {
	line := string(st.runes)
	cursor := st.cursor
	newLine, newCursor, display := completer(line, cursor)

	runes := []rune(newLine)
	newCursor = clampCursor(newCursor, len(runes))

	unchanged := newLine == line && newCursor == cursor && display == ""
	if unchanged {
		return
	}

	st.runes = runes
	st.cursor = newCursor
	st.histIdx = -1

	if display != "" {
		_, _ = io.WriteString(out, "\r\n")
		crlfDisplay := strings.ReplaceAll(display, "\n", "\r\n")
		_, _ = io.WriteString(out, crlfDisplay)
		if !strings.HasSuffix(display, "\n") {
			_, _ = io.WriteString(out, "\r\n")
		}
	}
	redrawLine(out, prompt, st.runes, st.cursor)
}

func clampCursor(cursor, n int) int {
	if cursor < 0 {
		return 0
	}
	if cursor > n {
		return n
	}
	return cursor
}
