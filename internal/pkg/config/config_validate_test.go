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

package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidate_OK(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	if _, err := EnsureFile(false); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(root, "config.yaml")
	report, err := Validate(path)
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK() {
		t.Fatalf("expected OK, issues=%v", report.Issues)
	}
}

func TestValidate_MissingFile(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	_, err := Validate(filepath.Join(root, "config.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "Try:") {
		t.Fatalf("expected Try block: %v", err)
	}
}

func TestValidate_BadYAML(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	path := filepath.Join(root, "config.yaml")
	if err := os.WriteFile(path, []byte("session: [\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	report, err := Validate(path)
	if err != nil {
		t.Fatal(err)
	}
	if report.OK() {
		t.Fatal("expected validation failure")
	}
	if !strings.Contains(report.Error(), "yaml") {
		t.Fatalf("issues=%q", report.Error())
	}
}

func TestValidate_BadTTL(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	path := filepath.Join(root, "config.yaml")
	yaml := "session:\n  auth_ttl: not-a-duration\nai:\n  default: deepseek\n  providers:\n    deepseek:\n      model: deepseek-chat\n"
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	report, err := Validate(path)
	if err != nil {
		t.Fatal(err)
	}
	if report.OK() {
		t.Fatal("expected failure")
	}
	if !strings.Contains(report.Error(), "session.auth_ttl") {
		t.Fatalf("issues=%q", report.Error())
	}
}

// AG-12: a malformed ai.agent.command_timeout is reported with its field path.
func TestValidate_BadAgentCommandTimeout(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	path := filepath.Join(root, "config.yaml")
	yaml := "ai:\n  agent:\n    command_timeout: nope\n  default: deepseek\n  providers:\n    deepseek:\n      model: deepseek-chat\n"
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	report, err := Validate(path)
	if err != nil {
		t.Fatal(err)
	}
	if report.OK() {
		t.Fatal("expected failure")
	}
	if !strings.Contains(report.Error(), "ai.agent.command_timeout") {
		t.Fatalf("issues=%q", report.Error())
	}
}

func TestValidate_BadMinimaxURL(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	path := filepath.Join(root, "config.yaml")
	yaml := "session:\n  auth_ttl: 168h\nai:\n  default: deepseek\n  providers:\n    deepseek:\n      model: deepseek-chat\n    minimax:\n      base_url: https://evil.example/v1\n      model: MiniMax-M2.5\n"
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	report, err := Validate(path)
	if err != nil {
		t.Fatal(err)
	}
	if report.OK() {
		t.Fatal("expected failure")
	}
	s := report.Error()
	if !strings.Contains(s, "ai.providers.minimax.base_url") || !strings.Contains(s, "api.minimaxi.com") {
		t.Fatalf("issues=%q", s)
	}
}

// TestValidate_BadRemotePromptPath (W9) rejects invalid session.remote_prompt.path.
func TestValidate_BadRemotePromptPath(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	path := filepath.Join(root, "config.yaml")
	yaml := "session:\n  auth_ttl: 168h\n  remote_prompt:\n    path: full\nai:\n  default: deepseek\n  providers:\n    deepseek:\n      model: deepseek-chat\n"
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	report, err := Validate(path)
	if err != nil {
		t.Fatal(err)
	}
	if report.OK() {
		t.Fatal("expected failure")
	}
	if !strings.Contains(report.Error(), "session.remote_prompt.path") {
		t.Fatalf("issues=%q", report.Error())
	}
}

func TestFormatValidation_OK(t *testing.T) {
	var buf bytes.Buffer
	FormatValidation(&buf, &ValidationReport{Path: "/tmp/config.yaml"})
	if !strings.Contains(buf.String(), "config OK:") {
		t.Fatalf("out=%q", buf.String())
	}
}

func TestReadRawFile(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	if _, err := EnsureFile(false); err != nil {
		t.Fatal(err)
	}

	path, b, err := ReadRawFile()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "session:") {
		t.Fatalf("content=%q", b)
	}
	if !strings.HasSuffix(path, "config.yaml") {
		t.Fatalf("path=%q", path)
	}
}
