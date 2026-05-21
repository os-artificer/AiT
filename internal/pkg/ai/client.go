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
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
	"github.com/os-artificer/AiT/internal/pkg/platform"

	openai "github.com/sashabaranov/go-openai"
)

// Client wraps Deepseek (OpenAI-compatible) API.
type Client struct {
	api *openai.Client
	cfg *config.Config
}

// NewClient builds a Deepseek (OpenAI-compatible) client from cfg and the keyring API key.
// cfg supplies BaseURL and Model. Returns (*Client, nil) on success.
// Errors: creds.GetDeepseekAPIKey failure wrapped with setup hint.
func NewClient(cfg *config.Config) (*Client, error) {
	key, err := creds.GetDeepseekAPIKey()
	if err != nil {
		return nil, fmt.Errorf("deepseek API key not set (ait config set deepseek.api_key): %w", err)
	}

	ocfg := openai.DefaultConfig(key)
	ocfg.BaseURL = cfg.Deepseek.BaseURL
	return &Client{
		api: openai.NewClientWithConfig(ocfg),
		cfg: cfg,
	}, nil
}

// Assist sends userPrompt to Deepseek with optional scrollback and remoteCwd context.
// ctx cancels the HTTP request. scrollback is truncated for the prompt; remoteCwd may be empty.
// Returns the assistant reply text. Errors: API/network failures; fmt.Errorf when the response has no choices.
func (c *Client) Assist(ctx context.Context, userPrompt, scrollback, remoteCwd string) (string, error) {
	sys := `You are AiT, an AI assistant inside an SSH terminal client.
Suggest shell commands but do NOT claim you executed them.
Be concise. For security, never ask the user to paste secrets.`

	var ctxParts []string
	if remoteCwd != "" {
		ctxParts = append(ctxParts, "remote_cwd: "+remoteCwd)
	}
	ctxParts = append(ctxParts, "local_os: "+platform.OSDescription())
	if scrollback != "" {
		ctxParts = append(ctxParts, "recent_terminal_output:\n"+truncate(scrollback, 4000))
	}

	user := userPrompt
	if len(ctxParts) > 0 {
		user = strings.Join(ctxParts, "\n\n") + "\n\n---\n\n" + userPrompt
	}

	resp, err := c.api.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.cfg.Deepseek.Model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: sys},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response from deepseek")
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

// SuggestInstall asks Deepseek for brief install instructions for command on the local OS.
// command is the missing binary name (e.g. "rsync"). Returns the same errors as Assist.
func (c *Client) SuggestInstall(ctx context.Context, command string) (string, error) {
	prompt := fmt.Sprintf(
		"The command %q is not installed on this system (%s). "+
			"Give brief install instructions for the most common package manager on this OS. "+
			"Do not run commands; only suggest.",
		command, platform.OSDescription(),
	)
	return c.Assist(ctx, prompt, "", "")
}
