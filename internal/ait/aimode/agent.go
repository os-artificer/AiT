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

	"github.com/os-artificer/AiT/internal/pkg/ai"
)

// runAgentTurn handles one user message in Agent interaction mode.
func runAgentTurn(ctx context.Context, opt Options, st *sessionState, question string) error {
	if opt.AI == nil {
		return ai.NotConfiguredErr()
	}

	env := assistEnv(opt, st)
	userMsg := question

	for step := 0; step < agentMaxSteps; step++ {
		reply, updated, err := opt.AI.ChatAgent(ctx, st.history, userMsg, env)
		if err != nil {
			return err
		}
		st.history = updated

		act, ok := parseAgentAction(reply)
		if !ok {
			aimWrite(opt.Out, reply+"\r\n")
			return nil
		}

		result, done, err := executeAgentAction(ctx, opt, st, act)
		if err != nil {
			return err
		}
		if done {
			return nil
		}

		userMsg = "tool_result:\n" + result
	}

	aimWrite(opt.Out, msgPrefixAiT+"agent step limit reached\r\n")
	return nil
}
