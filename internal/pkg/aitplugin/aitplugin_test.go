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

package aitplugin_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
)

func TestValidate_OK(t *testing.T) {
	m := aitplugin.Manifest{
		APIVersion: 1,
		Name:       "aissh",
		Binary:     "aissh",
		Short:      "SSH",
	}
	if err := aitplugin.Validate(&m); err != nil {
		t.Fatal(err)
	}
	if m.Order != 100 {
		t.Fatalf("Order = %d, want 100", m.Order)
	}
}

func TestValidate_ReservedName(t *testing.T) {
	m := aitplugin.Manifest{APIVersion: 1, Name: "help", Binary: "x", Short: "s"}
	if err := aitplugin.Validate(&m); err == nil {
		t.Fatal("expected reserved name error")
	}
}

func TestValidate_BadAPIVersion(t *testing.T) {
	m := aitplugin.Manifest{APIVersion: 2, Name: "aissh", Binary: "aissh", Short: "s"}
	if err := aitplugin.Validate(&m); err == nil {
		t.Fatal("expected apiVersion error")
	}
}

func TestValidate_UnsafeBinary(t *testing.T) {
	m := aitplugin.Manifest{APIVersion: 1, Name: "aissh", Binary: "../aissh", Short: "s"}
	if err := aitplugin.Validate(&m); err == nil {
		t.Fatal("expected unsafe binary error")
	}
}

func TestWriteManifest_Builtin(t *testing.T) {
	var buf bytes.Buffer
	if err := aitplugin.WriteManifest(&buf, "aissh"); err != nil {
		t.Fatal(err)
	}
	m, err := aitplugin.LoadBytes(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != "aissh" || m.Binary != "aissh" {
		t.Fatalf("manifest = %+v", m)
	}
}

func TestWriteHelp_IncludesFlags(t *testing.T) {
	data, err := aitplugin.BuiltinJSON("aissh")
	if err != nil {
		t.Fatal(err)
	}
	m, err := aitplugin.LoadBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := aitplugin.WriteHelp(&buf, m); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Flags:") || !strings.Contains(out, "--save") {
		t.Fatalf("help missing flags: %q", out)
	}
}

