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
	"github.com/mattn/go-runewidth"

	"github.com/os-artificer/AiT/internal/pkg/terminal/screen"
)

// AppendRedrawLine appends a carriage-return line redraw with cursor repositioning.
// prompt is the visible prefix before the edited line (no leading CR).
func AppendRedrawLine(dst []byte, prompt string, runes []rune, cursor int) []byte {
	line := string(runes)
	dst = append(dst, '\r')
	dst = append(dst, prompt...)
	dst = append(dst, line...)
	dst = append(dst, "\x1b[K"...)
	back := runewidth.StringWidth(string(runes[cursor:]))
	for i := 0; i < back; i++ {
		dst = append(dst, '\b')
	}
	return dst
}

// AppendClearAndRedrawLine appends a clear-screen sequence then redraws prompt and line.
func AppendClearAndRedrawLine(dst []byte, prompt string, runes []rune, cursor int) []byte {
	dst = append(dst, screen.ClearSeq()...)
	if prompt != "" || len(runes) > 0 {
		dst = AppendRedrawLine(dst, prompt, runes, cursor)
	}
	return dst
}
