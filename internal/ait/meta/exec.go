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
	"io"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

var metaBlockedInSession = map[string]string{
	"login":   "Exit the SSH session first, then log in from the local shell.",
	"load":    "Exit the SSH session first, then load a saved session locally.",
	"suspend": "Use !ait suspend inside this SSH session.",
	"ai":      "Use !ait ai or !! for AI mode inside SSH.",
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
	_ = WriteMetaHelp(&buf)
	_, _ = io.WriteString(w, strings.ReplaceAll(buf.String(), "\n", "\r\n"))
}

// Execute runs ait subcommands from an SSH meta line (!ait <args>).
// Empty args or "help" prints in-session help.
// Returns fmt.Errorf when args[0] is blocked in-session (login/load/suspend), or errors from the executor.
func Execute(args []string) error {
	if len(args) == 0 || (len(args) == 1 && (args[0] == replCmdHelp || args[0] == replArgH || args[0] == replArgHelp)) {
		PrintMetaHelp(os.Stdout)
		return nil
	}

	if detail, blocked := metaBlockedInSession[args[0]]; blocked {
		return metaBlockedErr(args[0], detail)
	}
	if executor == nil {
		return metaExecutorNotConfiguredErr()
	}
	return executor(metaCLIArgs(args), terminal.NewCRLFWriter(os.Stdout), os.Stderr)
}

func metaCLIArgs(args []string) []string {
	if len(args) > 0 && args[0] == "list" {
		return []string{"session", "--list"}
	}
	return args
}
