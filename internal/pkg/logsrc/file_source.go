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

package logsrc

import (
	"context"
	"fmt"
	"path/filepath"
)

// FileSource reads plain, rotated and gzipped text log files. Paths may
// contain globs; rotations next to a base path are picked up automatically.
type FileSource struct {
	Paths  []string
	Label  string
	Follow bool
}

// Name returns the source label, or the first path's basename.
func (s *FileSource) Name() string {
	if s.Label != "" {
		return s.Label
	}
	if len(s.Paths) > 0 {
		return filepath.Base(s.Paths[0])
	}
	return "file"
}

// SupportsReverse reports true: OpenReverse scans files backwards.
func (s *FileSource) SupportsReverse() bool { return true }

// Open opens a forward cursor over the resolved rotation files, oldest first.
// Returns ErrNoSource when no path exists, ErrPermission when a file cannot be
// read, or a glob error.
func (s *FileSource) Open(ctx context.Context, q Query) (Cursor, error) {
	paths, err := s.resolve()
	if err != nil {
		return nil, err
	}
	return openFileForward(ctx, paths, q.Follow || s.Follow)
}

// OpenReverse opens a newest-first cursor: the newest rotation file is scanned
// backwards in fixed-size chunks, then the next older one. Returns the same
// errors as Open.
func (s *FileSource) OpenReverse(ctx context.Context, q Query) (Cursor, error) {
	paths, err := s.resolve()
	if err != nil {
		return nil, err
	}
	newestFirst := make([]string, 0, len(paths))
	for i := len(paths) - 1; i >= 0; i-- {
		newestFirst = append(newestFirst, paths[i])
	}
	return openFileReverse(ctx, newestFirst)
}

func (s *FileSource) resolve() ([]string, error) {
	paths, err := ResolveLogFiles(s.Paths)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("%w: no files", ErrNoSource)
	}
	return paths, nil
}
