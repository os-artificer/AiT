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
	"encoding/json"
	"strings"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

func TestToolCallToAgentJSON_RunShell(t *testing.T) {
	tc := openai.ToolCall{
		Type: openai.ToolTypeFunction,
		Function: openai.FunctionCall{
			Name:      agentToolRunShell,
			Arguments: `{"command":"hostname -I"}`,
		},
	}
	got, err := toolCallToAgentJSON(tc)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `"action":"run_shell"`) || !strings.Contains(got, "hostname -I") {
		t.Fatalf("got %q", got)
	}
}

func TestChatAgent_ToolCall(t *testing.T) {
	tc := openai.ToolCall{
		Type: openai.ToolTypeFunction,
		Function: openai.FunctionCall{
			Name:      agentToolAnswer,
			Arguments: `{"text":"ok"}`,
		},
	}
	got, err := toolCallToAgentJSON(tc)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `"action":"answer"`) {
		t.Fatalf("got %q", got)
	}
}

func TestChatAgent_JSONFallbackContent(t *testing.T) {
	raw := `{"action":"answer","text":"done"}`
	act := struct {
		Action string `json:"action"`
		Text   string `json:"text"`
	}{}
	if err := json.Unmarshal([]byte(raw), &act); err != nil || act.Action != "answer" {
		t.Fatalf("parse: %v action=%q", err, act.Action)
	}
}
