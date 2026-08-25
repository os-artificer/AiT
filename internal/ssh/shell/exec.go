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
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
	"github.com/os-artificer/AiT/internal/ssh/shell/builtin"
)

var cliExecutor func([]string, io.Writer, io.Writer) error

// PrintHelp writes AiT shell REPL usage to w (CRLF when raw TTY mode is active).
// hasActiveSession controls whether ssh-style connect / load are blocked in help output.
// w must be non-nil. Does not return an error.
func PrintHelp(w io.Writer, hasActiveSession bool) {
	var buf bytes.Buffer
	_ = WriteShellREPLHelp(&buf, hasActiveSession)
	shellWriteTTY(w, buf.String())
}

// ExecuteLine runs one REPL line. Returns exitREPL true when the user typed exit/quit.
// hist holds session input history for the history builtin; may be nil in tests.
func ExecuteLine(ctx context.Context, mgr *session.Manager, line string, hist *terminal.LineHistory) (exitREPL bool, err error) {
	line = normalizeShellLine(line)
	words, err := parseShellLine(line)
	if err != nil {
		return false, err
	}

	if len(words) == 0 {
		return false, nil
	}

	env := builtin.Env{
		Out:       shellCLIOut,
		In:        os.Stdin,
		Mgr:       mgr,
		Hist:      hist,
		Line:      line,
		ParseLine: parseShellLine,
		PrintHelp: func(hasActiveSession bool) {
			PrintHelp(shellCLIOut, hasActiveSession)
		},
	}

	exitREPL, handled, err := builtin.Try(ctx, env, words)
	if handled {
		return exitREPL, err
	}

	return false, dispatchShellCommand(ctx, env.Mgr, words)
}

// SetCLIExecutor registers the function that runs Cobra subcommands from the REPL.
// Called from internal/ssh init_hooks to avoid an import cycle between aissh root and shell.
func SetCLIExecutor(fn func([]string, io.Writer, io.Writer) error) {
	cliExecutor = fn
}

// runCLI executes cobra subcommands (shared with meta, without SSH-only blocks).
func runCLI(argv []string) error {
	if cliExecutor == nil {
		return fmt.Errorf("%s", errShellCLIExecutorNotConfig)
	}
	return cliExecutor(argv, shellCLIOut, shellCLIErr)
}

// shellPrompt returns the REPL prompt string.
func shellPrompt() string {
	return client.BasePrompt
}
