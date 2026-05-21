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

	"golang.org/x/term"
)

func reportShellError(err error) {
	if err != nil {
		shellPrintErr(os.Stderr, err.Error())
	}
}

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

		_, _ = io.WriteString(out, shellPrompt())
		line, err := readTTYLine(br)
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

func readTTYLine(br *bufio.Reader) (string, error) {
	var line []byte
	for {
		b, err := br.ReadByte()
		if err != nil {
			res := ttyHandleEOF(err, line)
			return res.line, res.err
		}

		var res ttyReadResult
		switch b {
		case '\r':
			res = ttyHandleCR(br, line)
		case '\n':
			if len(line) > 0 {
				res = ttyReadResult{line: string(line), done: true}
			}
		case 3: // Ctrl+C
			_, _ = os.Stdout.WriteString("^C\r\n")
			return "", nil
		case 4: // Ctrl+D
			if len(line) == 0 {
				return "", io.EOF
			}
			return string(line), nil
		case 127, 8: // backspace
			res, line = ttyHandleBackspace(line)
		default:
			res, line = ttyHandlePrintable(b, line)
		}
		if res.done {
			return res.line, res.err
		}
	}
}
