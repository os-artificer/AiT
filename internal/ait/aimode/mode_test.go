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

	"github.com/os-artificer/AiT/internal/ait/aimode/agent"
	"github.com/os-artificer/AiT/internal/ait/aimode/approve"
	"github.com/os-artificer/AiT/internal/ait/aimode/risk"
	"github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
	"github.com/os-artificer/AiT/internal/pkg/ssh/client/pty"
)

func testAIClient(t *testing.T, messageContent string) *ai.Client {
	return testAIClientHandler(t, func(int) string { return messageContent })
}

func testAIClientHandler(t *testing.T, contentForCall func(call int) string) *ai.Client {
	t.Helper()
	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", t.TempDir())
	t.Setenv("AIT_ALLOW_CUSTOM_AI_URL", "1")
	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "k"); err != nil {
		t.Fatal(err)
	}
	n := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n++
		body, _ := json.Marshal(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"content": contentForCall(n)}}},
		})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
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
	return client
}

func TestSessionState_Defaults(t *testing.T) {
	st := newSessionState(Options{})
	if st.interaction != ModeAgent {
		t.Fatalf("interaction=%v", st.interaction)
	}
	if st.Trust.Policy != approve.PolicyPerCmd {
		t.Fatalf("policy=%v", st.Trust.Policy)
	}
}

func TestSetMode_Agent(t *testing.T) {
	var buf bytes.Buffer
	st := newSessionState(Options{})
	if err := setMode(context.Background(), Options{Out: &buf}, st, "agent", &buf); err != nil {
		t.Fatal(err)
	}
	if st.interaction != ModeAgent {
		t.Fatal("expected agent")
	}
	if !strings.Contains(buf.String(), "mode: agent") {
		t.Fatalf("out=%q", buf.String())
	}
}

func TestSetMode_Ask(t *testing.T) {
	var buf bytes.Buffer
	st := &sessionState{interaction: ModeAgent}
	if err := setMode(context.Background(), Options{Out: &buf}, st, "ask", &buf); err != nil {
		t.Fatal(err)
	}
	if st.interaction != ModeAsk {
		t.Fatal("expected ask")
	}
}

func TestAskAI_NoRunCapture(t *testing.T) {
	var local recordingLocal
	client := testAIClient(t, "use run hostname")
	opt := Options{
		Scope: ai.ScopeLocal,
		AI:    client,
		Local: &local,
		Out:   io.Discard,
	}
	st := newSessionState(opt)
	if err := askAI(context.Background(), opt, st, "what is my ip"); err != nil {
		t.Fatal(err)
	}
	if len(local.calls) != 0 {
		t.Fatal("ask mode should not run commands")
	}
}

func TestSetApprove_Each(t *testing.T) {
	var buf bytes.Buffer
	st := &sessionState{Trust: approve.Trust{Policy: approve.PolicySessionOnce, SessionTrusted: true}}
	if err := setApprove(st, "each", &buf, io.Discard); err != nil {
		t.Fatal(err)
	}
	if st.Trust.Policy != approve.PolicyPerCmd || st.Trust.SessionTrusted {
		t.Fatal("each should reset trust")
	}
}

func TestSetApprove_Session(t *testing.T) {
	var buf bytes.Buffer
	st := newSessionState(Options{})
	if err := setApprove(st, "session", &buf, io.Discard); err != nil {
		t.Fatal(err)
	}
	if st.Trust.Policy != approve.PolicySessionOnce {
		t.Fatal("expected session approve")
	}
}

func TestTrustAliasSession(t *testing.T) {
	var buf, errBuf bytes.Buffer
	st := newSessionState(Options{})
	_, err := ExecuteLine(context.Background(), Options{Out: &buf, ErrOut: &errBuf}, st, "trust", nil)
	if err != nil {
		t.Fatal(err)
	}
	if st.Trust.Policy != approve.PolicySessionOnce {
		t.Fatal("trust should set session approve")
	}
}

