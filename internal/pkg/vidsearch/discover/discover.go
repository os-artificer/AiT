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

// Package discover lists video files under a path.
package discover

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// DefaultExts are container extensions accepted when scanning directories.
var DefaultExts = []string{".mp4", ".mov", ".mkv", ".webm", ".avi", ".m4v"}

// ErrNoVideos is returned when a directory has no matching videos.
var ErrNoVideos = errhint.NewSentinel("no video files found")

// ErrTooManyVideos is returned when discovery exceeds an explicit MaxVideos > 0.
var ErrTooManyVideos = errhint.NewSentinel("too many video files")

// Options control ListVideos.
type Options struct {
	Recursive bool
	Exts      []string // empty → DefaultExts
	MaxVideos int      // <=0 unlimited; >0 hard cap (ErrTooManyVideos)
}

// ListVideos returns video paths under path (file or directory).
func ListVideos(path string, opts Options) ([]string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	st, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}

	maxV := opts.MaxVideos // <=0 means unlimited
	exts := opts.Exts
	if len(exts) == 0 {
		exts = DefaultExts
	}
	extSet := make(map[string]struct{}, len(exts))
	for _, e := range exts {
		e = strings.ToLower(e)
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		extSet[e] = struct{}{}
	}

	if !st.IsDir() {
		// Explicit file: always include regardless of extension whitelist.
		return []string{abs}, nil
	}

	var out []string
	walkFn := func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		name := d.Name()
		if strings.HasPrefix(name, ".") {
			if d.IsDir() && p != abs {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			if !opts.Recursive && p != abs {
				return fs.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(name))
		if _, ok := extSet[ext]; !ok {
			return nil
		}
		if maxV > 0 && len(out) >= maxV {
			return ErrTooManyVideos
		}
		out = append(out, p)
		return nil
	}

	if opts.Recursive {
		err = filepath.WalkDir(abs, walkFn)
	} else {
		entries, rdErr := os.ReadDir(abs)
		if rdErr != nil {
			return nil, rdErr
		}
		for _, e := range entries {
			p := filepath.Join(abs, e.Name())
			if walkErr := walkFn(p, e, nil); walkErr != nil {
				if errors.Is(walkErr, fs.SkipDir) {
					continue
				}
				err = walkErr
				break
			}
		}
	}
	if errors.Is(err, ErrTooManyVideos) || err == ErrTooManyVideos {
		return nil, errhint.WithTry(ErrTooManyVideos,
			fmt.Sprintf("Found more than %d videos under %s.", maxV, abs),
			toolname.VidCLI(fmt.Sprintf("search %s --max-videos %d", path, maxV*2)),
			toolname.VidCLI("search "+path+" --ext mp4"),
		)
	}
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, errhint.WithTry(ErrNoVideos,
			"No matching video files in the directory.",
			toolname.VidCLI("search "+path+" -r"),
			toolname.VidCLI("search "+path+" --ext mp4,mov,mkv"),
		)
	}
	return out, nil
}
