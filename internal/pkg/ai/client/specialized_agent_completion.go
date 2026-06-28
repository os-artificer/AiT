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
	"github.com/os-artificer/AiT/internal/pkg/ai/eino"

	"github.com/cloudwego/eino/schema"
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
		Type     string `json:"type"`
		Text     string `json:"text"`
		Thinking string `json:"thinking"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type openAIChatRequest struct {
	Model    string              `json:"model"`
	Messages []openAIChatMessage `json:"messages"`
	Tools    []openAITool        `json:"tools,omitempty"`
}

type openAIChatMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []openAIToolCall `json:"tool_calls,omitempty"`
}

type openAITool struct {
	Type     string             `json:"type"`
	Function openAIToolFunction `json:"function"`
}

type openAIToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type openAIToolCall struct {
	Type     string             `json:"type"`
	Function openAIFunctionCall `json:"function"`
}

type openAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message openAIChatMessage `json:"message"`
	} `json:"choices"`
}

var agentOpenAITools = openAIToolsFromDefs()

func openAIToolsFromDefs() []openAITool {
	defs := agent.ToolDefinitions()
	out := make([]openAITool, len(defs))
	for i, def := range defs {
		out[i] = openAITool{
			Type: "function",
			Function: openAIToolFunction{
				Name:        def.Name,
				Description: def.Desc,
				Parameters:  json.RawMessage(def.ParamJSON),
			},
		}
	}
	return out
}

func specializedChatAnthropic(ctx context.Context, slot *providerSlot, messages []completionMessage) (string, string, error) {
	return specializedAnthropic(ctx, slot, messages, false)
}

func specializedAgentAnthropic(ctx context.Context, slot *providerSlot, messages []completionMessage) (string, string, error) {
	return specializedAnthropic(ctx, slot, messages, true)
}

func specializedAnthropic(ctx context.Context, slot *providerSlot, messages []completionMessage, forAgent bool) (string, string, error) {
	system, msgs := splitAnthropicMessages(messages)
	body := anthropicRequest{
		Model:     slot.model,
		MaxTokens: anthropicMaxTokens,
		System:    system,
		Messages:  msgs,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", "", err
	}

	url := strings.TrimSuffix(slot.baseURL, "/") + anthropicMessagesPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", "", err
	}
	req.Header.Set(anthropicHeaderContentType, anthropicContentType)
	req.Header.Set(anthropicHeaderAPIKey, slot.key)
	req.Header.Set(anthropicHeaderVersion, anthropicVersion)

	httpClient := slot.http
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}
	if res.StatusCode == httpStatusTooManyRequests || res.StatusCode >= httpStatusInternalError {
		return "", "", fmt.Errorf("anthropic api %s: status %d", slot.id, res.StatusCode)
	}
	if res.StatusCode >= 400 {
		return "", "", fmt.Errorf("anthropic api %s: status %d: %s", slot.id, res.StatusCode, truncate(string(data), 200))
	}

	var parsed anthropicResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", "", err
	}
	if parsed.Error != nil {
		return "", "", fmt.Errorf("anthropic api %s: %s", slot.id, parsed.Error.Message)
	}
	if len(parsed.Content) == 0 {
		return "", "", fmt.Errorf(errEmptyResponseFmt, slot.id)
	}

	var parts []string
	var thinkingParts []string
	for _, block := range parsed.Content {
		switch block.Type {
		case anthropicBlockThinking:
			content := block.Thinking
			if content == "" {
				content = block.Text
			}
			if content != "" {
				thinkingParts = append(thinkingParts, content)
			}
		case anthropicBlockRedactedThinking:
		default:
			if block.Text != "" {
				parts = append(parts, block.Text)
			}
		}
	}
	if len(parts) == 0 {
		return "", "", fmt.Errorf(errEmptyResponseFmt, slot.id)
	}
	content := strings.TrimSpace(strings.Join(parts, "\n"))
	thinking := strings.TrimSpace(strings.Join(thinkingParts, "\n"))

	if forAgent {
		reply, agentThinking := eino.ProcessAnthropicAgentText(content, thinking)
		return reply, agentThinking, nil
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

func specializedAgentOpenAI(ctx context.Context, slot *providerSlot, messages []completionMessage) (string, string, error) {
	oaiMsgs := make([]openAIChatMessage, len(messages))
	for i, m := range messages {
		oaiMsgs[i] = openAIChatMessage{Role: m.Role, Content: m.Content}
	}
	body := openAIChatRequest{Model: slot.model, Messages: oaiMsgs, Tools: agentOpenAITools}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", "", err
	}

	url := strings.TrimSuffix(slot.baseURL, "/") + openAIChatCompletionsPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", "", err
	}
	req.Header.Set(openAIHeaderContentType, openAIContentType)
	req.Header.Set(openAIHeaderAuthorization, openAIBearerPrefix+slot.key)

	httpClient := slot.http
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}
	if res.StatusCode >= 400 {
		return "", "", fmt.Errorf("%s: status %d: %s", slot.id, res.StatusCode, truncate(string(data), 200))
	}

	var parsed openAIChatResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", "", err
	}
	if len(parsed.Choices) == 0 {
		return "", "", fmt.Errorf(errEmptyResponseFmt, slot.id)
	}

	msg := parsed.Choices[0].Message
	return eino.ProcessAgentMessage(openAIAgentMessageToSchema(msg))
}

func openAIAgentMessageToSchema(msg openAIChatMessage) *schema.Message {
	out := &schema.Message{
		Role:    schema.Assistant,
		Content: strings.TrimSpace(msg.Content),
	}
	if len(msg.ToolCalls) > 0 {
		out.ToolCalls = make([]schema.ToolCall, len(msg.ToolCalls))
		for i, tc := range msg.ToolCalls {
			out.ToolCalls[i] = schema.ToolCall{
				Function: schema.FunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
	}
	return out
}

func splitAnthropicMessages(messages []completionMessage) (system string, out []anthropicMessage) {
	for _, m := range messages {
		switch m.Role {
		case msgRoleSystem:
			if system != "" {
				system += "\n\n"
			}
			system += m.Content
		case msgRoleUser:
			out = append(out, anthropicMessage{Role: anthropicRoleUser, Content: m.Content})
		case msgRoleAssistant:
			out = append(out, anthropicMessage{Role: anthropicRoleAssistant, Content: m.Content})
		}
	}
	return system, out
}
