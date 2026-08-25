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

package agentcore

import (
	"context"
	"fmt"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/agentcore/approve"
	"github.com/os-artificer/AiT/internal/pkg/agentcore/risk"
	"github.com/os-artificer/AiT/internal/pkg/agentcore/runner"
	"github.com/os-artificer/AiT/internal/pkg/ai"
	pkgagent "github.com/os-artificer/AiT/internal/pkg/ai/agent"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func assistEnv(cfg Config, st *sessionState) ai.AssistEnv {
	scope, host, cwd, scrollFn, _ := resolveDynamic(cfg)
	env := ai.AssistEnv{
		Scope:       scope,
		HostLabel:   host,
		RemoteCwd:   cwd,
		AgentMode:   true,
		ApproveEach: st.Trust.Policy == approve.PolicyPerCmd,
		Domain:      cfg.Domain,
	}
	if scrollFn != nil {
		env.Scrollback = scrollFn()
	}
	if cfg.EnvExtra != "" {
		if env.Scrollback != "" {
			env.Scrollback = cfg.EnvExtra + "\n" + env.Scrollback
		} else {
			env.Scrollback = cfg.EnvExtra
		}
	}
	return env
}

func resolveDynamic(cfg Config) (scope ai.Scope, host, cwd string, scroll func() string, remote RemoteRunner) {
	scope, host, cwd = cfg.Scope, cfg.HostLabel, cfg.RemoteCwd
	scroll = cfg.Scrollback
	remote = cfg.Remote
	if cfg.Dynamic == nil {
		return scope, host, cwd, scroll, remote
	}
	snap := cfg.Dynamic.Snapshot()
	scope = snap.Scope
	if snap.HostLabel != "" {
		host = snap.HostLabel
	}
	if snap.RemoteCwd != "" {
		cwd = snap.RemoteCwd
	}
	if snap.Scrollback != nil {
		scroll = snap.Scrollback
	}
	if snap.Remote != nil {
		remote = snap.Remote
	}
	return scope, host, cwd, scroll, remote
}

func captureRunner(cfg Config) runner.CapturingRunner {
	if cfg.Dynamic != nil {
		if cap := cfg.Dynamic.Snapshot().Capture; cap != nil {
			return cap
		}
	}
	return cfg.Capture
}

func scopeLabel(cfg Config) string {
	scope, host, _, _, _ := resolveDynamic(cfg)
	if scope == ai.ScopeRemote && host != "" {
		return "remote:" + host
	}
	return "local"
}

func emit(cfg Config, e Event) {
	if cfg.Emitter != nil {
		cfg.Emitter.Emit(e)
	}
}

func executeAgentAction(ctx context.Context, cfg Config, st *sessionState, act runner.Action) (toolResult string, final bool, err error) {
	switch {
	case act.IsAnswer():
		text := act.Text
		if text == "" {
			text = "(done)"
		}
		emit(cfg, Event{Kind: EventMessageFinal, Detail: text})
		agentWrite(cfg.Out, text+"\r\n")
		return "", true, nil
	case act.IsRun():
		return runAgentShell(ctx, cfg, st, act.Command)
	case act.IsToolbox():
		return runAgentToolbox(ctx, cfg, st, act.Args)
	case act.IsAit():
		return runAgentAit(ctx, cfg, st, act.Args)
	default:
		return "error: unknown action " + act.Action, false, nil
	}
}

func runAgentShell(ctx context.Context, cfg Config, st *sessionState, cmd string) (string, bool, error) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return "error: empty command", false, nil
	}

	lvl := risk.Classify(cmd)
	if lvl == risk.Interactive {
		agentWrite(cfg.Out, msgInteractiveSkip)
		return "error: interactive_tty_not_supported", false, nil
	}

	label := "run " + cmd
	emit(cfg, Event{Kind: EventToolStarted, Tool: label})
	ok, err := confirmExecRisk(cfg, st, label, approve.PatternKey(cmd), lvl)
	if err != nil {
		return "", false, err
	}
	if !ok {
		emit(cfg, Event{Kind: EventToolFinished, Tool: label, Detail: "skipped"})
		agentWrite(cfg.Out, msgSkipped)
		return "skipped by user", false, nil
	}

	out, runErr := captureRun(ctx, cfg, cmd)
	if runErr != nil {
		result := fmt.Sprintf("exit error: %v\noutput:\n%s", runErr, out)
		emit(cfg, Event{Kind: EventToolOutput, Tool: label, Output: result})
		emit(cfg, Event{Kind: EventToolFinished, Tool: label})
		return result, false, nil
	}
	emit(cfg, Event{Kind: EventToolOutput, Tool: label, Output: out})
	emit(cfg, Event{Kind: EventToolFinished, Tool: label})
	return out, false, nil
}

