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

package pty

import "golang.org/x/term"

// ResizeHandler receives terminal dimension changes during an interactive session.
type ResizeHandler interface {
	OnResize(cols, rows int)
}

// TermSize returns the current terminal columns and rows for the given fd.
// Errors are swallowed; callers receive 0 in both dimensions when fd is invalid.
func TermSize(fd int) (cols, rows int) {
	cols, rows, _ = term.GetSize(fd)
	return cols, rows
}

// NotifyResize pushes the current terminal size to h when GetSize succeeds.
// h.OnResize is not called when fd is invalid.
func NotifyResize(h ResizeHandler, fd int) {
	cols, rows, err := term.GetSize(fd)
	if err != nil {
		return
	}

	h.OnResize(cols, rows)
}

// PollResize reads terminal size and notifies h when dimensions change from lastCols/lastRows.
// Returns the (possibly unchanged) current size plus a changed flag.
func PollResize(h ResizeHandler, fd int, lastCols, lastRows int) (cols, rows int, changed bool) {
	cols, rows, err := term.GetSize(fd)
	if err != nil {
		return lastCols, lastRows, false
	}
	if cols == lastCols && rows == lastRows {
		return lastCols, lastRows, false
	}

	h.OnResize(cols, rows)
	return cols, rows, true
}
