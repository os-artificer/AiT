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
	"strings"

	pkgagent "github.com/os-artificer/AiT/internal/pkg/ai/agent"
	"github.com/os-artificer/AiT/internal/pkg/ai/client"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
)

// ErrAgentMalformedReply is returned when Agent format recovery is exhausted.
var ErrAgentMalformedReply = errhint.NewSentinel("could not parse agent response")

type agentParseRetryState struct {
	turnStartHistory []client.ChatMessage
	formatRetryLeft  int
	formatErrorLeft  int
}

func newAgentParseRetryState(history []client.ChatMessage) agentParseRetryState {
	return agentParseRetryState{
		turnStartHistory: client.CloneHistory(history),
		formatRetryLeft:  1,
		formatErrorLeft:  1,
	}
}

func handleAgentParseFailure(
	st *sessionState,
	state *agentParseRetryState,
	userMsg, reply string,
	parseResult pkgagent.ParseResult,
) (newUserMsg string, err error) {
	if isAgentToolStepUserMsg(userMsg) {
		return handleToolStepParseFailure(st, state, userMsg, reply, parseResult)
	}

	if state.formatRetryLeft > 0 {
		state.formatRetryLeft--
		st.history = client.AppendUserTurn(st.history, userMsg)
		return pkgagent.FormatCorrectionPrompt(userMsg, reply, parseResult.Errs), nil
	}

	st.history = state.turnStartHistory
	return "", formatMalformedReplyErr()
}

func handleToolStepParseFailure(
	st *sessionState,
	state *agentParseRetryState,
	userMsg, reply string,
	parseResult pkgagent.ParseResult,
) (newUserMsg string, err error) {
	if state.formatErrorLeft > 0 {
		state.formatErrorLeft--
		st.history = client.AppendUserTurn(st.history, userMsg)
		return pkgagent.FormatToolResultError(reply, parseResult), nil
	}

	return pkgagent.ToolResultPrefix + pkgagent.FormatErrorPrefix +
		pkgagent.FormatErrorSummary(parseResult.Errs), nil
}

func isAgentToolStepUserMsg(msg string) bool {
	return strings.HasPrefix(msg, pkgagent.ToolResultPrefix)
}

func formatMalformedReplyErr() error {
	return errhint.WithTry(ErrAgentMalformedReply,
		"The model returned a response that could not be parsed as an agent action.",
	)
}
