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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/ai/agent"

	openai "github.com/sashabaranov/go-openai"
)

const anthropicVersion = "2023-06-01"

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type      string `json:"type"`
		Text      string `json:"text"`
		Thinking  string `json:"thinking"`
		Signature string `json:"signature,omitempty"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func (ep *endpoint) chatAnthropic(ctx context.Context, messages []openai.ChatCompletionMessage, forAgent bool) (string, string, error) {
	system, msgs := splitAnthropicMessages(messages)
	body := anthropicRequest{
		Model:     ep.model,
		MaxTokens: anthropicMaxTokens,
		System:    system,
		Messages:  msgs,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", "", err
	}

	url := strings.TrimSuffix(ep.baseURL, "/") + anthropicMessagesPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", "", err
	}
	req.Header.Set(anthropicHeaderContentType, anthropicContentType)
	req.Header.Set(anthropicHeaderAPIKey, ep.key)
	req.Header.Set(anthropicHeaderVersion, anthropicVersion)

	client := ep.http
	if client == nil {
		client = http.DefaultClient
	}
	res, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}
	if res.StatusCode == httpStatusTooManyRequests || res.StatusCode >= httpStatusInternalError {
		return "", "", fmt.Errorf("anthropic api %s: status %d", ep.id, res.StatusCode)
	}
	if res.StatusCode >= 400 {
		return "", "", fmt.Errorf("anthropic api %s: status %d: %s", ep.id, res.StatusCode, truncate(string(data), 200))
	}

	var parsed anthropicResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", "", err
	}
	if parsed.Error != nil {
		return "", "", fmt.Errorf("anthropic api %s: %s", ep.id, parsed.Error.Message)
	}
	if len(parsed.Content) == 0 {
		return "", "", fmt.Errorf(errEmptyResponseFmt, ep.id)
	}
	var parts []string
	var thinkingParts []string
	for _, block := range parsed.Content {
		switch block.Type {
		case anthropicBlockThinking:
			// MiniMax Anthropic uses "thinking" field; Anthropic SDK uses "thinking" field.
			// Fall back to "text" for providers that put content there.
			content := block.Thinking
			if content == "" {
				content = block.Text
			}
			if content != "" {
				thinkingParts = append(thinkingParts, content)
			}
		case anthropicBlockRedactedThinking:
			// Extended Thinking encrypted block — no readable content, skip silently.
		default:
			if block.Text != "" {
				parts = append(parts, block.Text)
			}
		}
	}
	if len(parts) == 0 {
		return "", "", fmt.Errorf(errEmptyResponseFmt, ep.id)
	}
	content := strings.TrimSpace(strings.Join(parts, "\n"))
	thinking := strings.TrimSpace(strings.Join(thinkingParts, "\n"))

	if forAgent {
		if toolJSON, ok := agent.XMLToolCallToJSON(content); ok {
			return toolJSON, thinking, nil
		}
	}

	reply, extractedThinking := ExtractReply(content)
	if extractedThinking != "" {
		if thinking != "" {
			thinking += "\n"
		}
		thinking += extractedThinking
	}

	return reply, thinking, nil
}

func splitAnthropicMessages(messages []openai.ChatCompletionMessage) (system string, out []anthropicMessage) {
	for _, m := range messages {
		switch m.Role {
		case openai.ChatMessageRoleSystem:
			if system != "" {
				system += "\n\n"
			}
			system += m.Content
		case openai.ChatMessageRoleUser:
			out = append(out, anthropicMessage{Role: anthropicRoleUser, Content: m.Content})
		case openai.ChatMessageRoleAssistant:
			out = append(out, anthropicMessage{Role: anthropicRoleAssistant, Content: m.Content})
		}
	}
	return system, out
}
