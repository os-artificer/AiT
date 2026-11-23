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

// Package format renders search hits for humans or NDJSON.
package format

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/os-artificer/AiT/internal/pkg/vidsearch/hit"
)

// VideoResult groups hits for one video in human output.
type VideoResult struct {
	Video    string
	CameraID string
	Hits     []hit.Hit
	Err      error
}

// Summary counts for the human footer.
type Summary struct {
	Videos int
	Hits   int
	Failed int
	OutDir string
}

// WriteHuman writes grouped human-readable results to w.
func WriteHuman(w io.Writer, results []VideoResult, sum Summary) error {
	for _, r := range results {
		cam := r.CameraID
		if cam == "" {
			cam = "-"
		}
		if _, err := fmt.Fprintf(w, "# %s  camera=%s\n", r.Video, cam); err != nil {
			return err
		}
		if len(r.Hits) == 0 {
			if _, err := fmt.Fprintf(w, "  (no hits)\n\n"); err != nil {
				return err
			}
			continue
		}
		for _, h := range r.Hits {
			frame := h.Frame
			if frame != "" {
				if rel, err := filepath.Rel(".", frame); err == nil && len(rel) < len(frame) {
					frame = rel
				}
			}
			if _, err := fmt.Fprintf(w, "  %.1fs  %s   %q  score=%.2f  frame=%s\n",
				h.TSec, h.Kind, h.Snippet, h.Score, frame); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}
	line := fmt.Sprintf("---\n%d videos, %d hits, %d failed", sum.Videos, sum.Hits, sum.Failed)
	if sum.OutDir != "" {
		line += "; saved " + sum.OutDir
	}
	_, err := fmt.Fprintln(w, line)
	return err
}

// WriteNDJSON writes one JSON object per hit (no wrapper).
func WriteNDJSON(w io.Writer, hits []hit.Hit) error {
	enc := json.NewEncoder(w)
	for _, h := range hits {
		if err := enc.Encode(h); err != nil {
			return err
		}
	}
	return nil
}
