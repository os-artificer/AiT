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

package connect

import (
	"bytes"
	"strings"
	"testing"
	"time"

	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func TestLooksLikeDestination_Table(t *testing.T) {
	reserved := []string{
		"ask", "agent", "connect", "version", "provider", "ollama", "model", "help",
		"mode", "approve", "trust", "policy", "run", "history", "clear", "exit", "quit", "disconnect",
		toolname.AISsh,
	}
	for _, tok := range reserved {
		if LooksLikeDestination(tok) {
			t.Errorf("%q must not look like destination", tok)
		}
	}

	destinations := []string{
		"user@host.example.com",
		"devhost",
		"root@host.example.com:2222",
	}
	for _, dest := range destinations {
		if !LooksLikeDestination(dest) {
			t.Errorf("%q should look like destination", dest)
		}
	}

	if LooksLikeDestination("-J") || LooksLikeDestination("--help") {
		t.Fatal("flags must not look like destination")
	}
}

// TestRemoteAgentOptions_ScopeRemote (OA5) remote connect loop uses ScopeRemote + ops domain.
func TestRemoteAgentOptions_ScopeRemote(t *testing.T) {
	var stdin strings.Builder
	cfg := buildRemoteAgentConfig(nil, sshclient.AgentSessionHandles{
		HostLabel: "u@h",
		Stdin:     &stdin,
		RunCapture: func(cmd string) (string, error) {
			return "out:" + cmd, nil
		},
	}, agentLimits{maxSteps: 5, commandTimeout: time.Minute}, nil)
	if cfg.CommandTimeout != time.Minute {
		t.Fatalf("CommandTimeout=%v want 1m", cfg.CommandTimeout)
	}
	if cfg.Scope != pkgai.ScopeRemote {
		t.Fatalf("Scope=%v want remote", cfg.Scope)
	}
	if cfg.HostLabel != "u@h" {
		t.Fatalf("HostLabel=%q", cfg.HostLabel)
	}
	if cfg.Domain != "ops" {
		t.Fatalf("Domain=%q want ops", cfg.Domain)
	}
	if cfg.Capture == nil || cfg.Remote == nil {
		t.Fatal("remote agent missing Capture or Remote runner")
	}
}

// TestRemoteAgentOptions_Capture (OA6) RunCapture is wired through Capture runner.
func TestRemoteAgentOptions_Capture(t *testing.T) {
	const want = "captured"
	cfg := buildRemoteAgentConfig(nil, sshclient.AgentSessionHandles{
		HostLabel: "u@h",
		RunCapture: func(cmd string) (string, error) {
			if cmd != "hostname" {
				t.Fatalf("cmd=%q", cmd)
			}
			return want, nil
		},
	}, agentLimits{}, nil)
	if cfg.Capture == nil {
		t.Fatal("missing Capture")
	}
	out, err := cfg.Capture.RunCapture(t.Context(), "hostname")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, want) {
		t.Fatalf("got %q", out)
	}
}

func TestPTYRemoteRunner_WritesNewline(t *testing.T) {
	var stdin bytes.Buffer
	r := &ptyRemoteRunner{stdin: &stdin}
	if err := r.Run("kubectl get pods"); err != nil {
		t.Fatal(err)
	}
	if stdin.String() != "kubectl get pods\n" {
		t.Fatalf("got %q", stdin.String())
	}
}
