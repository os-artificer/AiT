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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func (ep *endpoint) chatAnthropic(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error) {
	system, msgs := splitAnthropicMessages(messages)
	body := anthropicRequest{
		Model:     ep.model,
		MaxTokens: 4096,
		System:    system,
		Messages:  msgs,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	url := strings.TrimSuffix(ep.baseURL, "/") + "/v1/messages"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", ep.key)
	req.Header.Set("anthropic-version", anthropicVersion)

	client := ep.http
	if client == nil {
		client = http.DefaultClient
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode == 429 || res.StatusCode >= 500 {
		return "", fmt.Errorf("anthropic api %s: status %d", ep.id, res.StatusCode)
	}
	if res.StatusCode >= 400 {
		return "", fmt.Errorf("anthropic api %s: status %d: %s", ep.id, res.StatusCode, truncate(string(data), 200))
	}

	var parsed anthropicResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", err
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("anthropic api %s: %s", ep.id, parsed.Error.Message)
	}
	if len(parsed.Content) == 0 {
		return "", fmt.Errorf(errEmptyResponseFmt, ep.id)
	}
	var parts []string
	for _, block := range parsed.Content {
		if block.Text != "" {
			parts = append(parts, block.Text)
		}
	}
	if len(parts) == 0 {
		return "", fmt.Errorf(errEmptyResponseFmt, ep.id)
	}
	return strings.TrimSpace(strings.Join(parts, "\n")), nil
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
			out = append(out, anthropicMessage{Role: "user", Content: m.Content})
		case openai.ChatMessageRoleAssistant:
			out = append(out, anthropicMessage{Role: "assistant", Content: m.Content})
		}
	}
	return system, out
}
