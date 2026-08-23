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
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
)

func TestOfficialPluginNames_ManifestsAligned(t *testing.T) {
	t.Parallel()

	official := aitplugin.OfficialPluginNames()
	embedded, err := aitplugin.EmbeddedManifestNames()
	if err != nil {
		t.Fatal(err)
	}
	if len(official) != len(embedded) {
		t.Fatalf("official %v embedded %v", official, embedded)
	}
	for i := range official {
		if official[i] != embedded[i] {
			t.Fatalf("official=%v embedded=%v", official, embedded)
		}
	}
}

func TestPrimaryInstallDir_ResolvedSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink install root test skipped on windows")
	}

	root := t.TempDir()
	realDir := filepath.Join(root, "realbin")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	exePath := filepath.Join(realDir, "ait")
	if err := os.WriteFile(exePath, []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(root, "linkbin")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Fatal(err)
	}
	linkExe := filepath.Join(linkDir, "ait")

	aitplugin.SetInstallDirsForTest([]string{linkDir, realDir})
	aitplugin.ResetDiscoverForTest()
	t.Cleanup(func() {
		aitplugin.RestoreDiscoverHooksForTest()
		aitplugin.ResetDiscoverForTest()
	})

	got, err := aitplugin.PrimaryInstallDir()
	if err != nil {
		t.Fatal(err)
	}
	if got != realDir {
		t.Fatalf("PrimaryInstallDir=%q want %q", got, realDir)
	}
	_ = linkExe
}

func TestIsManageCommand(t *testing.T) {
	t.Parallel()

	if !aitplugin.IsManageCommand(aitplugin.CmdInstall) {
		t.Fatal("install should be manage command")
	}
	if aitplugin.IsManageCommand(aitplugin.NameSSH) {
		t.Fatal("aissh is not manage command")
	}
}

func TestIsReservedName_ManageCommands(t *testing.T) {
	t.Parallel()

	for _, name := range []string{aitplugin.CmdInstall, aitplugin.CmdUpdate, aitplugin.CmdUpgrade} {
		if !aitplugin.IsReservedName(name) {
			t.Fatalf("%q should be reserved", name)
		}
	}
}
