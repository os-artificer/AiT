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
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/platform"
	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/store"
)

func TestDispatchShellCommand_RunsBin(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(binDir, "myext")
	if err := os.WriteFile(script, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	var gotPath string
	var gotArgv []string
	prev := externalRunner
	SetExternalRunner(func(_ context.Context, path string, argv []string) error {
		gotPath = path
		gotArgv = append([]string(nil), argv...)
		return nil
	})
	defer SetExternalRunner(prev)

	if err := dispatchShellCommand(context.Background(), []string{"myext", "a", "b"}); err != nil {
		t.Fatal(err)
	}
	if gotPath != script {
		t.Fatalf("path=%q want %q", gotPath, script)
	}
	if len(gotArgv) != 3 || gotArgv[0] != "myext" || gotArgv[1] != "a" || gotArgv[2] != "b" {
		t.Fatalf("argv=%v", gotArgv)
	}
}

func TestDispatchShellCommand_AitCLIBeforePATH(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatal(err)
	}

	calledCLI := false
	prevCLI := cliExecutor
	cliExecutor = func(argv []string, out, errOut io.Writer) error {
		calledCLI = true
		if len(argv) < 1 || argv[0] != "config" {
			t.Fatalf("argv=%v", argv)
		}
		return nil
	}
	defer func() { cliExecutor = prevCLI }()

	prevExt := externalRunner
	SetExternalRunner(func(context.Context, string, []string) error {
		t.Fatal("external runner should not run for ai-ssh CLI")
		return nil
	})
	defer SetExternalRunner(prevExt)

	if err := dispatchShellCommand(context.Background(), []string{"config", "view"}); err != nil {
		t.Fatal(err)
	}
	if !calledCLI {
		t.Fatal("expected CLI executor")
	}
}

func TestDispatchShellCommand_PATHFallback(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	pathExe, err := platform.LookPath("true")
	if err != nil {
		t.Skip("true not on PATH")
	}

	prevCLI := cliExecutor
	cliExecutor = func([]string, io.Writer, io.Writer) error {
		return errors.New(`unknown command "true" for "ai-ssh"`)
	}
	defer func() { cliExecutor = prevCLI }()

	var gotPath string
	prevExt := externalRunner
	SetExternalRunner(func(_ context.Context, path string, argv []string) error {
		gotPath = path
		if argv[0] != "true" {
			t.Fatalf("argv=%v", argv)
		}
		return nil
	})
	defer SetExternalRunner(prevExt)

	if err := dispatchShellCommand(context.Background(), []string{"true"}); err != nil {
		t.Fatal(err)
	}
	if gotPath != pathExe {
		t.Fatalf("path=%q want %q", gotPath, pathExe)
	}
}

func TestDispatchShellCommand_PATHSkipsCLIError(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	prevCLI := cliExecutor
	cliExecutor = func([]string, io.Writer, io.Writer) error {
		return errors.New("usage: session requires an action")
	}
	defer func() { cliExecutor = prevCLI }()

	prevExt := externalRunner
	SetExternalRunner(func(context.Context, string, []string) error {
		t.Fatal("PATH fallback must not run on CLI business errors")
		return nil
	})
	defer SetExternalRunner(prevExt)

	err := dispatchShellCommand(context.Background(), []string{"session"})
	if err == nil || !strings.Contains(err.Error(), "usage: session") {
		t.Fatalf("got err=%v", err)
	}
}

func TestDispatchShellCommand_NotFoundHint(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	prevCLI := cliExecutor
	cliExecutor = func([]string, io.Writer, io.Writer) error {
		return errors.New(`unknown command "nosuchcmd" for "ai-ssh"`)
	}
	defer func() { cliExecutor = prevCLI }()

	prevExt := externalRunner
	SetExternalRunner(func(context.Context, string, []string) error {
		t.Fatal("external runner should not run")
		return nil
	})
	defer SetExternalRunner(prevExt)

	err := dispatchShellCommand(context.Background(), []string{"nosuchcmd"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrCommandNotFound) {
		t.Fatalf("errors.Is ErrCommandNotFound: %v", err)
	}
	if !strings.Contains(err.Error(), "Try:") {
		t.Fatalf("want Try block: %q", err.Error())
	}
	if !strings.Contains(err.Error(), filepath.Join(root, "bin")) {
		t.Fatalf("want bin path in hint: %q", err.Error())
	}
}

func TestExecuteLine_BuiltinWinsOverBin(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(binDir, "list")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho bin-list\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	prevExt := externalRunner
	SetExternalRunner(func(context.Context, string, []string) error {
		t.Fatal("builtin list must not run bin extension")
		return nil
	})
	defer SetExternalRunner(prevExt)

	oldOut := shellCLIOut
	var buf bytes.Buffer
	shellCLIOut = &buf
	defer func() { shellCLIOut = oldOut }()

	st, err := store.OpenPath(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	cfg := &config.Config{}
	cfg.Session.AuthTTL = "168h"
	mgr := &session.Manager{Store: st, Config: cfg}

	exit, err := ExecuteLine(context.Background(), mgr, "list", nil)
	if err != nil || exit {
		t.Fatalf("exit=%v err=%v", exit, err)
	}
	if !strings.Contains(buf.String(), "No saved sessions.") {
		t.Fatalf("out=%q want session list", buf.String())
	}
}
