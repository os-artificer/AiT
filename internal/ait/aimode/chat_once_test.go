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

package aimode

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
)

func TestParseChatCommand(t *testing.T) {
	msg, ok := parseChatCommand("chat hello world")
	if !ok || msg != "hello world" {
		t.Fatalf("got %q ok=%v", msg, ok)
	}

	_, ok = parseChatCommand("chat")
	if ok {
		t.Fatal("expected chat alone to fail parse")
	}

	_, ok = parseChatCommand("help")
	if ok {
		t.Fatal("expected non-chat to fail")
	}

	if !isChatCommandWithoutMessage("chat") {
		t.Fatal("expected chat without message")
	}
	if isChatCommandWithoutMessage("chat hi") {
		t.Fatal("expected chat with message not empty-cmd")
	}
}

func TestRunChatOnce_Mock(t *testing.T) {
	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", t.TempDir())
	t.Setenv("AIT_ALLOW_CUSTOM_AI_URL", "1")

	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(map[string]any{
		"choices": []map[string]any{{"message": map[string]string{"content": "once-reply"}}},
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	cfg := &config.Config{}
	cfg.AI.Default = config.ProviderDeepseek
	cfg.AI.Providers = map[string]config.ProviderConfig{
		config.ProviderDeepseek: {
			APIStyle: config.APIStyleOpenAI,
			BaseURL:  srv.URL + "/v1",
			Model:    "deepseek-chat",
		},
	}
	client, err := ai.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := RunChatOnce(context.Background(), Options{
		Scope: ai.ScopeLocal,
		AI:    client,
		Out:   &out,
	}, "hello"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "once-reply") {
		t.Fatalf("got %q", out.String())
	}
}

func TestRunChatOnce_EmptyMessage(t *testing.T) {
	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", t.TempDir())
	t.Setenv("AIT_ALLOW_CUSTOM_AI_URL", "1")
	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{
		"choices": []map[string]any{{"message": map[string]string{"content": "x"}}},
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()
	cfg := &config.Config{}
	cfg.AI.Providers = map[string]config.ProviderConfig{
		config.ProviderDeepseek: {APIStyle: config.APIStyleOpenAI, BaseURL: srv.URL + "/v1", Model: "m"},
	}
	client, err := ai.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	err = RunChatOnce(context.Background(), Options{AI: client, Out: &bytes.Buffer{}}, "  ")
	if !errors.Is(err, ai.ErrChatMessageRequired) {
		t.Fatalf("got %v", err)
	}
}
