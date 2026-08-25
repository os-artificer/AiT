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

package aitrelease

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
	"github.com/os-artificer/AiT/internal/pkg/aitrelease/testrelease"
)

func TestPlanInstall_AllOfficial(t *testing.T) {
	t.Parallel()

	plan, err := BuildPlan(context.Background(), Request{Action: ActionInstall})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Tools) != 5 {
		t.Fatalf("tools=%v", plan.Tools)
	}
}

func TestPlanUpdate_InstalledOnly(t *testing.T) {
	resetDiscoverHooks(t)
	root := t.TempDir()
	installStubPlugin(t, root, "aissh")
	aitplugin.ResetDiscoverForTest()

	plan, err := BuildPlan(context.Background(), Request{Action: ActionUpdate})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Tools) != 1 || plan.Tools[0] != "aissh" {
		t.Fatalf("plan=%+v", plan)
	}
}

func TestPlanUpdate_NothingInstalled(t *testing.T) {
	resetDiscoverHooks(t)
	_, err := BuildPlan(context.Background(), Request{Action: ActionUpdate})
	if !errors.Is(err, ErrNothingToUpdate) {
		t.Fatalf("got %v", err)
	}
}

func TestPlanUpdate_WithArgsSkipsNotInstalled(t *testing.T) {
	resetDiscoverHooks(t)
	root := t.TempDir()
	installStubPlugin(t, root, "aissh")
	aitplugin.ResetDiscoverForTest()

	plan, err := BuildPlan(context.Background(), Request{
		Action: ActionUpdate,
		Names:  []string{"aissh", "ailog"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Tools) != 1 || plan.Tools[0] != "aissh" {
		t.Fatalf("tools=%v", plan.Tools)
	}
	if len(plan.Skipped) != 1 || plan.Skipped[0] != "ailog" {
		t.Fatalf("skipped=%v", plan.Skipped)
	}
}

func TestPlanUpgrade(t *testing.T) {
	resetDiscoverHooks(t)
	root := t.TempDir()
	installStubPlugin(t, root, "aissh")
	aitplugin.ResetDiscoverForTest()

	plan, err := BuildPlan(context.Background(), Request{Action: ActionUpgrade})
	if err != nil {
		t.Fatal(err)
	}
	if !plan.IncludeAIT {
		t.Fatal("upgrade should include ait")
	}
	if len(plan.Tools) != 1 || plan.Tools[0] != "aissh" {
		t.Fatalf("tools=%v", plan.Tools)
	}
}

func TestPlanUpgrade_RejectsExtraArgs(t *testing.T) {
	t.Parallel()

	_, err := BuildPlan(context.Background(), Request{Action: ActionUpgrade, Names: []string{"aissh"}})
	if !errors.Is(err, ErrUpgradeExtraArgs) {
		t.Fatalf("got %v", err)
	}
}

func TestFetchLatestRelease_Errors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)

	_, err := FetchLatestRelease(context.Background(), srv.Client(), srv.URL, "o", "r", "test")
	if !errors.Is(err, ErrNoRelease) {
		t.Fatalf("got %v", err)
	}
}

func TestReleaseErr_IsSentinel(t *testing.T) {
	t.Parallel()

	if !errors.Is(ErrNoRelease, ErrNoRelease) {
		t.Fatal("ErrNoRelease sentinel")
	}
	if !errors.Is(ErrNothingToUpdate, ErrNothingToUpdate) {
		t.Fatal("ErrNothingToUpdate sentinel")
	}
}

func TestExtract_FlatRootBundle(t *testing.T) {
	archive := writeTestTarGz(t, map[string][]byte{
		"aissh":                     []byte("bin-aissh"),
		"plugins/aissh/aissh":       []byte("plug-aissh"),
		"plugins/aissh/plugin.json": []byte(`{"apiVersion":1,"name":"aissh","binary":"aissh","short":"s"}`),
	})
	dest := t.TempDir()
	if err := extractArchive(archive, dest); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dest, "aissh")); err != nil {
		t.Fatal(err)
	}
}

func TestExtract_RejectsTraversal(t *testing.T) {
	archive := writeTestTarGz(t, map[string][]byte{
		"../escape": []byte("bad"),
	})
	dest := t.TempDir()
	if err := extractArchive(archive, dest); err == nil {
		t.Fatal("expected traversal error")
	}
}

func TestApplyBundle_PartialInstall(t *testing.T) {
	extractDir := t.TempDir()
	installRoot := t.TempDir()
	testrelease.WriteBundleTree(extractDir, "aissh-v2", "ailog-v2")

	plan := Plan{Tools: []string{"aissh"}}
	if err := applyBundle(installRoot, extractDir, plan); err != nil {
		t.Fatal(err)
	}

	aisshData, err := os.ReadFile(filepath.Join(installRoot, "aissh"))
	if err != nil {
		t.Fatal(err)
	}
	if string(aisshData) != "aissh-v2" {
		t.Fatalf("aissh=%q", aisshData)
	}
	if _, err := os.Stat(filepath.Join(installRoot, "ailog")); !os.IsNotExist(err) {
		t.Fatalf("ailog should not be installed: %v", err)
	}
}

func TestExecute_InstallOne(t *testing.T) {
	root := t.TempDir()
	resetDiscoverHooks(t)
	aitplugin.SetInstallDirsForTest([]string{root})

	srv := testrelease.NewServer(t, "v0.9.0")
	t.Cleanup(srv.Close)

	var out bytes.Buffer
	err := Execute(context.Background(), Options{
		BaseURL:     srv.URL,
		InstallRoot: root,
		UserAgent:   "test",
	}, Request{Action: ActionInstall, Names: []string{"aissh"}}, &out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "installed aissh") {
		t.Fatalf("out=%q", out.String())
	}
	if _, err := os.Stat(filepath.Join(root, "plugins", "aissh", "aissh")); err != nil {
		t.Fatal(err)
	}
}

func TestExecute_RefreshesDiscover(t *testing.T) {
	root := t.TempDir()
	resetDiscoverHooks(t)
	aitplugin.SetInstallDirsForTest([]string{root})

	srv := testrelease.NewServer(t, "v0.9.0")
	t.Cleanup(srv.Close)

	err := Execute(context.Background(), Options{
		BaseURL:     srv.URL,
		InstallRoot: root,
		UserAgent:   "test",
	}, Request{Action: ActionInstall, Names: []string{"aissh"}}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := aitplugin.Lookup("aissh"); !ok {
		t.Fatal("expected aissh in Discover after install")
	}
}

func writeTestTarGz(t *testing.T, files map[string][]byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.tar.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)
	for name, data := range files {
		if err := tw.WriteHeader(&tar.Header{
			Name: name,
			Mode: 0o755,
			Size: int64(len(data)),
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func installStubPlugin(t *testing.T, root, binary string) {
	t.Helper()
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
	if err := os.WriteFile(filepath.Join(plugDir, name), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	aitplugin.SetInstallDirsForTest([]string{root})
}

func resetDiscoverHooks(t *testing.T) {
	t.Helper()
	aitplugin.ResetDiscoverForTest()
	aitplugin.RestoreDiscoverHooksForTest()
	aitplugin.SetLookPathForTest(func(string) (string, error) {
		return "", os.ErrNotExist
	})
	aitplugin.SetInstallDirsForTest([]string{t.TempDir()})
	aitplugin.SetWarnWriterForTest(io.Discard)
	t.Cleanup(func() {
		aitplugin.ResetDiscoverForTest()
		aitplugin.RestoreDiscoverHooksForTest()
	})
}
