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

package vidsearch_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/vidsearch"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/discover"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/frames"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/hit"
)

func TestRunSearch_Timeout_Cancels(t *testing.T) {
	dir := t.TempDir()
	writePlaceholderVideos(t, dir, "a.mp4", "b.mp4")

	err := vidsearch.RunSearch(context.Background(), vidsearch.SearchOptions{
		Path: dir, Query: "x", Mode: vidsearch.ModeText, JSON: true, SaveHits: false,
		Timeout:  40 * time.Millisecond,
		LookPath: lookAllOK,
		Extract: func(ctx context.Context, _ string, _ frames.Options) ([]frames.Frame, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
		OCRSearch: ocrMiss(),
		Stdout:    &bytes.Buffer{}, Stderr: &bytes.Buffer{},
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v", err)
	}
}

func TestRunSearch_ModeText_NoVision(t *testing.T) {
	vid := filepath.Join(t.TempDir(), "a.mp4")
	if err := os.WriteFile(vid, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	fv := &countingVision{hits: []hit.Hit{{Kind: hit.KindObject, Snippet: "nope"}}}
	var out bytes.Buffer
	err := vidsearch.RunSearch(context.Background(), vidsearch.SearchOptions{
		Path: vid, Query: "plate", Mode: vidsearch.ModeText, JSON: true, SaveHits: false,
		LookPath: lookAllOK, Vision: fv, Extract: stubExtract, OCRSearch: ocrHit("plate"),
		Stdout: &out, Stderr: &bytes.Buffer{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if fv.calls.Load() != 0 {
		t.Fatalf("vision calls=%d", fv.calls.Load())
	}
	if !bytes.Contains(out.Bytes(), []byte("plate-hit")) {
		t.Fatalf("stdout=%s", out.String())
	}
}

func TestRunSearch_ModeObject_WithVision(t *testing.T) {
	vid := filepath.Join(t.TempDir(), "a.mp4")
	if err := os.WriteFile(vid, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	fv := &countingVision{hits: []hit.Hit{{
		TSec: 1, Kind: hit.KindObject, Query: "bike", Snippet: "red bike", Score: 0.8, Mode: hit.ModeObject,
	}}}
	var out bytes.Buffer
	err := vidsearch.RunSearch(context.Background(), vidsearch.SearchOptions{
		Path: vid, Query: "bike", Mode: vidsearch.ModeObject, JSON: true, SaveHits: false,
		LookPath: lookAllOK, Vision: fv, Extract: stubExtract,
		Stdout: &out, Stderr: &bytes.Buffer{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if fv.calls.Load() != 1 {
		t.Fatalf("vision calls=%d", fv.calls.Load())
	}
	if !bytes.Contains(out.Bytes(), []byte("red bike")) {
		t.Fatalf("stdout=%s", out.String())
	}
}

func TestRunSearch_Auto_NoticeOCROnly(t *testing.T) {
	vid := filepath.Join(t.TempDir(), "a.mp4")
	if err := os.WriteFile(vid, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var errBuf bytes.Buffer
	err := vidsearch.RunSearch(context.Background(), vidsearch.SearchOptions{
		Path: vid, Query: "bike", Mode: vidsearch.ModeAuto, JSON: true, SaveHits: false,
		LookPath: lookAllOK, Vision: nil, Extract: stubExtract, OCRSearch: ocrMiss(),
		Stdout: &bytes.Buffer{}, Stderr: &errBuf,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errBuf.String(), "vision unavailable") {
		t.Fatalf("stderr=%s", errBuf.String())
	}
}

func TestRunSearch_SaveHits_Annotate(t *testing.T) {
	camDir := filepath.Join(t.TempDir(), "camera_7")
	if err := os.Mkdir(camDir, 0o755); err != nil {
		t.Fatal(err)
	}
	vid := filepath.Join(camDir, "clip.mp4")
	if err := os.WriteFile(vid, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	outDir := t.TempDir()
	err := vidsearch.RunSearch(context.Background(), vidsearch.SearchOptions{
		Path: vid, Query: "bike", Mode: vidsearch.ModeText, JSON: true,
		SaveHits: true, Annotate: true, OutDir: outDir,
		LookPath: lookAllOK, Extract: stubExtractJPEG(t), OCRSearch: ocrHit("bike"),
		Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{},
	})
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(outDir, "cam7"))
	if err != nil {
		t.Fatal(err)
	}
	var hasJPG, hasJSON bool
	for _, e := range entries {
		switch filepath.Ext(e.Name()) {
		case ".jpg":
			hasJPG = true
		case ".json":
			hasJSON = true
		}
	}
	if !hasJPG || !hasJSON {
		t.Fatalf("entries=%v", entries)
	}
}

func TestRunSearch_SaveHits_NoAnnotate(t *testing.T) {
	vid := filepath.Join(t.TempDir(), "clip.mp4")
	if err := os.WriteFile(vid, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	outDir := t.TempDir()
	err := vidsearch.RunSearch(context.Background(), vidsearch.SearchOptions{
		Path: vid, Query: "bike", Mode: vidsearch.ModeText, JSON: true,
		SaveHits: true, Annotate: false, OutDir: outDir, CameraID: "cam99",
		LookPath: lookAllOK, Extract: stubExtractJPEG(t), OCRSearch: ocrHit("bike"),
		Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{},
	})
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(outDir, "cam99"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 2 {
		t.Fatalf("entries=%v", entries)
	}
}

func TestRunSearch_CameraID_FromPath(t *testing.T) {
	camDir := filepath.Join(t.TempDir(), "camera_12")
	if err := os.Mkdir(camDir, 0o755); err != nil {
		t.Fatal(err)
	}
	vid := filepath.Join(camDir, "clip.mp4")
	if err := os.WriteFile(vid, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := vidsearch.RunSearch(context.Background(), vidsearch.SearchOptions{
		Path: vid, Query: "q", Mode: vidsearch.ModeText, JSON: true, SaveHits: false,
		LookPath: lookAllOK, Extract: stubExtract, OCRSearch: ocrHit("q"),
		Stdout: &out, Stderr: &bytes.Buffer{},
	})
	if err != nil {
		t.Fatal(err)
	}
	var row map[string]any
	if err := json.Unmarshal(out.Bytes(), &row); err != nil {
		t.Fatalf("json: %v stdout=%s", err, out.String())
	}
	if row["camera_id"] != "cam12" {
		t.Fatalf("camera_id=%v", row["camera_id"])
	}
}

func TestRunSearch_ErrNoVideos(t *testing.T) {
	err := vidsearch.RunSearch(context.Background(), vidsearch.SearchOptions{
		Path: t.TempDir(), Query: "x", Mode: vidsearch.ModeText, JSON: true, SaveHits: false,
		LookPath: lookAllOK, Extract: stubExtract, OCRSearch: ocrMiss(),
		Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{},
	})
	if !errors.Is(err, discover.ErrNoVideos) {
		t.Fatalf("got %v", err)
	}
	if !strings.Contains(err.Error(), "Try:") {
		t.Fatalf("missing Try block: %v", err)
	}
}
