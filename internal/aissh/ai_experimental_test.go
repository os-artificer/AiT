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

package aissh

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/aissh/ai"
	"github.com/os-artificer/AiT/internal/aissh/shell"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
	"github.com/os-artificer/AiT/internal/pkg/session"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
)

func disableAIExperimental(t *testing.T) {
	t.Helper()
	config.SetAIExperimentalCLI(false)
	t.Setenv(config.EnvAIExperimental, "")
	t.Cleanup(func() { config.SetAIExperimentalCLI(false) })
}

func assertExperimentalDisabled(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("want ErrAIExperimentalDisabled")
	}
	if !errors.Is(err, config.ErrAIExperimentalDisabled) {
		t.Fatalf("got %v, want ErrAIExperimentalDisabled", err)
	}
	if !strings.Contains(err.Error(), "Try:") {
		t.Fatalf("missing Try block in %q", err.Error())
	}
}

// EX4: aitssh ai / ask / provider / ollama / model list gated when experimental off.
// Gate must not depend on os.Args containing -h/--help (skipAIExperimentalRequire uses Cobra only).
func TestAICmd_ExperimentalDisabled(t *testing.T) {
	disableAIExperimental(t)
	t.Setenv("AIT_HOME", t.TempDir())

	cases := [][]string{
		{"ai"},
		{"ai", "ask", "hi"},
		{"ai", "ask", "please", "help", "me"}, // bare "help" in message must not skip gate
		{"ai", "ask", "help"},
		{"ai", "agent", "please", "help", "me"},
		{"ai", "agent", "help"},
		{"ai", "model", "list"},
		{"ai", "provider", "enable", "deepseek"},
		{"ai", "ollama", "status"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			err := ExecuteWriters(args, io.Discard, io.Discard)
			assertExperimentalDisabled(t, err)
		})
	}
}

// Stub always prints experimental-disabled even if process CLI gate flips later.
func TestDisabledRemoteAIRunner_IgnoresCLIFlip(t *testing.T) {
	disableAIExperimental(t)
	config.SetAIExperimentalCLI(true)

	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	runErr := disabledRemoteAIRunner(sshclient.RemoteAIParams{})
	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	os.Stderr = old

	if runErr != nil {
		t.Fatalf("stub must return nil: %v", runErr)
	}
	out := buf.String()
	if !strings.Contains(out, "experimental") {
		t.Fatalf("want experimental disabled hint, got %q", out)
	}
}

// EX4b: configured API keys must not bypass the experimental gate via LocalAIClient.
func TestAIAsk_ExperimentalDisabledWithKeys(t *testing.T) {
	disableAIExperimental(t)
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)
	t.Setenv("AIT_KEYRING_DIR", root)
	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "test-key"); err != nil {
		t.Fatal(err)
	}

	err := ExecuteWriters([]string{"ai", "ask", "ping"}, io.Discard, io.Discard)
	assertExperimentalDisabled(t, err)
}

// EX4 help: --help remains available when experimental is off.
func TestAIHelp_AllowedWhenDisabled(t *testing.T) {
	disableAIExperimental(t)
	t.Setenv("AIT_HOME", t.TempDir())

	var out bytes.Buffer
	err := ExecuteWriters([]string{"ai", "--help"}, &out, io.Discard)
	if err != nil {
		t.Fatalf("help should succeed: %v", err)
	}
	if !strings.Contains(out.String(), "Usage:") && !strings.Contains(out.String(), "ai") {
		t.Fatalf("unexpected help output: %q", out.String())
	}
}

// EX6: shell builtin ai uses the same Require gate.
func TestShellAI_ExperimentalDisabled(t *testing.T) {
	disableAIExperimental(t)
	t.Setenv("AIT_HOME", t.TempDir())

	shell.SetAIModeRunner(ai.RunLocalFromShell(ExecuteCobraOnly))
	mgr := &session.Manager{Config: &config.Config{}}
	_, err := shell.ExecuteLine(context.Background(), mgr, "ai", nil)
	assertExperimentalDisabled(t, err)
}

// EX10: process-entry --ai-experimental enables for ExecuteWriters and does not write config.
func TestAIExperimental_CLIFlag_ExecuteWriters(t *testing.T) {
	disableAIExperimental(t)
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	err := ExecuteWriters([]string{config.FlagAIExperimental, "ai", "ask", "hi"}, io.Discard, io.Discard)
	if errors.Is(err, config.ErrAIExperimentalDisabled) {
		t.Fatalf("CLI flag should enable experimental: %v", err)
	}
	if !config.AIExperimentalCLI() {
		t.Fatal("want AIExperimentalCLI true after ExecuteWriters saw flag")
	}
	if _, err := os.Stat(filepath.Join(root, "config.yaml")); err == nil {
		t.Fatal("CLI flag must not write config.yaml")
	}
}

// EX12: nested ExecuteCobraOnly must not flip the process CLI gate.
func TestExecuteCobraOnly_AIExperimentalFlagNoFlip(t *testing.T) {
	disableAIExperimental(t)
	t.Setenv("AIT_HOME", t.TempDir())

	err := ExecuteCobraOnly([]string{config.FlagAIExperimental, "ai", "ask", "hi"}, io.Discard, io.Discard)
	assertExperimentalDisabled(t, err)
	if config.AIExperimentalCLI() {
		t.Fatal("ExecuteCobraOnly must not SetAIExperimentalCLI")
	}
}

// EX13: ai subtree children inherit the root PersistentPreRunE gate (no per-child Require list).
func TestAISubtree_NewChildInheritsExperimentalGate(t *testing.T) {
	disableAIExperimental(t)
	t.Setenv("AIT_HOME", t.TempDir())

	cmd := ai.New(ai.Deps{
		Exec: ExecuteCobraOnly,
		RunWithMgr: func(fn func(context.Context, *session.Manager) error) error {
			return fn(context.Background(), &session.Manager{Config: &config.Config{}})
		},
	})
	if cmd.PersistentPreRunE == nil {
		t.Fatal("ai root must define PersistentPreRunE")
	}
	for _, child := range cmd.Commands() {
		if child.PersistentPreRunE != nil {
			t.Fatalf("child %q must not override PersistentPreRunE (inherit ai root gate)", child.Name())
		}
	}

	for _, args := range [][]string{
		{"ai", "ask", "x"},
		{"ai", "agent", "x"},
		{"ai", "ollama", "status"},
	} {
		err := ExecuteWriters(args, io.Discard, io.Discard)
		assertExperimentalDisabled(t, err)
	}
}
