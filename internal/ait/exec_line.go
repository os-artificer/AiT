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
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/os-artificer/AiT/internal/ait/version"
	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/pluginbin"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// ExecuteLine runs one REPL line.
// hist holds input history for the history builtin; may be nil in tests.
// out/errOut default to os.Stdout/os.Stderr when nil.
// Returns exitREPL true for exit/quit. Plugin *exec.ExitError is swallowed (AT14).
// Returns parse errors, unknown-command/help errhints, or non-exit exec failures.
func ExecuteLine(ctx context.Context, line string, hist *terminal.LineHistory, out, errOut io.Writer) (exitREPL bool, err error) {
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}

	words, err := parseShellLine(line)
	if err != nil {
		return false, err
	}
	if len(words) == 0 {
		return false, nil
	}

	switch words[0] {
	case cmdExit, cmdQuit:
		return true, nil
	case cmdHelp:
		if len(words) == 1 {
			return false, WriteRootHelp(out)
		}
		return false, writePluginHelp(out, words[1])
	case cmdVersion:
		_, err := fmt.Fprintln(out, version.Version())
		return false, err
	case cmdClear:
		terminal.ClearScreen(out)
		return false, nil
	case cmdHistory:
		return false, terminal.RunHistoryCommand(out, hist, words[1:])
	}

	if fn, ok := manageHandlers[words[0]]; ok {
		return false, fn(ctx, words, out, errOut)
	}

	p, ok := aitplugin.Lookup(words[0])
	if !ok {
		return false, errhint.WithTryMsg(
			unknownCmdMsg+": "+words[0],
			"Use help to list reserved commands and discovered plugins.",
			toolname.AIT,
			cmdHelp,
		)
	}

	runErr := runPlugin(ctx, p, words, os.Stdin, out, errOut)
	if isBareExitError(runErr) {
		return false, nil
	}
	return false, runErr
}

// isBareExitError reports whether err is solely a plugin process exit status (AT14).
// Joined or wrapped errors (e.g. MakeRaw failure + ExitError) must not be swallowed.
func isBareExitError(err error) bool {
	_, ok := err.(*exec.ExitError)
	return ok
}

// runPlugin executes a discovered plugin using its resolved Path (never LookPath).
func runPlugin(ctx context.Context, p aitplugin.Plugin, argv []string, in io.Reader, out, errOut io.Writer) error {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}

	execArgv := append([]string{p.Manifest.Binary}, argv[1:]...)
	return withAbsorbedInterrupt(func() error {
		return pluginbin.ExecPath(ctx, p.Path, execArgv, in, out, errOut)
	})
}
