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
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/aissh/ai/mode/agent"
	"github.com/os-artificer/AiT/internal/aissh/ai/mode/approve"
	"github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
)

// AF4: S2 exhausted → errhint, no raw JSON in output.
func TestRunAgentTurn_FormatRetryExhausted(t *testing.T) {
	invalid := "{broken-envelope-not-prose"
	n := 0
	client := testAIClientHandler(t, func(int) string {
		n++
		return invalid
	})
	var out strings.Builder
	opt := Options{Scope: ai.ScopeLocal, AI: client, Out: &out}
	st := newSessionState(opt)
	st.interaction = ModeAgent

	err := runAgentTurn(context.Background(), opt, st, "check host")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrAgentMalformedReply) {
		t.Fatalf("err=%v want ErrAgentMalformedReply", err)
	}
	got := out.String()
	if strings.Contains(got, invalid) {
		t.Fatalf("raw reply leaked to output: %q", got)
	}
}

// AF5: S2 correction retry recovers on second LLM response.
func TestRunAgentTurn_FormatRetryRecovers(t *testing.T) {
	invalid := "{broken-envelope-not-prose"
	fixed := `{"action":"answer","text":"recovered"}`
	n := 0
	client := testAIClientHandler(t, func(int) string {
		n++
		if n == 1 {
			return invalid
		}
		return fixed
	})
	var out strings.Builder
	opt := Options{Scope: ai.ScopeLocal, AI: client, Out: &out}
	st := newSessionState(opt)
	st.interaction = ModeAgent

	if err := runAgentTurn(context.Background(), opt, st, "analyze"); err != nil {
		t.Fatal(err)
	}
	if n < 2 {
		t.Fatal("expected S2 retry call")
	}
	if !strings.Contains(out.String(), "recovered") {
		t.Fatalf("out=%q", out.String())
	}
}

// AF7: resolved assistant in history is canonical JSON, not raw malformed reply.
func TestRunAgentTurn_HistoryNormalizedAssistant(t *testing.T) {
	broken := `{"action":"answer","text":"# Title` + "\n\nContent"
	client := testAIClientHandler(t, func(int) string { return broken })
	opt := Options{Scope: ai.ScopeLocal, AI: client, Out: io.Discard}
	st := newSessionState(opt)
	st.interaction = ModeAgent

	if err := runAgentTurn(context.Background(), opt, st, "report"); err != nil {
		t.Fatal(err)
	}
	if len(st.history) == 0 {
		t.Fatal("expected history")
	}
	last := st.history[len(st.history)-1]
	if last.Role != "assistant" {
		t.Fatalf("last role=%q", last.Role)
	}
	if strings.Contains(last.Content, "\n\nContent") && strings.HasPrefix(last.Content, `{"action":"answer","text":"# Title`) {
		t.Fatalf("malformed raw still in history: %q", last.Content)
	}
	if !strings.HasPrefix(last.Content, `{"action":"answer"`) {
		t.Fatalf("expected canonical envelope: %q", last.Content)
	}
}

// AF8: format correction (S2) only runs at step==0; step>0 uses format_error path.
func TestRunAgentTurn_FormatRetryOnlyFirstStep(t *testing.T) {
	runJSON := `{"action":"run_shell","command":"echo hi"}`
	brokenRun := `{"action":"run_shell","command":"broken`
	fixed := `{"action":"answer","text":"done"}`
	n := 0
	client := testAIClientHandler(t, func(int) string {
		n++
		switch n {
		case 1:
			return runJSON
		case 2:
			return brokenRun
		default:
			return fixed
		}
	})
	opt := Options{
		Scope: ai.ScopeLocal,
		AI:    client,
		Out:   io.Discard,
		Capture: agent.NewCaptureRunner(func(context.Context, string) (string, error) {
			return "hi\n", nil
		}),
		ConfirmRead: func(string) (approve.Result, error) {
			return approve.Result{Run: true}, nil
		},
	}
	st := newSessionState(opt)
	st.interaction = ModeAgent

	if err := runAgentTurn(context.Background(), opt, st, "run"); err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("calls=%d want 3 (no extra S2 at step>0)", n)
	}
}

