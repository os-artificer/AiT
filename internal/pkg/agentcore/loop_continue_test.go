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
)

// AG-13: the step-limit prompt maps answers to the next goal without recursion.
func TestParseContinueChoice(t *testing.T) {
	t.Parallel()
	cases := []struct {
		line    string
		goal    string
		decided bool
	}{
		{"continue", msgContinueGoal, true},
		{"y", msgContinueGoal, true},
		{"YES", msgContinueGoal, true},
		{"exit", "", true},
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

// AG-13: a CR-terminated answer (raw TTY Enter) is read instead of hanging.
func TestWaitContinueOrExit_CarriageReturnAnswer(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	cfg := Config{Out: &out}
	st := newSessionState(cfg)
	st.readLine = ttyLineReader(bufio.NewReader(strings.NewReader("continue\r")), &out)

	goal := waitWithDeadline(t, cfg, st)
	if goal != msgContinueGoal {
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

	goal := waitWithDeadline(t, cfg, st)
	if goal != "" {
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

// AG-13: a cancelled context returns immediately without reading input.
func TestWaitContinueOrExit_ContextCancelled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := Config{In: strings.NewReader("continue\n"), Out: &bytes.Buffer{}}
	if _, err := waitContinueOrExit(ctx, cfg, newSessionState(cfg)); err != context.Canceled {
		t.Fatalf("err = %v, want context.Canceled", err)
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
