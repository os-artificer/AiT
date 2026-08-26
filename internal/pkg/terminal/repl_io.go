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
	"bufio"
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

// WriteTTY writes s to w with LF converted to CRLF (raw TTY safe).
// Uses naive ReplaceAll (not crlf.NewWriter) to preserve existing shell/Agent Loop semantics.
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

	msg = collapseTTYMsg(msg)
	WriteTTY(w, errhint.PrefixAiT+msg+"\r\n")
}

// ReadTTYLine reads one line from br while stdin is in raw mode (Enter submits).
// out receives echoed input and editing control output.
// hist, when non-nil, enables Up/Down recall and records submitted lines.
func ReadTTYLine(br *bufio.Reader, out io.Writer, prompt string, hist *LineHistory) (string, error) {
	return ReadLine(br, out, LineEditOptions{
		Prompt:        prompt,
		History:       hist,
		RecordHistory: hist != nil,
	})
}

// ReadTTYLineNoHistory reads one line without history recall or recording (confirm prompts, pickers).
func ReadTTYLineNoHistory(br *bufio.Reader, out io.Writer, prompt string) (string, error) {
	return ReadLine(br, out, LineEditOptions{Prompt: prompt})
}

func collapseTTYMsg(msg string) string {
	msg = strings.TrimSpace(msg)
	msg = strings.ReplaceAll(msg, "\r\n", "\n")
	msg = strings.ReplaceAll(msg, "\r", "\n")
	return strings.Join(strings.Fields(msg), " ")
}
