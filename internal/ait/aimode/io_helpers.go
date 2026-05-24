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

package aimode

import (
	"io"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
)

func aimWrite(w io.Writer, s string) {
	s = strings.ReplaceAll(s, "\n", "\r\n")
	_, _ = io.WriteString(w, s)
}

func aimPrintErr(w io.Writer, msg string) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return
	}
	if strings.Contains(msg, "\n") {
		errhint.Print(w, errhint.SurfaceTTY, errhint.NewSentinel(msg))
		return
	}
	aimWrite(w, msgPrefixAiT+msg+"\r\n")
}

func normalizeLine(line string) string {
	line = strings.TrimSpace(line)
	line = strings.TrimSuffix(line, ";")
	return strings.TrimSpace(line)
}

func firstWord(line string) (word, rest string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", ""
	}
	i := strings.IndexByte(line, ' ')
	if i < 0 {
		return line, ""
	}
	return line[:i], strings.TrimSpace(line[i+1:])
}

// isMetaREPLCommand reports whether line is a built-in AI REPL command (not free-text chat).
func isMetaREPLCommand(line string) bool {
	word, _ := firstWord(normalizeLine(line))
	switch word {
	case replCmdHelp, replArgH, replArgHelp, replCmdClear, replCmdHistory, replCmdPolicy, replCmdModel, replCmdRun, replCmdAit:
		return true
	default:
		return false
	}
}
