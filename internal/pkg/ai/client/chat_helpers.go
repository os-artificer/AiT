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

package client

import (
	"github.com/os-artificer/AiT/internal/pkg/ai/eino"
)

// buildChatMessages constructs the full message list for a chat round via eino prompt_builder.
func buildChatMessages(scope Scope, history []ChatMessage, user string, env AssistEnv, meta PromptMeta) []completionMessage {
	return plainToCompletion(eino.BuildPlainChatMessages(scope, history, user, env, meta))
}

// assistMessages builds a minimal message pair for assist-style prompts via eino.
func assistMessages(userPrompt, scrollback, remoteCwd string, meta PromptMeta) []completionMessage {
	return plainToCompletion(eino.PlainFromSchema(eino.AssistMessages(userPrompt, scrollback, remoteCwd, meta)))
}

func plainToCompletion(msgs []eino.PlainMessage) []completionMessage {
	out := make([]completionMessage, len(msgs))
	for i, m := range msgs {
		out[i] = completionMessage{Role: m.Role, Content: m.Content}
	}
	return out
}
