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
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/store"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

func TestRunCLIRequiresExecutor(t *testing.T) {
	old := cliExecutor
	cliExecutor = nil
	defer func() { cliExecutor = old }()

	err := runCLI([]string{"list"})
	if err == nil || err.Error() != "AiT: shell CLI executor not configured" {
		t.Fatalf("got err=%v", err)
	}
}

func TestRunCLIWithMockExecutor(t *testing.T) {
	SetCLIExecutor(func(argv []string, out, errOut io.Writer) error {
		return fmt.Errorf("mock: %v", argv)
	})

	err := runCLI([]string{"list"})
	if err == nil || !strings.Contains(err.Error(), "mock") {
		t.Fatalf("got err=%v", err)
	}
}

func TestExecuteLineStripsTrailingSemicolon(t *testing.T) {
	words, err := parseShellLine(normalizeShellLine("list;"))
	if err != nil {
		t.Fatal(err)
	}
	if len(words) != 1 || words[0] != "list" {
		t.Errorf("words = %v want [list]", words)
	}

	words, err = parseShellLine(normalizeShellLine("clear;"))
	if err != nil {
		t.Fatal(err)
	}
	if len(words) != 1 || words[0] != replCmdClear {
		t.Errorf("words = %v want [clear]", words)
	}
}

func TestExecuteLineClear(t *testing.T) {
	oldOut := shellCLIOut
	var buf bytes.Buffer
	shellCLIOut = &buf
	defer func() { shellCLIOut = oldOut }()

	called := false
	old := cliExecutor
	cliExecutor = func([]string, io.Writer, io.Writer) error {
		called = true
		return nil
	}
	defer func() { cliExecutor = old }()

	exit, err := ExecuteLine(context.Background(), nil, "clear", nil)
	if err != nil || exit {
		t.Fatalf("exit=%v err=%v", exit, err)
	}
	if called {
		t.Fatal("clear must not invoke cliExecutor")
	}
	if !strings.Contains(buf.String(), terminal.ClearScreenSeq()) {
		t.Errorf("output = %q, want clear screen seq", buf.String())
	}
}

func TestExecuteLineList(t *testing.T) {
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
		t.Fatalf("out=%q", buf.String())
	}
}

func TestExecuteLineHistory_Shell(t *testing.T) {
	oldOut := shellCLIOut
	var buf bytes.Buffer
	shellCLIOut = &buf
	defer func() { shellCLIOut = oldOut }()

	hist := terminal.NewLineHistory()
	hist.Add("list")
	hist.Add("help")

	exit, err := ExecuteLine(context.Background(), nil, "history", hist)
	if err != nil || exit {
		t.Fatalf("exit=%v err=%v", exit, err)
	}
	if !strings.Contains(buf.String(), "      1  list") || !strings.Contains(buf.String(), "      2  help") {
		t.Fatalf("history output = %q", buf.String())
	}

	buf.Reset()
	exit, err = ExecuteLine(context.Background(), nil, "history 1", hist)
	if err != nil || exit {
		t.Fatalf("exit=%v err=%v", exit, err)
	}
	if strings.Contains(buf.String(), "list") || !strings.Contains(buf.String(), "help") {
		t.Fatalf("history 1 output = %q", buf.String())
	}
}

func TestShellCLIErrDefault(t *testing.T) {
	if shellCLIErr != io.Discard {
		t.Fatalf("shellCLIErr = %v, want io.Discard before SetREPLWriters", shellCLIErr)
	}
}
