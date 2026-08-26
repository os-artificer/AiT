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

package opsai_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/opsai"
	"github.com/os-artificer/AiT/internal/opsai/connect"
	"github.com/os-artificer/AiT/internal/pkg/agentcore"
	acrunner "github.com/os-artificer/AiT/internal/pkg/agentcore/runner"
	"github.com/os-artificer/AiT/internal/pkg/agentops"
	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
	"github.com/os-artificer/AiT/internal/pkg/clihints"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/ssh/openssh"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("AIT_AI_EXPERIMENTAL", "1")
	os.Exit(m.Run())
}

// TestOpsAI_ExperimentalDisabled (OA2) help/manifest work; gated commands fail when off.
func TestOpsAI_ExperimentalDisabled(t *testing.T) {
	t.Setenv("AIT_AI_EXPERIMENTAL", "")
	config.SetAIExperimentalCLI(false)

	var helpBuf bytes.Buffer
	if err := opsai.ExecuteWriters([]string{"--help"}, &helpBuf, &helpBuf); err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(helpBuf.String(), "Usage:") {
		t.Fatalf("help missing Usage: %s", helpBuf.String())
	}

	var manBuf bytes.Buffer
	if err := opsai.ExecuteWriters([]string{aitplugin.ManifestArg}, &manBuf, &manBuf); err != nil {
		t.Fatalf("manifest: %v", err)
	}
	if !strings.Contains(manBuf.String(), toolname.AI) {
		t.Fatalf("manifest missing ai: %s", manBuf.String())
	}

	err := opsai.ExecuteWriters([]string{"ask", "hi"}, ioDiscard{}, ioDiscard{})
	if err == nil {
		t.Fatal("expected experimental gate error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "experimental") {
		t.Fatalf("expected experimental hint, got %v", err)
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }

// TestOpsAI_ExperimentalCLIFlag (OA13).
func TestOpsAI_ExperimentalCLIFlag(t *testing.T) {
	t.Setenv("AIT_AI_EXPERIMENTAL", "")
	config.SetAIExperimentalCLI(false)
	// Flag alone should enable for this process strip path; ask still needs AI client —
	// we only assert strip sets the CLI gate without hanging on empty ask.
	err := opsai.ExecuteWriters([]string{"--ai-experimental", "ask"}, ioDiscard{}, ioDiscard{})
	if err == nil {
		t.Fatal("expected ask-requires-message (gate passed)")
	}
	if !errors.Is(err, pkgai.ErrChatMessageRequired) && !strings.Contains(err.Error(), "ask") {
		// Gate may surface ChatMessageRequired or similar.
		if strings.Contains(strings.ToLower(err.Error()), "experimental") {
			t.Fatalf("still experimental-blocked: %v", err)
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

// TestOpsAI_DefaultShowsHelp (OA1-v4) no args prints help and Try ait agent without hanging.
func TestOpsAI_DefaultShowsHelp(t *testing.T) {
	t.Setenv("AIT_AI_EXPERIMENTAL", "1")
	config.SetAIExperimentalCLI(true)

	var out bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- opsai.ExecuteWriters(nil, &out, ioDiscard{})
	}()
	select {
	case <-time.After(5 * time.Second):
		t.Fatal("default ai hung >5s")
	case err := <-done:
		if err == nil {
			t.Fatal("expected Try hint for ait agent")
		}
		if !strings.Contains(out.String(), "Usage:") {
			t.Fatalf("expected help on stdout, got %q", out.String())
		}
		if !strings.Contains(err.Error(), clihints.TryAitAgent) && !strings.Contains(out.String(), toolname.AITCLI("agent")) {
			t.Fatalf("expected ait agent Try, err=%v out=%q", err, out.String())
		}
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
	bin := filepath.Join(dir, toolname.AIHost)
	script := "#!/bin/sh\necho snapshot-ok\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	var out bytes.Buffer
	if err := agentops.DispatchToolbox([]string{toolname.AIHost, "snapshot"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "snapshot-ok") {
		t.Fatalf("got %q", out.String())
	}
}

// TestOpsAI_ModelListNoTTY (OA14) — non-TTY model list finishes quickly and is not unknown-command.
func TestOpsAI_ModelListNoTTY(t *testing.T) {
	done := make(chan error, 1)
	go func() {
		done <- opsai.ExecuteWriters([]string{"model", "list"}, ioDiscard{}, ioDiscard{})
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
