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
	"io"
	"runtime"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/ait/aimode/approve"
	"github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

type recordingRunner struct {
	calls []string
}

func (r *recordingRunner) Run(cmd string) error {
	r.calls = append(r.calls, cmd)
	return nil
}

type recordingLocal struct {
	recordingRunner
}

func (r *recordingLocal) Run(ctx context.Context, cmd string) error {
	_ = ctx
	return r.recordingRunner.Run(cmd)
}

func TestExecuteLine_PerCmdSkipsWithoutYes(t *testing.T) {
	var local recordingLocal
	opt := Options{
		Scope:  ai.ScopeLocal,
		Local:  &local,
		Out:    io.Discard,
		ErrOut: io.Discard,
		ConfirmRead: func(string) (approve.Result, error) {
			return approve.Result{}, nil
		},
	}
	st := &sessionState{Trust: approve.Trust{Policy: approve.PolicyPerCmd}}

	_, err := ExecuteLine(context.Background(), opt, st, "run echo hi", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(local.calls) != 0 {
		t.Fatalf("expected no run, got %v", local.calls)
	}
}

func TestExecuteLine_SessionOnceTrustsAfterYes(t *testing.T) {
	var remote recordingRunner
	n := 0
	opt := Options{
		Scope:  ai.ScopeRemote,
		Remote: &remote,
		Out:    io.Discard,
		ErrOut: io.Discard,
		ConfirmRead: func(string) (approve.Result, error) {
			n++
			return approve.Result{Run: true, TrustSession: true}, nil
		},
	}
	st := &sessionState{Trust: approve.Trust{Policy: approve.PolicySessionOnce}}

	for i := 0; i < 2; i++ {
		if _, err := ExecuteLine(context.Background(), opt, st, "run uptime", nil); err != nil {
			t.Fatal(err)
		}
	}
	if n != 1 {
		t.Fatalf("confirm asked %d times, want 1", n)
	}
	if len(remote.calls) != 2 {
		t.Fatalf("runs=%v", remote.calls)
	}
}

func TestExecuteLine_ExitClearsSessionTrust(t *testing.T) {
	st := &sessionState{
		Trust: approve.Trust{
			Policy:          approve.PolicySessionOnce,
			SessionTrusted:  true,
			TrustedPatterns: map[string]struct{}{"hostname": {}},
		},
	}
	exit, err := ExecuteLine(context.Background(), Options{}, st, "exit", nil)
	if err != nil || !exit {
		t.Fatalf("exit=%v err=%v", exit, err)
	}
	if st.Trust.SessionTrusted || st.Trust.TrustedPatterns != nil {
		t.Fatal("exec trust should clear on exit")
	}
}

func TestExecuteLine_AitBlockedOnRemote(t *testing.T) {
	_, err := ExecuteLine(context.Background(), Options{Scope: ai.ScopeRemote, Out: io.Discard, ErrOut: io.Discard}, &sessionState{}, "ait list", nil)
	if err == nil || !strings.Contains(err.Error(), "local-only") {
		t.Fatalf("err=%v", err)
	}
}

func TestExecuteLine_ApproveSession(t *testing.T) {
	var buf bytes.Buffer
	st := &sessionState{Trust: approve.Trust{Policy: approve.PolicyPerCmd}}

	if err := setApprove(st, "session", &buf, io.Discard); err != nil {
		t.Fatal(err)
	}
	if st.Trust.Policy != approve.PolicySessionOnce {
		t.Fatal("policy not session_once")
	}
	if !strings.Contains(buf.String(), "approve: session") {
		t.Fatalf("output=%q", buf.String())
	}
}

func TestExecuteLineHistory_AIMode(t *testing.T) {
	var out bytes.Buffer
	st := &sessionState{history: []ai.ChatMessage{{Role: "user", Content: "hi"}}}
	hist := terminal.NewLineHistory()
	hist.Add("run ls")

	exit, err := ExecuteLine(context.Background(), Options{Out: &out, ErrOut: io.Discard}, st, "clear", hist)
	if err != nil || exit {
		t.Fatalf("clear: exit=%v err=%v", exit, err)
	}
	if len(st.history) != 0 {
		t.Fatal("clear should wipe chat history")
	}

	out.Reset()
	exit, err = ExecuteLine(context.Background(), Options{Out: &out, ErrOut: io.Discard}, st, "history", hist)
	if err != nil || exit {
		t.Fatalf("history: exit=%v err=%v", exit, err)
	}
	if !strings.Contains(out.String(), "run ls") {
		t.Fatalf("input history missing after clear: %q", out.String())
	}
	if !strings.Contains(out.String(), "\r\n") {
		t.Fatalf("history output should use CRLF for raw-TTY safety, got %q", out.String())
	}
	if strings.Contains(out.String(), "\r\r\n") {
		t.Fatalf("history output must not double-wrap CRLF, got %q", out.String())
	}
}

// TestExecuteAitSubcmdAIMode_CRLF (F2) verifies that the local AI REPL's
// `ait <subcmd>` execution wraps both stdout and stderr in a CRLF writer so
// cobra/subcommand bare \n output renders correctly on the raw-mode TTY.
func TestExecuteAitSubcmdAIMode_CRLF(t *testing.T) {
	var out, errOut bytes.Buffer
	opt := Options{
		Scope:  ai.ScopeLocal,
		Out:    &out,
		ErrOut: &errOut,
		CLI: func(argv []string, w io.Writer, ew io.Writer) error {
			_, _ = io.WriteString(w, "line1\nline2\n")
			_, _ = io.WriteString(ew, "err1\nerr2\n")
			return nil
		},
		ConfirmRead: func(string) (approve.Result, error) {
			return approve.Result{Run: true}, nil
		},
	}
	st := newSessionState(opt)

	exit, err := ExecuteLine(context.Background(), opt, st, "ait list", terminal.NewLineHistory())
	if err != nil || exit {
		t.Fatalf("ait list: exit=%v err=%v", exit, err)
	}

	if got := out.String(); !strings.Contains(got, "line1\r\nline2\r\n") {
		t.Errorf("stdout missing CRLF: %q", got)
	}
	if strings.Contains(out.String(), "\r\r\n") {
		t.Errorf("stdout double-wrapped CRLF: %q", out.String())
	}
	if got := errOut.String(); !strings.Contains(got, "err1\r\nerr2\r\n") {
		t.Errorf("stderr missing CRLF: %q", got)
	}
	if strings.Contains(errOut.String(), "\r\r\n") {
		t.Errorf("stderr double-wrapped CRLF: %q", errOut.String())
	}
}

// TestExecuteLineHistory_AIMode_CRLF (F1) verifies that the multi-line history
// listing in the AI REPL terminates each entry with \r\n so it renders one row
// per line on the raw-mode local TTY used by ait ai and the !! remote AI REPL.
func TestExecuteLineHistory_AIMode_CRLF(t *testing.T) {
	var out bytes.Buffer
	st := &sessionState{}
	hist := terminal.NewLineHistory()
	hist.Add("first")
	hist.Add("second")
	hist.Add("third")

	exit, err := ExecuteLine(context.Background(), Options{Out: &out, ErrOut: io.Discard}, st, "history", hist)
	if err != nil || exit {
		t.Fatalf("history: exit=%v err=%v", exit, err)
	}

	got := out.String()
	if !strings.Contains(got, "first\r\n") {
		t.Errorf("missing CRLF after first entry: %q", got)
	}
	if !strings.Contains(got, "second\r\n") {
		t.Errorf("missing CRLF after second entry: %q", got)
	}
	if !strings.Contains(got, "third\r\n") {
		t.Errorf("missing CRLF after third entry: %q", got)
	}
	if strings.Contains(got, "\r\r\n") {
		t.Errorf("double-wrapped CRLF in history output: %q", got)
	}
}

func TestDefaultLocalRunner_Echo(t *testing.T) {
	r := DefaultLocalRunner()
	ctx := context.Background()
	cmd := "echo ait-ai-local-test"
	if runtime.GOOS == "windows" {
		cmd = "echo ait-ai-local-test"
	}
	if err := r.Run(ctx, cmd); err != nil {
		t.Fatal(err)
	}
}
