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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/platform"
	"github.com/os-artificer/AiT/internal/ssh/config/shared"
)

func TestConfigViewCmd(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	var out bytes.Buffer
	err := ExecuteWriters([]string{"config", "view"}, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "session:") {
		t.Fatalf("output=%q", out.String())
	}
}

func TestConfigViewCmd_Existing(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	custom := "session:\n  auth_ttl: 336h\n"
	path := filepath.Join(root, "config.yaml")
	if err := os.WriteFile(path, []byte(custom), 0o600); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := ExecuteWriters([]string{"config", "view"}, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if out.String() != custom {
		t.Fatalf("output=%q want %q", out.String(), custom)
	}
}

func TestConfigEditCmd(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	shared.OpenEditorFn = func(path string) error {
		if !strings.Contains(path, "config.yaml") {
			t.Fatalf("path=%q", path)
		}
		return nil
	}
	t.Cleanup(func() { shared.OpenEditorFn = platform.OpenEditor })

	if _, err := os.Stat(filepath.Join(root, "config.yaml")); !os.IsNotExist(err) {
		if err == nil {
			t.Fatal("config should not exist yet")
		}
	}

	var out bytes.Buffer
	err := ExecuteWriters([]string{"config", "edit"}, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "config OK:") {
		t.Fatalf("output=%q", out.String())
	}
}

func TestConfigEditCmd_InvalidYAML(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	var initOut bytes.Buffer
	if err := ExecuteWriters([]string{"config", "init"}, &initOut, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(root, "config.yaml")
	shared.OpenEditorFn = func(p string) error {
		if p != path {
			t.Fatalf("path=%q", p)
		}
		return os.WriteFile(p, []byte("session: [\n"), 0o600)
	}
	t.Cleanup(func() { shared.OpenEditorFn = platform.OpenEditor })

	var errOut bytes.Buffer
	err := ExecuteWriters([]string{"config", "edit"}, &bytes.Buffer{}, &errOut)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(errOut.String(), "validation failed") {
		t.Fatalf("stderr=%q", errOut.String())
	}
}

func TestConfigValidateCmd_OK(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	var initOut bytes.Buffer
	if err := ExecuteWriters([]string{"config", "init"}, &initOut, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := ExecuteWriters([]string{"config", "validate"}, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "config OK:") {
		t.Fatalf("output=%q", out.String())
	}
}

func TestConfigValidateCmd_MissingFile(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	var errOut bytes.Buffer
	err := ExecuteWriters([]string{"config", "validate"}, &bytes.Buffer{}, &errOut)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Try:") {
		t.Fatalf("err=%v", err)
	}
}

func TestConfigValidateCmd_BadYAML(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	path := filepath.Join(root, "config.yaml")
	if err := os.WriteFile(path, []byte("bad: [\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var errOut bytes.Buffer
	err := ExecuteWriters([]string{"config", "validate"}, &bytes.Buffer{}, &errOut)
	if err == nil {
		t.Fatal("expected validation failure")
	}
	s := errOut.String()
	if !strings.Contains(s, "yaml") && !strings.Contains(s, "validation failed") {
		t.Fatalf("stderr=%q", s)
	}
}

func TestConfigValidateCmd_BadTTL(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	yaml := "session:\n  auth_ttl: not-a-duration\nai:\n  default: deepseek\n  providers:\n    deepseek:\n      model: deepseek-chat\n"
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	var errOut bytes.Buffer
	err := ExecuteWriters([]string{"config", "validate"}, &bytes.Buffer{}, &errOut)
	if err == nil {
		t.Fatal("expected failure")
	}
	if !strings.Contains(errOut.String(), "session.auth_ttl") {
		t.Fatalf("stderr=%q", errOut.String())
	}
}

func TestConfigValidateCmd_BadURL(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	yaml := "session:\n  auth_ttl: 168h\nai:\n  default: deepseek\n  providers:\n    deepseek:\n      model: deepseek-chat\n    minimax:\n      base_url: https://evil.example/v1\n      model: MiniMax-M2.5\n"
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	var errOut bytes.Buffer
	err := ExecuteWriters([]string{"config", "validate"}, &bytes.Buffer{}, &errOut)
	if err == nil {
		t.Fatal("expected failure")
	}
	s := errOut.String()
	if !strings.Contains(s, "ai.providers.minimax.base_url") {
		t.Fatalf("stderr=%q", s)
	}
}
