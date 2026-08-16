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
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/imgmeta"
)

func copyTestdata(t *testing.T, name string) string {
	t.Helper()
	src := filepath.Join("testdata", name)
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(dst, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return dst
}

func TestPatch_SetMake(t *testing.T) {
	path := copyTestdata(t, "with_exif.jpg")
	makeVal := "AiTTestMake"
	if err := imgmeta.Patch(path, imgmeta.PatchFields{Make: &makeVal}); err != nil {
		t.Fatal(err)
	}
	info, err := imgmeta.Extract(path)
	if err != nil {
		t.Fatal(err)
	}
	// imagemeta normalizes Make to lowercase on decode.
	if !strings.EqualFold(info.Make, makeVal) {
		t.Fatalf("Make=%q want %q (case-insensitive)", info.Make, makeVal)
	}
}

func TestPatch_SetGPS(t *testing.T) {
	path := copyTestdata(t, "no_exif.jpg")
	lat, lon := 31.2304, 121.4737
	if err := imgmeta.Patch(path, imgmeta.PatchFields{Latitude: &lat, Longitude: &lon}); err != nil {
		t.Fatal(err)
	}
	info, err := imgmeta.Extract(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.HasGPS() {
		t.Fatal("expected GPS")
	}
	if math.Abs(*info.Latitude-lat) > 0.001 || math.Abs(*info.Longitude-lon) > 0.001 {
		t.Fatalf("gps=%v,%v want %v,%v", *info.Latitude, *info.Longitude, lat, lon)
	}
}

func TestClear_Make(t *testing.T) {
	path := copyTestdata(t, "with_exif.jpg")
	makeVal := "TempMake"
	if err := imgmeta.Patch(path, imgmeta.PatchFields{Make: &makeVal}); err != nil {
		t.Fatal(err)
	}
	if err := imgmeta.Clear(path, imgmeta.ClearMask{Make: true}); err != nil {
		t.Fatal(err)
	}
	info, err := imgmeta.Extract(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Make != "" {
		t.Fatalf("Make should be cleared, got %q", info.Make)
	}
}

func TestClear_All(t *testing.T) {
	path := copyTestdata(t, "no_exif.jpg")
	makeVal := "M"
	model := "X"
	lat, lon, alt := 1.0, 2.0, 10.0
	dt := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := imgmeta.Patch(path, imgmeta.PatchFields{
		Make: &makeVal, Model: &model,
		Latitude: &lat, Longitude: &lon, Altitude: &alt,
		DateTime: &dt,
	}); err != nil {
		t.Fatal(err)
	}
	if err := imgmeta.Clear(path, imgmeta.AllClearMask()); err != nil {
		t.Fatal(err)
	}
	info, err := imgmeta.Extract(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Make != "" || info.Model != "" || info.HasGPS() || info.Altitude != nil || !info.DateTime.IsZero() {
		t.Fatalf("expected cleared fields, got %+v", info)
	}
}

func TestPatch_RejectPNG(t *testing.T) {
	path := copyTestdata(t, "plain.png")
	makeVal := "X"
	err := imgmeta.Patch(path, imgmeta.PatchFields{Make: &makeVal})
	if !errors.Is(err, imgmeta.ErrUnsupported) {
		t.Fatalf("want ErrUnsupported, got %v", err)
	}
}

func TestPatch_Empty(t *testing.T) {
	err := imgmeta.Patch("x.jpg", imgmeta.PatchFields{})
	if !errors.Is(err, imgmeta.ErrNothingToChange) {
		t.Fatalf("got %v", err)
	}
}

func TestPatchTo_WritesToDst(t *testing.T) {
	src := copyTestdata(t, "with_exif.jpg")
	before, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(filepath.Dir(src), "out.jpg")
	makeVal := "DstOnlyMake"
	if err := imgmeta.PatchTo(src, dst, imgmeta.PatchFields{Make: &makeVal}); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("src bytes changed")
	}
	info, err := imgmeta.Extract(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(info.Make, makeVal) {
		t.Fatalf("dst Make=%q", info.Make)
	}
	srcInfo, err := imgmeta.Extract(src)
	if err != nil {
		t.Fatal(err)
	}
	if strings.EqualFold(srcInfo.Make, makeVal) {
		t.Fatal("src Make should be unchanged")
	}
}

func TestPatchTo_EmptyPaths(t *testing.T) {
	makeVal := "X"
	if err := imgmeta.PatchTo("", "a.jpg", imgmeta.PatchFields{Make: &makeVal}); !errors.Is(err, imgmeta.ErrEmptyPath) {
		t.Fatalf("got %v", err)
	}
	if err := imgmeta.PatchTo("a.jpg", "", imgmeta.PatchFields{Make: &makeVal}); !errors.Is(err, imgmeta.ErrEmptyPath) {
		t.Fatalf("got %v", err)
	}
}

func TestParseDateTime(t *testing.T) {
	t1, err := imgmeta.ParseDateTime("2020:06:07 01:30:00")
	if err != nil {
		t.Fatal(err)
	}
	if t1.Year() != 2020 || t1.Month() != 6 || t1.Day() != 7 {
		t.Fatalf("got %v", t1)
	}
	t2, err := imgmeta.ParseDateTime("2020-06-07T01:30:00Z")
	if err != nil {
		t.Fatal(err)
	}
	if !t2.Equal(time.Date(2020, 6, 7, 1, 30, 0, 0, time.UTC)) {
		t.Fatalf("got %v", t2)
	}
}
