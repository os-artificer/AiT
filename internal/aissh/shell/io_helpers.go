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

package shell

import (
	"io"

	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

// NormalizeLine trims space and an optional trailing ';' (Enter submits, ';' not required).
func NormalizeLine(line string) string {
	return terminal.NormalizeLine(line)
}

// WriteTTY writes s to w with LF converted to CRLF (raw TTY safe).
func WriteTTY(w io.Writer, s string) {
	terminal.WriteTTY(w, s)
}

// PrintErr writes user-facing errors with AiT: prefix and CRLF line endings.
// Multi-line hint errors keep their layout; single-line messages are collapsed for TTY width.
func PrintErr(w io.Writer, msg string) {
	terminal.PrintErr(w, msg)
}

func normalizeShellLine(line string) string {
	return NormalizeLine(line)
}

func shellWriteTTY(w io.Writer, s string) {
	WriteTTY(w, s)
}

func shellPrintErr(w io.Writer, msg string) {
	PrintErr(w, msg)
}
