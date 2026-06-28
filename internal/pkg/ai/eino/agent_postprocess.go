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

package eino

import (
	"fmt"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/ai/agent"

	"github.com/cloudwego/eino/schema"
)

// ProcessAgentMessage applies AiT Agent-mode reply extraction to an Eino assistant message.
// Order matches internal/pkg/ai/client/specialized_agent_completion.go (OpenAI legacy agent path)
// and internal/pkg/ai/eino/agent_completion.go (Eino GenerateAgent path):
// tool_calls → ToolCallToJSON; else content XML → XMLToolCallToJSON; else ExtractReply.
// ReasoningContent is returned as thinking and must not be stored in chat history (E9).
func ProcessAgentMessage(msg *schema.Message) (reply, thinking string, err error) {
	if msg == nil {
		return "", "", fmt.Errorf("nil assistant message")
	}

	if len(msg.ToolCalls) > 0 {
		tc := msg.ToolCalls[0]
		jsonReply, convErr := agent.ToolCallToJSON(agent.FCCall{
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
		if convErr != nil {
			return "", "", convErr
		}
		return jsonReply, strings.TrimSpace(msg.ReasoningContent), nil
	}

	content := strings.TrimSpace(msg.Content)
	if toolJSON, ok := agent.XMLToolCallToJSON(content); ok {
		return toolJSON, strings.TrimSpace(msg.ReasoningContent), nil
	}

	reply, thinking = extractAgentReplyContent(content)
	if rc := strings.TrimSpace(msg.ReasoningContent); rc != "" {
		thinking = rc
	}
	return reply, thinking, nil
}

// ProcessAnthropicAgentText applies Anthropic-style agent text post-processing (M6).
// Mirrors internal/pkg/ai/client/specialized_agent_completion.go forAgent on assembled text/thinking blocks.
func ProcessAnthropicAgentText(content, blockThinking string) (reply, thinking string) {
	content = strings.TrimSpace(content)
	if toolJSON, ok := agent.XMLToolCallToJSON(content); ok {
		return toolJSON, strings.TrimSpace(blockThinking)
	}

	reply, extracted := extractAgentReplyContent(content)
	thinking = strings.TrimSpace(blockThinking)
	if extracted != "" {
		if thinking != "" {
			thinking += "\n"
		}
		thinking += extracted
	}
	return reply, thinking
}

// extractAgentReplyContent runs ExtractReply and preserves raw content when stripping would leave an empty reply.
// ExtractReply removes trailing partial agent JSON; ResolveAgentReply heuristics need the raw bytes (C8).
func extractAgentReplyContent(content string) (reply, thinking string) {
	reply, thinking = ExtractReply(content)
	if strings.TrimSpace(reply) == "" && content != "" {
		reply = content
	}
	return reply, thinking
}
