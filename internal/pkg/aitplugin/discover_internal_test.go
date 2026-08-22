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

package aitplugin

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestUniqueInstallDirs_DedupesResolved(t *testing.T) {
	root := t.TempDir()
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(root, link); err != nil {
		t.Skipf("symlink: %v", err)
	}

	got := uniqueInstallDirs([]string{root, link, root})
	if len(got) != 1 {
		t.Fatalf("uniqueInstallDirs = %v, want 1 entry", got)
	}
}

func TestDiscover_IgnoresCancelledContext(t *testing.T) {
	ResetDiscoverForTest()
	RestoreDiscoverHooksForTest()
	t.Cleanup(func() {
		ResetDiscoverForTest()
		RestoreDiscoverHooksForTest()
	})

	root := t.TempDir()
	plugDir := filepath.Join(root, "plugins", "aissh")
	if err := os.MkdirAll(plugDir, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := BuiltinJSON("aissh")
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
	mode := os.FileMode(0o755)
	if runtime.GOOS == "windows" {
		mode = 0o666
	}
	if err := os.WriteFile(filepath.Join(plugDir, exeName), []byte("x"), mode); err != nil {
		t.Fatal(err)
	}

	SetInstallDirsForTest([]string{root})
	SetLookPathForTest(func(string) (string, error) {
		return "", os.ErrNotExist
	})
	SetWarnWriterForTest(io.Discard)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	plugins := Discover(ctx)
	if len(plugins) != 1 || plugins[0].Manifest.Name != NameSSH {
		t.Fatalf("plugins = %+v", plugins)
	}
}
