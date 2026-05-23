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

package aimode

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/os-artificer/AiT/internal/ait/shell"
	"github.com/os-artificer/AiT/internal/pkg/ai"

	"golang.org/x/term"
)

// ErrAINotConfigured is returned when RunREPL is called without a configured AI client.
var ErrAINotConfigured = fmt.Errorf("AiT: AI not configured — run: ait config set ai.deepseek.api_key <key> (or ai.minimax.api_key)")

// RunREPL runs the AI mode loop until exit/quit or ctx cancel.
// initialQuestion, if non-empty, is processed as the first user turn (!ait ai <q> / !! <q>).
// opt.AI must be non-nil. Returns ErrAINotConfigured when opt.AI is nil.
func RunREPL(ctx context.Context, opt Options, initialQuestion string) error {
	if opt.AI == nil {
		return ErrAINotConfigured
	}

	opt = fillOptionDefaults(opt)
	st := &sessionState{policy: opt.ExecPolicy}
	if st.policy == 0 {
		st.policy = ExecPolicyPerCmd
	}

	if q := stringsTrim(initialQuestion); q != "" {
		if isChatCommandWithoutMessage(q) {
			return ErrChatMessageRequired
		}
		if msg, ok := parseChatCommand(q); ok {
			return RunChatOnce(ctx, opt, msg)
		}

		exit, err := ExecuteLine(ctx, opt, st, q)
		if err != nil {
			return err
		}
		if exit || isMetaREPLCommand(q) {
			return nil
		}
	}

	fd := int(os.Stdin.Fd())
	if opt.In == os.Stdin && term.IsTerminal(fd) {
		return runAITTY(ctx, opt, st, fd)
	}
	return runAIPlain(ctx, opt, st)
}

func fillOptionDefaults(opt Options) Options {
	if opt.In == nil {
		opt.In = os.Stdin
	}
	if opt.Out == nil {
		opt.Out = os.Stdout
	}
	if opt.ErrOut == nil {
		opt.ErrOut = os.Stderr
	}
	if opt.Scope == ai.ScopeLocal && opt.Local == nil {
		opt.Local = DefaultLocalRunner()
	}
	return opt
}

func runAIPlain(ctx context.Context, opt Options, st *sessionState) error {
	sc := bufio.NewScanner(opt.In)
	prompt := Prompt(opt.Scope, opt.HostLabel)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		_, _ = io.WriteString(opt.Out, prompt)
		if !sc.Scan() {
			if err := sc.Err(); err != nil {
				return err
			}
			return nil
		}

		exit, err := ExecuteLine(ctx, opt, st, sc.Text())
		reportAIError(opt.ErrOut, err)
		if exit {
			return nil
		}
	}
}

func runAITTY(ctx context.Context, opt Options, st *sessionState, fd int) error {
	old, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("terminal: %w", err)
	}
	defer func() { _ = term.Restore(fd, old) }()

	aimWrite(opt.Out, "Type help for AI mode commands. Use exit or quit to leave.\r\n")

	br := bufio.NewReader(opt.In)
	opt = opt.withTTYConfirm(br)
	prompt := Prompt(opt.Scope, opt.HostLabel)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		_, _ = io.WriteString(opt.Out, prompt)
		line, err := shell.ReadTTYLine(br, opt.Out)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		exit, err := ExecuteLine(ctx, opt, st, line)
		reportAIError(opt.ErrOut, err)
		if exit {
			_, _ = io.WriteString(opt.Out, "\r\n")
			return nil
		}
	}
}

func reportAIError(w io.Writer, err error) {
	if err != nil {
		aimPrintErr(w, err.Error())
	}
}

func stringsTrim(s string) string {
	return normalizeLine(s)
}
