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

package ait

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var metaBlockedInSession = map[string]string{
	"login":   "exit the SSH session first, then run: ait login ...",
	"load":    "exit the SSH session first, then run: ait load <id|alias>",
	"suspend": "use !ait suspend (not !ait suspend via CLI runner)",
}

// PrintMetaHelp writes in-session !ait command help to w.
func PrintMetaHelp(w io.Writer) {
	const help = `AiT in-session commands (!ait prefix):
  help              show this help
  ai <question>     Deepseek assistant
  !! <question>     alias for !ait ai (space after !!)
  local             local mode (prompt: AiT Local>)
  remote            return to remote shell (prompt: AiT user@host>)
  suspend           suspend session and disconnect

Run ait CLI locally (examples):
  !ait list
  !ait config get <key>
  !ait sync <src> <dst>
  !ait session delete <id|alias>

For login/load use: ait shell (AiT> prompt), not from inside SSH.
`
	_, _ = io.WriteString(w, strings.ReplaceAll(help, "\n", "\r\n"))
}

// ExecuteMeta runs ait subcommands from an SSH meta line (!ait <args>).
// Empty args or "help" prints in-session help.
func ExecuteMeta(args []string) error {
	if len(args) == 0 || (len(args) == 1 && (args[0] == "help" || args[0] == "-h" || args[0] == "--help")) {
		PrintMetaHelp(os.Stdout)
		return nil
	}
	if msg, blocked := metaBlockedInSession[args[0]]; blocked {
		return fmt.Errorf("AiT: %q cannot run inside SSH — %s", args[0], msg)
	}
	return runCLI(args)
}
