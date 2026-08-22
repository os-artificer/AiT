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

package ait_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/ait"
	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/pluginbin"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func TestExecuteWriters_Help(t *testing.T) {
	resetPlugins(t)
	var out bytes.Buffer
	if err := ait.ExecuteWriters([]string{"-h"}, &out, io.Discard); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.Contains(s, "Usage:") || !strings.Contains(s, toolname.AIT) {
		t.Fatalf("help = %q", s)
	}
	if strings.Contains(s, "OpenSSH") {
		t.Fatal("root help must not include SSH connect section")
	}
}

func TestExecuteWriters_Version(t *testing.T) {
	var out bytes.Buffer
	if err := ait.ExecuteWriters([]string{"-V"}, &out, io.Discard); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.String()) == "" {
		t.Fatal("empty version")
	}
}

func TestExecuteWriters_InteractiveOnly(t *testing.T) {
	err := ait.ExecuteWriters([]string{"aissh", "host"}, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "interactive only") {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(err.Error(), "Try:") {
		t.Fatalf("missing Try: %v", err)
	}
}

func TestExecuteWriters_REPLExit(t *testing.T) {
	resetPlugins(t)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		_, _ = io.WriteString(w, "exit\n")
		_ = w.Close()
	}()
	var out bytes.Buffer
	if err := ait.RunREPL(context.Background(), r, &out, io.Discard); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteLine_VersionBuiltin(t *testing.T) {
	var out bytes.Buffer
	exit, err := ait.ExecuteLine(context.Background(), "version", nil, &out, io.Discard)
	if err != nil || exit {
		t.Fatalf("exit=%v err=%v", exit, err)
	}
	if strings.TrimSpace(out.String()) == "" {
		t.Fatal("empty version")
	}
}

func TestExecuteLine_HelpPlugin(t *testing.T) {
	resetPlugins(t)
	installBuiltinPlugin(t, "aissh")

	var out bytes.Buffer
	exit, err := ait.ExecuteLine(context.Background(), "help aissh", nil, &out, io.Discard)
	if err != nil || exit {
		t.Fatalf("exit=%v err=%v", exit, err)
	}
	if !strings.Contains(out.String(), "--save") {
		t.Fatalf("help = %q", out.String())
	}
}

func TestExecuteLine_HelpUnknown(t *testing.T) {
	resetPlugins(t)
	_, err := ait.ExecuteLine(context.Background(), "help nosuch", nil, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unknown help topic") {
		t.Fatalf("err = %v", err)
	}
}

func TestExecuteLine_UnknownCommand(t *testing.T) {
	resetPlugins(t)
	_, err := ait.ExecuteLine(context.Background(), "nosuch", nil, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("err = %v", err)
	}
}

func TestExecuteLine_RunPluginSwallowsExitError(t *testing.T) {
	resetPlugins(t)
	installBuiltinPlugin(t, "aissh")

	t.Cleanup(func() { pluginbin.SetRunner(nil) })
	pluginbin.SetRunner(func(context.Context, string, []string) error {
		return &exec.ExitError{}
	})

	exit, err := ait.ExecuteLine(context.Background(), "aissh --help", nil, io.Discard, io.Discard)
	if exit || err != nil {
		t.Fatalf("exit=%v err=%v", exit, err)
	}
}

func TestExecuteLine_JoinedExitErrorNotSwallowed(t *testing.T) {
	resetPlugins(t)
	installBuiltinPlugin(t, "aissh")

	t.Cleanup(func() { pluginbin.SetRunner(nil) })
	pluginbin.SetRunner(func(context.Context, string, []string) error {
		return errors.Join(fmt.Errorf("terminal: make raw"), &exec.ExitError{})
	})

	_, err := ait.ExecuteLine(context.Background(), "aissh --help", nil, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected MakeRaw+ExitError to surface")
	}
	if !strings.Contains(err.Error(), "terminal") {
		t.Fatalf("err = %v", err)
	}
}

func TestExecuteLine_UnclosedQuoteHint(t *testing.T) {
	_, err := ait.ExecuteLine(context.Background(), "aissh 'open", nil, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unclosed quote") || !strings.Contains(err.Error(), "Try:") {
		t.Fatalf("err = %v", err)
	}
}

func TestExecuteLine_DoubleQuotedArg(t *testing.T) {
	resetPlugins(t)
	installBuiltinPlugin(t, "aissh")

	t.Cleanup(func() { pluginbin.SetRunner(nil) })
	var gotArgv []string
	pluginbin.SetRunner(func(_ context.Context, _ string, argv []string) error {
		gotArgv = append([]string(nil), argv...)
		return nil
	})

	_, err := ait.ExecuteLine(context.Background(), `aissh "user@host.example.com"`, nil, io.Discard, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotArgv) < 2 || gotArgv[1] != "user@host.example.com" {
		t.Fatalf("argv = %#v", gotArgv)
	}
}

func TestExecuteLine_WrappedExitErrorNotSwallowed(t *testing.T) {
	resetPlugins(t)
	installBuiltinPlugin(t, "aissh")

	t.Cleanup(func() { pluginbin.SetRunner(nil) })
	pluginbin.SetRunner(func(context.Context, string, []string) error {
		return fmt.Errorf("wrap: %w", &exec.ExitError{})
	})

	_, err := ait.ExecuteLine(context.Background(), "aissh --help", nil, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected wrapped ExitError to surface")
	}
}

func TestWriteRootHelp_NoPlugins(t *testing.T) {
	resetPlugins(t)
	var out bytes.Buffer
	if err := ait.WriteRootHelp(&out); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.Contains(s, "no plugins found") || !strings.Contains(s, "Try:") {
		t.Fatalf("help = %q", s)
	}
	if !strings.Contains(s, "make install") {
		t.Fatalf("missing make install Try: %q", s)
	}
	if !strings.Contains(s, "help") || !strings.Contains(s, "exit") {
		t.Fatalf("missing reserved commands: %q", s)
	}
}

func TestWriteRootHelp_PluginsBeforeBuiltins(t *testing.T) {
	resetPlugins(t)
	installBuiltinPlugin(t, "aissh")

	var out bytes.Buffer
	if err := ait.WriteRootHelp(&out); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	cmdSSH := strings.Index(s, "  aissh ")
	cmdHelp := strings.Index(s, "  help ")
	if cmdSSH < 0 || cmdHelp < 0 {
		t.Fatalf("missing command rows in: %q", s)
	}
	if cmdSSH > cmdHelp {
		t.Fatalf("plugins must list before builtins: aissh@%d help@%d\n%s", cmdSSH, cmdHelp, s)
	}
}

func TestExecuteLine_RunPluginUsesPath(t *testing.T) {
	resetPlugins(t)
	path := installBuiltinPlugin(t, "ailog")

	t.Cleanup(func() { pluginbin.SetRunner(nil) })
	var gotPath string
	pluginbin.SetRunner(func(_ context.Context, p string, argv []string) error {
		gotPath = p
		if argv[0] != "ailog" {
			t.Fatalf("argv[0]=%q", argv[0])
		}
		return nil
	})

	_, err := ait.ExecuteLine(context.Background(), "ailog sources", nil, io.Discard, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != path {
		t.Fatalf("path = %q want %q", gotPath, path)
	}
}

func TestHistoryPath_NotReplHistory(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AIT_HOME", dir)
	var out bytes.Buffer
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		_, _ = io.WriteString(w, "exit\n")
		_ = w.Close()
	}()
	if err := ait.RunREPL(context.Background(), r, &out, io.Discard); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "repl_history")); err == nil {
		t.Fatal("must not use repl_history")
	}
}

