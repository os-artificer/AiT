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

package mode

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/os-artificer/AiT/internal/ai-ssh/ai/mode/approve"
	"github.com/os-artificer/AiT/internal/ai-ssh/ai/mode/risk"
	"github.com/os-artificer/AiT/internal/ai-ssh/shell"
	"github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/config"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// sessionState holds per-REPL AI chat history and embeds approve.Trust for exec confirmation state.
type sessionState struct {
	history       []ai.ChatMessage
	interaction   InteractionMode
	approve.Trust // embedded value: Policy / SessionTrusted / TrustedPatterns and helper methods
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
	case shell.ReplCmdExit, shell.ReplCmdQuit:
		st.Trust.Clear()
		return true, nil
	case pkghelpdoc.HelpArgCmd, pkghelpdoc.HelpArgH, pkghelpdoc.HelpArgHelp:
		PrintHelp(opt.Out, opt.Scope)
		return false, nil
	case shell.ReplCmdClear:
		st.history = nil
		aimWrite(opt.Out, msgConversationClearedCRLF)
		return false, nil
	case shell.ReplCmdHistory:
		parts := strings.Fields(line)
		var args []string
		if len(parts) > 1 {
			args = parts[1:]
		}
		return false, terminal.RunHistoryCommand(terminal.NewCRLFWriter(opt.Out), hist, args)
	case replCmdMode:
		return false, setMode(ctx, opt, st, rest, opt.Out)
	case replCmdApprove:
		return false, setApprove(st, rest, opt.Out, opt.ErrOut)
	case replCmdTrust:
		return false, setApprove(st, labelApproveSession, opt.Out, opt.ErrOut)
	case replCmdPolicy:
		return false, setPolicyDeprecated(st, rest, opt.Out, opt.ErrOut)
	case replCmdModel:
		return false, setModel(ctx, opt, rest, opt.Out)
	case replCmdRun:
		return false, execRun(ctx, opt, st, rest)
	case replCmdAISsh:
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

	lvl := risk.Classify(cmd)
	if lvl == risk.Interactive {
		aimWrite(opt.Out, msgInteractiveSkipCRLF)
		return nil
	}

	ok, err := confirmExecRisk(optWithPolicy(opt, st.Trust.Policy), st, cmd, approve.PatternKey(cmd), lvl)
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

	lvl := risk.ClassifyAit(argsLine)
	label := toolname.AISsh + " " + argsLine
	ok, err := confirmExecRisk(optWithPolicy(opt, st.Trust.Policy), st, label, approve.PatternKeyAit(argsLine), lvl)
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
	return opt.CLI(argv, terminal.NewCRLFWriter(opt.Out), terminal.NewCRLFWriter(opt.ErrOut))
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
	if rest == config.ProviderOllama {
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

	reply, _, updated, err := opt.AI.Chat(ctx, st.history, question, env)
	if err != nil {
		return err
	}

	st.history = updated
	aimWrite(opt.Out, reply+"\r\n")
	return nil
}

func optWithPolicy(opt Options, p approve.Policy) Options {
	opt.ExecPolicy = p
	return opt
}

// PrintHelp writes AI mode REPL help to w.
func PrintHelp(w io.Writer, scope ai.Scope) {
	var buf bytes.Buffer
	_ = WriteREPLHelp(&buf, scope)
	aimWrite(w, buf.String())
}

// confirmExec returns whether the user approved running cmd under opt's scope and policy.
func confirmExec(opt Options, st *sessionState, cmdLabel, patternKey string, lvl risk.Level) (bool, error) {
	return confirmExecRisk(opt, st, cmdLabel, patternKey, lvl)
}

// confirmExecRisk adapts aimode Options to approve.Inputs and delegates to approve.Confirm.
func confirmExecRisk(opt Options, st *sessionState, cmdLabel, patternKey string, lvl risk.Level) (bool, error) {
	return approve.Confirm(approve.Inputs{
		Trust:       &st.Trust,
		ScopeLabel:  scopeLabel(opt),
		CmdLabel:    cmdLabel,
		PatternKey:  patternKey,
		Risk:        lvl,
		Policy:      opt.ExecPolicy,
		ConfirmRead: opt.ConfirmRead,
		Out:         opt.Out,
		In:          opt.In,
		Write:       aimWrite,
	})
}

func scopeLabel(opt Options) string {
	if opt.Scope == ai.ScopeRemote {
		if opt.HostLabel != "" {
			return opt.HostLabel
		}
		return "remote"
	}
	return "local"
}
