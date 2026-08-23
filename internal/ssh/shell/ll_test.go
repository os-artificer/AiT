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

package shell

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteLine_LLWinsOverBin(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(binDir, "ll")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho from-bin\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	prevExt := externalRunner
	SetExternalRunner(func(context.Context, string, []string) error {
		t.Fatal("bin ll must not run when shell builtin handles ll")
		return nil
	})
	defer SetExternalRunner(prevExt)

	oldOut := shellCLIOut
	var buf bytes.Buffer
	shellCLIOut = &buf
	defer func() { shellCLIOut = oldOut }()

	exit, err := ExecuteLine(context.Background(), nil, "ll", nil)
	if err != nil || exit {
		t.Fatalf("exit=%v err=%v", exit, err)
	}
	if !strings.Contains(buf.String(), "total ") {
		t.Fatalf("want dirlist output not bin: %q", buf.String())
	}
}