func TestPolicyDeprecatedAlias_Each(t *testing.T) {
	var buf, errBuf bytes.Buffer
	st := newSessionState(Options{})
	if err := setPolicyDeprecated(st, "per-cmd", &buf, &errBuf); err != nil {
		t.Fatal(err)
	}
	if st.Trust.Policy != approve.PolicyPerCmd {
		t.Fatal("policy per-cmd -> each")
	}
	if !strings.Contains(errBuf.String(), "deprecated") {
		t.Fatalf("stderr=%q", errBuf.String())
	}
}

func TestPolicyDeprecatedAlias_Session(t *testing.T) {
	var buf, errBuf bytes.Buffer
	st := newSessionState(Options{})
	if err := setPolicyDeprecated(st, "once", &buf, &errBuf); err != nil {
		t.Fatal(err)
	}
	if st.Trust.Policy != approve.PolicySessionOnce {
		t.Fatal("policy once -> session")
	}
}

// TestPolicyDeprecated_CRLF (F3) verifies the deprecation notice for the
// legacy `policy` alias terminates with \r\n so it does not staircase on a
// raw-mode TTY, and that the underlying setApprove behavior still runs.
func TestPolicyDeprecated_CRLF(t *testing.T) {
	var buf, errBuf bytes.Buffer
	st := newSessionState(Options{})
	if err := setPolicyDeprecated(st, "once", &buf, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errBuf.String(), "deprecated") {
		t.Fatalf("missing deprecation notice: %q", errBuf.String())
	}
	if !strings.HasSuffix(errBuf.String(), "\r\n") {
		t.Fatalf("deprecation notice must end with CRLF: %q", errBuf.String())
	}
	if strings.Contains(errBuf.String(), "\r\r\n") {
		t.Fatalf("double-wrapped CRLF: %q", errBuf.String())
	}
	if st.Trust.Policy != approve.PolicySessionOnce {
		t.Fatal("policy once -> session not applied")
	}
}

func TestAgentTurn_RunActionMock(t *testing.T) {
	var capBuf strings.Builder
	runJSON := `{"action":"run_shell","command":"echo hi"}`
	answerJSON := `{"action":"answer","text":"done"}`
	client := testAIClientHandler(t, func(call int) string {
		if call == 1 {
			return runJSON
		}
		return answerJSON
	})
	opt := Options{
		Scope: ai.ScopeLocal,
		AI:    client,
		Capture: agent.NewCaptureRunner(func(ctx context.Context, cmd string) (string, error) {
			capBuf.WriteString(cmd)
			return "hi\n", nil
		}),
		Out: io.Discard,
		ConfirmRead: func(string) (approve.Result, error) {
			return approve.Result{Run: true}, nil
		},
	}
	st := newSessionState(opt)
	st.interaction = ModeAgent
	if err := runAgentTurn(context.Background(), opt, st, "run echo"); err != nil {
		t.Fatal(err)
	}
	if capBuf.String() != "echo hi" {
		t.Fatalf("capture cmd=%q", capBuf.String())
	}
}

func TestModeStatusLine_IncludesApprove(t *testing.T) {
	st := newSessionState(Options{})
	client := testAIClient(t, "x")
	line := formatModeStatus(Options{AI: client}, st)
	for _, want := range []string{"mode agent", "approve each", "deepseek"} {
		if !strings.Contains(line, want) {
			t.Fatalf("missing %q in %q", want, line)
		}
	}
}

func TestIsMetaREPLCommand_ModeApprove(t *testing.T) {
	for _, cmd := range []string{"mode", "mode agent", "approve", "approve session", "trust"} {
		if !isMetaREPLCommand(cmd) {
			t.Fatalf("%q should be meta", cmd)
		}
	}
}

func TestAgent_SkipsVim(t *testing.T) {
	client := testAIClientHandler(t, func(call int) string {
		if call == 1 {
			return `{"action":"run_shell","command":"vim foo"}`
		}
		return `{"action":"answer","text":"skipped"}`
	})
	var out bytes.Buffer
	opt := Options{
		Scope: ai.ScopeLocal,
		AI:    client,
		Out:   &out,
		ConfirmRead: func(string) (approve.Result, error) {
			return approve.Result{Run: true}, nil
		},
	}
	st := newSessionState(opt)
	st.interaction = ModeAgent
	if err := runAgentTurn(context.Background(), opt, st, "edit file"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "interactive") {
		t.Fatalf("out=%q", out.String())
	}
}

