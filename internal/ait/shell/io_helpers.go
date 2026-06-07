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
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
)

// NormalizeLine trims space and an optional trailing ';' (Enter submits, ';' not required).
func NormalizeLine(line string) string {
	line = strings.TrimSpace(line)
	line = strings.TrimSuffix(line, ";")

	return strings.TrimSpace(line)
}

// formatShellMsg collapses multi-line messages to one line for raw TTY display.
func formatShellMsg(msg string) string {
	msg = strings.TrimSpace(msg)
	msg = strings.ReplaceAll(msg, "\r\n", "\n")
	msg = strings.ReplaceAll(msg, "\r", "\n")
	msg = strings.Join(strings.Fields(msg), " ")

	return msg
}

// WriteTTY writes s to w with LF converted to CRLF (raw TTY safe).
func WriteTTY(w io.Writer, s string) {
	s = strings.ReplaceAll(s, "\n", "\r\n")
	_, _ = io.WriteString(w, s)
}

// PrintErr writes user-facing errors with AiT: prefix and CRLF line endings.
// Multi-line hint errors keep their layout; single-line messages are collapsed for TTY width.
func PrintErr(w io.Writer, msg string) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return
	}
	if strings.Contains(msg, "\n") {
		errhint.Print(w, errhint.SurfaceTTY, errhint.NewSentinel(msg))
		return
	}

	msg = formatShellMsg(msg)
	WriteTTY(w, errhint.PrefixAiT+msg+"\r\n")
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
