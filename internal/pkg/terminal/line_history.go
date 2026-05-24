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

import "strings"

const defaultLineHistoryCap = 500

// LineHistory holds in-memory REPL input lines (oldest first).
type LineHistory struct {
	lines []string
	cap   int
}

// NewLineHistory returns an empty history with the default capacity.
func NewLineHistory() *LineHistory {
	return &LineHistory{cap: defaultLineHistoryCap}
}

// Add appends line when non-empty and not identical to the newest entry.
func (h *LineHistory) Add(line string) {
	if h == nil {
		return
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	if len(h.lines) > 0 && h.lines[len(h.lines)-1] == line {
		return
	}
	h.lines = append(h.lines, line)
	if len(h.lines) > h.cap {
		h.lines = h.lines[len(h.lines)-h.cap:]
	}
}

// Len returns the number of stored lines.
func (h *LineHistory) Len() int {
	if h == nil {
		return 0
	}
	return len(h.lines)
}

// At returns the line at index i (0 = oldest). Returns "" when out of range.
func (h *LineHistory) At(i int) string {
	if h == nil || i < 0 || i >= len(h.lines) {
		return ""
	}
	return h.lines[i]
}
