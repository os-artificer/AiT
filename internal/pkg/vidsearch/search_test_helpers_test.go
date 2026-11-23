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
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/vidsearch/frames"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/hit"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/vision"
)

type countingVision struct {
	calls atomic.Int32
	hits  []hit.Hit
	err   error
}

func (c *countingVision) Analyze(ctx context.Context, req vision.Request) ([]hit.Hit, error) {
	c.calls.Add(1)
	if c.err != nil {
		return nil, c.err
	}
	return c.hits, nil
}

func lookAllOK(name string) (string, error) { return "/bin/" + name, nil }

func stubExtract(_ context.Context, _ string, opts frames.Options) ([]frames.Frame, error) {
	p := filepath.Join(opts.OutDir, "frame_0001.jpg")
	_ = os.WriteFile(p, []byte("x"), 0o644)
	return []frames.Frame{{Path: p, TSec: 1.0}}, nil
}

func stubExtractJPEG(t *testing.T) func(context.Context, string, frames.Options) ([]frames.Frame, error) {
	t.Helper()
	src := writeSolidJPEG(t)
	return func(_ context.Context, _ string, opts frames.Options) ([]frames.Frame, error) {
		p := filepath.Join(opts.OutDir, "frame_0001.jpg")
		data, err := os.ReadFile(src)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(p, data, 0o644); err != nil {
			return nil, err
		}
		return []frames.Frame{{Path: p, TSec: 1.0}}, nil
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

func writePlaceholderVideos(t *testing.T, dir string, names ...string) []string {
	t.Helper()
	var paths []string
	for _, name := range names {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, p)
	}
	return paths
}

func ocrHit(query string) func(context.Context, []frames.Frame, string, string) []hit.Hit {
	return func(_ context.Context, fr []frames.Frame, q, mode string) []hit.Hit {
		frame := ""
		if len(fr) > 0 {
			frame = fr[0].Path
		}
		use := query
		if use == "" {
			use = q
		}
		return []hit.Hit{{
			TSec: 1, Frame: frame, Kind: hit.KindText, Query: use,
			Snippet: use + "-hit", Score: 1, Mode: mode,
		}}
	}
}

func ocrMiss() func(context.Context, []frames.Frame, string, string) []hit.Hit {
	return func(context.Context, []frames.Frame, string, string) []hit.Hit { return nil }
}
