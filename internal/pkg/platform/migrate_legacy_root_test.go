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

func TestAppRoot_MigratesLegacyAitDir(t *testing.T) {
	isolatePlatformHome(t)

	home := os.Getenv("HOME")
	legacyRoot := filepath.Join(home, ".local", "share", legacyAppName)
	newRoot := filepath.Join(home, ".local", "share", appName)

	if err := os.MkdirAll(legacyRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(legacyRoot, dbFile)
	if err := os.WriteFile(dbPath, []byte("sqlite"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := AppRoot()
	if err != nil {
		t.Fatal(err)
	}
	if got != newRoot {
		t.Fatalf("AppRoot=%q want %q", got, newRoot)
	}
	if _, err := os.Stat(filepath.Join(newRoot, dbFile)); err != nil {
		t.Fatalf("migrated db missing: %v", err)
	}
	if _, err := os.Stat(legacyRoot); !os.IsNotExist(err) {
		t.Fatalf("legacy root should be removed after rename, stat err=%v", err)
	}
}

func TestAppRoot_MigratesIntoEmptyShellNewRoot(t *testing.T) {
	isolatePlatformHome(t)

	home := os.Getenv("HOME")
	legacyRoot := filepath.Join(home, ".local", "share", legacyAppName)
	newRoot := filepath.Join(home, ".local", "share", appName)

	if err := os.MkdirAll(legacyRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyRoot, ConfigFileName), []byte("session: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(newRoot, 0o700); err != nil {
		t.Fatal(err)
	}

	got, err := AppRoot()
	if err != nil {
		t.Fatal(err)
	}
	if got != newRoot {
		t.Fatalf("AppRoot=%q want %q", got, newRoot)
	}
	if _, err := os.Stat(filepath.Join(newRoot, ConfigFileName)); err != nil {
		t.Fatalf("migrated config missing: %v", err)
	}
}

func TestMaybeMigrateLegacyDataRoot_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin path layout")
	}
	testMigrateLegacyDataRootForOS(t, "darwin")
}

func TestMaybeMigrateLegacyDataRoot_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows path layout")
	}
	testMigrateLegacyDataRootForOS(t, "windows")
}

func testMigrateLegacyDataRootForOS(t *testing.T, goos string) {
	t.Helper()

	prev := runtimeGOOS
	runtimeGOOS = func() string { return goos }
	t.Cleanup(func() { runtimeGOOS = prev })

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv(envHome, "")
	t.Setenv(envDataDir, "")
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	t.Setenv("XDG_DATA_HOME", "")

	legacyRoot, err := legacyDataRootPath()
	if err != nil {
		t.Fatal(err)
	}
	newRoot, err := defaultAppRootPath()
	if err != nil {
		t.Fatal(err)
	}
	if legacyRoot == newRoot {
		t.Fatal("legacy and new roots must differ")
	}

	if err := os.MkdirAll(legacyRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyRoot, dbFile), []byte("sqlite"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := maybeMigrateLegacyDataRoot(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(newRoot, dbFile)); err != nil {
		t.Fatalf("migrated db missing: %v", err)
	}
}
