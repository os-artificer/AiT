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
	"io"
	"strings"

	"github.com/mattn/go-runewidth"
)

func redrawLine(out io.Writer, prompt string, runes []rune, cursor int) {
	line := string(runes)
	_, _ = io.WriteString(out, "\r"+prompt+line+"\x1b[K")
	back := runewidth.StringWidth(string(runes[cursor:]))
	for i := 0; i < back; i++ {
		_, _ = io.WriteString(out, "\b")
	}
}

func redrawLineAfterClear(out io.Writer, prompt string, runes []rune, cursor int) {
	ClearScreen(out)
	if prompt != "" || len(runes) > 0 {
		redrawLine(out, prompt, runes, cursor)
	}
}

func echoRune(out io.Writer, r rune) {
	_, _ = io.WriteString(out, string(r))
}

func eraseRuneBeforeCursor(out io.Writer, runes []rune, cursor int) int {
	if cursor <= 0 {
		return cursor
	}
	last := runes[cursor-1]
	width := runewidth.RuneWidth(last)
	if width < 1 {
		width = 1
	}
	var erase strings.Builder
	for i := 0; i < width; i++ {
		erase.WriteString("\b")
	}
	_, _ = io.WriteString(out, erase.String())
	return cursor - 1
}
