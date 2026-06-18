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
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ai/agent"
	"github.com/os-artificer/AiT/internal/pkg/ai/eino"
)

const xmlAgentReply = `<minimax:tool_call><invoke name="run"><parameter name="command">whoami</parameter></invoke></minimax:tool_call>`

func legacyProcessOpenAIAgentMessage(msg openAIChatMessage) (reply, thinking string, err error) {
	if len(msg.ToolCalls) > 0 {
		tc := msg.ToolCalls[0]
		toolJSON, convErr := agent.ToolCallToJSON(agent.FCCall{Name: tc.Function.Name, Arguments: tc.Function.Arguments})
		if convErr != nil {
			return "", "", convErr
		}
		return toolJSON, "", nil
	}

	content := strings.TrimSpace(msg.Content)
	if toolJSON, ok := agent.XMLToolCallToJSON(content); ok {
		return toolJSON, "", nil
	}
	reply, thinking = ExtractReply(content)
	return reply, thinking, nil
}

func legacyProcessAnthropicAgentContent(content, blockThinking string) (reply, thinking string) {
	content = strings.TrimSpace(content)
	if toolJSON, ok := agent.XMLToolCallToJSON(content); ok {
		return toolJSON, strings.TrimSpace(blockThinking)
	}
	reply, extracted := ExtractReply(content)
	thinking = strings.TrimSpace(blockThinking)
	if extracted != "" {
		if thinking != "" {
			thinking += "\n"
		}
		thinking += extracted
	}
	return reply, thinking
}

func TestSpecializedAgentOpenAI_DelegatesProcessAgentMessage(t *testing.T) {
	cases := []openAIChatMessage{
		{
			ToolCalls: []openAIToolCall{{
				Type: "function",
				Function: openAIFunctionCall{
					Name:      agent.ToolRunShell,
					Arguments: `{"command":"hostname -I"}`,
				},
			}},
		},
		{Content: xmlAgentReply},
		{Content: `{"action":"answer","text":"done"}`},
	}
	for i, msg := range cases {
		legacyReply, legacyThinking, legacyErr := legacyProcessOpenAIAgentMessage(msg)
		gotReply, gotThinking, gotErr := eino.ProcessAgentMessage(openAIAgentMessageToSchema(msg))
		if legacyErr != nil || gotErr != nil {
			if (legacyErr == nil) != (gotErr == nil) {
				t.Fatalf("case %d err legacy=%v eino=%v", i, legacyErr, gotErr)
			}
			continue
		}
		if legacyReply != gotReply || legacyThinking != gotThinking {
			t.Fatalf("case %d legacy reply=%q thinking=%q eino reply=%q thinking=%q", i, legacyReply, legacyThinking, gotReply, gotThinking)
		}
	}
}

func TestSpecializedAnthropicAgent_DelegatesProcessAnthropic(t *testing.T) {
	legacyReply, legacyThinking := legacyProcessAnthropicAgentContent(xmlAgentReply, "block-think")
	gotReply, gotThinking := eino.ProcessAnthropicAgentText(xmlAgentReply, "block-think")
	if legacyReply != gotReply || legacyThinking != gotThinking {
		t.Fatalf("legacy reply=%q thinking=%q eino reply=%q thinking=%q", legacyReply, legacyThinking, gotReply, gotThinking)
	}
	if !strings.Contains(gotReply, `"action":"run_shell"`) {
		t.Fatalf("reply=%q", gotReply)
	}
}

func TestSpecializedAnthropicChat_SkipsXMLToolParse(t *testing.T) {
	content := strings.TrimSpace(xmlAgentReply)
	reply, _ := ExtractReply(content)
	if strings.Contains(reply, `"action":"run_shell"`) {
		t.Fatalf("chat path must not emit agent JSON: %q", reply)
	}

	agentReply, _ := eino.ProcessAnthropicAgentText(xmlAgentReply, "")
	if !strings.Contains(agentReply, `"action":"run_shell"`) {
		t.Fatalf("agent postprocess should convert XML: %q", agentReply)
	}
}
