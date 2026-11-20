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
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
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

func TestCLI_Help(t *testing.T) {
	var out bytes.Buffer
	if err := aivid.ExecuteWriters([]string{"--help"}, &out, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	for _, want := range []string{"search", "frames", "deps", "version"} {
		if !strings.Contains(s, want) {
			t.Fatalf("help missing %q:\n%s", want, s)
		}
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
	msg := err.Error() + errBuf.String()
	if !strings.Contains(msg, "vision") && !strings.Contains(msg, "ffmpeg") && !strings.Contains(msg, "tesseract") {
		t.Fatalf("unexpected err: %v stderr=%s", err, errBuf.String())
	}
}