func runAgentAit(ctx context.Context, cfg Config, st *sessionState, argsLine string) (string, bool, error) {
	_ = ctx
	argsLine = strings.TrimSpace(argsLine)
	if argsLine == "" {
		return "error: empty " + toolname.AISsh + " args", false, nil
	}
	if cfg.CLI == nil {
		return "error: " + toolname.AISsh + " CLI not configured", false, nil
	}

	lvl := risk.ClassifyAit(argsLine)
	label := toolname.AISsh + " " + argsLine
	emit(cfg, Event{Kind: EventToolStarted, Tool: label})
	ok, err := confirmExecRisk(cfg, st, label, approve.PatternKeyAit(argsLine), lvl)
	if err != nil {
		return "", false, err
	}
	if !ok {
		emit(cfg, Event{Kind: EventToolFinished, Tool: label, Detail: "skipped"})
		agentWrite(cfg.Out, msgSkipped)
		return "skipped by user", false, nil
	}

	var buf strings.Builder
	err = cfg.CLI(strings.Fields(argsLine), &buf, cfg.ErrOut)
	if err != nil {
		result := fmt.Sprintf("error: %v", err)
		emit(cfg, Event{Kind: EventToolOutput, Tool: label, Output: result})
		emit(cfg, Event{Kind: EventToolFinished, Tool: label})
		return result, false, nil
	}
	out := buf.String()
	emit(cfg, Event{Kind: EventToolOutput, Tool: label, Output: out})
	emit(cfg, Event{Kind: EventToolFinished, Tool: label})
	return out, false, nil
}

func runAgentToolbox(ctx context.Context, cfg Config, st *sessionState, argsLine string) (string, bool, error) {
	_ = ctx
	argsLine = strings.TrimSpace(argsLine)
	if argsLine == "" {
		return "error: empty toolbox args", false, nil
	}
	if cfg.Toolbox == nil {
		return "error: toolbox executor not configured", false, nil
	}

	argv := strings.Fields(argsLine)
	lvl := risk.ClassifyToolbox(argsLine)
	if lvl == risk.Interactive {
		agentWrite(cfg.Out, msgInteractiveSkip)
		return "error: interactive_tty_not_supported", false, nil
	}

	label := "toolbox " + argsLine
	emit(cfg, Event{Kind: EventToolStarted, Tool: label})
	ok, err := confirmExecRisk(cfg, st, label, approve.PatternKey(argsLine), lvl)
	if err != nil {
		return "", false, err
	}
	if !ok {
		emit(cfg, Event{Kind: EventToolFinished, Tool: label, Detail: "skipped"})
		agentWrite(cfg.Out, msgSkipped)
		return "skipped by user", false, nil
	}

	var buf strings.Builder
	err = cfg.Toolbox(argv, &buf, cfg.ErrOut)
	if err != nil {
		result := fmt.Sprintf("error: %v", err)
		emit(cfg, Event{Kind: EventToolOutput, Tool: label, Output: result})
		emit(cfg, Event{Kind: EventToolFinished, Tool: label})
		return result, false, nil
	}
	out := buf.String()
	emit(cfg, Event{Kind: EventToolOutput, Tool: label, Output: out})
	emit(cfg, Event{Kind: EventToolFinished, Tool: label})
	return out, false, nil
}

func captureRun(ctx context.Context, cfg Config, cmd string) (string, error) {
	if cap := captureRunner(cfg); cap != nil {
		return cap.RunCapture(ctx, cmd)
	}
	scope, _, _, _, remote := resolveDynamic(cfg)
	if scope == ai.ScopeRemote && remote != nil {
		if err := remote.Run(cmd); err != nil {
			return "", err
		}
		return "(command sent to remote shell)", nil
	}
	if cfg.Local != nil {
		cap := runner.DefaultLocalCaptureRunner()
		return cap.RunCapture(ctx, cmd)
	}
	return "", fmt.Errorf("no runner configured")
}

func confirmExecRisk(cfg Config, st *sessionState, label, patternKey string, lvl risk.Level) (bool, error) {
	if lvl == risk.Read {
		emit(cfg, Event{Kind: EventApprovalResolved, Tool: label, Allowed: true, Detail: "auto-read"})
		return true, nil
	}

	needsPrompt := lvl == risk.High || lvl == risk.Low
	if !needsPrompt {
		return true, nil
	}

	emit(cfg, Event{Kind: EventApprovalRequired, Tool: label, Detail: label})

	in := approve.Inputs{
		Trust:       &st.Trust,
		ScopeLabel:  scopeLabel(cfg),
		CmdLabel:    label,
		PatternKey:  patternKey,
		Risk:        lvl,
		Policy:      st.Trust.Policy,
		ConfirmRead: cfg.ConfirmRead,
		Out:         cfg.Out,
		In:          cfg.In,
		Write:       agentWrite,
	}
	ok, err := approve.Confirm(in)
	emit(cfg, Event{Kind: EventApprovalResolved, Tool: label, Allowed: ok})
	return ok, err
}

func envelopeToAction(e pkgagent.AgentEnvelope) runner.Action {
	return runner.EnvelopeToAction(e)
}
