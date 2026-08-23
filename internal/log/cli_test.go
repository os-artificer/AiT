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

package ailog_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/log"
	"github.com/os-artificer/AiT/internal/log/view"
)

func TestWriteRootHelp(t *testing.T) {
	var buf bytes.Buffer
	if err := ailog.WriteRootHelp(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"Usage:", "Commands:", "search", "sources", "version", "Examples:",
		"search logs (keywords, --expr, or -e)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
	assertCommandsVersionLast(t, out, "  search ")
}

func assertCommandsVersionLast(t *testing.T, out, productRow string) {
	t.Helper()
	idx := strings.Index(out, "Commands:")
	if idx < 0 {
		t.Fatal("missing Commands:")
	}
	block := out[idx:]
	prod := strings.Index(block, productRow)
	ver := strings.Index(block, "  version ")
	if prod < 0 || ver < 0 {
		t.Fatalf("missing rows product=%q version in:\n%s", productRow, block)
	}
	if prod > ver {
		t.Fatalf("version must list after product commands: prod@%d ver@%d\n%s", prod, ver, block)
	}
}

func TestCLI_PluginManifest(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := ailog.ExecuteWriters([]string{"ait-plugin-manifest"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"binary": "ailog"`) {
		t.Fatalf("manifest = %q", out.String())
	}
}

func TestCLI_NoArgs(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := ailog.ExecuteWriters(nil, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("%s", out.String())
	}
}

func TestCLI_Version(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := ailog.ExecuteWriters([]string{"version"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Version:") {
		t.Fatalf("%q", out.String())
	}
}

func TestCLI_VersionFlag(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := ailog.ExecuteWriters([]string{"-V"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Version:") {
		t.Fatalf("%q", out.String())
	}
}

func TestCLI_BuiltinFileSearch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	content := "" +
		"2024-01-02T03:04:05Z alpha error disk\n" +
		"2024-01-02T03:04:06Z beta ok\n" +
		"2024-01-02T03:04:07Z gamma error disk full\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	args := []string{"--backend", "builtin", "-S", path, "--no-pager", "-n", "10", "error", "disk"}
	if err := ailog.ExecuteWriters(args, &out, &errBuf); err != nil {
		t.Fatalf("err=%v stderr=%s", err, errBuf.String())
	}
	got := out.String()
	if !strings.Contains(got, "alpha") || !strings.Contains(got, "gamma") {
		t.Fatalf("stdout=%q stderr=%q", got, errBuf.String())
	}
	if strings.Contains(got, "beta") {
		t.Fatal("beta should not match")
	}
}

func TestCLI_JSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	if err := os.WriteFile(path, []byte("2024-01-02T03:04:05Z hello world\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	args := []string{"--backend", "builtin", "-S", path, "--json", "--no-pager", "-n", "5"}
	if err := ailog.ExecuteWriters(args, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"message"`) {
		t.Fatalf("%q", out.String())
	}
}

func TestCLI_SearchRequiresKeyword(t *testing.T) {
	var out, errBuf bytes.Buffer
	err := ailog.ExecuteWriters([]string{"search"}, &out, &errBuf)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "search requires a keyword") {
		t.Fatalf("%v", err)
	}
}

func TestCLI_SearchVersionEscape(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	if err := os.WriteFile(path, []byte("2024-01-02T03:04:05Z version bump\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	args := []string{"search", "--backend", "builtin", "-S", path, "--no-pager", "version"}
	if err := ailog.ExecuteWriters(args, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "version bump") {
		t.Fatalf("%q", out.String())
	}
}

func TestCLI_Sources(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := ailog.ExecuteWriters([]string{"sources"}, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "system:") {
		t.Fatalf("%q", out.String())
	}
}

func TestCLI_BadSince(t *testing.T) {
	err := ailog.ExecuteWriters([]string{"--since", "nope", "--no-pager"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCLI_Head(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	var b strings.Builder
	for i := 0; i < 10; i++ {
		b.WriteString("2024-01-02T03:04:05Z line" + string(rune('A'+i)) + "\n")
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	args := []string{"--backend", "builtin", "-S", path, "--no-pager", "--head", "-n", "2"}
	if err := ailog.ExecuteWriters(args, &out, &errBuf); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("%q", out.String())
	}
}

// Rare keywords against the native macOS unified log must finish quickly via
// --predicate; without it Collect last-N scans multi-hour windows to EOF.
func TestCLI_NativeRareKeywordBounded(t *testing.T) {
	if testing.Short() {
		t.Skip("native log show")
	}
	if runtime.GOOS != "darwin" {
		t.Skip("macOS log show --predicate regression")
	}

	done := make(chan error, 1)
	go func() {
		var out, errBuf bytes.Buffer
		done <- ailog.ExecuteWriters(
			[]string{"--no-pager", "-n", "5", "ailog-rare-keyword-zz-9f3a"},
			&out, &errBuf,
		)
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(45 * time.Second):
		// Filtered windows still call `log show` up to 1h; under parallel
		// package load this often exceeds a tight 12s wall clock.
		t.Fatal("native rare-keyword query hung (>45s); expected --predicate + filtered windows")
	}
}

func TestCLI_ExprExclusive(t *testing.T) {
	err := ailog.ExecuteWriters(
		[]string{"--expr", "error", "fail", "--no-pager"},
		&bytes.Buffer{}, &bytes.Buffer{},
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "expr") {
		t.Fatalf("%v", err)
	}

	err = ailog.ExecuteWriters(
		[]string{"--expr", "error", "-e", "fail", "--no-pager"},
		&bytes.Buffer{}, &bytes.Buffer{},
	)
	if err == nil {
		t.Fatal("expected error with -e")
	}
}

func TestCLI_SearchExprOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	content := "" +
		"2024-01-02T03:04:05Z alpha error disk\n" +
		"2024-01-02T03:04:06Z beta ok\n" +
		"2024-01-02T03:04:07Z gamma fail soft\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	args := []string{"search", "--backend", "builtin", "-S", path, "--no-pager", "--expr", "error|fail"}
	if err := ailog.ExecuteWriters(args, &out, &errBuf); err != nil {
		t.Fatalf("err=%v stderr=%s", err, errBuf.String())
	}
	got := out.String()
	if !strings.Contains(got, "alpha") || !strings.Contains(got, "gamma") {
		t.Fatalf("stdout=%q", got)
	}
	if strings.Contains(got, "beta") {
		t.Fatal("beta should not match")
	}
}

func TestCLI_FollowExpr(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	if err := os.WriteFile(path, []byte("2024-01-02T03:04:05Z noise hello\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var out, errBuf bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- view.Run(ctx, &out, &errBuf, []string{
			"--backend", "builtin", "-S", path, "--no-pager", "-f",
			"--expr", "error|fail",
		})
	}()

	time.Sleep(300 * time.Millisecond)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString("2024-01-02T03:04:06Z got an error here\n")
	_, _ = f.WriteString("2024-01-02T03:04:07Z still noise\n")
	_ = f.Close()

	deadline := time.After(3 * time.Second)
	for {
		if strings.Contains(out.String(), "error") {
			break
		}
		select {
		case <-deadline:
			cancel()
			<-done
			t.Fatalf("stdout=%q stderr=%q", out.String(), errBuf.String())
		case <-time.After(50 * time.Millisecond):
		}
	}
	got := out.String()
	if strings.Contains(got, "noise") {
		t.Fatalf("noise should not print: %q", got)
	}
	cancel()
	<-done
}