func TestAgent_AitListMock(t *testing.T) {
	client := testAIClientHandler(t, func(call int) string {
		if call == 1 {
			return `{"action":"ait_cli","args":"list"}`
		}
		return `{"action":"answer","text":"ok"}`
	})
	opt := Options{
		Scope: ai.ScopeLocal,
		AI:    client,
		CLI: func(argv []string, out, errOut io.Writer) error {
			_, _ = io.WriteString(out, "sessions: 0\r\n")
			return nil
		},
		Out: io.Discard,
		ConfirmRead: func(string) (approve.Result, error) {
			return approve.Result{Run: true}, nil
		},
	}
	st := newSessionState(opt)
	st.interaction = ModeAgent
	if err := runAgentTurn(context.Background(), opt, st, "list sessions"); err != nil {
		t.Fatal(err)
	}
}

func TestParseAgentAction_Run(t *testing.T) {
	act, ok := agent.ParseAction(`{"action":"run_shell","command":"hostname"}`)
	if !ok || !act.IsRun() || act.Command != "hostname" {
		t.Fatalf("act=%+v ok=%v", act, ok)
	}
}

// A21a: prompt reflects interaction mode (Ask vs Agent).
func TestPrompt_ShowsInteractionMode(t *testing.T) {
	if !strings.Contains(Prompt(ai.ScopeLocal, "", ModeAgent), "(Agent)") {
		t.Fatal("expected Agent in prompt")
	}
	if !strings.Contains(Prompt(ai.ScopeLocal, "", ModeAsk), "(Ask)") {
		t.Fatal("expected Ask in prompt")
	}
}

// A21b: aimode-level CRLF conversion of the approve prompt template.
// approve.RenderPrompt content is asserted in approve/confirm_test.go.
func TestConfirmPromptCRLFConversion(t *testing.T) {
	p := approve.RenderPrompt("local", "run hostname", risk.Low)
	var buf strings.Builder
	aimWrite(&buf, p)
	if !strings.Contains(buf.String(), "\r\n") {
		t.Fatalf("expected CRLF output, got %q", buf.String())
	}
	if !strings.Contains(buf.String(), "Choice [y/c/s/N]:") {
		t.Fatalf("missing choice line: %q", buf.String())
	}
}

