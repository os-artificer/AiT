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
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/platform"
	"github.com/os-artificer/AiT/internal/pkg/pluginbin"
	"github.com/os-artificer/AiT/internal/pkg/shellext"
	"github.com/os-artificer/AiT/internal/pkg/terminal"

	"golang.org/x/term"
)

// RunREPL runs the ait> interactive loop until exit/quit, EOF, or ctx cancel.
// in is typically os.Stdin; when it is a TTY, raw mode + CRLF stdout and a
// Restore→exec→MakeRaw Runner are used. Returns ctx.Err on cancel, or line/IO errors.
func RunREPL(ctx context.Context, in io.Reader, out, errOut io.Writer) error {
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}

	fd := int(os.Stdin.Fd())
	if in == os.Stdin && term.IsTerminal(fd) {
		return runREPLTTY(ctx, fd, out, errOut)
	}
	return runREPLPlain(ctx, in, out, errOut)
}

func runREPLPlain(ctx context.Context, in io.Reader, out, errOut io.Writer) error {
	sc := bufio.NewScanner(in)
	hist := newAITHistory(errOut)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		_, _ = io.WriteString(out, replPrompt)
		if !sc.Scan() {
			if err := sc.Err(); err != nil {
				return err
			}
			return nil
		}

		exit, err := ExecuteLine(ctx, sc.Text(), hist, out, errOut)
		reportREPLError(errOut, err, false)
		if exit {
			return nil
		}
	}
}

func runREPLTTY(ctx context.Context, fd int, out, errOut io.Writer) error {
	old, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("terminal: %w", err)
	}
	defer func() { _ = term.Restore(fd, old) }()

	cliOut := terminal.NewCRLFWriter(out)

	prev := pluginbin.SetRunner(func(ctx context.Context, path string, argv []string) error {
		if restoreErr := term.Restore(fd, old); restoreErr != nil {
			return restoreErr
		}
		runErr := shellext.Run(ctx, path, argv, os.Stdin, os.Stdout, os.Stderr)
		_, rawErr := term.MakeRaw(fd)
		if rawErr != nil {
			if runErr != nil {
				return errors.Join(fmt.Errorf("terminal: %w", rawErr), runErr)
			}
			return fmt.Errorf("terminal: %w", rawErr)
		}
		return runErr
	})
	defer pluginbin.SetRunner(prev)

	_, _ = io.WriteString(cliOut, "Type help for commands. Use exit or quit to leave.\r\n")

	br := bufio.NewReader(os.Stdin)
	hist := newAITHistory(errOut)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		_, _ = io.WriteString(cliOut, replPrompt)
		line, err := terminal.ReadLine(br, cliOut, terminal.LineEditOptions{
			Prompt:        replPrompt,
			History:       hist,
			RecordHistory: true,
			Completer:     completeREPLLine,
		})
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		exit, err := ExecuteLine(ctx, line, hist, cliOut, errOut)
		reportREPLError(errOut, err, true)
		if exit {
			_, _ = io.WriteString(cliOut, "\r\n")
			return nil
		}
	}
}

func newAITHistory(warn io.Writer) *terminal.LineHistory {
	h := terminal.NewLineHistory()

	dir, err := platform.DataDir()
	if err != nil || dir == "" {
		return h
	}

	path := filepath.Join(dir, historyBasename)
	h.SetFilePath(path)
	if loadErr := h.LoadFromFile(path); loadErr != nil && warn != nil {
		_, _ = fmt.Fprintf(warn, "ait: history load: %v\n", loadErr)
	}
	return h
}

func reportREPLError(errOut io.Writer, err error, tty bool) {
	if err == nil {
		return
	}
	surface := errhint.SurfaceCLI
	if tty {
		surface = errhint.SurfaceTTY
	}
	errhint.Print(errOut, surface, err)
}
