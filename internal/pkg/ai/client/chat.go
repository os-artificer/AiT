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
	"context"
	"fmt"
	"strings"
)

// Chat sends history plus user to the configured AI provider(s) with failover.
// Returns the clean reply text, optional thinking/reasoning content, updated
// conversation history, and any error. The thinking field contains reasoning
// extracted from MiniMax in-content tags or DeepSeek reasoning_content; it is
// not stored in the updated history to avoid wasting tokens on subsequent turns.
func (c *Client) Chat(
	ctx context.Context, history []ChatMessage, user string, env AssistEnv,
) (reply, thinking string, updated []ChatMessage, err error) {
	user = strings.TrimSpace(user)
	if user == "" {
		return "", "", history, fmt.Errorf("%s", errEmptyPrompt)
	}

	reply, thinking, err = c.completion(ctx, user, len(history), func(meta PromptMeta) []completionMessage {
		return buildChatMessages(env.Scope, history, user, env, meta)
	})
	if err != nil {
		return "", "", history, err
	}

	updated = append(append([]ChatMessage{}, history...), ChatMessage{Role: "user", Content: user}, ChatMessage{Role: "assistant", Content: reply})
	return reply, thinking, updated, nil
}
