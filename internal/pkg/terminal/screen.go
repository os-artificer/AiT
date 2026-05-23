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

import "io"

// ansiClearScreen is VT100 erase display + cursor home (supported by modern Windows/macOS/Linux terminals).
const ansiClearScreen = "\x1b[2J\x1b[H"

// ClearScreenSeq returns the ANSI escape sequence written by ClearScreen.
func ClearScreenSeq() string {
	return ansiClearScreen
}

// ClearScreen writes a VT100 clear-and-home sequence to w.
// w must be non-nil. Does not return an error.
func ClearScreen(w io.Writer) {
	_, _ = io.WriteString(w, ansiClearScreen)
}
