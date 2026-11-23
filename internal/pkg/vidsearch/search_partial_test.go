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
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/vidsearch"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/frames"
)

func TestRunSearch_PartialFailure_SomeFail(t *testing.T) {
	dir := t.TempDir()
	writePlaceholderVideos(t, dir, "ok.mp4", "bad.mp4")

	var out, errBuf bytes.Buffer
	err := vidsearch.RunSearch(context.Background(), vidsearch.SearchOptions{
		Path: dir, Query: "京A", Mode: vidsearch.ModeText, JSON: true, SaveHits: false,
		LookPath: lookAllOK, Extract: func(ctx context.Context, video string, opts frames.Options) ([]frames.Frame, error) {
			if strings.HasSuffix(video, "bad.mp4") {
				return nil, fmt.Errorf("extract failed")
			}
			return stubExtract(ctx, video, opts)
		},
		OCRSearch: ocrHit("京A"),
		Stdout:    &out, Stderr: &errBuf,
	})
	if !errors.Is(err, vidsearch.ErrPartialFailure) {
		t.Fatalf("got %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("京A-hit")) {
		t.Fatalf("stdout missing hit: %s", out.String())
	}
	if !strings.Contains(errBuf.String(), "bad.mp4") {
		t.Fatalf("stderr=%s", errBuf.String())
	}
}

func TestRunSearch_PartialFailure_AllFail(t *testing.T) {
	dir := t.TempDir()
	writePlaceholderVideos(t, dir, "a.mp4", "b.mp4")

	err := vidsearch.RunSearch(context.Background(), vidsearch.SearchOptions{
		Path: dir, Query: "x", Mode: vidsearch.ModeText, JSON: true, SaveHits: false,
		LookPath: lookAllOK,
		Extract: func(context.Context, string, frames.Options) ([]frames.Frame, error) {
			return nil, fmt.Errorf("boom")
		},
		OCRSearch: ocrMiss(),
		Stdout:    &bytes.Buffer{}, Stderr: &bytes.Buffer{},
	})
	if !errors.Is(err, vidsearch.ErrPartialFailure) {
		t.Fatalf("got %v", err)
	}
	if !strings.Contains(err.Error(), "all videos failed") {
		t.Fatalf("want all videos failed: %v", err)
	}
}

func TestRunSearch_FailFast_StopsEarly(t *testing.T) {
	dir := t.TempDir()
	writePlaceholderVideos(t, dir, "a.mp4", "b.mp4")

	var calls atomic.Int32
	err := vidsearch.RunSearch(context.Background(), vidsearch.SearchOptions{
		Path: dir, Query: "x", Mode: vidsearch.ModeText, FailFast: true, JSON: true, SaveHits: false,
		LookPath: lookAllOK,
		Extract: func(_ context.Context, video string, _ frames.Options) ([]frames.Frame, error) {
			calls.Add(1)
			return nil, fmt.Errorf("fail %s", filepath.Base(video))
		},
		OCRSearch: ocrMiss(),
		Stdout:    &bytes.Buffer{}, Stderr: &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, vidsearch.ErrPartialFailure) {
		t.Fatalf("fail-fast should return first extract err, got %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("extract calls=%d want 1", got)
	}
}
