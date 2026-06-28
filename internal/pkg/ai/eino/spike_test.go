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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ai/agent"

	einoopenai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

func TestSpike_EinoOpenAI_ToolCallsFC(t *testing.T) {
	body := map[string]any{
		"choices": []map[string]any{{
			"message": map[string]any{
				"role": "assistant",
				"tool_calls": []map[string]any{{
					"id":   "call_1",
					"type": "function",
					"function": map[string]string{
						"name":      agent.ToolRunShell,
						"arguments": `{"command":"hostname -I"}`,
					},
				}},
			},
		}},
	}
	msg := runSpikeAgentGenerate(t, body, true)

	reply, thinking, err := ProcessAgentMessage(msg)
	if err != nil {
		t.Fatal(err)
	}
	if thinking != "" {
		t.Fatalf("thinking=%q", thinking)
	}
	if !strings.Contains(reply, `"action":"run_shell"`) || !strings.Contains(reply, "hostname -I") {
		t.Fatalf("reply=%q", reply)
	}
}

func TestSpike_EinoOpenAI_XMLContentM4(t *testing.T) {
	xmlReply := `<minimax:tool_call><invoke name="run"><parameter name="command">whoami</parameter></invoke></minimax:tool_call>`
	body := openAIContentBody(xmlReply)
	msg := runSpikeAgentGenerate(t, body, true)

	reply, _, err := ProcessAgentMessage(msg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(reply, `"action":"run_shell"`) || !strings.Contains(reply, "whoami") {
		t.Fatalf("reply=%q", reply)
	}
}

func TestSpike_EinoOpenAI_ReasoningContentE9(t *testing.T) {
	body := map[string]any{
		"choices": []map[string]any{{
			"message": map[string]any{
				"role":              "assistant",
				"content":           "The answer is 42",
				"reasoning_content": "Let me reason",
			},
		}},
	}
	msg := runSpikeAgentGenerate(t, body, false)

	reply, thinking, err := ProcessAgentMessage(msg)
	if err != nil {
		t.Fatal(err)
	}
	if reply != "The answer is 42" || thinking != "Let me reason" {
		t.Fatalf("reply=%q thinking=%q", reply, thinking)
	}
}

func TestSpike_AnthropicText_XMLM6(t *testing.T) {
	xmlReply := `<minimax:tool_call><invoke name="run"><parameter name="command">whoami</parameter></invoke></minimax:tool_call>`
	reply, thinking := ProcessAnthropicAgentText(xmlReply, "")
	if thinking != "" {
		t.Fatalf("thinking=%q", thinking)
	}
	if !strings.Contains(reply, `"action":"run_shell"`) || !strings.Contains(reply, "whoami") {
		t.Fatalf("reply=%q", reply)
	}
}

func TestSpike_OllamaCompatible_ShortPromptStrategyA(t *testing.T) {
	// Strategy A: tiered Agent on Ollama stays legacy endpoint in P3; this spike only
	// verifies eino-ext openai can bind tools on a short prompt (OpenAI-compatible Ollama path).
	body := map[string]any{
		"choices": []map[string]any{{
			"message": map[string]any{
				"role":    "assistant",
				"content": `{"action":"answer","text":"pong"}`,
			},
		}},
	}
	msg := runSpikeAgentGenerate(t, body, true)

	reply, _, err := ProcessAgentMessage(msg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(reply, "pong") {
		t.Fatalf("reply=%q", reply)
	}
}

func runSpikeAgentGenerate(t *testing.T, responseBody map[string]any, withTools bool) *schema.Message {
	t.Helper()
	raw, err := json.Marshal(responseBody)
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(raw)
	}))
	t.Cleanup(srv.Close)

	ctx := context.Background()
	cm, err := einoopenai.NewChatModel(ctx, &einoopenai.ChatModelConfig{
		APIKey:  "spike-key",
		BaseURL: srv.URL + "/v1",
		Model:   "spike-model",
	})
	if err != nil {
		t.Fatal(err)
	}

	var chatModel model.BaseChatModel = cm

	if withTools {
		tools, toolErr := AgentToolInfos()
		if toolErr != nil {
			t.Fatal(toolErr)
		}
		withToolsModel, bindErr := cm.WithTools(tools)
		if bindErr != nil {
			t.Fatal(bindErr)
		}
		chatModel = withToolsModel
	}

	msgs := []*schema.Message{
		schema.SystemMessage("You are AiT agent spike test."),
		schema.UserMessage("hi"),
	}

	out, err := chatModel.Generate(ctx, msgs)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func openAIContentBody(content string) map[string]any {
	return map[string]any{
		"choices": []map[string]any{{
			"message": map[string]any{
				"role":    "assistant",
				"content": content,
			},
		}},
	}
}
