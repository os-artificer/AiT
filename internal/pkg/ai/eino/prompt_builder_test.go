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
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ai/apitypes"
	"github.com/os-artificer/AiT/internal/pkg/config"

	"github.com/cloudwego/eino/schema"
)

func TestPromptBuilder_InjectsMeta(t *testing.T) {
	meta := apitypes.PromptMeta{Provider: config.ProviderDeepseek, Model: "deepseek-chat", Mode: "cloud_only"}
	sys := SystemPrompt(apitypes.ScopeLocal, apitypes.AssistEnv{}, meta)
	for _, want := range []string{
		promptKeyProvider + " " + config.ProviderDeepseek,
		promptKeyModel + " deepseek-chat",
		promptKeyMode + " cloud_only",
	} {
		if !strings.Contains(sys, want) {
			t.Fatalf("system prompt missing %q:\n%s", want, sys)
		}
	}
}

func TestPrependEnvContext_RemoteHost(t *testing.T) {
	got := PrependEnvContext("question", apitypes.AssistEnv{
		Scope: apitypes.ScopeRemote, HostLabel: "u@h", RemoteCwd: "/tmp",
	})
	for _, want := range []string{"remote_host: u@h", "remote_cwd: /tmp", "question"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}
}

func TestBuildChatMessages_FirstTurnOnlyPrependsEnv(t *testing.T) {
	msgs := BuildChatMessages(apitypes.ScopeLocal, nil, "q", apitypes.AssistEnv{Scrollback: "scroll"}, apitypes.PromptMeta{})
	if len(msgs) != 2 {
		t.Fatalf("len=%d", len(msgs))
	}
	if !strings.Contains(msgs[1].Content, "recent_terminal_output:") {
		t.Fatalf("user=%q", msgs[1].Content)
	}

	withHist := BuildChatMessages(apitypes.ScopeLocal, []apitypes.ChatMessage{{Role: "user", Content: "a"}}, "q", apitypes.AssistEnv{Scrollback: "scroll"}, apitypes.PromptMeta{})
	if withHist[len(withHist)-1].Content != "q" {
		t.Fatalf("history turn user=%q", withHist[len(withHist)-1].Content)
	}
}

func TestGenerateChat_Mock(t *testing.T) {
	m := &MockChatModel{Reply: "hello"}
	reply, thinking, err := GenerateChat(context.Background(), m, []*schema.Message{schema.UserMessage("ping")})
	if err != nil || reply != "hello" || thinking != "" {
		t.Fatalf("reply=%q thinking=%q err=%v", reply, thinking, err)
	}
}

func TestPlainFromSchema_Roles(t *testing.T) {
	plain := PlainFromSchema([]*schema.Message{
		schema.SystemMessage("sys"),
		schema.UserMessage("u"),
		schema.AssistantMessage("a", nil),
	})
	if len(plain) != 3 || plain[0].Role != RoleSystem || plain[2].Role != RoleAssistant {
		t.Fatalf("plain=%+v", plain)
	}
}
