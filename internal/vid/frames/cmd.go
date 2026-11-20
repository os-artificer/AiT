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

package frames

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/deps"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/discover"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/frames"

	"github.com/spf13/cobra"
)

// Name is the registered subcommand name.
const Name = "frames"

// New builds the `aitvid frames` command.
func New() *cobra.Command {
	var (
		outDir, ext          string
		recursive, jsonOut   bool
		fps                  float64
		maxFrames, maxVideos int
	)
	cmd := &cobra.Command{
		Use:           Name + " <path> [flags]",
		Short:         "Extract still frames from video file or directory",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if pkghelpdoc.HasHelpArg(args) {
				return WriteHelp(c.OutOrStdout())
			}
			if err := deps.MissingErr(deps.Check(deps.Need{Frames: true}, nil)); err != nil {
				return err
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
			videos, err := discover.ListVideos(args[0], discover.Options{
				Recursive: recursive,
				Exts:      exts,
				MaxVideos: maxVideos,
			})
			if err != nil {
				return err
			}
			baseOut := outDir
			if baseOut == "" {
				baseOut = filepath.Join(".", "aitvid-frames")
			}
			for _, video := range videos {
				dir := filepath.Join(baseOut, strings.TrimSuffix(filepath.Base(video), filepath.Ext(video)))
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return err
				}
				fr, err := frames.Extract(context.Background(), video, frames.Options{
					FPS:       fps,
					MaxFrames: maxFrames,
					OutDir:    dir,
				})
				if err != nil {
					fmt.Fprintf(c.ErrOrStderr(), "error: %s: %v\n", video, err)
					continue
				}
				times := make([]float64, len(fr))
				for i, f := range fr {
					times[i] = f.TSec
				}
				if jsonOut {
					enc := json.NewEncoder(c.OutOrStdout())
					_ = enc.Encode(map[string]any{
						"video":      video,
						"frames_dir": dir,
						"count":      len(fr),
						"times_sec":  times,
					})
					continue
				}
				fmt.Fprintf(c.OutOrStdout(), "%s\t%s\t%d\n", video, dir, len(fr))
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&outDir, "out-dir", "o", "./aitvid-frames", "output directory")
	cmd.Flags().Float64Var(&fps, "fps", 0.5, "frames per second")
	cmd.Flags().IntVar(&maxFrames, "max-frames", 30, "max frames per video")
	cmd.Flags().IntVar(&maxVideos, "max-videos", 0, "max videos (0 = unlimited)")
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "recurse directories")
	cmd.Flags().StringVar(&ext, "ext", "", "extension whitelist")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "JSON per video")
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
