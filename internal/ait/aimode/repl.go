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

	"github.com/os-artificer/AiT/internal/ait/aimode/chat"
	"github.com/os-artificer/AiT/internal/ait/shell"
	"github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/terminal"

	"golang.org/x/term"
)

// ErrAINotConfigured is returned when RunREPL is called without a configured AI client.
var ErrAINotConfigured = ai.ErrNotConfigured

// RunREPL runs the AI mode loop until exit/quit or ctx cancel.
// initialQuestion, if non-empty, is processed as the first user turn (!ait ai <q> / !! <q>).
// opt.AI must be non-nil. Returns ErrAINotConfigured when opt.AI is nil.
func RunREPL(ctx context.Context, opt Options, initialQuestion string) error {
	if opt.AI == nil {
		return ai.NotConfiguredErr()
	}

	opt = fillOptionDefaults(opt)
	st := newSessionState(opt)

	if q := stringsTrim(initialQuestion); q != "" {
		if chat.IsEmptyCommand(q) || chat.IsEmptyAgentCommand(q) {
			return ai.ChatMessageRequiredErr()
		}
	}

	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) && (opt.In == os.Stdin || opt.UseTTYLineEdit) {
		return runAITTY(ctx, opt, st, fd, initialQuestion)
	}

	if q := stringsTrim(initialQuestion); q != "" {
		if msg, ok := chat.ParseCommand(q); ok {
			return runChatOnceFromOpt(ctx, opt, msg)
		}
		if msg, ok := chat.ParseAgentCommand(q); ok {
			return runAgentOnceFromOpt(ctx, opt, msg)
		}

		exit, err := ExecuteLine(ctx, opt, st, q, nil)
		if err != nil {
			return err
		}
		if exit || isMetaREPLCommand(q) {
			return nil
		}
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
	hist := terminal.NewReplLineHistory(opt.ErrOut)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		prompt := Prompt(opt.Scope, opt.HostLabel, st.interaction)
		_, _ = io.WriteString(opt.Out, prompt)
		if !sc.Scan() {
			if err := sc.Err(); err != nil {
				return err
			}
			return nil
		}

		exit, err := ExecuteLine(ctx, opt, st, sc.Text(), hist)
		reportAIError(opt.ErrOut, err)
		if exit {
			return nil
		}
	}
}

func runAITTY(ctx context.Context, opt Options, st *sessionState, fd int, initialQuestion string) error {
	old, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("terminal: %w", err)
	}
	defer func() { _ = term.Restore(fd, old) }()

	aimWrite(opt.Out, "Type help for AI mode commands. Use exit or quit to leave.\r\n")

	br := bufio.NewReader(opt.In)
	opt = opt.withTTYConfirm(br)
	hist := terminal.NewReplLineHistory(opt.ErrOut)

	if q := stringsTrim(initialQuestion); q != "" {
		if msg, ok := chat.ParseCommand(q); ok {
			return runChatOnceFromOpt(ctx, opt, msg)
		}
		if msg, ok := chat.ParseAgentCommand(q); ok {
			return runAgentOnceFromOpt(ctx, opt, msg)
		}

		exit, err := ExecuteLine(ctx, opt, st, q, hist)
		if err != nil {
			return err
		}
		if exit || isMetaREPLCommand(q) {
			return nil
		}
		hist.Add(q)
	}

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		prompt := Prompt(opt.Scope, opt.HostLabel, st.interaction)
		_, _ = io.WriteString(opt.Out, prompt)
		line, err := shell.ReadTTYLine(br, opt.Out, prompt, hist)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		exit, err := ExecuteLine(ctx, opt, st, line, hist)
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

// runChatOnceFromOpt adapts Options to chat.Input for the meta one-shot path.
// On error, prints via aimPrintErr; returns the underlying error to the caller.
func runChatOnceFromOpt(ctx context.Context, opt Options, message string) error {
	opt = fillOptionDefaults(opt)
	err := chat.RunOnce(ctx, chat.Input{
		AI:         opt.AI,
		Message:    message,
		Scope:      opt.Scope,
		HostLabel:  opt.HostLabel,
		RemoteCwd:  opt.RemoteCwd,
		Scrollback: opt.Scrollback,
		Out:        opt.Out,
	})
	if err != nil {
		reportAIError(opt.ErrOut, err)
	}
	return err
}

// runAgentOnceFromOpt runs a single agent turn for the meta one-shot path.
// On error, prints via aimPrintErr; returns the underlying error to the caller.
func runAgentOnceFromOpt(ctx context.Context, opt Options, question string) error {
	err := RunAgentOnce(ctx, opt, question)
	if err != nil {
		reportAIError(opt.ErrOut, err)
	}
	return err
}
