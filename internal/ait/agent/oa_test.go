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

package agent_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/ait/agent"
	"github.com/os-artificer/AiT/internal/ait/agent/connect"
	"github.com/os-artificer/AiT/internal/pkg/agentcore"
	acrunner "github.com/os-artificer/AiT/internal/pkg/agentcore/runner"
	"github.com/os-artificer/AiT/internal/pkg/agentops"
	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/clihints"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/ssh/openssh"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("AIT_AI_EXPERIMENTAL", "1")
	// Never touch the macOS login keychain during go test (avoids interactive prompts).
	_ = os.Setenv("AIT_KEYRING_BACKEND", "file")
	_ = os.Setenv("AIT_KEYRING_DIR", os.TempDir())
	_ = os.Setenv("AIT_KEYRING_FILE_PASSWORD", "ait-test")
	os.Exit(m.Run())
}

// TestOpsAI_ExperimentalDisabled (OA2) model list is not blocked by experimental gate when off.
func TestOpsAI_ExperimentalDisabled(t *testing.T) {
	t.Setenv("AIT_HOME", t.TempDir())
	t.Setenv("AIT_AI_EXPERIMENTAL", "0")
	config.SetAIExperimentalCLI(false)

	done := make(chan error, 1)
	go func() {
		done <- agent.RunAdmin([]string{"model", "list"}, ioDiscard{}, ioDiscard{})
	}()
	select {
	case <-time.After(5 * time.Second):
		t.Fatal("model list hung >5s")
	case err := <-done:
		if err != nil && strings.Contains(strings.ToLower(err.Error()), "experimental") {
			t.Fatalf("model list should not be experimental-blocked: %v", err)
		}
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }

// TestOpsAI_ExperimentalCLIFlag (OA13) agent admin no longer uses --ai-experimental gate.
func TestOpsAI_ExperimentalCLIFlag(t *testing.T) {
	t.Setenv("AIT_HOME", t.TempDir())
	t.Setenv("AIT_AI_EXPERIMENTAL", "0")
	config.SetAIExperimentalCLI(false)
	done := make(chan error, 1)
	go func() {
		done <- agent.RunAdmin([]string{"model", "list"}, ioDiscard{}, ioDiscard{})
	}()
	select {
	case <-time.After(5 * time.Second):
		t.Fatal("model list hung >5s")
	case err := <-done:
		if err != nil && strings.Contains(strings.ToLower(err.Error()), "experimental") {
			t.Fatalf("still experimental-blocked: %v", err)
		}
	}
}

// TestOpsAI_AskAgentRemoved (RO-1) one-shot ask/agent are unknown; Try points at ait agent.
func TestOpsAI_AskAgentRemoved(t *testing.T) {
	for _, argv := range [][]string{{"ask", "hi"}, {"agent", "hi"}} {
		err := agent.RunAdmin(argv, ioDiscard{}, ioDiscard{})
		if err == nil {
			t.Fatalf("%v: expected unknown command", argv)
		}
		msg := err.Error()
		if !strings.Contains(msg, "unknown agent command") {
			t.Fatalf("%v: want unknown, got %v", argv, err)
		}
		if !strings.Contains(msg, clihints.TryAitAgent()) {
			t.Fatalf("%v: expected ait agent Try, got %v", argv, err)
		}
	}
}

// TestOpsAI_AgentNotConnect (OA17).
func TestOpsAI_AgentNotConnect(t *testing.T) {
	reserved := []string{"agent", "ask", "mode", "approve", "run", "history", "clear", "exit", "disconnect"}
	for _, tok := range reserved {
		if connect.LooksLikeDestination(tok) {
			t.Fatalf("%q must not look like connect destination", tok)
		}
	}
	if !connect.LooksLikeDestination("user@host") {
		t.Fatal("user@host should look like destination")
	}
}

// TestOpsAI_DefaultShowsHelp (OA1-v4) empty admin args prints help and Try ait agent.
func TestOpsAI_DefaultShowsHelp(t *testing.T) {
	var out bytes.Buffer
	err := agent.RunAdmin(nil, &out, ioDiscard{})
	if err == nil {
		t.Fatal("expected Try hint for ait agent")
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("expected help on stdout, got %q", out.String())
	}
	if !strings.Contains(err.Error(), clihints.TryAitAgent()) && !strings.Contains(out.String(), toolname.AITCLI("agent")) {
		t.Fatalf("expected ait agent Try, err=%v out=%q", err, out.String())
	}
}

// TestOpsAI_ConnectRemoteAgent (OA5) bare destination is connect-shaped, not REPL verb.
func TestOpsAI_ConnectRemoteAgent(t *testing.T) {
	if !connect.LooksLikeDestination("root@host.example.com") {
		t.Fatal("bare destination should route to connect")
	}
	if connect.LooksLikeDestination("connect") {
		t.Fatal("connect subcommand must stay reserved")
	}
}

// TestOpsAI_RemoteCapture (OA6) connect options builder wires capture (see connect/run_test.go).
func TestOpsAI_RemoteCapture(t *testing.T) {
	// Smoke: package connect tests cover Capture wiring; ensure LooksLikeDestination does not block host labels.
	if connect.LooksLikeDestination("devhost") {
		return
	}
	t.Fatal("devhost should look like destination for connect routing")
}

// TestOpsAI_ConnectProxyJump (OA15) -J is a flag, not a bare destination.
func TestOpsAI_ConnectProxyJump(t *testing.T) {
	if connect.LooksLikeDestination("-J") {
		t.Fatal("-J must not bare-route to connect")
	}
	spec, err := openssh.ParseLogin([]string{"-J", "jump@host.example.com", "root@final.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.ProxyJump) != 1 || spec.ProxyJump[0] != "jump@host.example.com" {
		t.Fatalf("ProxyJump=%v", spec.ProxyJump)
	}
}

// TestOpsAI_Disconnect (OA7) DynamicRemote.SetLocal clears remote capture.
func TestOpsAI_Disconnect(t *testing.T) {
	dyn := &agentcore.DynamicRemote{}
	dyn.SetRemote("u@h", acrunner.NewCaptureRunner(func(context.Context, string) (string, error) {
		return "ok", nil
	}), nil, nil)
	dyn.SetLocal()
	snap := dyn.Snapshot()
	if snap.Scope != pkgai.ScopeLocal || snap.Capture != nil {
		t.Fatal("DynamicRemote SetLocal should clear remote capture")
	}
}

// TestToolbox_DispatchAIHost (OA3) uses agentops.DispatchToolbox (single SSOT).
func TestToolbox_DispatchAIHost(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, toolname.AITHost)
	script := "#!/bin/sh\necho snapshot-ok\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	var out bytes.Buffer
	if err := agentops.DispatchToolbox([]string{toolname.AITHost, "snapshot"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "snapshot-ok") {
		t.Fatalf("got %q", out.String())
	}
}

// TestOpsAI_ModelListNoTTY (OA14) — non-TTY model list finishes quickly and is not unknown-command.
func TestOpsAI_ModelListNoTTY(t *testing.T) {
	t.Setenv("AIT_HOME", t.TempDir())
	done := make(chan error, 1)
	go func() {
		done <- agent.RunAdmin([]string{"model", "list"}, ioDiscard{}, ioDiscard{})
	}()
	select {
	case <-time.After(5 * time.Second):
		t.Fatal("model list hung >5s")
	case err := <-done:
		if err != nil && strings.Contains(err.Error(), `unknown command "model"`) {
			t.Fatalf("model list must reach model subcommand, got %v", err)
		}
		if err != nil && strings.Contains(strings.ToLower(err.Error()), "experimental") {
			t.Fatalf("model list blocked by experimental gate: %v", err)
		}
	}
}
