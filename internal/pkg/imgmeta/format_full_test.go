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
	"path/filepath"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/imgmeta"
)

func TestFormatFullHuman_WithExif(t *testing.T) {
	info, err := imgmeta.Extract(filepath.Join("testdata", "with_exif.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := imgmeta.FormatFullHuman(&buf, info.Exif); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "--- EXIF ---\n") {
		t.Fatalf("header: %q", out)
	}
	for _, want := range []string{"Field", "Value", "ExifIFD", "ExposureTime", "FNumber", "ISOSpeedRatings"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "\nIFD1\n") {
		t.Fatalf("unexpected IFD1 on fixture without thumbnail IFD: %s", out)
	}
}

func TestFormatFullHuman_Sparse(t *testing.T) {
	var buf bytes.Buffer
	if err := imgmeta.FormatFullHuman(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buf.String()) != "--- EXIF ---" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestFormatFullHuman_SectionOrder(t *testing.T) {
	info, err := imgmeta.Extract(filepath.Join("testdata", "with_gps.exif"))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := imgmeta.FormatFullHuman(&buf, info.Exif); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "IFD0\n") || !strings.Contains(out, "ExifIFD\n") || !strings.Contains(out, "GPS\n") {
		t.Fatalf("missing sections in %q", out)
	}
	pos := 0
	for _, key := range []string{"IFD0", "ExifIFD", "GPS"} {
		i := strings.Index(out[pos:], key+"\n")
		if i < 0 {
			t.Fatalf("missing %s in order; out=%q", key, out)
		}
		pos += i + len(key)
	}
}

func TestFormatFullHuman_TableHeaders(t *testing.T) {
	info, err := imgmeta.Extract(filepath.Join("testdata", "with_exif.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := imgmeta.FormatFullHuman(&buf, info.Exif); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Field") || !strings.Contains(out, "Value") {
		t.Fatalf("missing table headers: %s", out)
	}
	if !strings.Contains(out, "-----") {
		t.Fatalf("missing table separator: %s", out)
	}
}
