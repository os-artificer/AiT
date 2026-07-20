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
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/aitimg"
	"github.com/os-artificer/AiT/internal/pkg/geocode"
)

func TestWriteRootHelp(t *testing.T) {
	var buf bytes.Buffer
	if err := aitimg.WriteRootHelp(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"Usage:", "Examples:", "--json", "--no-geocode", "--version", "Location"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %s", want, out)
		}
	}
}

func TestCLI_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := aitimg.ExecuteWriters(nil, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "image path required") {
		t.Fatalf("err=%v", err)
	}
}

func TestCLI_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"display_name":"Test Address"}`))
	}))
	t.Cleanup(srv.Close)

	aitimg.SetTestGeocodeClient(&geocode.Client{BaseURL: srv.URL, HTTP: srv.Client(), UserAgent: "test"})
	aitimg.SetTestGeocodeSleep(func(context.Context) error { return nil })
	t.Cleanup(func() {
		aitimg.SetTestGeocodeClient(nil)
		aitimg.SetTestGeocodeSleep(nil)
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

	aitimg.SetTestGeocodeClient(&geocode.Client{BaseURL: srv.URL, HTTP: srv.Client()})
	aitimg.SetTestGeocodeSleep(func(context.Context) error { return nil })
	t.Cleanup(func() {
		aitimg.SetTestGeocodeClient(nil)
		aitimg.SetTestGeocodeSleep(nil)
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
