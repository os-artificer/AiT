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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
)

func TestRunREPL_NoAI(t *testing.T) {
	err := RunREPL(context.Background(), Options{Scope: ai.ScopeLocal}, "")
	if err == nil || !strings.Contains(err.Error(), "AI not configured") {
		t.Fatalf("got %v", err)
	}
}

func TestPrompt_LocalAndRemote(t *testing.T) {
	if got := Prompt(ai.ScopeLocal, "", ModeAsk); got != "AiT AI(Ask)> " {
		t.Fatalf("local ask prompt %q", got)
	}
	if got := Prompt(ai.ScopeLocal, "", ModeAgent); got != "AiT AI(Agent)> " {
		t.Fatalf("local agent prompt %q", got)
	}
	if got := Prompt(ai.ScopeRemote, "u@h", ModeAgent); got != "AiT AI(Agent) u@h> " {
		t.Fatalf("remote prompt %q", got)
	}
}

func TestRunREPL_InitialMetaExits(t *testing.T) {
	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", t.TempDir())
	t.Setenv("AIT_ALLOW_CUSTOM_AI_URL", "1")

	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.AI.Default = config.ProviderDeepseek
	cfg.AI.Providers = map[string]config.ProviderConfig{
		config.ProviderDeepseek: {
			APIStyle: config.APIStyleOpenAI,
			BaseURL:  "http://127.0.0.1:9/v1",
			Model:    "deepseek-chat",
		},
	}
	client, err := ai.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- RunREPL(context.Background(), Options{
			Scope:  ai.ScopeLocal,
			AI:     client,
			In:     strings.NewReader(""),
			Out:    &out,
			ErrOut: io.Discard,
		}, "help")
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunREPL blocked after meta initial line")
	}
	if !strings.Contains(out.String(), "help") && !strings.Contains(out.String(), "Commands") {
		t.Fatalf("expected help output, got %q", out.String())
	}
}

func TestRunREPL_InitialMetaModelList(t *testing.T) {
	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", t.TempDir())
	t.Setenv("AIT_ALLOW_CUSTOM_AI_URL", "1")

	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.AI.Default = config.ProviderDeepseek
	cfg.AI.Providers = map[string]config.ProviderConfig{
		config.ProviderDeepseek: {
			APIStyle: config.APIStyleOpenAI,
			BaseURL:  "http://127.0.0.1:9/v1",
			Model:    "deepseek-chat",
		},
	}
	client, err := ai.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err = RunREPL(context.Background(), Options{
		Scope:  ai.ScopeLocal,
		AI:     client,
		In:     strings.NewReader(""),
		Out:    &out,
		ErrOut: io.Discard,
	}, "model list")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "PROVIDER") && !strings.Contains(out.String(), "deepseek") {
		t.Fatalf("expected model list output, got %q", out.String())
	}
}

func TestRunREPL_InitialChatOnceExits(t *testing.T) {
	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", t.TempDir())
	t.Setenv("AIT_ALLOW_CUSTOM_AI_URL", "1")

	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(map[string]any{
		"choices": []map[string]any{{"message": map[string]string{"content": "chat-ok"}}},
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
	done := make(chan error, 1)
	go func() {
		done <- RunREPL(context.Background(), Options{
			Scope:  ai.ScopeLocal,
			AI:     client,
			In:     strings.NewReader(""),
			Out:    &out,
			ErrOut: io.Discard,
		}, "chat hi")
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunREPL blocked after chat initial line")
	}
	if !strings.Contains(out.String(), "chat-ok") {
		t.Fatalf("expected chat reply, got %q", out.String())
	}
}

func TestRunREPL_InitialChatEmptyMessage(t *testing.T) {
	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", t.TempDir())
	t.Setenv("AIT_ALLOW_CUSTOM_AI_URL", "1")

	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.AI.Default = config.ProviderDeepseek
	cfg.AI.Providers = map[string]config.ProviderConfig{
		config.ProviderDeepseek: {
			APIStyle: config.APIStyleOpenAI,
			BaseURL:  "http://127.0.0.1:9/v1",
			Model:    "deepseek-chat",
		},
	}
	client, err := ai.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- RunREPL(context.Background(), Options{
			Scope:  ai.ScopeLocal,
			AI:     client,
			In:     strings.NewReader(""),
			Out:    io.Discard,
			ErrOut: io.Discard,
		}, "chat")
	}()

	select {
	case err := <-done:
		if !errors.Is(err, ErrChatMessageRequired) {
			t.Fatalf("got %v, want ErrChatMessageRequired", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunREPL blocked after empty chat initial line")
	}
}

func TestRunREPL_InitialNonChatShowsPrompt(t *testing.T) {
	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", t.TempDir())
	t.Setenv("AIT_ALLOW_CUSTOM_AI_URL", "1")

	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(map[string]any{
		"choices": []map[string]any{{"message": map[string]string{"content": "repl-ok"}}},
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
	err = RunREPL(context.Background(), Options{
		Scope:  ai.ScopeLocal,
		AI:     client,
		In:     strings.NewReader(""),
		Out:    &out,
		ErrOut: io.Discard,
	}, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "repl-ok") {
		t.Fatalf("expected first-turn reply, got %q", out.String())
	}
	if !strings.Contains(out.String(), "AiT AI(Ask)>") {
		t.Fatalf("expected REPL prompt after non-chat initial, got %q", out.String())
	}
}

func TestRunREPL_InitialChatOnceRemoteCRLF(t *testing.T) {
	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", t.TempDir())
	t.Setenv("AIT_ALLOW_CUSTOM_AI_URL", "1")

	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(map[string]any{
		"choices": []map[string]any{{"message": map[string]string{"content": "remote-chat"}}},
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
	done := make(chan error, 1)
	go func() {
		done <- RunREPL(context.Background(), Options{
			Scope:     ai.ScopeRemote,
			HostLabel: "u@h",
			AI:        client,
			In:        strings.NewReader(""),
			Out:       &out,
			ErrOut:    io.Discard,
		}, "chat hi")
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunREPL blocked after remote chat initial line")
	}
	got := out.String()
	if !strings.Contains(got, "remote-chat") {
		t.Fatalf("expected chat reply, got %q", got)
	}
	if !strings.Contains(got, "\r\n") {
		t.Fatalf("expected CRLF remote output, got %q", got)
	}
	if strings.Contains(got, "AiT AI(Ask)>") {
		t.Fatalf("chat-once should not print local REPL prompt, got %q", got)
	}
}
