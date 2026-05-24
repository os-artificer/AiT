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
	"context"
	"fmt"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/ai"
)

func assistEnv(opt Options, st *sessionState) ai.AssistEnv {
	env := ai.AssistEnv{
		Scope:       opt.Scope,
		HostLabel:   opt.HostLabel,
		RemoteCwd:   opt.RemoteCwd,
		AgentMode:   st.interaction == ModeAgent,
		ApproveEach: st.policy == ExecPolicyPerCmd,
	}
	if opt.Scrollback != nil {
		env.Scrollback = opt.Scrollback()
	}
	return env
}

func executeAgentAction(ctx context.Context, opt Options, st *sessionState, act AgentAction) (toolResult string, final bool, err error) {
	switch {
	case actionIsAnswer(act):
		text := act.Text
		if text == "" {
			text = "(done)"
		}
		aimWrite(opt.Out, text+"\r\n")
		return "", true, nil
	case actionIsRun(act):
		return runAgentShell(ctx, opt, st, act.Command)
	case actionIsAit(act):
		return runAgentAit(ctx, opt, st, act.Args)
	default:
		return "error: unknown action " + act.Action, false, nil
	}
}

func runAgentShell(ctx context.Context, opt Options, st *sessionState, cmd string) (string, bool, error) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return "error: empty command", false, nil
	}

	risk := ClassifyRisk(cmd)
	if risk == RiskInteractive {
		aimWrite(opt.Out, msgInteractiveSkipCRLF)
		return "error: interactive_tty_not_supported", false, nil
	}

	label := "run " + cmd
	ok, err := confirmExecRisk(optWithPolicy(opt, st.policy), st, label, CommandPatternKey(cmd), risk)
	if err != nil {
		return "", false, err
	}
	if !ok {
		aimWrite(opt.Out, msgSkippedCRLF)
		return "skipped by user", false, nil
	}

	out, runErr := captureRun(ctx, opt, cmd)
	if runErr != nil {
		return fmt.Sprintf("exit error: %v\noutput:\n%s", runErr, out), false, nil
	}
	return out, false, nil
}

func runAgentAit(ctx context.Context, opt Options, st *sessionState, argsLine string) (string, bool, error) {
	argsLine = strings.TrimSpace(argsLine)
	if argsLine == "" {
		return "error: empty ait args", false, nil
	}
	if opt.CLI == nil {
		return "error: ait CLI not configured", false, nil
	}

	risk := ClassifyAitRisk(argsLine)
	label := "ait " + argsLine
	ok, err := confirmExecRisk(optWithPolicy(opt, st.policy), st, label, CommandPatternKeyAit(argsLine), risk)
	if err != nil {
		return "", false, err
	}
	if !ok {
		aimWrite(opt.Out, msgSkippedCRLF)
		return "skipped by user", false, nil
	}

	var buf strings.Builder
	err = opt.CLI(strings.Fields(argsLine), &buf, opt.ErrOut)
	if err != nil {
		return fmt.Sprintf("error: %v", err), false, nil
	}
	return buf.String(), false, nil
}

func captureRun(ctx context.Context, opt Options, cmd string) (string, error) {
	if opt.Capture != nil {
		return opt.Capture.RunCapture(ctx, cmd)
	}
	if opt.Scope == ai.ScopeRemote && opt.Remote != nil {
		if err := opt.Remote.Run(cmd); err != nil {
			return "", err
		}
		return "(command sent to remote shell)", nil
	}
	if opt.Local != nil {
		cap := DefaultLocalCaptureRunner()
		return cap.RunCapture(ctx, cmd)
	}
	return "", fmt.Errorf("no runner configured")
}
