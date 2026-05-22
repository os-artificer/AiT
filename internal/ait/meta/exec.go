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

package meta

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

var metaBlockedInSession = map[string]string{
	"login":   "exit the SSH session first, then run: ait login ...",
	"load":    "exit the SSH session first, then run: ait load <id|alias>",
	"suspend": "use !ait suspend (not !ait suspend via CLI runner)",
	"ai":      "use !ait ai or !! inside this SSH session (not !ait ai via CLI runner)",
}

var executor func([]string, io.Writer, io.Writer) error

// SetExecutorWriters registers the function that runs ait subcommands from meta lines.
// Called from internal/ait/cli init to avoid an import cycle between cli and meta.
func SetExecutorWriters(fn func([]string, io.Writer, io.Writer) error) {
	executor = fn
}

// PrintMetaHelp writes in-session !ait meta-command help to w (CRLF line endings).
// w must be non-nil. Does not return an error.
func PrintMetaHelp(w io.Writer) {
	var buf bytes.Buffer
	_ = helpdoc.WriteMetaHelp(&buf)
	_, _ = io.WriteString(w, strings.ReplaceAll(buf.String(), "\n", "\r\n"))
}

// Execute runs ait subcommands from an SSH meta line (!ait <args>).
// Empty args or "help" prints in-session help.
// Returns fmt.Errorf when args[0] is blocked in-session (login/load/suspend), or errors from the executor.
func Execute(args []string) error {
	if len(args) == 0 || (len(args) == 1 && (args[0] == "help" || args[0] == "-h" || args[0] == "--help")) {
		PrintMetaHelp(os.Stdout)
		return nil
	}

	if msg, blocked := metaBlockedInSession[args[0]]; blocked {
		return fmt.Errorf("AiT: %q cannot run inside SSH — %s", args[0], msg)
	}
	if executor == nil {
		return fmt.Errorf("AiT: meta executor not configured")
	}
	return executor(args, terminal.NewCRLFWriter(os.Stdout), os.Stderr)
}
