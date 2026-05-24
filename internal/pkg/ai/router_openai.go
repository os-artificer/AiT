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

	openai "github.com/sashabaranov/go-openai"
)

func (ep *endpoint) chatOpenAI(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error) {
	return ep.chatOpenAIRequest(ctx, messages, nil)
}

func (ep *endpoint) chatOpenAIAgent(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error) {
	tools := agentTools()
	return ep.chatOpenAIRequest(ctx, messages, tools)
}

func (ep *endpoint) chatOpenAIRequest(ctx context.Context, messages []openai.ChatCompletionMessage, tools []openai.Tool) (string, error) {
	req := openai.ChatCompletionRequest{
		Model:    ep.model,
		Messages: messages,
	}
	if len(tools) > 0 {
		req.Tools = tools
	}
	resp, err := ep.openai.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf(errEmptyResponseFmt, ep.id)
	}
	msg := resp.Choices[0].Message
	if len(msg.ToolCalls) > 0 {
		return toolCallToAgentJSON(msg.ToolCalls[0])
	}
	return strings.TrimSpace(msg.Content), nil
}
