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
	"os"
	"path/filepath"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/vidsearch"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/deps"
)

func TestRunSearch_MissingFFmpeg(t *testing.T) {
	look := func(string) (string, error) { return "", errors.New("x") }
	err := vidsearch.RunSearch(context.Background(), vidsearch.SearchOptions{
		Path: t.TempDir(), Query: "x", Mode: vidsearch.ModeText, LocalOnly: true,
		LookPath: look, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{},
	})
	if !errors.Is(err, deps.ErrMissingFFmpeg) {
		t.Fatalf("got %v", err)
	}
}

func TestRunSearch_ObjectNeedsVision(t *testing.T) {
	look := func(name string) (string, error) {
		return "/bin/" + name, nil
	}
	err := vidsearch.RunSearch(context.Background(), vidsearch.SearchOptions{
		Path: t.TempDir(), Query: "bike", Mode: vidsearch.ModeObject, LocalOnly: true,
		LookPath: look, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{},
	})
	if !errors.Is(err, vidsearch.ErrVisionRequired) {
		t.Fatalf("got %v", err)
	}
}

func TestRunSearch_JSONZeroHitsEmptyStdout(t *testing.T) {
	// Early-fail path (missing deps): stdout stays empty.
	look := func(string) (string, error) { return "", errors.New("x") }
	var out, errBuf bytes.Buffer
	_ = vidsearch.RunSearch(context.Background(), vidsearch.SearchOptions{
		Path: t.TempDir(), Query: "x", Mode: vidsearch.ModeText, LocalOnly: true, JSON: true,
		LookPath: look, Stdout: &out, Stderr: &errBuf,
	})
	if out.Len() != 0 {
		t.Fatalf("stdout=%q", out.String())
	}
}

func TestRunSearch_JSONZeroHits_AfterPipeline(t *testing.T) {
	vid := filepath.Join(t.TempDir(), "a.mp4")
	if err := os.WriteFile(vid, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := vidsearch.RunSearch(context.Background(), vidsearch.SearchOptions{
		Path: vid, Query: "nomatch", Mode: vidsearch.ModeText, JSON: true, SaveHits: false,
		LookPath: lookAllOK, Extract: stubExtract, OCRSearch: ocrMiss(),
		Stdout: &out, Stderr: &bytes.Buffer{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 {
		t.Fatalf("stdout=%q", out.String())
	}
}

func TestLocalOnly_MissingTesseract(t *testing.T) {
	look := func(name string) (string, error) {
		if name == "tesseract" {
			return "", errors.New("missing")
		}
		return "/bin/" + name, nil
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "a.mp4")
	if err := os.WriteFile(p, []byte("not-a-video"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := vidsearch.RunSearch(context.Background(), vidsearch.SearchOptions{
		Path: p, Query: "x", Mode: vidsearch.ModeText, LocalOnly: true,
		LookPath: look, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{},
	})
	if !errors.Is(err, deps.ErrMissingTesseract) {
		t.Fatalf("got %v", err)
	}
}
