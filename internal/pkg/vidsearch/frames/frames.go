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

// Package frames extracts still JPEGs from videos via ffmpeg.
package frames

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Options control Extract.
type Options struct {
	FPS         float64 // default 0.5
	MaxFrames   int     // default 30
	OutDir      string  // required temp or keep dir
	FFmpegPath  string
	FFprobePath string
	FileTimeout time.Duration // default 2m
}

// Frame is one extracted still.
type Frame struct {
	Path string
	TSec float64
}

// ProbeInfo is duration metadata from ffprobe.
type ProbeInfo struct {
	DurationSec float64
}

// Probe returns video duration via ffprobe.
func Probe(ctx context.Context, video string, ffprobe string) (ProbeInfo, error) {
	if ffprobe == "" {
		ffprobe = "ffprobe"
	}
	cmd := exec.CommandContext(ctx, ffprobe,
		"-v", "error",
		"-print_format", "json",
		"-show_format",
		video,
	)
	out, err := cmd.Output()
	if err != nil {
		return ProbeInfo{}, fmt.Errorf("ffprobe: %w", err)
	}
	var parsed struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		return ProbeInfo{}, err
	}
	dur, _ := strconv.ParseFloat(parsed.Format.Duration, 64)
	return ProbeInfo{DurationSec: dur}, nil
}

// Extract writes JPEG frames under opts.OutDir.
func Extract(ctx context.Context, video string, opts Options) ([]Frame, error) {
	if opts.OutDir == "" {
		return nil, fmt.Errorf("frames: OutDir required")
	}
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return nil, err
	}
	fps := opts.FPS
	if fps <= 0 {
		fps = 0.5
	}
	maxF := opts.MaxFrames
	if maxF <= 0 {
		maxF = 30
	}
	ffmpeg := opts.FFmpegPath
	if ffmpeg == "" {
		ffmpeg = "ffmpeg"
	}
	ffprobe := opts.FFprobePath
	if ffprobe == "" {
		ffprobe = "ffprobe"
	}
	to := opts.FileTimeout
	if to <= 0 {
		to = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, to)
	defer cancel()

	info, err := Probe(ctx, video, ffprobe)
	if err != nil {
		return nil, err
	}
	dur := info.DurationSec
	if dur <= 0 {
		dur = float64(maxF) / fps
	}

	interval := 1.0 / fps
	times := make([]float64, 0, maxF)
	for t := 0.0; t < dur && len(times) < maxF; t += interval {
		times = append(times, t)
	}
	if len(times) == 0 {
		times = []float64{0}
	}

	pattern := filepath.Join(opts.OutDir, "frame_%04d.jpg")
	// fps filter capped by selecting every Nth via fps=
	vf := fmt.Sprintf("fps=%g", fps)
	cmd := exec.CommandContext(ctx, ffmpeg,
		"-hide_banner", "-loglevel", "error",
		"-i", video,
		"-vf", vf,
		"-frames:v", strconv.Itoa(maxF),
		"-q:v", "2",
		pattern,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ffmpeg extract: %w (%s)", err, strings.TrimSpace(string(out)))
	}

	entries, err := filepath.Glob(filepath.Join(opts.OutDir, "frame_*.jpg"))
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("ffmpeg produced no frames")
	}
	// Sort by name (frame_0001…)
	sortPaths(entries)
	frames := make([]Frame, 0, len(entries))
	for i, p := range entries {
		t := 0.0
		if i < len(times) {
			t = times[i]
		} else {
			t = float64(i) * interval
		}
		frames = append(frames, Frame{Path: p, TSec: round1(t)})
	}
	return frames, nil
}

func sortPaths(paths []string) {
	for i := 0; i < len(paths); i++ {
		for j := i + 1; j < len(paths); j++ {
			if paths[j] < paths[i] {
				paths[i], paths[j] = paths[j], paths[i]
			}
		}
	}
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
