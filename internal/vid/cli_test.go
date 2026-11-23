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

package aivid_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/deps"
	"github.com/os-artificer/AiT/internal/vid"
)

func TestCLI_PluginManifest(t *testing.T) {
	var out bytes.Buffer
	if err := aivid.ExecuteWriters([]string{aitplugin.ManifestArg}, &out, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(out.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	if m["name"] != "vid" || m["binary"] != "aitvid" {
		t.Fatalf("manifest=%v", m)
	}
}

func TestCLI_NoArgs(t *testing.T) {
	var out bytes.Buffer
	if err := aivid.ExecuteWriters(nil, &out, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	for _, want := range []string{"Usage:", "Commands:", "Examples:", "search", "frames", "deps", "version"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q:\n%s", want, s)
		}
	}
}

func TestCLI_Help(t *testing.T) {
	var out bytes.Buffer
	if err := aivid.ExecuteWriters([]string{"--help"}, &out, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	for _, want := range []string{"Usage:", "Commands:", "Examples:", "search", "frames", "deps", "version"} {
		if !strings.Contains(s, want) {
			t.Fatalf("help missing %q:\n%s", want, s)
		}
	}
}

func TestCLI_Version(t *testing.T) {
	var out bytes.Buffer
	if err := aivid.ExecuteWriters([]string{"version"}, &out, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.String()) == "" {
		t.Fatal("empty version")
	}
}

func TestCLI_Deps(t *testing.T) {
	var out bytes.Buffer
	if err := aivid.ExecuteWriters([]string{"deps"}, &out, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.Contains(s, "ffmpeg") {
		t.Fatalf("deps status=%s", s)
	}
}

func TestCLI_DepsJSON(t *testing.T) {
	var out bytes.Buffer
	if err := aivid.ExecuteWriters([]string{"deps", "--json"}, &out, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var m map[string]string
	if err := json.Unmarshal(out.Bytes(), &m); err != nil {
		t.Fatalf("json: %v body=%s", err, out.String())
	}
	if _, ok := m["ffmpeg"]; !ok {
		t.Fatalf("map=%v", m)
	}
}

func TestCLI_SearchRequiresQuery(t *testing.T) {
	err := aivid.ExecuteWriters([]string{"search", t.TempDir()}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "query") {
		t.Fatalf("got %v", err)
	}
}

func TestCLI_ObjectModeLocalOnlyHint(t *testing.T) {
	var out, errBuf bytes.Buffer
	err := aivid.ExecuteWriters([]string{
		"search", t.TempDir(), "--query", "bike", "--mode", "object", "--local-only",
	}, &out, &errBuf)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, vidsearch.ErrVisionRequired) &&
		!errors.Is(err, deps.ErrMissingFFmpeg) &&
		!errors.Is(err, deps.ErrMissingTesseract) {
		t.Fatalf("unexpected err: %v stderr=%s", err, errBuf.String())
	}
}

func TestCLI_SearchHelp(t *testing.T) {
	var out bytes.Buffer
	if err := aivid.ExecuteWriters([]string{"search", "--help"}, &out, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	for _, want := range []string{"Usage:", "--query", "--mode", "--local-only"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q:\n%s", want, s)
		}
	}
}

func TestCLI_FramesHelp(t *testing.T) {
	var out bytes.Buffer
	if err := aivid.ExecuteWriters([]string{"frames", "--help"}, &out, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("frames help=%s", out.String())
	}
}

func TestCLI_DepsHelp(t *testing.T) {
	var out bytes.Buffer
	if err := aivid.ExecuteWriters([]string{"deps", "--help"}, &out, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	for _, want := range []string{"Usage:", "--install", "--yes", "--json"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q:\n%s", want, s)
		}
	}
}
