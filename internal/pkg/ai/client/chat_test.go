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
)

func TestSystemPrompt_Scopes(t *testing.T) {
	meta := PromptMeta{Provider: ProviderMinimax, Model: "MiniMax-M2.5", Mode: aiModeCloudOnly}
	local := systemPrompt(ScopeLocal, AssistEnv{}, meta)
	if !strings.Contains(local, "local AI mode") || !strings.Contains(local, "ai_provider: minimax") {
		t.Fatalf("local: %q", local)
	}
	remote := systemPrompt(ScopeRemote, AssistEnv{}, meta)
	if !strings.Contains(remote, "remote AI mode") {
		t.Fatalf("remote: %q", remote)
	}
	agent := systemPrompt(ScopeLocal, AssistEnv{AgentMode: true}, meta)
	if !strings.Contains(agent, "run_shell") {
		t.Fatalf("agent prompt: %q", agent)
	}
}

func TestPrependEnvContext_RemoteHost(t *testing.T) {
	got := prependEnvContext("question", AssistEnv{Scope: ScopeRemote, HostLabel: "u@h", RemoteCwd: "/tmp"})
	for _, want := range []string{"remote_host: u@h", "remote_cwd: /tmp", "question"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
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