// AF10: ChatAgent hard error returns immediately without format retry.
func TestRunAgentTurn_FCErrorNoFormatRetry(t *testing.T) {
	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", t.TempDir())
	t.Setenv("AIT_ALLOW_CUSTOM_AI_URL", "1")
	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}
	n := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

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

	opt := Options{Scope: ai.ScopeLocal, AI: client, Out: io.Discard}
	st := newSessionState(opt)
	st.interaction = ModeAgent

	err = runAgentTurn(context.Background(), opt, st, "hello")
	if err == nil {
		t.Fatal("expected error")
	}
	if n != 1 {
		t.Fatalf("calls=%d want 1 (no S2 on hard error)", n)
	}
}

// AF11: S2 retry preserves original user question in committed history.
func TestRunAgentTurn_S2PreservesUserTurn(t *testing.T) {
	invalid := "{broken-envelope-not-prose"
	fixed := `{"action":"answer","text":"ok"}`
	n := 0
	client := testAIClientHandler(t, func(int) string {
		n++
		if n == 1 {
			return invalid
		}
		return fixed
	})
	opt := Options{Scope: ai.ScopeLocal, AI: client, Out: io.Discard}
	st := newSessionState(opt)
	st.interaction = ModeAgent

	if err := runAgentTurn(context.Background(), opt, st, "original question"); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, msg := range st.history {
		if msg.Role == "user" && msg.Content == "original question" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("history missing original user: %+v", st.history)
	}
}

// AF13: S2 exhausted restores history to turn start (C19/C20).
func TestRunAgentTurn_S2ExhaustRestoresHistory(t *testing.T) {
	seed := []ai.ChatMessage{
		{Role: "user", Content: "prior"},
		{Role: "assistant", Content: `{"action":"answer","text":"old"}`},
	}
	invalid := "{broken-envelope-not-prose"
	client := testAIClientHandler(t, func(int) string { return invalid })
	opt := Options{Scope: ai.ScopeLocal, AI: client, Out: io.Discard}
	st := newSessionState(opt)
	st.interaction = ModeAgent
	st.history = append([]ai.ChatMessage{}, seed...)

	err := runAgentTurn(context.Background(), opt, st, "new question")
	if !errors.Is(err, ErrAgentMalformedReply) {
		t.Fatalf("err=%v", err)
	}
	if len(st.history) != len(seed) {
		t.Fatalf("history len=%d want %d", len(st.history), len(seed))
	}
	for i := range seed {
		if st.history[i].Role != seed[i].Role || st.history[i].Content != seed[i].Content {
			t.Fatalf("history[%d]=%+v want %+v", i, st.history[i], seed[i])
		}
	}
}

func TestFormatMalformedReplyErr_HintShape(t *testing.T) {
	err := formatMalformedReplyErr()
	if !errors.Is(err, ErrAgentMalformedReply) {
		t.Fatal("expected sentinel")
	}
	s := err.Error()
	if !strings.Contains(s, "Try:") {
		t.Fatalf("missing Try block: %q", s)
	}
	for _, line := range []string{"mode ask", "help"} {
		if !strings.Contains(s, line) {
			t.Fatalf("missing try line %q in %q", line, s)
		}
	}
	if strings.HasPrefix(s, errhint.PrefixAiT) {
		t.Fatalf("CLI err should not have AiT prefix: %q", s)
	}
}

func TestRunAgentTurn_MultilineAnswerNoJSONLeak(t *testing.T) {
	body := "# Report\n\n## CPU\n8 cores"
	raw := `{"action":"answer","text":"` + body
	client := testAIClientHandler(t, func(int) string { return raw })
	var out strings.Builder
	opt := Options{Scope: ai.ScopeLocal, AI: client, Out: &out}
	st := newSessionState(opt)
	st.interaction = ModeAgent

	if err := runAgentTurn(context.Background(), opt, st, "hardware"); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if strings.Contains(got, `"action":"answer"`) {
		t.Fatalf("JSON prefix leaked: %q", got)
	}
	if !strings.Contains(got, "# Report") {
		t.Fatalf("missing report body: %q", got)
	}
}

func TestRunAgentTurn_ProseAnswer(t *testing.T) {
	client := testAIClientHandler(t, func(int) string { return "Plain text reply." })
	var out strings.Builder
	opt := Options{Scope: ai.ScopeLocal, AI: client, Out: &out}
	st := newSessionState(opt)
	st.interaction = ModeAgent

	if err := runAgentTurn(context.Background(), opt, st, "hello"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Plain text reply.") {
		t.Fatalf("out=%q", out.String())
	}
	last := st.history[len(st.history)-1]
	var parsed map[string]string
	if err := json.Unmarshal([]byte(last.Content), &parsed); err != nil {
		t.Fatalf("history not canonical JSON: %q err=%v", last.Content, err)
	}
}
