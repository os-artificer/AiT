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
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/ai"
)

// sessionState holds per-REPL AI chat history and exec trust.
type sessionState struct {
	history        []ai.ChatMessage
	sessionTrusted bool
	policy         ExecPolicy
}

// ExecuteLine handles one AI mode line. Returns exitAI true when the user typed exit/quit.
func ExecuteLine(ctx context.Context, opt Options, st *sessionState, line string) (exitAI bool, err error) {
	line = normalizeLine(line)
	if line == "" {
		return false, nil
	}

	word, rest := firstWord(line)
	switch word {
	case "exit", "quit":
		st.sessionTrusted = false
		return true, nil
	case "help", "-h", "--help":
		PrintHelp(opt.Out, opt.Scope)
		return false, nil
	case "clear":
		st.history = nil
		aimWrite(opt.Out, "conversation cleared\r\n")
		return false, nil
	case "policy":
		return false, setPolicy(st, rest, opt.Out)
	case "model":
		return false, setModel(ctx, opt, rest, opt.Out)
	case "run":
		return false, execRun(ctx, opt, st, rest)
	case "ait":
		if opt.Scope != ai.ScopeLocal {
			return false, fmt.Errorf("AiT: ait subcommands are local-only; use run for remote shell commands")
		}
		return false, execAit(ctx, opt, st, rest)
	default:
		return false, askAI(ctx, opt, st, line)
	}
}

func setPolicy(st *sessionState, rest string, out io.Writer) error {
	switch strings.ToLower(strings.TrimSpace(rest)) {
	case "per-cmd", "per_cmd", "":
		st.policy = ExecPolicyPerCmd
		st.sessionTrusted = false
		aimWrite(out, "exec policy: per_cmd (confirm each run)\r\n")
	case "once", "session_once", "trust":
		st.policy = ExecPolicySessionOnce
		st.sessionTrusted = false
		aimWrite(out, "exec policy: session_once (confirm once per AI session)\r\n")
	default:
		return fmt.Errorf("AiT: unknown policy %q (try per-cmd or once)", rest)
	}
	return nil
}

func execRun(ctx context.Context, opt Options, st *sessionState, cmd string) error {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return fmt.Errorf("AiT: usage: run <command>")
	}

	ok, err := confirmExec(optWithPolicy(opt, st.policy), cmd, &st.sessionTrusted)
	if err != nil {
		return err
	}
	if !ok {
		aimWrite(opt.Out, "skipped\r\n")
		return nil
	}

	if opt.Scope == ai.ScopeRemote {
		if opt.Remote == nil {
			return fmt.Errorf("AiT: remote runner not configured")
		}
		return opt.Remote.Run(cmd)
	}

	if opt.Local == nil {
		return fmt.Errorf("AiT: local runner not configured")
	}
	return opt.Local.Run(ctx, cmd)
}

func execAit(ctx context.Context, opt Options, st *sessionState, argsLine string) error {
	_ = ctx
	argsLine = strings.TrimSpace(argsLine)
	if argsLine == "" {
		return fmt.Errorf("AiT: usage: ait <subcommand> [args...]")
	}

	ok, err := confirmExec(optWithPolicy(opt, st.policy), "ait "+argsLine, &st.sessionTrusted)
	if err != nil {
		return err
	}
	if !ok {
		aimWrite(opt.Out, "skipped\r\n")
		return nil
	}
	if opt.CLI == nil {
		return fmt.Errorf("AiT: CLI executor not configured")
	}

	argv := strings.Fields(argsLine)
	return opt.CLI(argv, opt.Out, opt.ErrOut)
}

func setModel(ctx context.Context, opt Options, rest string, out io.Writer) error {
	if opt.AI == nil {
		return ErrAINotConfigured
	}
	rest = strings.ToLower(strings.TrimSpace(rest))
	if rest == "" || rest == "status" {
		meta := opt.AI.StatusLine()
		aimWrite(out, meta+"\r\n")
		return nil
	}
	if rest == "list" {
		report, err := opt.AI.ListModels(ctx)
		if err != nil {
			return err
		}
		aimWrite(out, strings.ReplaceAll(report.Format(), "\n", "\r\n")+"\r\n")
		return nil
	}
	if rest == "ollama" {
		return fmt.Errorf("AiT: cannot set session provider to ollama (use tiered routing for local models)")
	}
	if err := opt.AI.SetSessionProvider(rest); err != nil {
		return err
	}
	aimWrite(out, fmt.Sprintf("AI provider set to %s\r\n", rest))
	return nil
}

func askAI(ctx context.Context, opt Options, st *sessionState, question string) error {
	if opt.AI == nil {
		return ErrAINotConfigured
	}

	env := ai.AssistEnv{Scope: opt.Scope, HostLabel: opt.HostLabel, RemoteCwd: opt.RemoteCwd}
	if opt.Scrollback != nil {
		env.Scrollback = opt.Scrollback()
	}

	reply, updated, err := opt.AI.Chat(ctx, st.history, question, env)
	if err != nil {
		return err
	}

	st.history = updated
	aimWrite(opt.Out, reply+"\r\n")
	return nil
}

func optWithPolicy(opt Options, p ExecPolicy) Options {
	opt.ExecPolicy = p
	return opt
}

// PrintHelp writes AI mode REPL help to w.
func PrintHelp(w io.Writer, scope ai.Scope) {
	var buf bytes.Buffer
	_ = helpdoc.WriteAIREPLHelp(&buf, scope)
	aimWrite(w, buf.String())
}
