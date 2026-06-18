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

	"github.com/os-artificer/AiT/internal/pkg/ai/eino"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
)

func TestSystemPrompt_Scopes(t *testing.T) {
	meta := PromptMeta{Provider: ProviderMinimax, Model: "MiniMax-M2.5", Mode: aiModeCloudOnly}
	local := eino.SystemPrompt(ScopeLocal, AssistEnv{}, meta)
	if !strings.Contains(local, "local AI mode") || !strings.Contains(local, "ai_provider: minimax") {
		t.Fatalf("local: %q", local)
	}
	remote := eino.SystemPrompt(ScopeRemote, AssistEnv{}, meta)
	if !strings.Contains(remote, "remote AI mode") {
		t.Fatalf("remote: %q", remote)
	}
	agent := eino.SystemPrompt(ScopeLocal, AssistEnv{AgentMode: true}, meta)
	if !strings.Contains(agent, "run_shell") {
		t.Fatalf("agent prompt: %q", agent)
	}
}

func TestBuildChatMessages_History(t *testing.T) {
	hist := []ChatMessage{{Role: "user", Content: "a"}, {Role: "assistant", Content: "b"}}
	msgs := buildChatMessages(ScopeLocal, hist, "c", AssistEnv{}, PromptMeta{Provider: ProviderDeepseek, Model: "m", Mode: aiModeCloudOnly})
	if len(msgs) != 4 {
		t.Fatalf("len=%d", len(msgs))
	}
	if msgs[3].Content != "c" {
		t.Fatalf("user=%q", msgs[3].Content)
	}
}

func TestBuildChatMessages_FirstTurnPrependsEnv(t *testing.T) {
	msgs := buildChatMessages(ScopeRemote, nil, "question", AssistEnv{
		Scope: ScopeRemote, HostLabel: "u@h", RemoteCwd: "/tmp",
	}, PromptMeta{Provider: ProviderDeepseek, Model: "m", Mode: aiModeCloudOnly})
	if len(msgs) != 2 {
		t.Fatalf("len=%d", len(msgs))
	}
	user := msgs[1].Content
	for _, want := range []string{"remote_host: u@h", "remote_cwd: /tmp", "question"} {
		if !strings.Contains(user, want) {
			t.Fatalf("missing %q in %q", want, user)
		}
	}
}

func TestClient_StatusLine(t *testing.T) {
	testAIEnv(t)
	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}
	cfg := testCfgDefaultDeepseek()
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	line := client.StatusLine()
	for _, want := range []string{"AI provider: deepseek", "model: deepseek-chat", "mode: cloud_only", "config default: deepseek"} {
		if !strings.Contains(line, want) {
			t.Fatalf("line=%q missing %q", line, want)
		}
	}
}
