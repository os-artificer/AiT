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

	"github.com/os-artificer/AiT/internal/ait/version"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// Execute runs the top-level ait CLI with args on os.Stdout and os.Stderr.
// Returns the same errors as ExecuteWriters (interactive-only hint, help/REPL failures).
func Execute(args []string) error {
	return ExecuteWriters(args, os.Stdout, os.Stderr)
}

// ExecuteWriters runs the top-level ait CLI with custom stdout and stderr.
// Empty argv opens the REPL; -h/--help/help print root help; -V/--version/version print
// the ait build string; any other argv returns an interactive-only errhint.
// stderr is reserved for API parity with other tools (REPL errors use errOut inside RunREPL).
func ExecuteWriters(args []string, stdout, stderr io.Writer) error {
	_ = stderr

	if len(args) == 0 {
		return RunREPL(context.Background(), os.Stdin, stdout, stderr)
	}

	if pkghelpdoc.HelpRequested(args) {
		return WriteRootHelp(stdout)
	}

	if len(args) == 1 && (args[0] == argVersionShort || args[0] == argVersionLong || args[0] == cmdVersion) {
		_, err := fmt.Fprintln(stdout, version.Version())
		return err
	}

	return errhint.WithTryMsg(
		interactiveOnlyMsg,
		interactiveOnlyExpl,
		toolname.AIT,
	)
}
