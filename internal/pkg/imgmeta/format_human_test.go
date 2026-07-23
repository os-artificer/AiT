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
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/imgmeta"
)

func TestFormatHuman_LocationAndGPS(t *testing.T) {
	lat, lon := 31.24, 121.49
	info := imgmeta.Info{
		File:      "photo.jpg",
		Format:    "jpeg",
		Address:   "上海市黄浦区外滩, 上海, 中国",
		Latitude:  &lat,
		Longitude: &lon,
	}
	var buf bytes.Buffer
	if err := imgmeta.FormatHuman(&buf, info); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Location:") || !strings.Contains(out, "上海市") {
		t.Fatalf("missing Location: %s", out)
	}
	if !strings.Contains(out, "GPS:") || !strings.Contains(out, "31.240000") {
		t.Fatalf("missing GPS: %s", out)
	}
}

func TestFormatHuman_EmptyShowsZeros(t *testing.T) {
	info := imgmeta.Info{File: "a.jpg", Format: "jpeg"}
	var buf bytes.Buffer
	if err := imgmeta.FormatHuman(&buf, info); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Contains(out, "No EXIF metadata found.") {
		t.Fatalf("sparse note must not appear: %q", out)
	}
	for _, want := range []string{
		"Width:       0",
		"Height:      0",
		"Make:        0",
		"Model:       0",
		"Software:    0",
		"LensMake:    0",
		"LensModel:   0",
		"SerialNumber: 0",
		"DateTime:    0",
		"Location:    0",
		"GPS:         0.000000, 0.000000",
		"Altitude:    0 m",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %q", want, out)
		}
	}
}

func TestFormatHuman_FieldOrder(t *testing.T) {
	lat, lon := 1.0, 2.0
	alt := 12.5
	info := imgmeta.Info{
		File: "f.jpg", Format: "jpeg", Width: 10, Height: 20,
		Make: "M", Model: "N", Software: "S", LensMake: "LM", LensModel: "L",
		SerialNumber: "SN", DateTime: time.Date(2024, 6, 1, 12, 34, 56, 0, time.UTC),
		Address: "Addr", Latitude: &lat, Longitude: &lon, Altitude: &alt,
	}
	var buf bytes.Buffer
	if err := imgmeta.FormatHuman(&buf, info); err != nil {
		t.Fatal(err)
	}
	order := []string{"File:", "Format:", "Width:", "Height:", "Make:", "Model:", "Software:",
		"LensMake:", "LensModel:", "SerialNumber:", "DateTime:", "Location:", "GPS:", "Altitude:"}
	pos := 0
	for _, key := range order {
		i := strings.Index(buf.String()[pos:], key)
		if i < 0 {
			t.Fatalf("missing %s in %q", key, buf.String())
		}
		pos += i + len(key)
	}
}

func TestFormat_GPSDateTimeAltitude(t *testing.T) {
	lat, lon := 14.204164, 120.969578
	alt := -3.5
	info := imgmeta.Info{
		File: "x", Format: "jpeg",
		DateTime:  time.Date(2024, 6, 1, 12, 34, 56, 0, time.UTC),
		Latitude:  &lat,
		Longitude: &lon,
		Altitude:  &alt,
	}
	var buf bytes.Buffer
	if err := imgmeta.FormatHuman(&buf, info); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "2024-06-01T12:34:56") {
		t.Fatalf("datetime: %s", out)
	}
	if !strings.Contains(out, "14.204164") {
		t.Fatalf("gps: %s", out)
	}
	if !strings.Contains(out, "Altitude:") || !strings.Contains(out, "-3.5 m") {
		t.Fatalf("altitude: %s", out)
	}
}