func TestDiscover_PluginJSON(t *testing.T) {
	resetDiscover(t)

	root := t.TempDir()
	plugDir := filepath.Join(root, "plugins", "aissh")
	if err := os.MkdirAll(plugDir, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := aitplugin.BuiltinJSON("aissh")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plugDir, "plugin.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	exeName := "aissh"
	if runtime.GOOS == "windows" {
		exeName = "aissh.exe"
	}
	exePath := filepath.Join(plugDir, exeName)
	writeFakeExe(t, exePath)

	aitplugin.SetInstallDirsForTest([]string{root})
	aitplugin.SetLookPathForTest(func(string) (string, error) {
		return "", os.ErrNotExist
	})
	aitplugin.SetWarnWriterForTest(ioDiscard{})

	plugins := aitplugin.Discover(context.Background())
	if len(plugins) != 1 {
		t.Fatalf("len = %d, want 1; %+v", len(plugins), plugins)
	}
	if plugins[0].Manifest.Name != "aissh" {
		t.Fatalf("name = %q", plugins[0].Manifest.Name)
	}
	if plugins[0].Path == "" {
		t.Fatal("empty path")
	}
}

func TestDiscover_JSONWithoutBinarySkipped(t *testing.T) {
	resetDiscover(t)

	root := t.TempDir()
	plugDir := filepath.Join(root, "plugins", "aissh")
	if err := os.MkdirAll(plugDir, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := aitplugin.BuiltinJSON("aissh")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plugDir, "plugin.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	var warn bytes.Buffer
	aitplugin.SetInstallDirsForTest([]string{root})
	aitplugin.SetLookPathForTest(func(string) (string, error) {
		return "", os.ErrNotExist
	})
	aitplugin.SetWarnWriterForTest(&warn)

	plugins := aitplugin.Discover(context.Background())
	if len(plugins) != 0 {
		t.Fatalf("want empty, got %+v", plugins)
	}
	if !strings.Contains(warn.String(), "not found") {
		t.Fatalf("warn = %q", warn.String())
	}
}

func TestDiscover_ScatterFileIgnored(t *testing.T) {
	resetDiscover(t)

	root := t.TempDir()
	pluginsRoot := filepath.Join(root, "plugins")
	if err := os.MkdirAll(pluginsRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := aitplugin.BuiltinJSON("aissh")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginsRoot, "plugin.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	aitplugin.SetInstallDirsForTest([]string{root})
	aitplugin.SetLookPathForTest(func(string) (string, error) {
		return "", os.ErrNotExist
	})
	aitplugin.SetWarnWriterForTest(ioDiscard{})

	if got := aitplugin.Discover(context.Background()); len(got) != 0 {
		t.Fatalf("scatter file must be ignored: %+v", got)
	}
}

func TestDiscover_ManifestFallback(t *testing.T) {
	resetDiscover(t)

	root := t.TempDir()
	plugDir := filepath.Join(root, "plugins", "ailog")
	if err := os.MkdirAll(plugDir, 0o755); err != nil {
		t.Fatal(err)
	}
	exeName := "ailog"
	if runtime.GOOS == "windows" {
		exeName = "ailog.exe"
	}
	exePath := filepath.Join(plugDir, exeName)
	writeFakeExe(t, exePath)

	data, err := aitplugin.BuiltinJSON("ailog")
	if err != nil {
		t.Fatal(err)
	}

	aitplugin.SetInstallDirsForTest([]string{root})
	aitplugin.SetLookPathForTest(func(string) (string, error) {
		return "", os.ErrNotExist
	})
	aitplugin.SetRunManifestForTest(func(context.Context, string) ([]byte, error) {
		return data, nil
	})
	aitplugin.SetWarnWriterForTest(ioDiscard{})

	plugins := aitplugin.Discover(context.Background())
	if len(plugins) != 1 || plugins[0].Manifest.Name != "ailog" {
		t.Fatalf("plugins = %+v", plugins)
	}
}

func TestDiscover_ManifestFallback_RejectsBinaryMismatch(t *testing.T) {
	resetDiscover(t)

	root := t.TempDir()
	plugDir := filepath.Join(root, "plugins", "custom")
	if err := os.MkdirAll(plugDir, 0o755); err != nil {
		t.Fatal(err)
	}
	exeName := "custom"
	if runtime.GOOS == "windows" {
		exeName = "custom.exe"
	}
	writeFakeExe(t, filepath.Join(plugDir, exeName))

	data, err := aitplugin.BuiltinJSON("ailog")
	if err != nil {
		t.Fatal(err)
	}

	aitplugin.SetInstallDirsForTest([]string{root})
	aitplugin.SetLookPathForTest(func(string) (string, error) {
		return "", os.ErrNotExist
	})
	aitplugin.SetRunManifestForTest(func(context.Context, string) ([]byte, error) {
		return data, nil
	})
	aitplugin.SetWarnWriterForTest(ioDiscard{})

	plugins := aitplugin.Discover(context.Background())
	if len(plugins) != 0 {
		t.Fatalf("expected skip on binary mismatch, got %+v", plugins)
	}
}

func TestDiscover_PATHFill_RejectsNameMismatch(t *testing.T) {
	resetDiscover(t)

	root := t.TempDir()
	exePath := filepath.Join(root, "aiimg")
	if runtime.GOOS == "windows" {
		exePath += ".exe"
	}
	writeFakeExe(t, exePath)

	// Wrong short name / binary in returned JSON.
	data := []byte(`{"apiVersion":1,"name":"evil","binary":"aiimg","short":"x"}`)

	aitplugin.SetInstallDirsForTest([]string{t.TempDir()})
	aitplugin.SetLookPathForTest(func(name string) (string, error) {
		if name == "aiimg" {
			return exePath, nil
		}
		return "", os.ErrNotExist
	})
	aitplugin.SetRunManifestForTest(func(context.Context, string) ([]byte, error) {
		return data, nil
	})
	aitplugin.SetWarnWriterForTest(ioDiscard{})

	plugins := aitplugin.Discover(context.Background())
	for _, p := range plugins {
		if p.Manifest.Name == "evil" || p.Manifest.Name == "aiimg" {
			t.Fatalf("mismatched PATH fill should skip aiimg, got %+v", plugins)
		}
	}
}

func TestDiscover_PATHFill(t *testing.T) {
	resetDiscover(t)

	root := t.TempDir()
	exePath := filepath.Join(root, "aiimg")
	if runtime.GOOS == "windows" {
		exePath += ".exe"
	}
	writeFakeExe(t, exePath)

	data, err := aitplugin.BuiltinJSON("aiimg")
	if err != nil {
		t.Fatal(err)
	}

	aitplugin.SetInstallDirsForTest([]string{t.TempDir()})
	aitplugin.SetLookPathForTest(func(name string) (string, error) {
		if name == "aiimg" {
			return exePath, nil
		}
		return "", os.ErrNotExist
	})
	aitplugin.SetRunManifestForTest(func(_ context.Context, path string) ([]byte, error) {
		if path != exePath {
			t.Fatalf("path = %q", path)
		}
		return data, nil
	})
	aitplugin.SetWarnWriterForTest(ioDiscard{})

	plugins := aitplugin.Discover(context.Background())
	if len(plugins) != 1 || plugins[0].Manifest.Name != "aiimg" {
		t.Fatalf("plugins = %+v", plugins)
	}
	if plugins[0].Path != exePath {
		t.Fatalf("Path = %q", plugins[0].Path)
	}
}

func TestDiscover_OnceCached(t *testing.T) {
	resetDiscover(t)

	calls := 0
	aitplugin.SetInstallDirsForTest([]string{t.TempDir()})
	aitplugin.SetLookPathForTest(func(string) (string, error) {
		calls++
		return "", os.ErrNotExist
	})
	aitplugin.SetWarnWriterForTest(ioDiscard{})

	_ = aitplugin.Discover(context.Background())
	_ = aitplugin.Discover(context.Background())
	if calls != 4 {
		// four builtins looked up once
		t.Fatalf("LookPath calls = %d, want 4", calls)
	}
}

func TestDiscover_Lookup(t *testing.T) {
	resetDiscover(t)

	root := t.TempDir()
	plugDir := filepath.Join(root, "plugins", "aissh")
	if err := os.MkdirAll(plugDir, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := aitplugin.BuiltinJSON("aissh")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plugDir, "plugin.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	exeName := "aissh"
	if runtime.GOOS == "windows" {
		exeName = "aissh.exe"
	}
	writeFakeExe(t, filepath.Join(plugDir, exeName))

	aitplugin.SetInstallDirsForTest([]string{root})
	aitplugin.SetLookPathForTest(func(string) (string, error) {
		return "", os.ErrNotExist
	})
	aitplugin.SetWarnWriterForTest(ioDiscard{})

	p, ok := aitplugin.Lookup(aitplugin.NameSSH)
	if !ok || p.Manifest.Binary != "aissh" {
		t.Fatalf("Lookup ssh = %+v ok=%v", p, ok)
	}
	if _, ok := aitplugin.Lookup("nosuch"); ok {
		t.Fatal("expected miss")
	}
}

func TestIsReservedName(t *testing.T) {
	if !aitplugin.IsReservedName(aitplugin.CmdHelp) || aitplugin.IsReservedName(aitplugin.NameSSH) {
		t.Fatal("reserved map wrong")
	}
}

func resetDiscover(t *testing.T) {
	t.Helper()
	aitplugin.ResetDiscoverForTest()
	aitplugin.RestoreDiscoverHooksForTest()
	t.Cleanup(func() {
		aitplugin.ResetDiscoverForTest()
		aitplugin.RestoreDiscoverHooksForTest()
	})
}

func writeFakeExe(t *testing.T, path string) {
	t.Helper()
	mode := os.FileMode(0o755)
	if runtime.GOOS == "windows" {
		mode = 0o666
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), mode); err != nil {
		t.Fatal(err)
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
