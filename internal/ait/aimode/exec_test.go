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
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ai"
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
		ConfirmRead: func(string) (bool, error) {
			return false, nil
		},
	}
	st := &sessionState{policy: ExecPolicyPerCmd}

	_, err := ExecuteLine(context.Background(), opt, st, "run echo hi")
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
		ConfirmRead: func(string) (bool, error) {
			n++
			return true, nil
		},
	}
	st := &sessionState{policy: ExecPolicySessionOnce}

	for i := 0; i < 2; i++ {
		if _, err := ExecuteLine(context.Background(), opt, st, "run uptime"); err != nil {
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
	st := &sessionState{policy: ExecPolicySessionOnce, sessionTrusted: true}
	exit, err := ExecuteLine(context.Background(), Options{}, st, "exit")
	if err != nil || !exit {
		t.Fatalf("exit=%v err=%v", exit, err)
	}
	if st.sessionTrusted {
		t.Fatal("sessionTrusted should clear on exit")
	}
}

func TestExecuteLine_AitBlockedOnRemote(t *testing.T) {
	_, err := ExecuteLine(context.Background(), Options{Scope: ai.ScopeRemote, Out: io.Discard, ErrOut: io.Discard}, &sessionState{}, "ait list")
	if err == nil || !strings.Contains(err.Error(), "local-only") {
		t.Fatalf("err=%v", err)
	}
}

func TestExecuteLine_PolicyOnce(t *testing.T) {
	var buf bytes.Buffer
	st := &sessionState{policy: ExecPolicyPerCmd}

	if err := setPolicy(st, "once", &buf); err != nil {
		t.Fatal(err)
	}
	if st.policy != ExecPolicySessionOnce {
		t.Fatal("policy not session_once")
	}
	if !strings.Contains(buf.String(), "session_once") {
		t.Fatalf("output=%q", buf.String())
	}
}
