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

package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/config"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/vision"

	"github.com/spf13/cobra"
)

// Name is the registered subcommand name.
const Name = "search"

// New builds the `aitvid search` command.
func New() *cobra.Command {
	var (
		query, mode, outDir, cameraID, ext                           string
		localOnly, recursive, failFast, keepFrames, jsonOut, verbose bool
		saveHits, annotate, noSaveHits, noAnnotate                   bool
		fps                                                          float64
		maxFrames, maxVideos                                         int
		timeout                                                      time.Duration
	)
	cmd := &cobra.Command{
		Use:           Name + " <path> [flags]",
		Short:         "Search video file or directory for query content",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if pkghelpdoc.HasHelpArg(args) {
				return WriteHelp(c.OutOrStdout())
			}
			if strings.TrimSpace(query) == "" {
				return fmt.Errorf("--query is required")
			}
			save := true
			if c.Flags().Changed("save-hits") {
				save = saveHits
			}
			if noSaveHits {
				save = false
			}
			ann := true
			if c.Flags().Changed("annotate") {
				ann = annotate
			}
			if noAnnotate {
				ann = false
			}
			var exts []string
			if ext != "" {
				for _, e := range strings.Split(ext, ",") {
					e = strings.TrimSpace(e)
					if e != "" {
						exts = append(exts, e)
					}
				}
			}
			opts := vidsearch.SearchOptions{
				Path:       args[0],
				Query:      query,
				Mode:       mode,
				LocalOnly:  localOnly,
				FPS:        fps,
				MaxFrames:  maxFrames,
				MaxVideos:  maxVideos,
				Recursive:  recursive,
				Exts:       exts,
				FailFast:   failFast,
				KeepFrames: keepFrames,
				SaveHits:   save,
				OutDir:     outDir,
				CameraID:   cameraID,
				Annotate:   ann,
				Timeout:    timeout,
				JSON:       jsonOut,
				Verbose:    verbose,
				Stdout:     c.OutOrStdout(),
				Stderr:     c.ErrOrStderr(),
			}
			if !localOnly && (mode == vidsearch.ModeObject || mode == vidsearch.ModeAuto || mode == "") {
				if va, err := maybeVision(); err == nil {
					opts.Vision = va
				}
			}
			return vidsearch.RunSearch(context.Background(), opts)
		},
	}
	cmd.Flags().StringVar(&query, "query", "", "search query (required)")
	cmd.Flags().StringVar(&mode, "mode", vidsearch.ModeAuto, "text|object|auto")
	cmd.Flags().BoolVar(&localOnly, "local-only", false, "OCR only; never call vision models")
	cmd.Flags().Float64Var(&fps, "fps", 0.5, "frames per second to sample")
	cmd.Flags().IntVar(&maxFrames, "max-frames", 30, "max frames per video")
	cmd.Flags().IntVar(&maxVideos, "max-videos", 0, "max videos in a directory (0 = unlimited)")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "wall-clock timeout for the whole search")
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "recurse into subdirectories")
	cmd.Flags().StringVar(&ext, "ext", "", "comma-separated extensions for directory scan")
	cmd.Flags().BoolVar(&failFast, "fail-fast", false, "stop on first video error")
	cmd.Flags().BoolVar(&keepFrames, "keep-frames", false, "keep temporary extracted frames")
	cmd.Flags().BoolVar(&saveHits, "save-hits", true, "save hit keyframes under --out-dir")
	cmd.Flags().BoolVar(&noSaveHits, "no-save-hits", false, "do not save hit keyframes")
	cmd.Flags().StringVarP(&outDir, "out-dir", "o", "./aitvid-hits", "hit frame output directory")
	cmd.Flags().StringVar(&cameraID, "camera-id", "", "camera id override")
	cmd.Flags().BoolVar(&annotate, "annotate", true, "burn provenance text onto saved frames")
	cmd.Flags().BoolVar(&noAnnotate, "no-annotate", false, "disable pixel annotation")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "NDJSON hit output")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "extra progress on stderr")
	_ = cmd.MarkFlagRequired("query")
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}

func maybeVision() (vision.Analyzer, error) {
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		return nil, fmt.Errorf("no config")
	}
	model := strings.TrimSpace(cfg.AI.Vision.Model)
	if model == "" {
		return nil, fmt.Errorf("no vision model")
	}
	// Reuse first enabled provider base URL / key when present.
	base, key := visionEndpoint(cfg)
	return vision.NewClient(vision.Config{
		BaseURL: base,
		APIKey:  key,
		Model:   model,
	})
}
