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
	"os"
	"path/filepath"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/vidsearch"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/frames"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/hit"
)

func TestAuto_OCRHitSkipsVLM(t *testing.T) {
	vid := filepath.Join(t.TempDir(), "a.mp4")
	if err := os.WriteFile(vid, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	fv := &countingVision{hits: []hit.Hit{{Kind: hit.KindObject, Snippet: "should-not"}}}
	var out bytes.Buffer
	err := vidsearch.RunSearch(context.Background(), vidsearch.SearchOptions{
		Path: vid, Query: "京A", Mode: vidsearch.ModeAuto, JSON: true, SaveHits: false,
		LookPath: lookAllOK, Vision: fv, Extract: stubExtract,
		OCRSearch: func(ctx context.Context, fr []frames.Frame, query, mode string) []hit.Hit {
			return []hit.Hit{{TSec: 1, Frame: fr[0].Path, Kind: hit.KindText, Query: query, Snippet: "京A12345", Score: 1, Mode: mode}}
		},
		Stdout: &out, Stderr: &bytes.Buffer{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if fv.calls.Load() != 0 {
		t.Fatalf("vision calls=%d", fv.calls.Load())
	}
	if !bytes.Contains(out.Bytes(), []byte("京A12345")) {
		t.Fatalf("stdout=%s", out.String())
	}
}

func TestAuto_OCRMissCallsVLM(t *testing.T) {
	vid := filepath.Join(t.TempDir(), "a.mp4")
	if err := os.WriteFile(vid, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	fv := &countingVision{hits: []hit.Hit{{
		TSec: 1, Kind: hit.KindObject, Query: "bike", Snippet: "red bike", Score: 0.8, Mode: hit.ModeAuto,
	}}}
	var out bytes.Buffer
	err := vidsearch.RunSearch(context.Background(), vidsearch.SearchOptions{
		Path: vid, Query: "bike", Mode: vidsearch.ModeAuto, JSON: true, SaveHits: false,
		LookPath: lookAllOK, Vision: fv, Extract: stubExtract,
		OCRSearch: ocrMiss(),
		Stdout:    &out, Stderr: &bytes.Buffer{},
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
