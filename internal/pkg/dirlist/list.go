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

package dirlist

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// WriteLongListing writes an ls -la style directory listing for dir to w.
// dir must be an existing directory path.
func WriteLongListing(w io.Writer, dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", dir)
	}

	entries, err := collectEntries(dir)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "total %d\n", totalBlocks(entries)); err != nil {
		return err
	}

	for _, e := range entries {
		path := e.name
		if e.name != dirDot && e.name != dirDotDot {
			path = filepath.Join(dir, e.name)
		} else if e.name == dirDotDot {
			path = filepath.Dir(dir)
		} else {
			path = dir
		}
		if err := formatLine(w, path, e.name, e.info); err != nil {
			return err
		}
	}
	return nil
}

func collectEntries(dir string) ([]entry, error) {
	var entries []entry

	dotInfo, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	entries = append(entries, entry{name: dirDot, info: dotInfo})

	parent := filepath.Dir(dir)
	parentInfo, err := os.Stat(parent)
	if err != nil {
		return nil, err
	}
	entries = append(entries, entry{name: dirDotDot, info: parentInfo})

	names, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, de := range names {
		fi, err := de.Info()
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry{name: de.Name(), info: fi})
	}

	sortEntries(entries)
	return entries, nil
}
