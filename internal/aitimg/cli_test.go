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

package aitimg_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/aitimg"
	"github.com/os-artificer/AiT/internal/aitimg/clear"
	"github.com/os-artificer/AiT/internal/aitimg/extract"
	"github.com/os-artificer/AiT/internal/aitimg/outpath"
	"github.com/os-artificer/AiT/internal/aitimg/set"
	"github.com/os-artificer/AiT/internal/aitimg/version"
	"github.com/os-artificer/AiT/internal/pkg/geocode"
	"github.com/os-artificer/AiT/internal/pkg/imgmeta"
)

func TestWriteRootHelp(t *testing.T) {
	var buf bytes.Buffer
	if err := aitimg.WriteRootHelp(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"Usage:", "Examples:", "--json", "--all", "--no-geocode", "--version",
		"Location", "Commands:", "version", "set", "clear",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %s", want, out)
		}
	}
}

func TestCLI_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := aitimg.ExecuteWriters(nil, &stdout, &stderr); err != nil {
		t.Fatalf("empty argv should print root help: %v", err)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("want root help, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr=%s", stderr.String())
	}
}

func TestCLI_EmptyPaths(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := aitimg.ExecuteWriters([]string{"--json"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, imgmeta.ErrEmptyPath) {
		t.Fatalf("want ErrEmptyPath, got %v", err)
	}
}

func TestCLI_VersionSubcommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := aitimg.ExecuteWriters([]string{"version"}, &stdout, &stderr); err != nil {
		t.Fatalf("err=%v stderr=%s", err, stderr.String())
	}
	out := stdout.String()
	want := version.Version()
	if !strings.Contains(out, want) && want != "" {
		// Version() may be multi-line; at least one known prefix.
		if !strings.Contains(out, "Version:") && !strings.Contains(out, "GitHash:") && !strings.Contains(out, "Build:") {
			t.Fatalf("want version output, got %q", out)
		}
	}
	if out == "" {
		t.Fatal("empty version output")
	}
}

func TestCLI_VersionTakesPrecedenceOverFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := aitimg.ExecuteWriters([]string{"version"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if errors.Is(err, imgmeta.ErrEmptyPath) {
		t.Fatal("version must not be treated as a file path")
	}
	if strings.Contains(stderr.String(), "not found") || strings.Contains(stderr.String(), "File not found") {
		t.Fatalf("must not be file-not-found: stderr=%s", stderr.String())
	}
	if stdout.Len() == 0 {
		t.Fatal("expected version text on stdout")
	}
}

func TestCLI_VersionFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := aitimg.ExecuteWriters([]string{"-V"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected -V version output")
	}
}

func TestCLI_VersionFlagWithFile(t *testing.T) {
	path := filepath.Join("..", "pkg", "imgmeta", "testdata", "no_exif.jpg")
	var stdout, stderr bytes.Buffer
	if err := aitimg.ExecuteWriters([]string{"-V", path}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stdout.String(), "File:") {
		t.Fatalf("-V with file must not extract: %s", stdout.String())
	}
	if stdout.Len() == 0 {
		t.Fatal("expected version output")
	}
}

func TestCLI_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"display_name":"Test Address"}`))
	}))
	t.Cleanup(srv.Close)

	extract.SetTestGeocodeClient(&geocode.Client{BaseURL: srv.URL, HTTP: srv.Client(), UserAgent: "test"})
	extract.SetTestGeocodeSleep(func(context.Context) error { return nil })
	t.Cleanup(func() {
		extract.SetTestGeocodeClient(nil)
		extract.SetTestGeocodeSleep(nil)
	})

	path := filepath.Join("..", "pkg", "imgmeta", "testdata", "with_gps.exif")
	var stdout, stderr bytes.Buffer
	if err := aitimg.ExecuteWriters([]string{"--json", path}, &stdout, &stderr); err != nil {
		t.Fatalf("err=%v stderr=%s", err, stderr.String())
	}
	var arr []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &arr); err != nil {
		t.Fatalf("json: %v body=%s", err, stdout.String())
	}
	if len(arr) != 1 {
		t.Fatalf("len=%d", len(arr))
	}
	if arr[0]["address"] != "Test Address" {
		t.Fatalf("address=%v", arr[0]["address"])
	}
	gps, ok := arr[0]["gps"].(map[string]any)
	if !ok || gps["lat"] == nil {
		t.Fatalf("gps=%v", arr[0]["gps"])
	}
	if _, has := arr[0]["exif"]; has {
		t.Fatal("default --json must not include exif")
	}
}

func TestCLI_JSONAll(t *testing.T) {
	path := filepath.Join("..", "pkg", "imgmeta", "testdata", "with_gps.exif")
	var stdout, stderr bytes.Buffer
	if err := aitimg.ExecuteWriters([]string{"--json", "--all", "--no-geocode", path}, &stdout, &stderr); err != nil {
		t.Fatalf("err=%v stderr=%s", err, stderr.String())
	}
	var arr []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &arr); err != nil {
		t.Fatalf("json: %v body=%s", err, stdout.String())
	}
	if len(arr) != 1 {
		t.Fatalf("len=%d", len(arr))
	}
	ex, ok := arr[0]["exif"].(map[string]any)
	if !ok {
		t.Fatalf("missing exif: %v", arr[0])
	}
	if _, ok := ex["IFD0"]; !ok {
		t.Fatalf("missing IFD0: %v", ex)
	}
	if _, ok := ex["ExifIFD"]; !ok {
		t.Fatalf("missing ExifIFD: %v", ex)
	}
}

func TestCLI_NoGeocodeKeepsGPS(t *testing.T) {
	path := filepath.Join("..", "pkg", "imgmeta", "testdata", "with_gps.exif")
	var stdout, stderr bytes.Buffer
	if err := aitimg.ExecuteWriters([]string{"--no-geocode", path}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	if !strings.Contains(out, "Location:    0") {
		t.Fatalf("want Location: 0: %s", out)
	}
	if !strings.Contains(out, "GPS:") {
		t.Fatalf("missing GPS: %s", out)
	}
}

func TestCLI_MultiFilePartialFail(t *testing.T) {
	good := filepath.Join("..", "pkg", "imgmeta", "testdata", "no_exif.jpg")
	bad := filepath.Join("..", "pkg", "imgmeta", "testdata", "not-an-image.txt")

	var stdout, stderr bytes.Buffer
	err := aitimg.ExecuteWriters([]string{"--json", good, bad}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected partial failure")
	}
	if !errors.Is(err, imgmeta.ErrPartialFailure) {
		t.Fatalf("want ErrPartialFailure, got %v", err)
	}
	if stderr.Len() == 0 {
		t.Fatal("expected stderr hint")
	}
	var arr []map[string]any
	if jerr := json.Unmarshal(stdout.Bytes(), &arr); jerr != nil {
		t.Fatalf("json: %v body=%s", jerr, stdout.String())
	}
	if len(arr) != 1 {
		t.Fatalf("want 1 success, got %d", len(arr))
	}
}

func TestCLI_GeocodeErrorKeepsGPS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	extract.SetTestGeocodeClient(&geocode.Client{BaseURL: srv.URL, HTTP: srv.Client()})
	extract.SetTestGeocodeSleep(func(context.Context) error { return nil })
	t.Cleanup(func() {
		extract.SetTestGeocodeClient(nil)
		extract.SetTestGeocodeSleep(nil)
	})

	path := filepath.Join("..", "pkg", "imgmeta", "testdata", "with_gps.exif")
	var stdout, stderr bytes.Buffer
	if err := aitimg.ExecuteWriters([]string{path}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	if !strings.Contains(out, "Location:    0") {
		t.Fatalf("want Location: 0 on geocode fail: %s", out)
	}
	if !strings.Contains(out, "GPS:") {
		t.Fatalf("missing GPS: %s", out)
	}
}

func TestSetCmd_NoFlags(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.jpg")
	src := filepath.Join("..", "pkg", "imgmeta", "testdata", "no_exif.jpg")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err = aitimg.ExecuteWriters([]string{"set", path}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, imgmeta.ErrNothingToChange) {
		t.Fatalf("want ErrNothingToChange, got %v", err)
	}
}

func TestClearCmd_NoFlags(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.jpg")
	src := filepath.Join("..", "pkg", "imgmeta", "testdata", "no_exif.jpg")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err = aitimg.ExecuteWriters([]string{"clear", path}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, imgmeta.ErrNothingToChange) {
		t.Fatalf("want ErrNothingToChange, got %v", err)
	}
}

func TestCLI_SetAndClearMake(t *testing.T) {
	src := filepath.Join("..", "pkg", "imgmeta", "testdata", "with_exif.jpg")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "photo.jpg")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if err := aitimg.ExecuteWriters([]string{"set", "--make", "CliMake", "--inplace", path}, &stdout, &stderr); err != nil {
		t.Fatalf("set: %v stderr=%s", err, stderr.String())
	}
	info, err := imgmeta.Extract(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(info.Make, "CliMake") {
		t.Fatalf("Make=%q", info.Make)
	}

	stdout.Reset()
	stderr.Reset()
	if err := aitimg.ExecuteWriters([]string{"clear", "--make", "--inplace", path}, &stdout, &stderr); err != nil {
		t.Fatalf("clear: %v stderr=%s", err, stderr.String())
	}
	info, err = imgmeta.Extract(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Make != "" {
		t.Fatalf("Make still set: %q", info.Make)
	}
}

func TestCLI_SetPartialFailure(t *testing.T) {
	src := filepath.Join("..", "pkg", "imgmeta", "testdata", "no_exif.jpg")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	good := filepath.Join(t.TempDir(), "good.jpg")
	if err := os.WriteFile(good, data, 0o600); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join("..", "pkg", "imgmeta", "testdata", "plain.png")

	var stdout, stderr bytes.Buffer
	err = aitimg.ExecuteWriters([]string{"set", "--make", "X", "--inplace", good, bad}, &stdout, &stderr)
	if !errors.Is(err, imgmeta.ErrPartialFailure) {
		t.Fatalf("want partial failure, got %v", err)
	}
	if stderr.Len() == 0 {
		t.Fatal("expected stderr hint for PNG")
	}
	if !strings.Contains(stdout.String(), "updated "+good) {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestCLI_SetDefaultNewFile_NonTTY(t *testing.T) {
	outpath.SetStdinIsTTYForTest(t, false)
	src := filepath.Join("..", "pkg", "imgmeta", "testdata", "with_exif.jpg")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "photo.jpg")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if err := aitimg.ExecuteWriters([]string{"set", "--make", "NewFileMake", path}, &stdout, &stderr); err != nil {
		t.Fatalf("set: %v stderr=%s", err, stderr.String())
	}
	dst := filepath.Join(dir, "photo_aitimg.jpg")
	if !strings.Contains(stdout.String(), "wrote "+dst) {
		t.Fatalf("stdout=%q", stdout.String())
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("source changed")
	}
	info, err := imgmeta.Extract(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(info.Make, "NewFileMake") {
		t.Fatalf("Make=%q", info.Make)
	}
}

func TestCLI_ClearDefaultNewFile_NonTTY(t *testing.T) {
	outpath.SetStdinIsTTYForTest(t, false)
	src := filepath.Join("..", "pkg", "imgmeta", "testdata", "with_exif.jpg")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "photo.jpg")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := imgmeta.Patch(path, imgmeta.PatchFields{Make: strPtr("Temp")}); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if err := aitimg.ExecuteWriters([]string{"clear", "--make", path}, &stdout, &stderr); err != nil {
		t.Fatalf("clear: %v stderr=%s", err, stderr.String())
	}
	dst := filepath.Join(dir, "photo_aitimg.jpg")
	if !strings.Contains(stdout.String(), "wrote "+dst) {
		t.Fatalf("stdout=%q", stdout.String())
	}
	info, err := imgmeta.Extract(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Make != "" {
		t.Fatalf("Make=%q", info.Make)
	}
}

func strPtr(s string) *string { return &s }

func TestCLI_SetOutputFlag(t *testing.T) {
	srcData, err := os.ReadFile(filepath.Join("..", "pkg", "imgmeta", "testdata", "no_exif.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "photo.jpg")
	out := filepath.Join(dir, "out.jpg")
	if err := os.WriteFile(path, srcData, 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if err := aitimg.ExecuteWriters([]string{"set", "--make", "OutMake", "-o", out, path}, &stdout, &stderr); err != nil {
		t.Fatalf("%v stderr=%s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "wrote "+out) {
		t.Fatalf("stdout=%q", stdout.String())
	}
	info, err := imgmeta.Extract(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(info.Make, "OutMake") {
		t.Fatalf("Make=%q", info.Make)
	}
}

func TestCLI_SetInplace(t *testing.T) {
	srcData, err := os.ReadFile(filepath.Join("..", "pkg", "imgmeta", "testdata", "no_exif.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "photo.jpg")
	if err := os.WriteFile(path, srcData, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if err := aitimg.ExecuteWriters([]string{"set", "--make", "Y", "-y", path}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "updated "+path) {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestCLI_SetOutputMultiFileFails(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := aitimg.ExecuteWriters([]string{"set", "--make", "X", "-o", "out.jpg", "a.jpg", "b.jpg"}, &stdout, &stderr)
	if !errors.Is(err, outpath.ErrOutputMultiFile) {
		t.Fatalf("got %v", err)
	}
}

func TestCLI_SetOutputInplaceConflict(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := aitimg.ExecuteWriters([]string{"set", "--make", "X", "-o", "out.jpg", "--inplace", "a.jpg"}, &stdout, &stderr)
	if !errors.Is(err, outpath.ErrOutputInplaceConflict) {
		t.Fatalf("got %v", err)
	}
}

func TestCLI_SetOutputOverwritesExisting(t *testing.T) {
	srcData, err := os.ReadFile(filepath.Join("..", "pkg", "imgmeta", "testdata", "no_exif.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "photo.jpg")
	out := filepath.Join(dir, "out.jpg")
	if err := os.WriteFile(path, srcData, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(out, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if err := aitimg.ExecuteWriters([]string{"set", "--make", "Ow", "-o", out, path}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	info, err := imgmeta.Extract(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(info.Make, "Ow") {
		t.Fatalf("Make=%q", info.Make)
	}
}

func TestCLI_SetOutputIsDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "photo.jpg")
	if err := os.WriteFile(path, []byte{0xff, 0xd8, 0xff}, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	err := aitimg.ExecuteWriters([]string{"set", "--make", "X", "-o", dir, path}, &stdout, &stderr)
	if !errors.Is(err, outpath.ErrOutputIsDir) {
		t.Fatalf("got %v", err)
	}
}

func TestCLI_SetOutputEmpty(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := aitimg.ExecuteWriters([]string{"set", "--make", "X", "-o", "", "a.jpg"}, &stdout, &stderr)
	if !errors.Is(err, outpath.ErrOutputEmpty) {
		t.Fatalf("got %v", err)
	}
}

func TestWriteSetHelp_OutputFlags(t *testing.T) {
	var buf bytes.Buffer
	if err := set.WriteHelp(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"--output", "--inplace", "-y", "_aitimg.jpg", "overwrite"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in help", want)
		}
	}
}

func TestWriteClearHelp_OutputFlags(t *testing.T) {
	var buf bytes.Buffer
	if err := clear.WriteHelp(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"--output", "--inplace", "-y", "_aitimg.jpg"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in help", want)
		}
	}
}
