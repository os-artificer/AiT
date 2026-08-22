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

package builtin

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShellLL_CurrentDir(t *testing.T) {
	root := t.TempDir()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "sample.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	env := Env{Out: &buf}
	_, _, err := Try(context.Background(), env, []string{CmdLL})
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "total ") {
		t.Fatalf("want total line: %q", out)
	}
	if !strings.Contains(out, "sample.txt") {
		t.Fatalf("want file entry: %q", out)
	}
}

func TestShellLL_TargetPath(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "inner")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	env := Env{Out: &buf}
	_, _, err := Try(context.Background(), env, []string{CmdLL, sub})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), " f\n") && !strings.Contains(buf.String(), " f\r\n") {
		t.Fatalf("want f entry: %q", buf.String())
	}
}

func TestShellLL_NotFoundHint(t *testing.T) {
	env := Env{Out: ioDiscard{}}
	_, _, err := Try(context.Background(), env, []string{CmdLL, filepath.Join(t.TempDir(), "missing-dir")})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Try:") {
		t.Fatalf("want Try block: %q", err.Error())
	}
}

func TestShellLL_TooManyArgs(t *testing.T) {
	env := Env{Out: ioDiscard{}}
	_, _, err := Try(context.Background(), env, []string{CmdLL, "a", "b"})
	if err == nil || !strings.Contains(err.Error(), "usage: ll") {
		t.Fatalf("got err=%v", err)
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
