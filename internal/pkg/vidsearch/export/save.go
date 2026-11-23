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

// Package export persists hit keyframes with provenance metadata.
package export

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/imgmeta"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/hit"
)

// Options control SaveHitFrame.
type Options struct {
	OutDir   string
	Annotate bool
	CameraID string // pre-resolved
}

// Sidecar is the machine-readable SSOT next to a saved JPEG.
type Sidecar struct {
	Video     string  `json:"video"`
	VideoName string  `json:"video_name"`
	CameraID  string  `json:"camera_id"`
	TSec      float64 `json:"t_sec"`
	Query     string  `json:"query"`
	Kind      string  `json:"kind"`
	Snippet   string  `json:"snippet"`
	Score     float64 `json:"score"`
	Mode      string  `json:"mode"`
	SavedAt   string  `json:"saved_at"`
	Frame     string  `json:"frame"`
}

// SaveHitFrame copies hit.Frame into OutDir with sidecar (+ optional annotate/EXIF).
func SaveHitFrame(h hit.Hit, srcFrame string, opts Options) (framePath, sidecarPath string, err error) {
	if opts.OutDir == "" {
		return "", "", fmt.Errorf("export: OutDir required")
	}
	cam := opts.CameraID
	if cam == "" {
		cam = "unknown"
	}
	stem := strings.TrimSuffix(filepath.Base(h.VideoName), filepath.Ext(h.VideoName))
	if stem == "" {
		stem = "video"
	}
	short := shortID(h.Video, h.TSec, h.Query)
	base := fmt.Sprintf("%s_t%.1fs_%s", sanitize(stem), h.TSec, short)
	dir := filepath.Join(opts.OutDir, sanitize(cam))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	framePath = filepath.Join(dir, base+".jpg")
	sidecarPath = filepath.Join(dir, base+".json")

	if err := copyFile(srcFrame, framePath); err != nil {
		return "", "", err
	}

	if opts.Annotate {
		lines := []string{
			fmt.Sprintf("video=%s camera=%s t=%.1fs", h.VideoName, cam, h.TSec),
			fmt.Sprintf("query=%s kind=%s", h.Query, h.Kind),
		}
		if annErr := AnnotateJPEG(framePath, lines); annErr != nil {
			// Keep unannotated frame.
			_ = annErr
		}
	}

	now := time.Now()
	soft := "aitvid"
	_ = imgmeta.Patch(framePath, imgmeta.PatchFields{
		Software: &soft,
		DateTime: &now,
	})

	sc := Sidecar{
		Video:     h.Video,
		VideoName: h.VideoName,
		CameraID:  cam,
		TSec:      h.TSec,
		Query:     h.Query,
		Kind:      h.Kind,
		Snippet:   h.Snippet,
		Score:     h.Score,
		Mode:      h.Mode,
		SavedAt:   now.Format(time.RFC3339),
		Frame:     framePath,
	}
	data, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return framePath, "", err
	}
	if err := os.WriteFile(sidecarPath, append(data, '\n'), 0o644); err != nil {
		return framePath, "", err
	}
	return framePath, sidecarPath, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func shortID(video string, t float64, query string) string {
	h := sha1.Sum([]byte(fmt.Sprintf("%s|%.1f|%s", video, t, query)))
	return hex.EncodeToString(h[:4])
}

func sanitize(s string) string {
	s = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			return r
		default:
			return '_'
		}
	}, s)
	if s == "" {
		return "x"
	}
	return s
}
