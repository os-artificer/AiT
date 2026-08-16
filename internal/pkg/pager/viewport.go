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

import "strings"

// Viewport holds buffered lines and a scroll offset.
type Viewport struct {
	lines  []string
	offset int
	rows   int
}

// NewViewport returns a Viewport that shows rows lines at a time.
func NewViewport(rows int) *Viewport {
	if rows < 1 {
		rows = 1
	}
	return &Viewport{rows: rows}
}

// Append adds a line to the buffer.
func (v *Viewport) Append(line string) {
	v.lines = append(v.lines, line)
}

// Len returns buffered line count.
func (v *Viewport) Len() int { return len(v.lines) }

// SetRows updates the visible row count.
func (v *Viewport) SetRows(rows int) {
	if rows < 1 {
		rows = 1
	}
	v.rows = rows
	v.clamp()
}

// Visible returns the currently displayed lines.
func (v *Viewport) Visible() []string {
	if len(v.lines) == 0 {
		return nil
	}
	end := v.offset + v.rows
	if end > len(v.lines) {
		end = len(v.lines)
	}
	return v.lines[v.offset:end]
}

// Down scrolls down by n lines.
func (v *Viewport) Down(n int) {
	v.offset += n
	v.clamp()
}

// Up scrolls up by n lines.
func (v *Viewport) Up(n int) {
	v.offset -= n
	v.clamp()
}

// HalfDown scrolls down by half a screen.
func (v *Viewport) HalfDown() { v.Down(v.rows / 2) }

// HalfUp scrolls up by half a screen.
func (v *Viewport) HalfUp() { v.Up(v.rows / 2) }

// Top jumps to the first line.
func (v *Viewport) Top() { v.offset = 0 }

// Bottom jumps so the last line is visible.
func (v *Viewport) Bottom() {
	v.offset = len(v.lines) - v.rows
	v.clamp()
}

// Percent returns scroll position 0–100.
func (v *Viewport) Percent() int {
	if len(v.lines) <= v.rows {
		return 100
	}
	max := len(v.lines) - v.rows
	if max <= 0 {
		return 100
	}
	return (v.offset * 100) / max
}

// Search finds the next (or previous) line containing q from the current view.
func (v *Viewport) Search(q string, backward bool) bool {
	if q == "" || len(v.lines) == 0 {
		return false
	}
	q = strings.ToLower(q)
	if backward {
		start := v.offset - 1
		for i := start; i >= 0; i-- {
			if strings.Contains(strings.ToLower(v.lines[i]), q) {
				v.offset = i
				v.clamp()
				return true
			}
		}
		return false
	}
	start := v.offset + 1
	for i := start; i < len(v.lines); i++ {
		if strings.Contains(strings.ToLower(v.lines[i]), q) {
			v.offset = i
			v.clamp()
			return true
		}
	}
	return false
}

func (v *Viewport) clamp() {
	if v.offset < 0 {
		v.offset = 0
	}
	max := len(v.lines) - v.rows
	if max < 0 {
		max = 0
	}
	if v.offset > max {
		v.offset = max
	}
}
