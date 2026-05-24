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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
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
	if st.interaction != ModeAsk {
		t.Fatalf("interaction=%v", st.interaction)
	}
	if st.policy != ExecPolicyPerCmd {
		t.Fatalf("policy=%v", st.policy)
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
	st := &sessionState{policy: ExecPolicySessionOnce, sessionTrusted: true}
	if err := setApprove(st, "each", &buf, io.Discard); err != nil {
		t.Fatal(err)
	}
	if st.policy != ExecPolicyPerCmd || st.sessionTrusted {
		t.Fatal("each should reset trust")
	}
}

func TestSetApprove_Session(t *testing.T) {
	var buf bytes.Buffer
	st := newSessionState(Options{})
	if err := setApprove(st, "session", &buf, io.Discard); err != nil {
		t.Fatal(err)
	}
	if st.policy != ExecPolicySessionOnce {
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
	if st.policy != ExecPolicySessionOnce {
		t.Fatal("trust should set session approve")
	}
}

func TestPolicyDeprecatedAlias_Each(t *testing.T) {
	var buf, errBuf bytes.Buffer
	st := newSessionState(Options{})
	if err := setPolicyDeprecated(st, "per-cmd", &buf, &errBuf); err != nil {
		t.Fatal(err)
	}
	if st.policy != ExecPolicyPerCmd {
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
	if st.policy != ExecPolicySessionOnce {
		t.Fatal("policy once -> session")
	}
}

func TestConfirm_HighRiskIgnoresSession(t *testing.T) {
	st := &sessionState{policy: ExecPolicySessionOnce, sessionTrusted: true}
	opt := Options{Scope: ai.ScopeLocal, ExecPolicy: ExecPolicySessionOnce, Out: io.Discard, In: strings.NewReader("")}
	ok, err := confirmExecRisk(opt, st, "rm -rf /tmp/x", "rm", RiskHigh)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("empty confirm should not approve high risk")
	}
}

func TestApproveSession_LowRiskSkipsSecond(t *testing.T) {
	n := 0
	st := &sessionState{policy: ExecPolicySessionOnce}
	opt := Options{
		Scope:      ai.ScopeLocal,
		ExecPolicy: ExecPolicySessionOnce,
		Out:        io.Discard,
		ConfirmRead: func(string) (ConfirmResult, error) {
			n++
			return ConfirmResult{Run: true, TrustSession: true}, nil
		},
	}
	if _, err := confirmExecRisk(opt, st, "hostname", "hostname", RiskLow); err != nil {
		t.Fatal(err)
	}
	if !st.sessionTrusted {
		t.Fatal("should trust after session choice")
	}
	ok, err := confirmExecRisk(opt, st, "hostname", "hostname", RiskLow)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if n != 1 {
		t.Fatalf("confirm asked %d times, want 1", n)
	}
}

func TestClassifyRisk_HighPatterns(t *testing.T) {
	if ClassifyRisk("rm -rf /tmp/x") != RiskHigh {
		t.Fatal("expected high")
	}
}

func TestClassifyRisk_Low(t *testing.T) {
	if ClassifyRisk("hostname") != RiskLow {
		t.Fatal("expected low")
	}
}

func TestClassifyRisk_Interactive(t *testing.T) {
	if ClassifyRisk("vim /etc/hosts") != RiskInteractive {
		t.Fatal("expected interactive")
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
		Capture: NewCaptureRunner(func(ctx context.Context, cmd string) (string, error) {
			capBuf.WriteString(cmd)
			return "hi\n", nil
		}),
		Out: io.Discard,
		ConfirmRead: func(string) (ConfirmResult, error) {
			return ConfirmResult{Run: true}, nil
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
	for _, want := range []string{"mode ask", "approve each", "deepseek"} {
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
		ConfirmRead: func(string) (ConfirmResult, error) {
			return ConfirmResult{Run: true}, nil
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
		ConfirmRead: func(string) (ConfirmResult, error) {
			return ConfirmResult{Run: true}, nil
		},
	}
	st := newSessionState(opt)
	st.interaction = ModeAgent
	if err := runAgentTurn(context.Background(), opt, st, "list sessions"); err != nil {
		t.Fatal(err)
	}
}

func TestParseAgentAction_Run(t *testing.T) {
	act, ok := parseAgentAction(`{"action":"run_shell","command":"hostname"}`)
	if !ok || !actionIsRun(act) || act.Command != "hostname" {
		t.Fatalf("act=%+v ok=%v", act, ok)
	}
}
