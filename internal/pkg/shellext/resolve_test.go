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

package shellext

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAppBin_Found(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(binDir, "foo")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho hi\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, ok := ResolveAppBin("foo")
	if !ok {
		t.Fatal("expected foo in bin")
	}
	if got != script {
		t.Fatalf("got %q want %q", got, script)
	}
}

func TestResolveAppBin_RejectsUnsafeName(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"", "a/b", `a\b`, "..", "../x"} {
		if _, ok := ResolveAppBin(name); ok {
			t.Fatalf("name %q should not resolve", name)
		}
	}
}

func TestResolveAppBin_SkipsNonExecutable(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(binDir, "ro")
	if err := os.WriteFile(script, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := ResolveAppBin("ro"); ok {
		t.Fatal("non-executable should not resolve")
	}
}
