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

package errhint

import (
	"fmt"
	"io"
	"strings"
)

// Surface selects how user-facing errors are printed.
type Surface int

const (
	// SurfaceCLI is plain stderr for top-level aissh commands (no AiT: prefix).
	SurfaceCLI Surface = iota
	// SurfaceTTY is REPL/SSH-style output: AiT: on the first line, CRLF, continuation indent.
	SurfaceTTY
)

// Print writes err to w using the layout for surface.
func Print(w io.Writer, surface Surface, err error) {
	if err == nil {
		return
	}

	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return
	}
	msg = strings.TrimPrefix(msg, msgPrefixAiT)

	switch surface {
	case SurfaceCLI:
		_, _ = fmt.Fprintln(w, msg)
	case SurfaceTTY:
		printTTY(w, msg)
	default:
		_, _ = fmt.Fprintln(w, msg)
	}
}

func printTTY(w io.Writer, msg string) {
	msg = strings.ReplaceAll(msg, "\r\n", "\n")
	msg = strings.ReplaceAll(msg, "\r", "\n")
	lines := strings.Split(msg, "\n")
	indent := strings.Repeat(" ", len(msgPrefixAiT))

	for i, line := range lines {
		if i == 0 {
			writeCRLF(w, msgPrefixAiT+line)
			continue
		}
		if line == "" {
			writeCRLF(w, "")
			continue
		}
		writeCRLF(w, indent+line)
	}
}

func writeCRLF(w io.Writer, line string) {
	_, _ = io.WriteString(w, strings.ReplaceAll(line, "\n", "\r\n")+"\r\n")
}
