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

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
	"github.com/os-artificer/AiT/internal/ssh/session"
	"github.com/os-artificer/AiT/internal/ssh/shell"
)

var metaBlockedInSession = map[string]string{
	shell.ReplCmdLoad:    "Exit the SSH session first, then load a saved session locally.",
	shell.ReplCmdSuspend: "Use " + toolname.MetaLine("suspend") + " inside this SSH session.",
}

var executor func([]string, io.Writer, io.Writer) error

// SetExecutorWriters registers the function that runs aissh subcommands from meta lines.
// Called from internal/ssh init_hooks to avoid an import cycle between aissh root and meta.
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

// Execute runs aissh subcommands from an SSH meta line (!ait <args>).
// Empty args or "help" prints in-session help.
// Returns fmt.Errorf when args[0] is blocked in-session (load/suspend), or errors from the executor.
func Execute(args []string) error {
	if len(args) == 0 || (len(args) == 1 && helpArgv(args[0])) {
		PrintMetaHelp(os.Stdout)
		return nil
	}

	if detail, blocked := metaBlockedInSession[args[0]]; blocked {
		return metaBlockedErr(args[0], detail)
	}
	if executor == nil {
		return metaExecutorNotConfiguredErr()
	}
	err := executor(metaCLIArgs(args), terminal.NewCRLFWriter(os.Stdout), os.Stderr)
	return wrapMetaAIUnknown(args, err)
}

func helpArgv(arg string) bool {
	return arg == pkghelpdoc.HelpArgCmd || arg == pkghelpdoc.HelpArgH || arg == pkghelpdoc.HelpArgHelp
}

func metaCLIArgs(args []string) []string {
	if len(args) > 0 && args[0] == shell.ReplCmdList {
		return []string{session.Name, session.FlagList}
	}
	return args
}
