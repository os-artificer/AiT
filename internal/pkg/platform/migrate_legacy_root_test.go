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
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func isolateLinuxPlatformHome(t *testing.T) {
	t.Helper()

	prev := runtimeGOOS
	runtimeGOOS = func() string { return "linux" }
	t.Cleanup(func() { runtimeGOOS = prev })

	isolatePlatformHome(t)
}

func requireDistinctMixedCaseRoots(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	legacy := filepath.Join(home, legacyMixedCaseAppName)
	current := filepath.Join(home, appName)
	if err := os.MkdirAll(legacy, 0o700); err != nil {
		t.Fatal(err)
	}
	if pathsSameDir(legacy, current) {
		t.Skip("case-insensitive filesystem merges AiT and ait paths")
	}
}

func TestAppRoot_MigratesMixedCaseDir(t *testing.T) {
	isolatePlatformHome(t)
	requireDistinctMixedCaseRoots(t)

	legacyRoot, err := legacyDataRootPath()
	if err != nil {
		t.Fatal(err)
	}
	newRoot, err := defaultAppRootPath()
	if err != nil {
		t.Fatal(err)
	}

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
	requireDistinctMixedCaseRoots(t)

	legacyRoot, err := legacyDataRootPath()
	if err != nil {
		t.Fatal(err)
	}
	newRoot, err := defaultAppRootPath()
	if err != nil {
		t.Fatal(err)
	}

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

func TestPathsSameDir_SamePath(t *testing.T) {
	dir := t.TempDir()
	if !pathsSameDir(dir, dir) {
		t.Fatal("same path should match")
	}
}

func TestPathsSameDir_OneMissing(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing")
	if pathsSameDir(dir, missing) {
		t.Fatal("missing path should not match")
	}
}

func TestPathsSameDir_CaseInsensitive(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
		t.Skip("case-insensitive same inode requires darwin or windows")
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
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
		t.Fatal("legacy and new roots must differ as strings")
	}

	if err := os.MkdirAll(newRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if !pathsSameDir(legacyRoot, newRoot) {
		t.Fatalf("expected same inode for %q and %q", legacyRoot, newRoot)
	}
}

func TestMaybeMigrateLegacyDataRoot_SameDirNoWarn(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
		t.Skip("case-insensitive same inode requires darwin or windows")
	}

	isolatePlatformHome(t)
	resetLegacyDataConflictWarn(t)

	newRoot, err := defaultAppRootPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(newRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newRoot, dbFile), []byte("sqlite"), 0o600); err != nil {
		t.Fatal(err)
	}

	stderr, restore, err := captureStderr(t)
	if err != nil {
		t.Fatal(err)
	}
	defer restore()

	if _, err := AppRoot(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stderr.String(), legacyConflictIntro) {
		t.Fatalf("unexpected conflict warning for same inode: %s", stderr.String())
	}
}

func resetLegacyDataConflictWarn(t *testing.T) {
	t.Helper()
	warnLegacyDataConflictOnce = sync.Once{}
	t.Cleanup(func() {
		warnLegacyDataConflictOnce = sync.Once{}
	})
}

func setupLegacyDataConflict(t *testing.T) (legacyRoot, newRoot string) {
	t.Helper()
	isolatePlatformHome(t)
	requireDistinctMixedCaseRoots(t)
	resetLegacyDataConflictWarn(t)

	var err error
	legacyRoot, err = legacyDataRootPath()
	if err != nil {
		t.Fatal(err)
	}
	newRoot, err = defaultAppRootPath()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(legacyRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyRoot, dbFile), []byte("sqlite"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(newRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newRoot, dbFile), []byte("sqlite"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newRoot, ConfigFileName), []byte("session: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return legacyRoot, newRoot
}

func TestMaybeMigrateLegacyDataRoot_ConflictWarnsOnce(t *testing.T) {
	setupLegacyDataConflict(t)

	stderr, restore, err := captureStderr(t)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		if _, err := AppRoot(); err != nil {
			restore()
			t.Fatal(err)
		}
		if _, err := DataDir(); err != nil {
			restore()
			t.Fatal(err)
		}
		if _, err := DBPath(); err != nil {
			restore()
			t.Fatal(err)
		}
	}

	restore()
	out := stderr.String()
	if got := strings.Count(out, legacyConflictIntro); got != 1 {
		t.Fatalf("warning count=%d want 1; output:\n%s", got, out)
	}
}

func TestMaybeMigrateLegacyDataRoot_ConflictMessageIncludesPaths(t *testing.T) {
	legacyRoot, newRoot := setupLegacyDataConflict(t)

	var buf bytes.Buffer
	writeLegacyDataConflictWarn(&buf, legacyRoot, newRoot)
	out := buf.String()

	for _, want := range []string{
		"mixed-case",
		"legacy:",
		"current:",
		dbFile,
		ConfigFileName,
		legacyConflictAction,
		legacyConflictHint,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, formatDisplayPath(legacyRoot)) {
		t.Fatalf("output missing legacy path %q:\n%s", formatDisplayPath(legacyRoot), out)
	}
	if !strings.Contains(out, formatDisplayPath(newRoot)) {
		t.Fatalf("output missing current path %q:\n%s", formatDisplayPath(newRoot), out)
	}
	if !strings.Contains(out, "\r\n") {
		t.Fatalf("output should use CRLF line endings:\n%s", out)
	}
}

func TestAppRoot_AITHomeOverrideSkipsConflictWarn(t *testing.T) {
	resetLegacyDataConflictWarn(t)

	override := t.TempDir()
	t.Setenv(envHome, override)
	t.Setenv(envDataDir, "")

	stderr, restore, err := captureStderr(t)
	if err != nil {
		t.Fatal(err)
	}
	defer restore()

	if got, err := AppRoot(); err != nil {
		t.Fatal(err)
	} else if got != override {
		t.Fatalf("AppRoot=%q want %q", got, override)
	}
	if strings.Contains(stderr.String(), legacyConflictIntro) {
		t.Fatalf("unexpected conflict warning with AIT_HOME: %s", stderr.String())
	}
}

func captureStderr(t *testing.T) (*bytes.Buffer, func(), error) {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	prev := os.Stderr
	os.Stderr = w
	buf := &bytes.Buffer{}
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(buf, r)
		close(done)
	}()

	restore := func() {
		os.Stderr = prev
		_ = w.Close()
		<-done
		_ = r.Close()
	}
	return buf, restore, nil
}
