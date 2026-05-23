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

	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	"github.com/os-artificer/AiT/internal/ait/login"
	"github.com/os-artificer/AiT/internal/pkg/session"
	sshargs "github.com/os-artificer/AiT/internal/pkg/ssh/args"
	"github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

var cliExecutor func([]string, io.Writer, io.Writer) error

var aiModeRunner func(context.Context, *session.Manager) error

// PrintHelp writes AiT shell REPL usage to w (CRLF when raw TTY mode is active).
// w must be non-nil. Does not return an error.
func PrintHelp(w io.Writer) {
	var buf bytes.Buffer
	_ = helpdoc.WriteShellREPLHelp(&buf)
	shellWriteTTY(w, buf.String())
}

// ExecuteLine runs one REPL line. Returns exitREPL true when the user typed exit/quit.
// mgr must be non-nil. Returns parse errors, or errors from login/load/list/session subcommands.
func ExecuteLine(ctx context.Context, mgr *session.Manager, line string) (exitREPL bool, err error) {
	line = normalizeShellLine(line)
	words, err := parseShellLine(line)
	if err != nil {
		return false, err
	}

	if len(words) == 0 {
		return false, nil
	}

	switch words[0] {
	case replCmdExit, replCmdQuit:
		return true, nil
	case replCmdHelp, replArgH, replArgHelp:
		PrintHelp(shellCLIOut)
		return false, nil
	case replCmdClear:
		terminal.ClearScreen(shellCLIOut)
		return false, nil
	case replCmdAI:
		if aiModeRunner == nil {
			return false, fmt.Errorf("%s", errAIModeRunnerNotConfig)
		}
		return false, aiModeRunner(ctx, mgr)
	case replCmdLogin:
		return false, shellLogin(ctx, mgr, words[1:])
	case replCmdLoad:
		return false, shellLoad(ctx, mgr, words[1:])
	default:
		return false, runCLI(words)
	}
}

func shellLogin(ctx context.Context, mgr *session.Manager, argv []string) error {
	spec, err := sshargs.Parse(argv)
	if err != nil {
		return err
	}

	resolved, err := sshargs.Resolve(spec)
	if err != nil {
		return err
	}

	return mgr.Login(ctx, login.OptsFromResolved(resolved))
}

func shellLoad(ctx context.Context, mgr *session.Manager, argv []string) error {
	id, password, tmux, err := parseLoadArgs(argv)
	if err != nil {
		return err
	}

	return mgr.Load(ctx, id, tmux, password)
}

// SetAIModeRunner registers the function that enters local AI mode from the REPL.
// Called from internal/ait/cli init to avoid an import cycle between shell and aimode.
func SetAIModeRunner(fn func(context.Context, *session.Manager) error) {
	aiModeRunner = fn
}

// SetCLIExecutor registers the function that runs Cobra subcommands from the REPL.
// Called from internal/ait/cli init to avoid an import cycle between cli and shell.
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
