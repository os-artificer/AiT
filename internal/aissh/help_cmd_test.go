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
	"strings"
	"testing"

	_ "github.com/os-artificer/AiT/internal/aissh/registry/all"

	"github.com/os-artificer/AiT/internal/aissh/get"
	"github.com/os-artificer/AiT/internal/aissh/put"
	"github.com/os-artificer/AiT/internal/aissh/registry"
)

func testDeps() registry.Deps {
	return registry.Deps{RunWithMgr: runWithManager}
}

func TestGetCmdHelpRunE(t *testing.T) {
	cmd := get.New(testDeps().RunWithMgr)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if err := cmd.RunE(cmd, []string{"--help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "OpenSSH-compatible flags:") {
		t.Fatalf("get --help: %s", buf.String())
	}
}

func TestGetCmdHelpRunEPosArg(t *testing.T) {
	cmd := get.New(testDeps().RunWithMgr)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if err := cmd.RunE(cmd, []string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "OpenSSH-compatible flags:") {
		t.Fatalf("get help: %s", buf.String())
	}
}

func TestPutCmdHelpRunE(t *testing.T) {
	cmd := put.New(testDeps().RunWithMgr)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if err := cmd.RunE(cmd, []string{"--help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "help for put") {
		t.Fatalf("put --help: %s", buf.String())
	}
}

func TestPutCmdHelpRunEPosArg(t *testing.T) {
	cmd := put.New(testDeps().RunWithMgr)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if err := cmd.RunE(cmd, []string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "help for put") {
		t.Fatalf("put help: %s", buf.String())
	}
}

func helpExecuteMustContain(t *testing.T, args []string, wants ...string) {
	t.Helper()
	root := newRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Fatalf("%v help missing %q:\n%s", args, want, out)
		}
	}
}

func TestRootCmdHelpExecute(t *testing.T) {
	var buf bytes.Buffer
	if err := ExecuteWriters([]string{"--help"}, &buf, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"Usage:", "Examples:", "help for aitssh", "Commands:", "session", "OpenSSH-compatible flags:"} {
		if !strings.Contains(out, want) {
			t.Fatalf("root help missing %q:\n%s", want, out)
		}
	}
	for _, gone := range []string{"  list       ", "  load       ", "  suspend    ", "  login       "} {
		if strings.Contains(out, gone) {
			t.Fatalf("root help still lists removed command %q", gone)
		}
	}
}

func TestShellCmdHelpExecute(t *testing.T) {
	helpExecuteMustContain(t, []string{"shell", "--help"}, "Usage:", "Examples:", "help for shell")
}

func TestShellCmdHelpPosArgExecute(t *testing.T) {
	helpExecuteMustContain(t, []string{"shell", "help"}, "Usage:", "Examples:", "help for shell")
}

func TestConfigCmdHelpExecute(t *testing.T) {
	helpExecuteMustContain(t, []string{"config", "--help"}, "Usage:", "Examples:", "help for config", "init")
	helpExecuteMustContain(t, []string{"config", "set", "--help"}, "help for set")
	helpExecuteMustContain(t, []string{"config", "get", "--help"}, "help for get")
	helpExecuteMustContain(t, []string{"config", "init", "--help"}, "help for init")
	helpExecuteMustContain(t, []string{"config", "view", "--help"}, "help for view")
	helpExecuteMustContain(t, []string{"config", "edit", "--help"}, "help for edit")
	helpExecuteMustContain(t, []string{"config", "validate", "--help"}, "help for validate")
}

func TestSessionCmdHelpExecute(t *testing.T) {
	helpExecuteMustContain(t, []string{"session", "--help"},
		"Usage:", "Examples:", "help for session", "--list", "--load", "--switch", "--delete", "--tmux", "--force")
}

func TestSessionCmdHelpPosArgExecute(t *testing.T) {
	helpExecuteMustContain(t, []string{"session", "help"},
		"Usage:", "Examples:", "help for session", "--list", "--load", "--switch", "--delete", "--tmux", "--force")
}

func TestVersionCmdHelpExecute(t *testing.T) {
	helpExecuteMustContain(t, []string{"version", "--help"}, "Usage:", "help for version")
}

func TestVersionCmdHelpPosArgExecute(t *testing.T) {
	helpExecuteMustContain(t, []string{"version", "help"}, "Usage:", "help for version")
}
