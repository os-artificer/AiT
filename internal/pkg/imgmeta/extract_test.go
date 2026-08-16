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
	"path/filepath"
	"strings"
	"testing"

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
