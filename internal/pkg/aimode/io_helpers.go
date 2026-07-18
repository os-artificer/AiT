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
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/aimode/approve"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/replcmd"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

func aimWrite(w io.Writer, s string) {
	terminal.WriteTTY(w, s)
}

func aimPrintErr(w io.Writer, msg string) {
	terminal.PrintErr(w, msg)
}

func normalizeLine(line string) string {
	return terminal.NormalizeLine(line)
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
	case pkghelpdoc.HelpArgCmd, pkghelpdoc.HelpArgH, pkghelpdoc.HelpArgHelp,
		replcmd.Clear, replcmd.History,
		replCmdMode, replCmdApprove, replCmdTrust, replCmdPolicy, replCmdModel, replCmdRun, replCmdAISsh:
		return true
	default:
		return false
	}
}

func (o Options) withTTYConfirm(br *bufio.Reader) Options {
	if o.ConfirmRead != nil {
		return o
	}

	out := o.Out
	if out == nil {
		out = os.Stdout
	}

	o.ConfirmRead = func(prompt string) (approve.Result, error) {
		aimWrite(out, prompt)
		line, err := terminal.ReadTTYLineNoHistory(br, out, "")
		if err != nil {
			return approve.Result{}, err
		}
		allowPattern := true
		allowSession := true
		if strings.Contains(prompt, "HIGH-RISK") {
			allowPattern = false
			allowSession = false
		}
		return approve.ParseChoice(line, allowPattern, allowSession), nil
	}
	return o
}
