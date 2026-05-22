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
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

// setShellTTYIO configures CRLF stdout and silenced stderr for the duration of ait shell.
func setShellTTYIO(enabled bool) {
	if enabled {
		shellCLIOut = terminal.NewCRLFWriter(os.Stdout)
		shellCLIErr = io.Discard
		return
	}

	shellCLIOut = os.Stdout
	shellCLIErr = os.Stderr
}

// normalizeShellLine trims space and an optional trailing ';' (Enter submits, ';' not required).
func normalizeShellLine(line string) string {
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

func shellWriteTTY(w io.Writer, s string) {
	s = strings.ReplaceAll(s, "\n", "\r\n")
	_, _ = io.WriteString(w, s)
}

// shellPrintErr writes a single-line AiT-prefixed error with CRLF line endings.
func shellPrintErr(w io.Writer, msg string) {
	msg = formatShellMsg(msg)
	if msg == "" {
		return
	}

	shellWriteTTY(w, "AiT: "+msg+"\r\n")
}
