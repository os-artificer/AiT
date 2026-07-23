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
	"github.com/os-artificer/AiT/internal/aitimg/extract"
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
	if err := aitimg.ExecuteWriters([]string{"set", "--make", "CliMake", path}, &stdout, &stderr); err != nil {
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
	if err := aitimg.ExecuteWriters([]string{"clear", "--make", path}, &stdout, &stderr); err != nil {
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
	err = aitimg.ExecuteWriters([]string{"set", "--make", "X", good, bad}, &stdout, &stderr)
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
