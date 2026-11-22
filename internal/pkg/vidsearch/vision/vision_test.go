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

package vision_test

import (
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/vidsearch/frames"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/hit"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/vision"
)

func TestNewClient_RejectsTextModel(t *testing.T) {
	c, err := vision.NewClient(vision.Config{
		BaseURL: "http://example.invalid",
		Model:   "deepseek-chat",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Analyze(context.Background(), vision.Request{
		Query:  "bike",
		Frames: []frames.Frame{{Path: writeJPEG(t), TSec: 1}},
		Mode:   hit.ModeObject,
	})
	if !errors.Is(err, vision.ErrNotVisionModel) {
		t.Fatalf("got %v", err)
	}
}

func TestAnalyze_MockHTTPHit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]string{
					"content": `[{"i":0,"hit":true,"snippet":"red bike"}]`,
				},
			}},
		})
	}))
	defer srv.Close()

	c, err := vision.NewClient(vision.Config{
		BaseURL:     srv.URL,
		Model:       "gpt-4o-mini",
		ForceVision: true,
		HTTPClient:  srv.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	hits, err := c.Analyze(context.Background(), vision.Request{
		Query:  "bike",
		Frames: []frames.Frame{{Path: writeJPEG(t), TSec: 2.5}},
		Mode:   hit.ModeObject,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].Kind != hit.KindObject || hits[0].Snippet != "red bike" {
		t.Fatalf("hits=%+v", hits)
	}
}

func TestAnalyze_BatchesFour(t *testing.T) {
	var requests int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]string{"content": `[]`},
			}},
		})
	}))
	defer srv.Close()

	c, err := vision.NewClient(vision.Config{
		BaseURL:     srv.URL,
		Model:       "gpt-4o-mini",
		ForceVision: true,
		HTTPClient:  srv.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	fr := make([]frames.Frame, 9)
	for i := range fr {
		fr[i] = frames.Frame{Path: writeJPEG(t), TSec: float64(i)}
	}
	if _, err := c.Analyze(context.Background(), vision.Request{
		Query: "x", Frames: fr, Mode: hit.ModeObject,
	}); err != nil {
		t.Fatal(err)
	}
	if requests != 3 {
		t.Fatalf("batches=%d want 3", requests)
	}
}

func writeJPEG(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "f.jpg")
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	img.Set(0, 0, color.RGBA{1, 2, 3, 255})
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
