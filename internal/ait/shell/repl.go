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
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/terminal"

	"golang.org/x/term"
)

// RunREPL runs the AiT> interactive command loop until exit/quit or ctx cancel.
// mgr must be non-nil; in is typically os.Stdin and out os.Stdout.
// Returns ctx.Err on cancel, or errors from ExecuteLine (login/load/list failures).
func RunREPL(ctx context.Context, mgr *session.Manager, in io.Reader, out io.Writer) error {
	fd := int(os.Stdin.Fd())
	if in == os.Stdin && term.IsTerminal(fd) {
		return runShellREPLTTY(ctx, mgr, fd, out)
	}
	return runShellREPLPlain(ctx, mgr, in, out)
}

// ReadTTYLine reads one line from br while stdin is in raw mode (Enter submits).
// out receives echoed input and editing control output.
// prompt is redrawn after Ctrl+L when non-empty.
func ReadTTYLine(br *bufio.Reader, out io.Writer, prompt string) (string, error) {
	return terminal.ReadLineRaw(br, out, prompt)
}

func reportShellError(err error) {
	if err != nil {
		shellPrintErr(os.Stderr, err.Error())
	}
}

func runShellREPLPlain(ctx context.Context, mgr *session.Manager, in io.Reader, out io.Writer) error {
	sc := bufio.NewScanner(in)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		_, _ = io.WriteString(out, shellPrompt())
		if !sc.Scan() {
			if err := sc.Err(); err != nil {
				return err
			}
			return nil
		}

		exit, err := ExecuteLine(ctx, mgr, sc.Text())
		reportShellError(err)
		if exit {
			return nil
		}
	}
}

func runShellREPLTTY(ctx context.Context, mgr *session.Manager, fd int, out io.Writer) error {
	old, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("terminal: %w", err)
	}

	defer func() { _ = term.Restore(fd, old) }()

	setShellTTYIO(true)
	defer setShellTTYIO(false)

	_, _ = io.WriteString(out, "Type help for commands. Use exit or quit to leave.\r\n")

	br := bufio.NewReader(os.Stdin)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		prompt := shellPrompt()
		_, _ = io.WriteString(out, prompt)
		line, err := ReadTTYLine(br, out, prompt)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		exit, err := ExecuteLine(ctx, mgr, line)
		reportShellError(err)
		if exit {
			_, _ = io.WriteString(out, "\r\n")
			return nil
		}
	}
}
