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

package mode

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/aissh/shell"
	"github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
	"github.com/os-artificer/AiT/internal/pkg/ssh/client/pty"
	"github.com/os-artificer/AiT/internal/pkg/terminal"

	"golang.org/x/term"
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
		"choices": []map[string]any{{"message": map[string]string{"content": "ask-ok"}}},
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
		}, "ask hi")
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunREPL blocked after ask initial line")
	}
	if !strings.Contains(out.String(), "ask-ok") {
		t.Fatalf("expected ask reply, got %q", out.String())
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
		}, "ask")
	}()

	select {
	case err := <-done:
		if !errors.Is(err, ai.ErrChatMessageRequired) {
			t.Fatalf("got %v, want ai.ErrChatMessageRequired", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunREPL blocked after empty ask initial line")
	}
}

// RA1/RA2: bare meta "agent" enters Agent REPL (ScopeRemote) without ChatMessageRequired.
func TestRunREPL_InitialEmptyAgentEntersREPL(t *testing.T) {
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
		Scope:     ai.ScopeRemote,
		HostLabel: "u@h",
		AI:        client,
		In:        strings.NewReader("exit\n"),
		Out:       &out,
		ErrOut:    io.Discard,
	}, "agent")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "ask requires a message") {
		t.Fatal("bare agent must not trigger ask-required hint")
	}
	if !strings.Contains(out.String(), "AiT AI(Agent) u@h>") {
		t.Fatalf("expected remote Agent REPL prompt, got %q", out.String())
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
	if !strings.Contains(out.String(), "AiT AI(Agent)>") {
		t.Fatalf("expected REPL prompt after non-ask initial, got %q", out.String())
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
		}, "ask hi")
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunREPL blocked after remote ask initial line")
	}
	got := out.String()
	if !strings.Contains(got, "remote-chat") {
		t.Fatalf("expected ask reply, got %q", got)
	}
	if !strings.Contains(got, "\r\n") {
		t.Fatalf("expected CRLF remote output, got %q", got)
	}
	if strings.Contains(got, "AiT AI(Ask)>") {
		t.Fatalf("ask-once should not print local REPL prompt, got %q", got)
	}
}

func TestIsMetaREPLCommand(t *testing.T) {
	meta := []string{"help", "model list", "model", "clear", "policy once", "run echo", "ai-ssh list", "-h", "--help"}
	for _, line := range meta {
		if !isMetaREPLCommand(line) {
			t.Errorf("%q should be meta", line)
		}
	}
	nonMeta := []string{"how do I grep logs", "what is docker", "ask hello world", ""}
	for _, line := range nonMeta {
		if isMetaREPLCommand(line) {
			t.Errorf("%q should not be meta", line)
		}
	}
}

func TestAIMode_ConfirmNoHistory(t *testing.T) {
	hist := terminal.NewLineHistory()
	hist.Add("run ls")

	br := bufio.NewReader(bytes.NewReader([]byte("y\r\x1b[A\r")))
	_, err := shell.ReadTTYLineNoHistory(br, io.Discard, "")
	if err != nil {
		t.Fatal(err)
	}

	line, err := shell.ReadTTYLine(br, io.Discard, "AiT AI> ", hist)
	if err != nil {
		t.Fatal(err)
	}
	if line != "run ls" {
		t.Fatalf("line = %q, want run ls (confirm y must not pollute history)", line)
	}
}

// RC2: meta initial question on TTY path confirms via withTTYConfirm before agent runs.
func TestRunREPL_InitialAgentConfirm_UseTTYLineEdit(t *testing.T) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		t.Skip("requires stdin tty")
	}

	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", t.TempDir())
	t.Setenv("AIT_ALLOW_CUSTOM_AI_URL", "1")
	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}

	call := 0
	bodyRun, _ := json.Marshal(map[string]any{
		"choices": []map[string]any{{"message": map[string]string{"content": `{"action":"ait_cli","args":"list"}`}}},
	})
	bodyAns, _ := json.Marshal(map[string]any{
		"choices": []map[string]any{{"message": map[string]string{"content": `{"action":"answer","text":"ok"}`}}},
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		call++
		w.Header().Set("Content-Type", "application/json")
		if call == 1 {
			_, _ = w.Write(bodyRun)
			return
		}
		_, _ = w.Write(bodyAns)
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

	var cliOut strings.Builder
	sr := pty.StartFrom(bytes.NewReader([]byte("y\r\x04")))
	replIn := sr.BeginREPL()
	defer sr.EndREPL()

	err = RunREPL(context.Background(), Options{
		Scope:          ai.ScopeLocal,
		AI:             client,
		In:             replIn,
		Out:            io.Discard,
		ErrOut:         io.Discard,
		UseTTYLineEdit: true,
		CLI: func(argv []string, out, errOut io.Writer) error {
			cliOut.WriteString(strings.Join(argv, " "))
			return nil
		},
	}, "list sessions")
	if err != nil {
		t.Fatal(err)
	}
	if cliOut.String() != "list" {
		t.Fatalf("cli=%q", cliOut.String())
	}
}

// RC3: non-TTY initial agent confirm still works with newline-delimited stdin.
func TestRunREPL_InitialAgentConfirm_PlainPath(t *testing.T) {
	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", t.TempDir())
	t.Setenv("AIT_ALLOW_CUSTOM_AI_URL", "1")
	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}

	call := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		call++
		w.Header().Set("Content-Type", "application/json")
		var body []byte
		if call == 1 {
			body, _ = json.Marshal(map[string]any{
				"choices": []map[string]any{{"message": map[string]string{"content": `{"action":"ait_cli","args":"list"}`}}},
			})
		} else {
			body, _ = json.Marshal(map[string]any{
				"choices": []map[string]any{{"message": map[string]string{"content": `{"action":"answer","text":"ok"}`}}},
			})
		}
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

	var cliOut strings.Builder
	err = RunREPL(context.Background(), Options{
		Scope:  ai.ScopeLocal,
		AI:     client,
		In:     strings.NewReader("y\n"),
		Out:    io.Discard,
		ErrOut: io.Discard,
		CLI: func(argv []string, out, errOut io.Writer) error {
			cliOut.WriteString(strings.Join(argv, " "))
			return nil
		},
	}, "list sessions")
	if err != nil {
		t.Fatal(err)
	}
	if cliOut.String() != "list" {
		t.Fatalf("cli=%q", cliOut.String())
	}
}
