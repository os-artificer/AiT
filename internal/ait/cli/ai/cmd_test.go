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

package ai

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/ait/helpdoc"
)

func TestAIHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := helpdoc.WriteAIHelp(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"Usage:", "Examples:", "help for ai", "ait ai", "ait ai model list", "ait ai ollama status", "ait ai ollama setup",
		"ait ai --model minimax", "Subcommands:", "ask", "provider enable", "ollama status", "start", "stop", "disable",
		"--model", "deepseek|minimax", "run <command>", "mode ask|agent", "approve each|session", "trust", "model list", "!ait ai",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestLocalAIModeOptions_HasCaptureAndCLI(t *testing.T) {
	exec := func(args []string, out, errOut io.Writer) error { return nil }
	opt := localAIModeOptions(nil, 0, exec)
	if opt.Capture == nil {
		t.Fatal("local AI mode missing Capture runner")
	}
	if opt.CLI == nil {
		t.Fatal("local AI mode missing CLI executor")
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
