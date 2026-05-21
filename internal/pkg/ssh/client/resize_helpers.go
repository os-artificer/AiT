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

package client

import "golang.org/x/term"

func termSize(fd int) (cols, rows int) {
	cols, rows, _ = term.GetSize(fd)
	return cols, rows
}

// notifyTerminalResize pushes the current terminal size to h when GetSize succeeds.
func notifyTerminalResize(h resizeHandler, fd int) {
	cols, rows, err := term.GetSize(fd)
	if err != nil {
		return
	}

	h.onResize(cols, rows)
}

// pollWindowResize reads terminal size and notifies h when dimensions change.
func pollWindowResize(h resizeHandler, fd int, lastCols, lastRows int) (cols, rows int, changed bool) {
	cols, rows, err := term.GetSize(fd)
	if err != nil {
		return lastCols, lastRows, false
	}
	if cols == lastCols && rows == lastRows {
		return lastCols, lastRows, false
	}

	h.onResize(cols, rows)
	return cols, rows, true
}
