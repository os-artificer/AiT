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
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
)

func setupTieredClient(t *testing.T, ollamaStatus, cloudStatus int, ollamaReply, cloudReply string) (*Client, *mockOpenAIServer, *mockOpenAIServer) {
	t.Helper()
	testAIEnv(t)

	ollama := newMockOpenAIServer(t, ollamaStatus, ollamaReply)
	cloud := newMockOpenAIServer(t, cloudStatus, cloudReply)

	cfg := testCfgDefaultDeepseek()
	cfg.AI.Tiered.Enabled = true
	cfg.AI.Tiered.MaxPromptChars = config.DefaultTieredMaxPromptChars
	cfg.AI.Tiered.MaxHistoryTurns = config.DefaultTieredMaxHistoryTurns
	cfg.AI.Providers[config.ProviderOllama] = config.ProviderConfig{
		APIStyle: config.APIStyleOpenAI,
		BaseURL:  ollama.srv.URL + "/v1",
		Model:    "llama3.2",
	}
	cfg.AI.Providers[config.ProviderDeepseek] = config.ProviderConfig{
		APIStyle: config.APIStyleOpenAI,
		BaseURL:  cloud.srv.URL + "/v1",
		Model:    "deepseek-chat",
	}

	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return client, ollama, cloud
}

func tieredCfg() *config.Config {
	cfg := &config.Config{}
	cfg.AI.Tiered.Enabled = true
	cfg.AI.Tiered.MaxPromptChars = 280
	cfg.AI.Tiered.MaxHistoryTurns = 0
	cfg.AI.Tiered.ComplexKeywords = []string{"implement", "debug"}
	cfg.AI.Providers = map[string]config.ProviderConfig{
		config.ProviderOllama: {Model: "llama3.2"},
	}
	return cfg
}

func TestTiered_SimpleUsesOllamaT1(t *testing.T) {
	client, ollama, cloud := setupTieredClient(t, 200, 200, "local-ok", "cloud-ok")

	reply, _, _, err := client.Chat(context.Background(), nil, "hi", AssistEnv{Scope: ScopeLocal})
	if err != nil || reply != "local-ok" {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
	if ollama.calls.Load() != 1 {
		t.Fatalf("ollama calls=%d", ollama.calls.Load())
	}
	if cloud.calls.Load() != 0 {
		t.Fatalf("cloud calls=%d", cloud.calls.Load())
	}
}

func TestTiered_AgentLegacy(t *testing.T) {
	agentReply := `{"action":"answer","text":"local-agent"}`
	client, ollama, cloud := setupTieredClient(t, 200, 200, agentReply, "cloud-agent")

	reply, _, _, err := client.ChatAgent(context.Background(), nil, "hi", AssistEnv{Scope: ScopeLocal})
	if err != nil || reply != agentReply {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
	if ollama.calls.Load() != 1 {
		t.Fatalf("ollama calls=%d", ollama.calls.Load())
	}
	if cloud.calls.Load() != 0 {
		t.Fatalf("cloud calls=%d", cloud.calls.Load())
	}
}

func TestTiered_AgentLegacy_RequestIncludesTools(t *testing.T) {
	testAIEnv(t)
	agentReply := `{"action":"answer","text":"local-agent"}`
	var lastBody atomic.Pointer[string]

	ollamaSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			return
		}
		s := string(body)
		lastBody.Store(&s)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = io.WriteString(w, openAISuccessBody(agentReply))
	}))
	t.Cleanup(ollamaSrv.Close)

	cloud := newMockOpenAIServer(t, 200, "cloud-agent")

	cfg := testCfgDefaultDeepseek()
	cfg.AI.Tiered.Enabled = true
	cfg.AI.Tiered.MaxPromptChars = config.DefaultTieredMaxPromptChars
	cfg.AI.Tiered.MaxHistoryTurns = config.DefaultTieredMaxHistoryTurns
	cfg.AI.Providers[config.ProviderOllama] = config.ProviderConfig{
		APIStyle: config.APIStyleOpenAI,
		BaseURL:  ollamaSrv.URL + "/v1",
		Model:    "llama3.2",
	}
	cfg.AI.Providers[config.ProviderDeepseek] = config.ProviderConfig{
		APIStyle: config.APIStyleOpenAI,
		BaseURL:  cloud.srv.URL + "/v1",
		Model:    "deepseek-chat",
	}
	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}

	reply, _, _, err := client.ChatAgent(context.Background(), nil, "hi", AssistEnv{Scope: ScopeLocal})
	if err != nil || reply != agentReply {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
	bodyPtr := lastBody.Load()
	if bodyPtr == nil || !strings.Contains(*bodyPtr, `"tools"`) {
		t.Fatalf("agent request body missing tools: %v", bodyPtr)
	}
}

