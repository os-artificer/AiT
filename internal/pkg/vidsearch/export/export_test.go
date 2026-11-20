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

package export_test

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/vidsearch"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/export"
)

func TestResolveCameraID(t *testing.T) {
	if got := export.ResolveCameraID("cam9", "/x/y/z.mp4"); got != "cam9" {
		t.Fatalf("override: %s", got)
	}
	dir := t.TempDir()
	camDir := filepath.Join(dir, "camera_12")
	if err := os.Mkdir(camDir, 0o755); err != nil {
		t.Fatal(err)
	}
	vid := filepath.Join(camDir, "clip.mp4")
	if got := export.ResolveCameraID("", vid); got != "cam12" {
		t.Fatalf("heuristic: %s", got)
	}
}

func TestSaveHitFrame_NoAnnotate(t *testing.T) {
	src := writeSolidJPEG(t)
	out := t.TempDir()
	h := vidsearch.Hit{
		Video: "/abs/clip1.mp4", VideoName: "clip1.mp4", TSec: 3,
		Query: "q", Kind: vidsearch.KindText, Snippet: "q", Score: 1, Mode: vidsearch.ModeText,
	}
	frame, _, err := export.SaveHitFrame(h, src, export.Options{OutDir: out, Annotate: false, CameraID: "unknown"})
	if err != nil {
		t.Fatal(err)
	}
	if bottomDiffers(t, src, frame) {
		t.Fatal("expected no annotation bar when Annotate=false")
	}
}

func TestSaveHitFrame_Annotate(t *testing.T) {
	src := writeSolidJPEG(t)
	out := t.TempDir()
	hit := vidsearch.Hit{
		Video:     "/abs/clip1.mp4",
		VideoName: "clip1.mp4",
		TSec:      12.0,
		Query:     "bike",
		Kind:      vidsearch.KindText,
		Snippet:   "bike",
		Score:     1,
		Mode:      vidsearch.ModeText,
	}
	frame, side, err := export.SaveHitFrame(hit, src, export.Options{
		OutDir:   out,
		Annotate: true,
		CameraID: "cam03",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(filepath.Base(frame), "clip1_t12.0s_") {
		t.Fatalf("name: %s", frame)
	}
	if !strings.Contains(frame, "cam03") {
		t.Fatalf("camera dir: %s", frame)
	}
	data, err := os.ReadFile(side)
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	for _, want := range []string{`"camera_id": "cam03"`, `"t_sec": 12`, "clip1.mp4"} {
		if !strings.Contains(body, want) {
			t.Fatalf("sidecar missing %q in %s", want, body)
		}
	}
	// Annotation bar changes bottom pixels vs solid red source.
	if !bottomDiffers(t, src, frame) {
		t.Fatal("expected annotated pixels to differ at bottom")
	}
}

func writeSolidJPEG(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "src.jpg")
	img := image.NewRGBA(image.Rect(0, 0, 160, 120))
	for y := 0; y < 120; y++ {
		for x := 0; x < 160; x++ {
			img.Set(x, y, color.RGBA{200, 0, 0, 255})
		}
	}
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(f, img, nil); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	return p
}

func bottomDiffers(t *testing.T, a, b string) bool {
	t.Helper()
	ia := decode(t, a)
	ib := decode(t, b)
	ba, bb := ia.Bounds(), ib.Bounds()
	y := ba.Max.Y - 2
	for x := ba.Min.X; x < ba.Max.X && x < bb.Max.X; x++ {
		ca := color.RGBAModel.Convert(ia.At(x, y)).(color.RGBA)
		cb := color.RGBAModel.Convert(ib.At(x, y)).(color.RGBA)
		if ca != cb {
			return true
		}
	}
	return false
}

func decode(t *testing.T, path string) image.Image {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := jpeg.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	return img
}
