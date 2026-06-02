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
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/os-artificer/AiT/internal/ait/ai/mode/approve"
	"github.com/os-artificer/AiT/internal/pkg/ai"
)

// InteractionMode is Ask (Q&A only) or Agent (plan and execute commands).
type InteractionMode int

const (
	// ModeAsk answers questions without auto-running shell commands.
	ModeAsk InteractionMode = iota
	// ModeAgent may invoke run_shell / ait_cli tools after confirmation.
	ModeAgent
)

// setMode handles "mode", "mode ask", "mode agent".
func setMode(ctx context.Context, opt Options, st *sessionState, rest string, out io.Writer) error {
	_ = ctx
	rest = strings.ToLower(strings.TrimSpace(rest))
	switch rest {
	case "":
		aimWrite(out, formatModeStatus(opt, st)+"\r\n")
		return nil
	case "ask":
		st.interaction = ModeAsk
		aimWrite(out, msgModeAskCRLF)
		return nil
	case "agent":
		st.interaction = ModeAgent
		aimWrite(out, msgModeAgentCRLF)
		return nil
	default:
		return unknownModeErr(rest)
	}
}

// setApprove handles approve each|session and deprecated policy aliases.
func setApprove(st *sessionState, rest string, out io.Writer, errOut io.Writer) error {
	rest = strings.ToLower(strings.TrimSpace(rest))
	switch rest {
	case "", "status":
		aimWrite(out, formatApproveStatus(st)+"\r\n")
		return nil
	case "each", "per-cmd", "per_cmd":
		st.Trust.Policy = approve.PolicyPerCmd
		st.Trust.Clear()
		aimWrite(out, msgApproveEachCRLF)
		return nil
	case "session", "once", "session_once", "trust":
		st.Trust.Policy = approve.PolicySessionOnce
		st.Trust.SessionTrusted = false
		st.Trust.TrustedPatterns = nil
		aimWrite(out, msgApproveSessionCRLF)
		return nil
	default:
		return unknownApproveErr(rest)
	}
}

func setPolicyDeprecated(st *sessionState, rest string, out io.Writer, errOut io.Writer) error {
	aimWrite(errOut, msgPolicyDeprecated+"\n")
	return setApprove(st, mapPolicyToApprove(rest), out, errOut)
}

func mapPolicyToApprove(rest string) string {
	switch strings.ToLower(strings.TrimSpace(rest)) {
	case "per-cmd", "per_cmd", "":
		return "each"
	case "once", "session_once", "trust":
		return "session"
	default:
		return rest
	}
}

func formatModeStatus(opt Options, st *sessionState) string {
	return msgPrefixAiT + "mode " + interactionLabel(st.interaction) +
		" · approve " + approveLabel(st.Trust.Policy) +
		" · " + providerStatusSuffix(opt)
}

func formatApproveStatus(st *sessionState) string {
	return msgPrefixAiT + "approve: " + approveLabel(st.Trust.Policy)
}

func interactionLabel(m InteractionMode) string {
	if m == ModeAgent {
		return "agent"
	}
	return "ask"
}

func approveLabel(p approve.Policy) string {
	if p == approve.PolicySessionOnce {
		return "session"
	}
	return "each"
}

func providerStatusSuffix(opt Options) string {
	if opt.AI == nil {
		return "ai not configured"
	}
	line := opt.AI.StatusLine()
	line = strings.TrimPrefix(line, "AI provider: ")
	return line
}

func newSessionState(opt Options) *sessionState {
	st := &sessionState{
		interaction: ModeAgent,
		Trust:       approve.Trust{Policy: opt.ExecPolicy},
	}
	if st.Trust.Policy == 0 {
		st.Trust.Policy = approve.PolicyPerCmd
	}
	return st
}

// Prompt returns the AI mode prompt for scope, host label, and interaction mode.
func Prompt(scope ai.Scope, hostLabel string, mode InteractionMode) string {
	modeLabel := interactionPromptLabel(mode)
	if scope == ai.ScopeRemote {
		if hostLabel != "" {
			return fmt.Sprintf("AiT AI(%s) %s> ", modeLabel, hostLabel)
		}
		return fmt.Sprintf("AiT AI(%s) remote> ", modeLabel)
	}
	return fmt.Sprintf("AiT AI(%s)> ", modeLabel)
}

func interactionPromptLabel(m InteractionMode) string {
	if m == ModeAgent {
		return "Agent"
	}
	return "Ask"
}
