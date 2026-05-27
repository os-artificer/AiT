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
	openai "github.com/sashabaranov/go-openai"
)

func assistMessages(userPrompt, scrollback, remoteCwd string, meta PromptMeta) []openai.ChatCompletionMessage {
	env := AssistEnv{Scope: ScopeLocal, Scrollback: scrollback, RemoteCwd: remoteCwd}
	user := prependEnvContext(userPrompt, env)
	return []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: systemPrompt(ScopeLocal, env, meta)},
		{Role: openai.ChatMessageRoleUser, Content: user},
	}
}
