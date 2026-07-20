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
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/imgmeta"
)

func TestExtract_JPEGWithEXIF(t *testing.T) {
	info, err := imgmeta.Extract(filepath.Join("testdata", "with_exif.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Make != "Canon" {
		t.Fatalf("Make=%q", info.Make)
	}
	if !strings.Contains(info.Model, "Canon") {
		t.Fatalf("Model=%q", info.Model)
	}
	if info.LensModel == "" {
		t.Fatal("expected LensModel")
	}
	if info.SerialNumber == "" {
		t.Fatal("expected SerialNumber")
	}
	if info.DateTime.IsZero() {
		t.Fatal("expected DateTime")
	}
	if info.Format != "jpeg" {
		t.Fatalf("Format=%q", info.Format)
	}
}

func TestExtract_JPEGGPS(t *testing.T) {
	info, err := imgmeta.Extract(filepath.Join("testdata", "with_gps.exif"))
	if err != nil {
		t.Fatal(err)
	}
	if !info.HasGPS() {
		t.Fatal("expected GPS")
	}
	if *info.Latitude == 0 || *info.Longitude == 0 {
		t.Fatalf("lat=%v lon=%v", *info.Latitude, *info.Longitude)
	}
}

func TestExtract_PNG(t *testing.T) {
	info, err := imgmeta.Extract(filepath.Join("testdata", "plain.png"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Format != "png" && info.Format != "image" {
		t.Fatalf("Format=%q", info.Format)
	}
}

func TestExtract_Unsupported(t *testing.T) {
	_, err := imgmeta.Extract(filepath.Join("testdata", "not-an-image.txt"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, imgmeta.ErrUnsupported) {
		t.Fatalf("want ErrUnsupported, got %v", err)
	}
}

func TestExtract_UnsupportedMissingFile(t *testing.T) {
	path := filepath.Join("testdata", "missing-xyz.jpg")
	_, err := imgmeta.Extract(path)
	if err == nil {
		t.Fatal("expected error")
	}
}

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

func TestFormatJSON(t *testing.T) {
	lat, lon := 1.5, 2.5
	infos := []imgmeta.Info{{
		File: "a.jpg", Format: "jpeg", Address: "Somewhere",
		Latitude: &lat, Longitude: &lon,
	}}
	var buf bytes.Buffer
	if err := imgmeta.FormatJSON(&buf, infos); err != nil {
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
	if err := imgmeta.FormatJSON(&buf, infos); err != nil {
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
	if err := imgmeta.FormatJSON(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buf.String()) != "[]" {
		t.Fatalf("got %q", buf.String())
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
