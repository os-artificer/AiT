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

package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAppRoot_AITHomeOverride(t *testing.T) {
	root := t.TempDir()
	t.Setenv(envHome, root)
	t.Setenv(envDataDir, filepath.Join(t.TempDir(), "ignored"))

	data, err := DataDir()
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if data != root || cfg != root {
		t.Fatalf("data=%q cfg=%q want %q", data, cfg, root)
	}

	db, err := DBPath()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(db) != root {
		t.Fatalf("db=%q not under root", db)
	}
}

func TestBinDir_UnderAppRoot(t *testing.T) {
	root := t.TempDir()
	t.Setenv(envHome, root)

	got, err := BinDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "bin")
	if got != want {
		t.Fatalf("BinDir=%q want %q", got, want)
	}
}

func TestConfigDir_EqualsDataDir(t *testing.T) {
	isolatePlatformHome(t)

	data, err := DataDir()
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if data != cfg {
		t.Fatalf("DataDir=%q ConfigDir=%q", data, cfg)
	}
}

func TestAppRoot_EmptyIgnores(t *testing.T) {
	root := t.TempDir()
	t.Setenv(envHome, "   ")
	t.Setenv(envDataDir, "")
	isolatePlatformHome(t)

	got, err := AppRoot()
	if err != nil {
		t.Fatal(err)
	}
	if got == root {
		t.Fatalf("empty AIT_HOME must not use temp dir %q", root)
	}
}

func isolatePlatformHome(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(envHome, "")
	t.Setenv(envDataDir, "")
	if runtime.GOOS == "linux" {
		t.Setenv("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	}
}

func TestLegacyConfigPath_LinuxOnly(t *testing.T) {
	if runtime.GOOS != "linux" {
		legacy, err := LegacyConfigPath()
		if err != nil || legacy != "" {
			t.Fatalf("legacy=%q err=%v", legacy, err)
		}
		return
	}
	isolatePlatformHome(t)
	legacy, err := LegacyConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), appName, ConfigFileName)
	if legacy != want {
		t.Fatalf("legacy=%q want %q", legacy, want)
	}
}
