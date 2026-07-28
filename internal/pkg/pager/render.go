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

package pager

import (
	"fmt"
	"io"
	"strings"

	"github.com/mattn/go-runewidth"
)

const (
	// clearSequence homes the cursor and clears the screen.
	clearSequence = "\x1b[H\x1b[2J"

	statusPositionFmt = "--More--(%d%%)"
	statusTruncated   = "truncated"
	statusEnd         = "END"
	statusHintSuffix  = "…"
)

// helpLines is the key reference shown for h and ?.
var helpLines = []string{
	"Keys:",
	"  space, f    forward one screen      b        back one screen",
	"  j, Enter    forward one line        k        back one line",
	"  d, u        forward/back half       g, G     top/end",
	"  /, \\        search forward/back     n, N     next/previous match",
	"  q, Ctrl+C   quit                    h, ?     this help",
	"",
	"Press any key to continue…",
}

// writeFrame clears the screen and paints the visible lines.
func writeFrame(w io.Writer, lines []string) {
	_, _ = io.WriteString(w, clearSequence)
	for _, line := range lines {
		_, _ = fmt.Fprintln(w, line)
	}
}

// writeStatus paints text on the status row, padded to cols so that leftovers
// from a longer previous status are erased, and leaves the cursor at column 0.
func writeStatus(w io.Writer, text string, cols int) {
	_, _ = fmt.Fprintf(w, "\r%s\r", padDisplay(text, cols))
}

// writeHelp replaces the screen with the key reference.
func writeHelp(w io.Writer) {
	writeFrame(w, helpLines)
}

func writeNewline(w io.Writer) {
	_, _ = io.WriteString(w, "\n")
}

// formatStatus builds the status line: scroll position, progress marker while
// a source is still producing, and the end state.
func formatStatus(percent int, done, truncated bool, hint string) string {
	parts := []string{fmt.Sprintf(statusPositionFmt, percent)}
	if hint != "" && !done {
		parts = append(parts, hint+statusHintSuffix)
	}
	if truncated {
		parts = append(parts, statusTruncated)
	}
	if done {
		parts = append(parts, statusEnd)
	}
	return strings.Join(parts, " ")
}

// padDisplay truncates or pads s to cols terminal cells. Width is measured in
// display cells, not bytes, so wide characters are counted correctly and never
// cut mid-rune.
func padDisplay(s string, cols int) string {
	if cols <= 0 {
		return s
	}
	width := runewidth.StringWidth(s)
	if width == cols {
		return s
	}
	if width > cols {
		return runewidth.Truncate(s, cols, "")
	}
	return s + strings.Repeat(" ", cols-width)
}
