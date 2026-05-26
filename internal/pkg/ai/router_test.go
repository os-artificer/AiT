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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ai/router"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("AIT_KEYRING_BACKEND", "file")
	os.Exit(m.Run())
}

func testAIEnv(t *testing.T) {
	t.Helper()
	t.Setenv("AIT_ALLOW_CUSTOM_AI_URL", "1")
	t.Setenv("AIT_KEYRING_DIR", t.TempDir())
}

func openAISuccessBody(reply string) string {
	b, _ := json.Marshal(map[string]any{
		"choices": []map[string]any{{"message": map[string]string{"content": reply}}},
	})
	return string(b)
}

type mockOpenAIServer struct {
	srv   *httptest.Server
	calls atomic.Int32
}

func newMockOpenAIServer(t *testing.T, status int, reply string) *mockOpenAIServer {
	t.Helper()
	m := &mockOpenAIServer{}
	body := `{"error":{"message":"fail"}}`
	if status >= 200 && status < 300 {
		body = openAISuccessBody(reply)
	}
	m.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(m.srv.Close)
	return m
}

func setupTwoProviderClient(t *testing.T, cfg *config.Config, deep, mini *mockOpenAIServer) *Client {
	t.Helper()
	testAIEnv(t)

	cfg.AI.Providers[config.ProviderDeepseek] = config.ProviderConfig{
		APIStyle: config.APIStyleOpenAI,
		BaseURL:  deep.srv.URL + "/v1",
		Model:    "deepseek-chat",
	}
	cfg.AI.Providers[config.ProviderMinimax] = config.ProviderConfig{
		APIStyle: config.APIStyleOpenAI,
		BaseURL:  mini.srv.URL + "/v1",
		Model:    "MiniMax-M2.5",
	}

	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "deep-key"); err != nil {
		t.Fatal(err)
	}
	if err := creds.SetProviderAPIKey(config.ProviderMinimax, "mini-key"); err != nil {
		t.Fatal(err)
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func testCfgDefaultDeepseek() *config.Config {
	cfg := &config.Config{}
	cfg.AI.Default = config.ProviderDeepseek
	cfg.AI.Providers = map[string]config.ProviderConfig{
		config.ProviderDeepseek: {APIStyle: config.APIStyleOpenAI, Model: "deepseek-chat"},
		config.ProviderMinimax:  {APIStyle: config.APIStyleOpenAI, Model: "MiniMax-M2.5"},
	}
	return cfg
}

func TestFailover_DefaultSuccessM1(t *testing.T) {
	deep := newMockOpenAIServer(t, 200, "ok-deep")
	mini := newMockOpenAIServer(t, 200, "ok-mini")
	client := setupTwoProviderClient(t, testCfgDefaultDeepseek(), deep, mini)

	reply, err := client.Assist(context.Background(), "hi", "", "")
	if err != nil || reply != "ok-deep" {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
	if deep.calls.Load() != 1 || mini.calls.Load() != 0 {
		t.Fatalf("calls deep=%d mini=%d", deep.calls.Load(), mini.calls.Load())
	}
}

func TestFailover_Default503ThenMinimaxM2(t *testing.T) {
	deep := newMockOpenAIServer(t, 503, "")
	mini := newMockOpenAIServer(t, 200, "ok-mini")
	client := setupTwoProviderClient(t, testCfgDefaultDeepseek(), deep, mini)

	reply, err := client.Assist(context.Background(), "hi", "", "")
	if err != nil || reply != "ok-mini" {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
	if deep.calls.Load() != 1 || mini.calls.Load() != 1 {
		t.Fatalf("calls deep=%d mini=%d", deep.calls.Load(), mini.calls.Load())
	}
}

func TestFailover_OnlyDefaultHasKeyM3(t *testing.T) {
	testAIEnv(t)
	deep := newMockOpenAIServer(t, 200, "only")
	mini := newMockOpenAIServer(t, 200, "never")

	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}

	cfg := testCfgDefaultDeepseek()
	cfg.AI.Providers[config.ProviderDeepseek] = config.ProviderConfig{
		APIStyle: config.APIStyleOpenAI,
		BaseURL:  deep.srv.URL + "/v1",
		Model:    "m",
	}
	cfg.AI.Providers[config.ProviderMinimax] = config.ProviderConfig{
		BaseURL: mini.srv.URL + "/v1",
		Model:   "m",
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Assist(context.Background(), "x", "", ""); err != nil {
		t.Fatal(err)
	}
	if mini.calls.Load() != 0 {
		t.Fatal("minimax should not be called without key")
	}
}

func TestFailover_DefaultMinimaxM4(t *testing.T) {
	cfg := testCfgDefaultDeepseek()
	cfg.AI.Default = config.ProviderMinimax
	deep := newMockOpenAIServer(t, 200, "deep")
	mini := newMockOpenAIServer(t, 200, "mini-ok")
	client := setupTwoProviderClient(t, cfg, deep, mini)

	reply, err := client.Assist(context.Background(), "q", "", "")
	if err != nil || reply != "mini-ok" {
		t.Fatalf("got %q err=%v", reply, err)
	}
	if mini.calls.Load() != 1 || deep.calls.Load() != 0 {
		t.Fatalf("calls deep=%d mini=%d", deep.calls.Load(), mini.calls.Load())
	}
}

func TestFailover_ExplicitFallbackOrderM7(t *testing.T) {
	cfg := testCfgDefaultDeepseek()
	cfg.AI.Fallback = []string{config.ProviderMinimax, config.ProviderDeepseek}
	deep := newMockOpenAIServer(t, 503, "")
	mini := newMockOpenAIServer(t, 503, "")
	client := setupTwoProviderClient(t, cfg, deep, mini)

	if _, err := client.Assist(context.Background(), "q", "", ""); err == nil {
		t.Fatal("expected error when all fail")
	}
	if deep.calls.Load() != 1 || mini.calls.Load() != 1 {
		t.Fatalf("calls deep=%d mini=%d", deep.calls.Load(), mini.calls.Load())
	}
}

func TestNewClient_ProviderOverrideM11(t *testing.T) {
	testAIEnv(t)
	if err := creds.SetProviderAPIKey(config.ProviderMinimax, "k"); err != nil {
		t.Fatal(err)
	}
	cfg := testCfgDefaultDeepseek()
	client, err := NewClient(cfg, WithProviderOverride(config.ProviderMinimax))
	if err != nil {
		t.Fatal(err)
	}
	if client.sessionProvider != config.ProviderMinimax {
		t.Fatalf("sessionProvider=%q", client.sessionProvider)
	}
}

func TestClient_SetSessionProviderM12(t *testing.T) {
	deep := newMockOpenAIServer(t, 200, "a")
	mini := newMockOpenAIServer(t, 200, "b")
	client := setupTwoProviderClient(t, testCfgDefaultDeepseek(), deep, mini)
	if err := client.SetSessionProvider(config.ProviderMinimax); err != nil {
		t.Fatal(err)
	}
	if client.SessionProvider() != config.ProviderMinimax {
		t.Fatal("expected minimax session override")
	}
}

func TestResolveTryOrder_AutoFallbackM8(t *testing.T) {
	cfg := testCfgDefaultDeepseek()
	keyed := map[ProviderID]struct{}{
		config.ProviderDeepseek: {},
		config.ProviderMinimax:  {},
	}
	order := router.ResolveTryOrder(cfg, "", keyed)
	if len(order) != 2 || order[0] != config.ProviderDeepseek || order[1] != config.ProviderMinimax {
		t.Fatalf("order=%v", order)
	}
}

func TestFailover_AnthropicStyleM10(t *testing.T) {
	testAIEnv(t)
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		if r.Header.Get("anthropic-version") == "" {
			t.Error("missing anthropic-version header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"content":[{"type":"text","text":"anth-ok"}]}`)
	}))
	t.Cleanup(srv.Close)

	if err := creds.SetProviderAPIKey(config.ProviderMinimax, "mk"); err != nil {
		t.Fatal(err)
	}

	cfg := testCfgDefaultDeepseek()
	cfg.AI.Default = config.ProviderMinimax
	cfg.AI.Providers[config.ProviderMinimax] = config.ProviderConfig{
		APIStyle: config.APIStyleAnthropic,
		BaseURL:  srv.URL,
		Model:    "MiniMax-M2.5",
	}

	client, err := NewClient(cfg, WithProviderOverride(config.ProviderMinimax))
	if err != nil {
		t.Fatal(err)
	}
	reply, err := client.Assist(context.Background(), "hi", "", "")
	if err != nil || reply != "anth-ok" {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
	if path != "/v1/messages" {
		t.Fatalf("path=%q want /v1/messages", path)
	}
}

func TestChat_OpenAI(t *testing.T) {
	deep := newMockOpenAIServer(t, 200, "chat-reply")
	mini := newMockOpenAIServer(t, 200, "x")
	client := setupTwoProviderClient(t, testCfgDefaultDeepseek(), deep, mini)
	reply, hist, err := client.Chat(context.Background(), nil, "hello", AssistEnv{Scope: ScopeLocal})
	if err != nil || reply != "chat-reply" || len(hist) != 2 {
		t.Fatalf("reply=%q hist=%d err=%v", reply, len(hist), err)
	}
}

func TestNewClient_NoKeys(t *testing.T) {
	testAIEnv(t)
	_, err := NewClient(testCfgDefaultDeepseek())
	if err == nil {
		t.Fatal("expected error without keys")
	}
}
