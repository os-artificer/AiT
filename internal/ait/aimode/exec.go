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
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

// sessionState holds per-REPL AI chat history and exec trust.
type sessionState struct {
	history         []ai.ChatMessage
	sessionTrusted  bool
	trustedPatterns map[string]struct{}
	policy          ExecPolicy
	interaction     InteractionMode
}

// ExecuteLine handles one AI mode line. Returns exitAI true when the user typed exit/quit.
// hist holds REPL input history for the history builtin; may be nil in tests.
func ExecuteLine(ctx context.Context, opt Options, st *sessionState, line string, hist *terminal.LineHistory) (exitAI bool, err error) {
	line = normalizeLine(line)
	if line == "" {
		return false, nil
	}

	word, rest := firstWord(line)
	switch word {
	case replCmdExit, replCmdQuit:
		clearExecTrust(st)
		return true, nil
	case replCmdHelp, replArgH, replArgHelp:
		PrintHelp(opt.Out, opt.Scope)
		return false, nil
	case replCmdClear:
		st.history = nil
		aimWrite(opt.Out, msgConversationClearedCRLF)
		return false, nil
	case replCmdHistory:
		parts := strings.Fields(line)
		var args []string
		if len(parts) > 1 {
			args = parts[1:]
		}
		return false, terminal.RunHistoryCommand(opt.Out, hist, args)
	case replCmdMode:
		return false, setMode(ctx, opt, st, rest, opt.Out)
	case replCmdApprove:
		return false, setApprove(st, rest, opt.Out, opt.ErrOut)
	case replCmdTrust:
		return false, setApprove(st, "session", opt.Out, opt.ErrOut)
	case replCmdPolicy:
		return false, setPolicyDeprecated(st, rest, opt.Out, opt.ErrOut)
	case replCmdModel:
		return false, setModel(ctx, opt, rest, opt.Out)
	case replCmdRun:
		return false, execRun(ctx, opt, st, rest)
	case replCmdAit:
		if opt.Scope != ai.ScopeLocal {
			return false, aitSubcmdLocalOnlyErr()
		}
		return false, execAit(ctx, opt, st, rest)
	default:
		if st.interaction == ModeAgent {
			return false, runAgentTurn(ctx, opt, st, line)
		}
		return false, askAI(ctx, opt, st, line)
	}
}

func execRun(ctx context.Context, opt Options, st *sessionState, cmd string) error {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return usageRunCommandErr()
	}

	risk := ClassifyRisk(cmd)
	if risk == RiskInteractive {
		aimWrite(opt.Out, msgInteractiveSkipCRLF)
		return nil
	}

	ok, err := confirmExecRisk(optWithPolicy(opt, st.policy), st, cmd, CommandPatternKey(cmd), risk)
	if err != nil {
		return err
	}
	if !ok {
		aimWrite(opt.Out, msgSkippedCRLF)
		return nil
	}

	if opt.Scope == ai.ScopeRemote {
		if opt.Remote == nil {
			return remoteRunnerNotConfiguredErr()
		}
		return opt.Remote.Run(cmd)
	}

	if opt.Local == nil {
		return localRunnerNotConfiguredErr()
	}
	return opt.Local.Run(ctx, cmd)
}

func execAit(ctx context.Context, opt Options, st *sessionState, argsLine string) error {
	_ = ctx
	argsLine = strings.TrimSpace(argsLine)
	if argsLine == "" {
		return usageAitSubcommandErr()
	}

	risk := ClassifyAitRisk(argsLine)
	label := "ait " + argsLine
	ok, err := confirmExecRisk(optWithPolicy(opt, st.policy), st, label, CommandPatternKeyAit(argsLine), risk)
	if err != nil {
		return err
	}
	if !ok {
		aimWrite(opt.Out, msgSkippedCRLF)
		return nil
	}
	if opt.CLI == nil {
		return cliExecutorNotConfiguredErr()
	}

	argv := strings.Fields(argsLine)
	return opt.CLI(argv, opt.Out, opt.ErrOut)
}

func setModel(ctx context.Context, opt Options, rest string, out io.Writer) error {
	if opt.AI == nil {
		return ai.NotConfiguredErr()
	}
	rest = strings.ToLower(strings.TrimSpace(rest))
	if rest == "" || rest == replModelStatus {
		aimWrite(out, opt.AI.StatusLine()+"\r\n")
		return nil
	}
	if rest == replModelList {
		report, err := opt.AI.ListModels(ctx)
		if err != nil {
			return err
		}
		aimWrite(out, strings.ReplaceAll(report.Format(), "\n", "\r\n")+"\r\n")
		return nil
	}
	if rest == "ollama" {
		return fmt.Errorf("%s", errCannotSetOllamaProvider)
	}
	if err := opt.AI.SetSessionProvider(rest); err != nil {
		return err
	}
	aimWrite(out, fmt.Sprintf(msgAIProviderSetFmt, rest))
	return nil
}

func askAI(ctx context.Context, opt Options, st *sessionState, question string) error {
	if opt.AI == nil {
		return ai.NotConfiguredErr()
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
