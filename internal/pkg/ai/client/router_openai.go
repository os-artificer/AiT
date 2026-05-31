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

	"github.com/os-artificer/AiT/internal/pkg/ai/agent"

	openai "github.com/sashabaranov/go-openai"
)

func (ep *endpoint) chatOpenAI(ctx context.Context, messages []openai.ChatCompletionMessage) (string, string, error) {
	return ep.chatOpenAIRequest(ctx, messages, nil)
}

func (ep *endpoint) chatOpenAIAgent(ctx context.Context, messages []openai.ChatCompletionMessage) (string, string, error) {
	return ep.chatOpenAIRequest(ctx, messages, agent.Tools())
}

func (ep *endpoint) chatOpenAIRequest(ctx context.Context, messages []openai.ChatCompletionMessage, tools []openai.Tool) (string, string, error) {
	req := openai.ChatCompletionRequest{
		Model:    ep.model,
		Messages: messages,
	}
	if len(tools) > 0 {
		req.Tools = tools
	}
	resp, err := ep.openai.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", "", err
	}
	if len(resp.Choices) == 0 {
		return "", "", fmt.Errorf(errEmptyResponseFmt, ep.id)
	}
	msg := resp.Choices[0].Message
	if len(msg.ToolCalls) > 0 {
		toolJSON, err := agent.ToolCallToJSON(msg.ToolCalls[0])
		if err != nil {
			return "", "", err
		}
		return toolJSON, "", nil
	}

	content := strings.TrimSpace(msg.Content)
	if len(tools) > 0 {
		if toolJSON, ok := agent.XMLToolCallToJSON(content); ok {
			return toolJSON, "", nil
		}
	}
	reply, thinking := ExtractReply(content)

	// DeepSeek reasoner puts reasoning in a separate field; prefer it over
	// in-content tags when present.
	if msg.ReasoningContent != "" {
		thinking = strings.TrimSpace(msg.ReasoningContent)
	}

	return reply, thinking, nil
}