// A21 (aimode setApprove integration): "each" clears trust set.
func TestSetApprove_EachClearsPatterns(t *testing.T) {
	st := &sessionState{
		Trust: approve.Trust{
			SessionTrusted:  true,
			TrustedPatterns: map[string]struct{}{"hostname": {}},
			Policy:          approve.PolicySessionOnce,
		},
	}
	if err := setApprove(st, "each", io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	if st.Trust.SessionTrusted || st.Trust.TrustedPatterns != nil {
		t.Fatal("each should clear trust")
	}
}

func TestSetModel_M12(t *testing.T) {
	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", t.TempDir())
	t.Setenv("AIT_ALLOW_CUSTOM_AI_URL", "1")

	if err := creds.SetProviderAPIKey(config.ProviderMinimax, "k"); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.AI.Default = config.ProviderDeepseek
	cfg.AI.Providers = map[string]config.ProviderConfig{
		config.ProviderMinimax: {APIStyle: config.APIStyleOpenAI, BaseURL: "http://127.0.0.1/v1", Model: "m"},
	}

	client, err := ai.NewClient(cfg, ai.WithProviderOverride(config.ProviderMinimax))
	if err != nil {
		t.Fatal(err)
	}

	var showBuf bytes.Buffer
	if err := setModel(context.Background(), Options{AI: client, Out: &showBuf}, "", &showBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(showBuf.String(), "minimax") {
		t.Fatalf("show: %q", showBuf.String())
	}

	var setBuf bytes.Buffer
	if err := setModel(context.Background(), Options{AI: client, Out: &setBuf}, config.ProviderMinimax, &setBuf); err != nil {
		t.Fatal(err)
	}
	if client.SessionProvider() != config.ProviderMinimax {
		t.Fatalf("session provider=%q", client.SessionProvider())
	}
}

func TestSetModel_ErrNoAI(t *testing.T) {
	var buf bytes.Buffer
	err := setModel(context.Background(), Options{Out: &buf}, "", &buf)
	if !errors.Is(err, ai.ErrNotConfigured) {
		t.Fatalf("got %v", err)
	}
	if !strings.Contains(err.Error(), "Try:") {
		t.Fatalf("expected Try block: %v", err)
	}
}

func TestRunAgentOnce_Answer(t *testing.T) {
	client := testAIClientHandler(t, func(call int) string {
		return `{"action":"answer","text":"the system is Linux"}`
	})
	var out bytes.Buffer
	opt := Options{
		Scope: ai.ScopeLocal,
		AI:    client,
		Out:   &out,
	}
	if err := RunAgentOnce(context.Background(), opt, "what is the system?"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "the system is Linux") {
		t.Fatalf("out=%q", out.String())
	}
}

func TestRunAgentOnce_RunAndAnswer(t *testing.T) {
	var capBuf strings.Builder
	client := testAIClientHandler(t, func(call int) string {
		if call == 1 {
			return `{"action":"run_shell","command":"uname"}`
		}
		return `{"action":"answer","text":"Linux"}`
	})
	var out bytes.Buffer
	opt := Options{
		Scope: ai.ScopeLocal,
		AI:    client,
		Capture: agent.NewCaptureRunner(func(_ context.Context, cmd string) (string, error) {
			capBuf.WriteString(cmd)
			return "Linux\n", nil
		}),
		Out: &out,
		ConfirmRead: func(string) (approve.Result, error) {
			return approve.Result{Run: true}, nil
		},
	}
	if err := RunAgentOnce(context.Background(), opt, "check OS"); err != nil {
		t.Fatal(err)
	}
	if capBuf.String() != "uname" {
		t.Fatalf("capture cmd=%q", capBuf.String())
	}
	if !strings.Contains(out.String(), "Linux") {
		t.Fatalf("out=%q", out.String())
	}
}

func TestRunAgentOnce_EmptyMessage(t *testing.T) {
	err := RunAgentOnce(context.Background(), Options{AI: testAIClient(t, "x"), Out: io.Discard}, "  ")
	if !errors.Is(err, ai.ErrChatMessageRequired) {
		t.Fatalf("got %v", err)
	}
}

func TestRunAgentOnce_NilAI(t *testing.T) {
	err := RunAgentOnce(context.Background(), Options{Out: io.Discard}, "hello")
	if !errors.Is(err, ai.ErrNotConfigured) {
		t.Fatalf("got %v", err)
	}
}

// RC1: withTTYConfirm on ReplIn accepts CR-terminated confirm (SSH raw TTY path).
func TestRunAgentOnce_TTYConfirmAcceptsCR(t *testing.T) {
	client := testAIClientHandler(t, func(call int) string {
		if call == 1 {
			return `{"action":"ait_cli","args":"list"}`
		}
		return `{"action":"answer","text":"ok"}`
	})

	var cliOut strings.Builder
	sr := pty.StartFrom(bytes.NewReader([]byte("y\r")))
	replIn := sr.BeginREPL()
	defer sr.EndREPL()

	br := bufio.NewReader(replIn)
	opt := Options{
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
	}
	opt = opt.withTTYConfirm(br)

	if err := RunAgentOnce(context.Background(), opt, "list sessions"); err != nil {
		t.Fatal(err)
	}
	if cliOut.String() != "list" {
		t.Fatalf("cli=%q", cliOut.String())
	}
}

func TestMain(m *testing.M) {
	_ = os.Setenv("AIT_KEYRING_BACKEND", "file")
	os.Exit(m.Run())
}
