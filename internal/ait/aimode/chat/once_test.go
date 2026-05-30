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

package chat

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

// R18d: ParseCommand recognises "ask <msg>" and rejects bare "ask".
func TestParseCommand(t *testing.T) {
	msg, ok := ParseCommand("ask hello world")
	if !ok || msg != "hello world" {
		t.Fatalf("got %q ok=%v", msg, ok)
	}

	if _, ok := ParseCommand("ask"); ok {
		t.Fatal("expected ask alone to fail parse")
	}

	if _, ok := ParseCommand("help"); ok {
		t.Fatal("expected non-ask to fail")
	}

	if !IsEmptyCommand("ask") {
		t.Fatal("expected ask without message")
	}
	if IsEmptyCommand("ask hi") {
		t.Fatal("expected ask with message not empty-cmd")
	}
}

// ParseCommand trims trailing ";" and surrounding whitespace.
func TestParseCommand_TrailingSemicolon(t *testing.T) {
	msg, ok := ParseCommand("  ask hello ;  ")
	if !ok || msg != "hello" {
		t.Fatalf("msg=%q ok=%v", msg, ok)
	}
}

// R18a/c: RunOnce sends one user message and writes the reply to Out.
func TestRunOnce_Mock(t *testing.T) {
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
	if err := RunOnce(context.Background(), Input{
		AI:      client,
		Message: "hello",
		Scope:   ai.ScopeLocal,
		Out:     &out,
	}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "once-reply") {
		t.Fatalf("got %q", out.String())
	}
}

// R18b: empty message returns ai.ErrChatMessageRequired without calling the model.
func TestRunOnce_EmptyMessage(t *testing.T) {
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
	err = RunOnce(context.Background(), Input{AI: client, Out: &bytes.Buffer{}, Message: "  "})
	if !errors.Is(err, ai.ErrChatMessageRequired) {
		t.Fatalf("got %v", err)
	}
}

// R18a: nil AI client returns ai.ErrNotConfigured before any I/O.
func TestRunOnce_NilAI(t *testing.T) {
	err := RunOnce(context.Background(), Input{Message: "hi"})
	if !errors.Is(err, ai.ErrNotConfigured) {
		t.Fatalf("got %v", err)
	}
}

// R18f: remote scope writes reply with CRLF line terminator.
func TestRunOnce_RemoteCRLF(t *testing.T) {
	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", t.TempDir())
	t.Setenv("AIT_ALLOW_CUSTOM_AI_URL", "1")
	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{
		"choices": []map[string]any{{"message": map[string]string{"content": "line1\nline2"}}},
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
	var out bytes.Buffer
	if err := RunOnce(context.Background(), Input{
		AI:      client,
		Message: "hi",
		Scope:   ai.ScopeRemote,
		Out:     &out,
	}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(out.String(), "\r\n") || !strings.Contains(out.String(), "line1\r\nline2\r\n") {
		t.Fatalf("expected CRLF output, got %q", out.String())
	}
}

func TestParseAgentCommand(t *testing.T) {
	msg, ok := ParseAgentCommand("agent list files")
	if !ok || msg != "list files" {
		t.Fatalf("got %q ok=%v", msg, ok)
	}

	if _, ok := ParseAgentCommand("agent"); ok {
		t.Fatal("expected agent alone to fail parse")
	}

	if _, ok := ParseAgentCommand("ask hello"); ok {
		t.Fatal("expected ask verb to fail agent parse")
	}
}

func TestIsEmptyAgentCommand(t *testing.T) {
	if !IsEmptyAgentCommand("agent") {
		t.Fatal("expected agent without message to be empty")
	}
	if IsEmptyAgentCommand("agent do stuff") {
		t.Fatal("expected agent with message not empty-cmd")
	}
}
