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

package ait

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/ait/ai/ollama"
)

func TestOllamaSetupHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := ollama.WriteOllamaSetupHelp(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"Usage:", "Examples:", "help for setup", "ait ai ollama setup",
		"[y/N]", "--yes", "--no-install", "--dry-run", "--models-dir",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestAIOllamaHelpSections(t *testing.T) {
	var buf bytes.Buffer
	if err := ollama.WriteAIOllamaHelp(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"ait ai ollama", "status", "setup", "start", "stop", "disable",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestOllamaSetupDryRun(t *testing.T) {
	var out, errOut bytes.Buffer
	err := ExecuteWriters([]string{"ai", "ollama", "setup", "--dry-run"}, &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.Contains(s, "dry-run") || !strings.Contains(s, "models_dir=") {
		t.Fatalf("stdout=%q", s)
	}
}

func TestOllamaStatusSmoke(t *testing.T) {
	var out bytes.Buffer
	err := ExecuteWriters([]string{"ai", "ollama", "status"}, &out, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Ollama status:") {
		t.Fatalf("stdout=%q", out.String())
	}
}

func TestTopLevelOllamaRemoved(t *testing.T) {
	err := ExecuteWriters([]string{"ollama", "setup"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected top-level ollama to fail")
	}
}

func TestTopLevelModelsRemoved(t *testing.T) {
	err := ExecuteWriters([]string{"models"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected top-level models to fail")
	}
}
