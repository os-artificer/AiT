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

package agentcore

import (
	"bufio"
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/ai"
)

// AG-13: the step-limit prompt maps answers to the next goal.
func TestParseContinueChoice(t *testing.T) {
	t.Parallel()
	cases := []struct {
		line    string
		goal    string
		decided bool
	}{
		{metaContinue, msgContinueGoal, true},
		{"y", msgContinueGoal, true},
		{"YES", msgContinueGoal, true},
		{metaExit, "", true},
		{"n", "", true},
		{"quit", "", true},
		{"list the files", "list the files", true},
		{"", "", false},
		{"   ", "", false},
	}

	for _, tc := range cases {
		goal, decided := parseContinueChoice(tc.line)
		if goal != tc.goal || decided != tc.decided {
			t.Fatalf("parseContinueChoice(%q) = (%q, %v), want (%q, %v)", tc.line, goal, decided, tc.goal, tc.decided)
		}
	}
}

// AG-13: a carriage-return answer (raw-mode Enter) is read instead of hanging.
func TestWaitContinueOrExit_CarriageReturnAnswer(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	cfg := Config{Out: &out}
	st := newSessionState(cfg)
	st.readLine = ttyLineReader(bufio.NewReader(strings.NewReader(metaContinue+"\r")), &out)

	if goal := waitWithDeadline(t, cfg, st); goal != msgContinueGoal {
		t.Fatalf("goal = %q, want %q", goal, msgContinueGoal)
	}
	if !strings.Contains(out.String(), promptAgentContinue) {
		t.Fatalf("prompt missing from output %q", out.String())
	}
}

// AG-13: blank lines re-prompt in a loop and exhausted input ends the session.
func TestWaitContinueOrExit_BlankThenEOF(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	cfg := Config{In: strings.NewReader("\n\n"), Out: &out}
	st := newSessionState(cfg)

	if goal := waitWithDeadline(t, cfg, st); goal != "" {
		t.Fatalf("goal = %q, want empty on EOF", goal)
	}
	if got := strings.Count(out.String(), promptAgentContinue); got != 3 {
		t.Fatalf("prompted %d times, want 3 (two blank lines + EOF)", got)
	}
}

// AG-13: a free-form answer becomes the next goal.
func TestWaitContinueOrExit_FreeFormGoal(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	cfg := Config{In: strings.NewReader("check disk usage\n"), Out: &out}
	st := newSessionState(cfg)

	if goal := waitWithDeadline(t, cfg, st); goal != "check disk usage" {
		t.Fatalf("goal = %q", goal)
	}
}

// AG-13: exit at the prompt ends the session without a follow-up goal.
func TestWaitContinueOrExit_ExitEndsSession(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	cfg := Config{In: strings.NewReader(metaExit + "\n"), Out: &out}
	st := newSessionState(cfg)

	if goal := waitWithDeadline(t, cfg, st); goal != "" {
		t.Fatalf("goal = %q, want empty on exit", goal)
	}
}

// AG-13: a cancelled context returns immediately without reading input.
func TestWaitContinueOrExit_ContextCancelled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := Config{In: strings.NewReader(metaContinue + "\n"), Out: &bytes.Buffer{}}
	if _, err := waitContinueOrExit(ctx, cfg, newSessionState(cfg)); err != context.Canceled {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

// AG-13: continuing past the step limit iterates instead of recursing, so a
// missing AI client surfaces once per goal rather than growing the stack.
func TestRunAgentTurn_ContinueIsIterative(t *testing.T) {
	t.Parallel()
	cfg := Config{In: strings.NewReader(""), Out: &bytes.Buffer{}}
	st := newSessionState(cfg)

	if err := runAgentTurn(context.Background(), cfg, st, "goal"); err == nil {
		t.Fatal("expected not-configured error without an AI client")
	}
	if err := runAgentTurn(context.Background(), cfg, st, ""); err != nil {
		t.Fatalf("empty goal should be a no-op, got %v", err)
	}
}

// AG-13: the loop prompt reflects the session scope.
func TestLoopPrompt_ScopeAndHost(t *testing.T) {
	t.Parallel()
	if got := loopPrompt(ai.ScopeLocal, ""); got != promptAgent {
		t.Fatalf("local prompt = %q", got)
	}
	if got := loopPrompt(ai.ScopeRemote, "u@h"); !strings.Contains(got, "u@h") {
		t.Fatalf("remote prompt = %q, want host label", got)
	}
	if got := loopPrompt(ai.ScopeRemote, ""); got != promptAgent {
		t.Fatalf("remote prompt without host = %q", got)
	}
}

// waitWithDeadline runs waitContinueOrExit and fails the test if it blocks,
// which is the regression this suite guards (raw-mode Enter used to hang).
func waitWithDeadline(t *testing.T, cfg Config, st *sessionState) string {
	t.Helper()

	type result struct {
		goal string
		err  error
	}
	done := make(chan result, 1)
	go func() {
		goal, err := waitContinueOrExit(context.Background(), cfg, st)
		done <- result{goal: goal, err: err}
	}()

	select {
	case res := <-done:
		if res.err != nil {
			t.Fatalf("waitContinueOrExit: %v", res.err)
		}
		return res.goal
	case <-time.After(2 * time.Second):
		t.Fatal("waitContinueOrExit blocked on input")
		return ""
	}
}
