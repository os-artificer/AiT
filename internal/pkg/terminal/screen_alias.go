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

	"github.com/os-artificer/AiT/internal/pkg/terminal/crlf"
	"github.com/os-artificer/AiT/internal/pkg/terminal/screen"
)

// ClearScreenSeq returns the ANSI escape sequence written by ClearScreen.
func ClearScreenSeq() string {
	return screen.ClearSeq()
}

// ClearScreen writes a VT100 clear-and-home sequence to w.
// w must be non-nil. Does not return an error.
func ClearScreen(w io.Writer) {
	screen.Clear(w)
}

// NewCRLFWriter returns a Writer that translates LF to CRLF before writing to w.
// Write returns len(p) on success so callers see the original byte count.
func NewCRLFWriter(w io.Writer) io.Writer {
	return crlf.NewWriter(w)
}
