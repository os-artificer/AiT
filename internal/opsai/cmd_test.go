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

package opsai

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/opsai/runner"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func TestWriteAIHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteAIHelp(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"Usage:", "Examples:", "help for ai", toolname.AICLI(""), toolname.AICLI("model list"), toolname.AICLI("ollama status"), toolname.AICLI("ollama setup"),
		toolname.AICLI("--model minimax"), "Subcommands:", "ask", "agent", "provider enable", "ollama status", "start", "stop", "disable",
		"--model", toolname.AICLI("connect"), toolname.AITCLI("agent"),
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestAIHelpSections(t *testing.T) {
	TestWriteAIHelpSections(t)
}

func TestLocalAgentConfig_HasCaptureAndCLI(t *testing.T) {
	exec := func(args []string, out, errOut io.Writer) error { return nil }
	cfg := withOpsDomain(runner.LocalAgentConfig(nil, 0, exec))
	if cfg.Capture == nil {
		t.Fatal("local agent missing Capture runner")
	}
	if cfg.CLI == nil {
		t.Fatal("local agent missing CLI executor")
	}
	if cfg.Domain != promptDomainOps {
		t.Fatalf("Domain=%q want %q", cfg.Domain, promptDomainOps)
	}
	if cfg.Toolbox == nil {
		t.Fatal("ops domain missing Toolbox")
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
