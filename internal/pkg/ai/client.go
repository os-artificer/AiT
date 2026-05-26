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

package ai

import (
	"context"
	"fmt"

	"github.com/os-artificer/AiT/internal/pkg/platform"

	openai "github.com/sashabaranov/go-openai"
)

// Assist sends userPrompt with optional scrollback and remoteCwd context via cloud providers only.
func (c *Client) Assist(ctx context.Context, userPrompt, scrollback, remoteCwd string) (string, error) {
	reply, _, err := c.chatWithFailover(ctx, func(meta PromptMeta) []openai.ChatCompletionMessage {
		return assistMessages(userPrompt, scrollback, remoteCwd, meta)
	})
	return reply, err
}

// SuggestInstall asks the AI for brief install instructions for command on the local OS (cloud only).
func (c *Client) SuggestInstall(ctx context.Context, command string) (string, error) {
	prompt := fmt.Sprintf(
		"The command %q is not installed on this system (%s). "+
			"Give brief install instructions for the most common package manager on this OS. "+
			"Do not run commands; only suggest.",
		command, platform.OSDescription(),
	)
	return c.Assist(ctx, prompt, "", "")
}
