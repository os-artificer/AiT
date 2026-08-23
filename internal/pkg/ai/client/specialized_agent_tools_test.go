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
	"encoding/json"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ai/agent"
)

// goldenAgentOpenAITools preserves pre-refactor agentOpenAITools JSON for byte-level SSOT checks.
var goldenAgentOpenAITools = []openAITool{
	{
		Type: "function",
		Function: openAIToolFunction{
			Name:        agent.ToolRunShell,
			Description: "Run one shell command on the active host",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"command":{"type":"string"}},"required":["command"]}`),
		},
	},
	{
		Type: "function",
		Function: openAIToolFunction{
			Name:        agent.ToolAISshCLI,
			Description: "Run aissh CLI subcommand on the local machine",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"args":{"type":"string"}},"required":["args"]}`),
		},
	},
	{
		Type: "function",
		Function: openAIToolFunction{
			Name:        agent.ToolAitCLI,
			Description: "Run aissh CLI subcommand on the local machine (legacy alias)",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"args":{"type":"string"}},"required":["args"]}`),
		},
	},
	{
		Type: "function",
		Function: openAIToolFunction{
			Name:        agent.ToolAnswer,
			Description: "Reply to the user and finish the tool chain",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"]}`),
		},
	},
}

func TestToolDefinitions_OpenAIJSONMatchesGolden(t *testing.T) {
	got := openAIToolsFromDefs()
	if len(got) != len(goldenAgentOpenAITools) {
		t.Fatalf("len=%d want %d", len(got), len(goldenAgentOpenAITools))
	}
	for i := range goldenAgentOpenAITools {
		gotRaw, err := json.Marshal(got[i])
		if err != nil {
			t.Fatal(err)
		}
		wantRaw, err := json.Marshal(goldenAgentOpenAITools[i])
		if err != nil {
			t.Fatal(err)
		}
		if string(gotRaw) != string(wantRaw) {
			t.Fatalf("tool[%d] got %s want %s", i, gotRaw, wantRaw)
		}
	}
}
