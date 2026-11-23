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

package vidsearch

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/deps"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/discover"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/export"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/format"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/frames"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/ocr"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/vision"
)

// ErrVisionRequired is returned for object mode without a vision backend.
var ErrVisionRequired = errhint.NewSentinel("vision model required")

// ErrPartialFailure indicates some videos failed while others succeeded.
var ErrPartialFailure = errhint.NewSentinel("partial video search failure")

// SearchOptions configure RunSearch.
type SearchOptions struct {
	Path       string
	Query      string
	Mode       string
	LocalOnly  bool
	FPS        float64
	MaxFrames  int
	MaxVideos  int
	Recursive  bool
	Exts       []string
	FailFast   bool
	KeepFrames bool
	SaveHits   bool
	OutDir     string
	CameraID   string
	Annotate   bool
	Timeout    time.Duration
	JSON       bool
	Verbose    bool
	Stdout     io.Writer
	Stderr     io.Writer
	LookPath   deps.LookPathFunc
	Vision     vision.Analyzer // optional; nil skips VLM
	// Extract overrides frames.Extract (tests).
	Extract func(ctx context.Context, video string, opts frames.Options) ([]frames.Frame, error)
	// OCRSearch overrides tesseract OCR (tests). When set, hasTess is treated as true.
	OCRSearch func(ctx context.Context, fr []frames.Frame, query, mode string) []Hit
}

// RunSearch discovers videos, extracts frames, OCR/VLM searches, and formats output.
func RunSearch(ctx context.Context, opts SearchOptions) error {
	if opts.Stdout == nil {
		opts.Stdout = io.Discard
	}
	if opts.Stderr == nil {
		opts.Stderr = io.Discard
	}
	mode := opts.Mode
	if mode == "" {
		mode = ModeAuto
	}

	needOCRHard := opts.LocalOnly || mode == ModeText
	rep := deps.Check(deps.Need{Frames: true, OCR: true, OCRHard: needOCRHard}, opts.LookPath)
	if err := deps.MissingErr(rep); err != nil {
		return err
	}
	hasTess := rep.HasTesseract() || opts.OCRSearch != nil

	if mode == ModeObject && (opts.LocalOnly || opts.Vision == nil) {
		return errhint.WithTry(ErrVisionRequired,
			"Object search needs a configured vision model (P1) or omit --local-only.",
			"ait agent config set ai.vision.model <vision-model>",
			"aitvid search … --mode text --local-only",
		)
	}

	to := opts.Timeout
	if to <= 0 {
		to = 10 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, to)
	defer cancel()

	videos, err := discover.ListVideos(opts.Path, discover.Options{
		Recursive: opts.Recursive,
		Exts:      opts.Exts,
		MaxVideos: opts.MaxVideos,
	})
	if err != nil {
		return err
	}

	var (
		results    []format.VideoResult
		allHits    []Hit
		failCount  int
		hitCount   int
		noticeAuto bool
	)

	for _, video := range videos {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		cam := export.ResolveCameraID(opts.CameraID, video)
		vr, hits, vErr := searchOne(ctx, video, cam, hasTess, &noticeAuto, opts)
		results = append(results, vr)
		if vErr != nil {
			failCount++
			fmt.Fprintf(opts.Stderr, "error: %s: %v\n", video, vErr)
			if opts.FailFast {
				return vErr
			}
			continue
		}
		allHits = append(allHits, hits...)
		hitCount += len(hits)
	}

	if noticeAuto {
		fmt.Fprintf(opts.Stderr, "notice: vision unavailable; auto mode used OCR only\n")
	}

	sum := format.Summary{
		Videos: len(videos),
		Hits:   hitCount,
		Failed: failCount,
	}
	if opts.SaveHits {
		sum.OutDir = opts.OutDir
	}

	if opts.JSON {
		if err := format.WriteNDJSON(opts.Stdout, allHits); err != nil {
			return err
		}
	} else {
		if err := format.WriteHuman(opts.Stdout, results, sum); err != nil {
			return err
		}
	}

	if failCount > 0 && failCount < len(videos) {
		return ErrPartialFailure
	}
	if failCount > 0 {
		return fmt.Errorf("%w: all videos failed", ErrPartialFailure)
	}
	return nil
}

func searchOne(ctx context.Context, video, cam string, hasTess bool, noticeAuto *bool, opts SearchOptions) (format.VideoResult, []Hit, error) {
	vr := format.VideoResult{Video: video, CameraID: cam}
	tmp, err := os.MkdirTemp("", "aitvid-frames-*")
	if err != nil {
		return vr, nil, err
	}
	defer func() {
		if !opts.KeepFrames {
			_ = os.RemoveAll(tmp)
		}
	}()

	extract := opts.Extract
	if extract == nil {
		extract = frames.Extract
	}
	fr, err := extract(ctx, video, frames.Options{
		FPS:       opts.FPS,
		MaxFrames: opts.MaxFrames,
		OutDir:    tmp,
	})
	if err != nil {
		vr.Err = err
		return vr, nil, err
	}
	if opts.Verbose {
		fmt.Fprintf(opts.Stderr, "frames: %s count=%d dir=%s\n", video, len(fr), tmp)
	}

	mode := opts.Mode
	if mode == "" {
		mode = ModeAuto
	}
	var hits []Hit

	if hasTess && mode != ModeObject {
		if opts.OCRSearch != nil {
			hits = opts.OCRSearch(ctx, fr, opts.Query, mode)
		} else {
			idx, _ := ocr.Build(ctx, fr, "tesseract")
			hits = ocr.Search(idx, opts.Query, mode)
		}
	} else if mode == ModeText || opts.LocalOnly {
		return vr, nil, deps.MissingErr(deps.Check(deps.Need{OCRHard: true}, opts.LookPath))
	}

	needVision := mode == ModeObject || (mode == ModeAuto && len(hits) == 0)
	if needVision {
		if opts.LocalOnly || opts.Vision == nil {
			if mode == ModeAuto {
				*noticeAuto = true
			}
		} else {
			vh, vErr := opts.Vision.Analyze(ctx, vision.Request{
				Query:  opts.Query,
				Frames: fr,
				Mode:   mode,
			})
			if vErr != nil {
				return vr, nil, vErr
			}
			hits = append(hits, vh...)
		}
	}

	base := filepath.Base(video)
	outHits := make([]Hit, 0, len(hits))
	for _, h := range hits {
		h.Video = video
		h.VideoName = base
		h.CameraID = cam
		if h.Mode == "" {
			h.Mode = mode
		}
		srcFrame := h.Frame
		if opts.SaveHits && srcFrame != "" {
			outDir := opts.OutDir
			if outDir == "" {
				outDir = "./aitvid-hits"
			}
			fp, sp, sErr := export.SaveHitFrame(h, srcFrame, export.Options{
				OutDir:   outDir,
				Annotate: opts.Annotate,
				CameraID: cam,
			})
			if sErr != nil {
				fmt.Fprintf(opts.Stderr, "error: save frame %s: %v\n", video, sErr)
			} else {
				h.Frame = fp
				h.Sidecar = sp
			}
		} else if !opts.KeepFrames {
			h.Frame = ""
		}
		outHits = append(outHits, h)
	}
	vr.Hits = outHits
	return vr, outHits, nil
}