func TestToolnameAIT(t *testing.T) {
	if toolname.AIT != "ait" {
		t.Fatalf("AIT = %q", toolname.AIT)
	}
}

func TestErrhintPreservesIs(t *testing.T) {
	cause := errors.New("interactive only")
	err := errhint.WithTry(cause, "x", "ait")
	if !errors.Is(err, cause) {
		t.Fatal("errors.Is broken")
	}
}

func resetPlugins(t *testing.T) {
	t.Helper()
	aitplugin.ResetDiscoverForTest()
	aitplugin.RestoreDiscoverHooksForTest()
	t.Cleanup(func() {
		aitplugin.ResetDiscoverForTest()
		aitplugin.RestoreDiscoverHooksForTest()
	})
	aitplugin.SetLookPathForTest(func(string) (string, error) {
		return "", os.ErrNotExist
	})
	aitplugin.SetWarnWriterForTest(io.Discard)
	aitplugin.SetInstallDirsForTest([]string{t.TempDir()})
}

func installBuiltinPlugin(t *testing.T, binary string) string {
	t.Helper()
	root := t.TempDir()
	plugDir := filepath.Join(root, "plugins", binary)
	if err := os.MkdirAll(plugDir, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := aitplugin.BuiltinJSON(binary)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plugDir, "plugin.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	name := binary
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	path := filepath.Join(plugDir, name)
	mode := os.FileMode(0o755)
	if runtime.GOOS == "windows" {
		mode = 0o666
	}
	if err := os.WriteFile(path, []byte("x"), mode); err != nil {
		t.Fatal(err)
	}
	aitplugin.SetInstallDirsForTest([]string{root})
	aitplugin.ResetDiscoverForTest()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}
