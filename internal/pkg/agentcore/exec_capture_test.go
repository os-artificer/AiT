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
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/agentcore/runner"
	"github.com/os-artificer/AiT/internal/pkg/ai"
)

// AG-12: a runner that ignores ctx cannot block the loop past the timeout.
func TestRunBoundedCapture_TimeoutOnBlockedRunner(t *testing.T) {
	t.Parallel()
	cfg := Config{
		CommandTimeout: 50 * time.Millisecond,
		Capture: runner.NewCaptureRunner(func(context.Context, string) (string, error) {
			time.Sleep(5 * time.Second)
			return "late", nil
		}),
	}

	start := time.Now()
	out, err := runBoundedCapture(context.Background(), cfg, "sleep 5")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}
	if out != "" {
		t.Fatalf("expected no output, got %q", out)
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout message, got %v", err)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("bounded capture waited %s; timeout not enforced", elapsed)
	}
}

// AG-12: a context-aware runner still sees the derived deadline.
func TestRunBoundedCapture_PassesDeadlineToRunner(t *testing.T) {
	t.Parallel()
	var hasDeadline bool
	cfg := Config{
		CommandTimeout: time.Minute,
		Capture: runner.NewCaptureRunner(func(ctx context.Context, _ string) (string, error) {
			_, hasDeadline = ctx.Deadline()
			return "ok", nil
		}),
	}

	out, err := runBoundedCapture(context.Background(), cfg, "ls")
	if err != nil || out != "ok" {
		t.Fatalf("out=%q err=%v", out, err)
	}
	if !hasDeadline {
		t.Fatal("runner context has no deadline")
	}
}

// AG-12: caller cancellation is reported as ctx.Err, not as a timeout observation.
func TestRunBoundedCapture_CallerCancel(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cfg := Config{
		CommandTimeout: time.Minute,
		Capture: runner.NewCaptureRunner(func(ctx context.Context, _ string) (string, error) {
			<-ctx.Done()
			return "", ctx.Err()
		}),
	}

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	if _, err := runBoundedCapture(ctx, cfg, "sleep 60"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

// AG-12: the default bound applies when Config leaves CommandTimeout unset.
func TestAgentCommandTimeout_Default(t *testing.T) {
	t.Parallel()
	if got := (Config{}).AgentCommandTimeout(); got != defaultCommandTimeout {
		t.Fatalf("got %s, want %s", got, defaultCommandTimeout)
	}
	if got := (Config{CommandTimeout: time.Second}).AgentCommandTimeout(); got != time.Second {
		t.Fatalf("got %s, want 1s", got)
	}
}

// AG-12: the terminal is released for a local child and restored afterwards.
func TestWithTTYHandoff_SuspendsAndResumes(t *testing.T) {
	t.Parallel()
	var suspended, resumed, ranInside bool
	cfg := Config{
		SuspendTTY: func() func() {
			suspended = true
			return func() { resumed = true }
		},
	}

	err := withTTYHandoff(cfg, func() error {
		ranInside = suspended && !resumed
		return nil
	})
	if err != nil {
		t.Fatalf("withTTYHandoff: %v", err)
	}
	if !ranInside {
		t.Fatal("child ran without the terminal being suspended")
	}
	if !resumed {
		t.Fatal("terminal was not resumed")
	}
}

// AG-12: a failed handoff (nil resume) must not stop the command from running.
func TestWithTTYHandoff_NilResume(t *testing.T) {
	t.Parallel()
	cfg := Config{SuspendTTY: func() func() { return nil }}

	var ran bool
	if err := withTTYHandoff(cfg, func() error { ran = true; return nil }); err != nil {
		t.Fatalf("withTTYHandoff: %v", err)
	}
	if !ran {
		t.Fatal("command did not run")
	}
}

// AG-12: a remote capture keeps raw mode so Ctrl+C can reach the remote PTY.
func TestWithLocalChildTTY_RemoteScopeKeepsRawMode(t *testing.T) {
	t.Parallel()
	var suspended, ran bool
	suspend := func() func() {
		suspended = true
		return func() {}
	}

	remote := Config{Scope: ai.ScopeRemote, SuspendTTY: suspend}
	if err := withLocalChildTTY(remote, func() error { ran = true; return nil }); err != nil {
		t.Fatalf("withLocalChildTTY: %v", err)
	}
	if !ran {
		t.Fatal("remote command did not run")
	}
	if suspended {
		t.Fatal("remote capture must not suspend the local terminal")
	}

	local := Config{Scope: ai.ScopeLocal, SuspendTTY: suspend}
	if err := withLocalChildTTY(local, func() error { return nil }); err != nil {
		t.Fatalf("withLocalChildTTY: %v", err)
	}
	if !suspended {
		t.Fatal("local command should suspend the terminal")
	}
}

// AG-12: Dynamic remote state also suppresses the handoff after a hot connect.
func TestWithLocalChildTTY_DynamicRemoteKeepsRawMode(t *testing.T) {
	t.Parallel()
	dyn := &DynamicRemote{}
	dyn.SetRemote("u@h", nil, nil, nil)

	var suspended bool
	cfg := Config{
		Scope:      ai.ScopeLocal,
		Dynamic:    dyn,
		SuspendTTY: func() func() { suspended = true; return func() {} },
	}

	if err := withLocalChildTTY(cfg, func() error { return nil }); err != nil {
		t.Fatalf("withLocalChildTTY: %v", err)
	}
	if suspended {
		t.Fatal("dynamic remote scope must not suspend the local terminal")
	}
}
