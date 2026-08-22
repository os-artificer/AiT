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

// Package history implements the bash-style REPL input history model used by
// `aissh shell` and `aissh ai`: bounded in-memory ring, optional file backing
// (load/append/write/reload), reverse search, and the `history` command.
package history

import (
	"fmt"
	"io"
	"strings"
)

const defaultLineHistoryCap = 500

// LineHistory holds in-memory REPL input lines (oldest first) with optional file backing.
type LineHistory struct {
	lines    []string
	cap      int
	filePath string
	flushed  int
}

// NewLineHistory returns an empty history with the default capacity.
func NewLineHistory() *LineHistory {
	return &LineHistory{cap: defaultLineHistoryCap}
}

// SetFilePath sets the persistence file used by LoadFromFile, Add, and history -a/-w.
func (h *LineHistory) SetFilePath(path string) {
	if h == nil {
		return
	}
	h.filePath = path
}

// FilePath returns the configured persistence file path, or "" when unset.
func (h *LineHistory) FilePath() string {
	if h == nil {
		return ""
	}
	return h.filePath
}

// Add appends line when non-empty and not identical to the newest entry.
// When filePath is set, the line is appended to the file (histappend-style).
func (h *LineHistory) Add(line string) {
	h.addLine(line, true)
}

// AddMemoryOnly appends a line without writing to the persistence file (history -s).
func (h *LineHistory) AddMemoryOnly(line string) {
	h.addLine(line, false)
}

func (h *LineHistory) addLine(line string, persistFile bool) {
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
		drop := len(h.lines) - h.cap
		h.lines = h.lines[drop:]
		if h.flushed > drop {
			h.flushed -= drop
		} else {
			h.flushed = 0
		}
	}
	if persistFile && h.filePath != "" {
		if err := appendLineToFile(h.filePath, line); err == nil {
			h.flushed = len(h.lines)
		}
	}
}

// Clear removes all in-memory lines (history -c). The file is unchanged until history -w.
func (h *LineHistory) Clear() {
	if h == nil {
		return
	}
	h.lines = nil
	h.flushed = 0
}

// Delete removes the 1-based display entry (history -d offset).
func (h *LineHistory) Delete(displayNum int) error {
	if h == nil {
		return fmt.Errorf("history: empty")
	}
	idx := displayNum - 1
	if idx < 0 || idx >= len(h.lines) {
		return fmt.Errorf("history: %d not found", displayNum)
	}
	h.lines = append(h.lines[:idx], h.lines[idx+1:]...)
	if h.flushed > idx {
		h.flushed--
	}
	if h.flushed > len(h.lines) {
		h.flushed = len(h.lines)
	}
	return nil
}

// SetLine replaces or appends a line at displayNum (history -s uses append).
func (h *LineHistory) SetLine(line string) {
	if h == nil {
		return
	}
	h.AddMemoryOnly(line)
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

// Format writes numbered history lines to w. lastN <= 0 means all entries.
func (h *LineHistory) Format(w io.Writer, lastN int) error {
	if h == nil || len(h.lines) == 0 {
		return nil
	}
	start := 0
	if lastN > 0 && lastN < len(h.lines) {
		start = len(h.lines) - lastN
	}
	for i := start; i < len(h.lines); i++ {
		if _, err := fmt.Fprintf(w, "  %5d  %s\n", i+1, h.lines[i]); err != nil {
			return err
		}
	}
	return nil
}

// ReverseSearch finds the previous line before beforeIdx containing query (case-sensitive).
// beforeIdx is a line index; pass len(lines) to start from the newest match.
func (h *LineHistory) ReverseSearch(query string, beforeIdx int) (idx int, line string, ok bool) {
	if h == nil || len(h.lines) == 0 {
		return -1, "", false
	}
	if beforeIdx <= 0 || beforeIdx > len(h.lines) {
		beforeIdx = len(h.lines)
	}
	for i := beforeIdx - 1; i >= 0; i-- {
		if strings.Contains(h.lines[i], query) {
			return i, h.lines[i], true
		}
	}
	return -1, "", false
}