func TestTiered_ComplexUsesCloudT2(t *testing.T) {
	client, ollama, cloud := setupTieredClient(t, 200, 200, "local", "cloud-ok")

	prompt := "please ```implement``` a full system"
	reply, _, _, err := client.Chat(context.Background(), nil, prompt, AssistEnv{Scope: ScopeLocal})
	if err != nil || reply != "cloud-ok" {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
	if ollama.calls.Load() != 0 {
		t.Fatalf("ollama calls=%d", ollama.calls.Load())
	}
	if cloud.calls.Load() != 1 {
		t.Fatalf("cloud calls=%d", cloud.calls.Load())
	}
}

func TestTiered_OllamaFailoverT3(t *testing.T) {
	client, ollama, cloud := setupTieredClient(t, 503, 200, "", "cloud-ok")

	reply, _, _, err := client.Chat(context.Background(), nil, "hi", AssistEnv{Scope: ScopeLocal})
	if err != nil || reply != "cloud-ok" {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
	if ollama.calls.Load() != 1 || cloud.calls.Load() != 1 {
		t.Fatalf("calls ollama=%d cloud=%d", ollama.calls.Load(), cloud.calls.Load())
	}
}

func TestTiered_SessionOverrideSkipsOllamaT4(t *testing.T) {
	client, ollama, _ := setupTieredClient(t, 200, 200, "local", "cloud-ok")
	if err := client.SetSessionProvider(config.ProviderDeepseek); err != nil {
		t.Fatal(err)
	}

	reply, _, _, err := client.Chat(context.Background(), nil, "hi", AssistEnv{Scope: ScopeLocal})
	if err != nil || reply != "cloud-ok" {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
	if ollama.calls.Load() != 0 {
		t.Fatalf("ollama calls=%d", ollama.calls.Load())
	}
}

func TestTiered_OnlyOllamaNoCloudKeyT7(t *testing.T) {
	testAIEnv(t)
	ollama := newMockOpenAIServer(t, 200, "ok")
	cfg := &config.Config{}
	cfg.AI.Tiered.Enabled = true
	cfg.AI.Tiered.MaxPromptChars = config.DefaultTieredMaxPromptChars
	cfg.AI.Providers = map[string]config.ProviderConfig{
		config.ProviderOllama: {
			BaseURL: ollama.srv.URL + "/v1",
			Model:   "llama3.2",
		},
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, err = client.Chat(context.Background(), nil, strings.Repeat("x", 400), AssistEnv{})
	if err == nil {
		t.Fatal("expected error for complex prompt without cloud key")
	}
}

func stringsRepeat(s string, n int) string {
	var b []byte
	for i := 0; i < n; i++ {
		b = append(b, s...)
	}
	return string(b)
}

func TestOllamaTagsL2(t *testing.T) {
	testAIEnv(t)
	tagSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"models":[{"name":"llama3.2"},{"name":"qwen2.5:7b"}]}`)
	}))
	t.Cleanup(tagSrv.Close)

	cfg := tieredCfg()
	cfg.AI.Providers[config.ProviderOllama] = config.ProviderConfig{
		BaseURL: tagSrv.URL + "/v1",
		Model:   "llama3.2",
	}
	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}
	cfg.AI.Providers[config.ProviderDeepseek] = config.ProviderConfig{BaseURL: "http://127.0.0.1:9/v1", Model: "d"}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	report, err := client.ListModels(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	foundStar := false
	for _, tag := range report.OllamaTags {
		if strings.Contains(tag, "llama3.2") && strings.Contains(tag, "configured") {
			foundStar = true
		}
	}
	if !foundStar {
		t.Fatalf("tags=%v", report.OllamaTags)
	}
}

func TestListModelsFormatL4(t *testing.T) {
	testAIEnv(t)
	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}
	cfg := testCfgDefaultDeepseek()
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	report, err := client.ListModels(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	formatted := report.Format()
	if !strings.Contains(formatted, "PROVIDER") || !strings.Contains(formatted, "deepseek") {
		t.Fatalf("format=%q", formatted)
	}
}
