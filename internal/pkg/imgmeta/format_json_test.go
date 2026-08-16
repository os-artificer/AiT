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

package imgmeta_test

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/imgmeta"
)

func TestFormatJSON(t *testing.T) {
	lat, lon := 1.5, 2.5
	infos := []imgmeta.Info{{
		File: "a.jpg", Format: "jpeg", Address: "Somewhere",
		Latitude: &lat, Longitude: &lon,
	}}
	var buf bytes.Buffer
	if err := imgmeta.FormatJSON(&buf, infos, false); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.HasPrefix(strings.TrimSpace(s), "[") {
		t.Fatalf("want array: %s", s)
	}
	if !strings.Contains(s, `"address"`) || !strings.Contains(s, `"gps"`) {
		t.Fatalf("got %s", s)
	}
	if strings.Contains(s, "No EXIF") {
		t.Fatal("JSON must not include sparse note")
	}
}

func TestFormatJSON_MissingFieldsAreZero(t *testing.T) {
	infos := []imgmeta.Info{{File: "a.jpg", Format: "jpeg"}}
	var buf bytes.Buffer
	if err := imgmeta.FormatJSON(&buf, infos, false); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for _, want := range []string{
		`"make":"0"`,
		`"model":"0"`,
		`"software":"0"`,
		`"lens_make":"0"`,
		`"lens_model":"0"`,
		`"serial_number":"0"`,
		`"datetime":"0"`,
		`"address":"0"`,
		`"width":0`,
		`"height":0`,
		`"altitude":0`,
		`"lat":0`,
		`"lon":0`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in %s", want, s)
		}
	}
}

func TestFormatJSON_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := imgmeta.FormatJSON(&buf, nil, false); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buf.String()) != "[]" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestFormatJSON_All(t *testing.T) {
	info, err := imgmeta.Extract(filepath.Join("testdata", "with_gps.exif"))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := imgmeta.FormatJSON(&buf, []imgmeta.Info{info}, true); err != nil {
		t.Fatal(err)
	}
	var arr []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &arr); err != nil {
		t.Fatalf("json: %v body=%s", err, buf.String())
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
	gps, ok := ex["GPS"].(map[string]any)
	if !ok || gps["Latitude"] == nil {
		t.Fatalf("GPS=%v", ex["GPS"])
	}
}
