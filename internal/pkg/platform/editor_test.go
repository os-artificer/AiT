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

package platform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveEditor_Env(t *testing.T) {
	t.Setenv(envVisual, "")
	t.Setenv(envEditor, "")

	t.Setenv(envVisual, "myvisual")
	got, err := ResolveEditor()
	if err != nil {
		t.Fatal(err)
	}
	if got != "myvisual" {
		t.Fatalf("got %q", got)
	}

	t.Setenv(envVisual, "")
	t.Setenv(envEditor, "myeditor")
	got, err = ResolveEditor()
	if err != nil {
		t.Fatal(err)
	}
	if got != "myeditor" {
		t.Fatalf("got %q", got)
	}
}

func TestShellQuote(t *testing.T) {
	in := "/Library/Application Support/ait/config.yaml"
	got := shellQuote(in)
	if !strings.Contains(got, "Application Support") {
		t.Fatalf("quote=%q", got)
	}
	if !strings.HasPrefix(got, "'") || !strings.HasSuffix(got, "'") {
		t.Fatalf("quote=%q", got)
	}
}

func TestOpenEditor_Unix(t *testing.T) {
	if os.Getenv("GOOS") == "windows" {
		t.Skip("unix-only")
	}

	root := t.TempDir()
	marker := filepath.Join(root, "opened.txt")
	script := filepath.Join(root, "fake-editor.sh")
	content := "#!/bin/sh\necho \"$@\" > \"" + marker + "\"\n"
	if err := os.WriteFile(script, []byte(content), 0o700); err != nil {
		t.Fatal(err)
	}

	t.Setenv(envVisual, "")
	t.Setenv(envEditor, script)

	cfgPath := filepath.Join(root, "Application Support", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfgPath, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := OpenEditor(cfgPath); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "Application Support") {
		t.Fatalf("marker=%q", b)
	}
}
