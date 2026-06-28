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

// ChatAgent sends a turn in Agent mode (tool-style JSON actions). Uses native tools when supported.
// Returns the clean reply text, optional thinking/reasoning content, updated
// conversation history, and any error. Thinking is not stored in updated history.
func (c *Client) ChatAgent(
	ctx context.Context, history []ChatMessage, user string, env AssistEnv,
) (reply, thinking string, updated []ChatMessage, err error) {
	user = strings.TrimSpace(user)
	if user == "" {
		return "", "", history, fmt.Errorf("%s", errEmptyPrompt)
	}

	env.AgentMode = true
	reply, thinking, err = c.completionAgent(ctx, user, len(history), func(meta PromptMeta) []completionMessage {
		return buildChatMessages(env.Scope, history, user, env, meta)
	})
	if err != nil {
		return "", "", history, err
	}

	updated = append(append([]ChatMessage{}, history...), ChatMessage{Role: "user", Content: user}, ChatMessage{Role: "assistant", Content: reply})
	return reply, thinking, updated, nil
}
