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
	"io/fs"
	"os"
	"sort"
	"strings"
)

type entry struct {
	name string
	info os.FileInfo
}

func formatMode(info os.FileInfo) string {
	var prefix byte = '-'
	switch {
	case info.IsDir():
		prefix = 'd'
	case info.Mode()&fs.ModeSymlink != 0:
		prefix = 'l'
	}
	perm := info.Mode().Perm().String()
	if len(perm) > 1 {
		perm = perm[1:]
	}
	return string(prefix) + perm
}

func formatLine(w io.Writer, path, displayName string, info os.FileInfo) error {
	if info.Mode()&fs.ModeSymlink != 0 {
		if target, err := os.Readlink(path); err == nil {
			displayName = displayName + symlinkArrow + target
		}
	}

	owner, group := fileOwnerGroup(info)
	_, err := fmt.Fprintf(w, "%s %3d %-8s %-8s %8d %s %s\n",
		formatMode(info),
		fileNLink(info),
		owner,
		group,
		info.Size(),
		info.ModTime().Format(modTimeLayout),
		displayName,
	)
	return err
}

func totalBlocks(entries []entry) int64 {
	var sum int64
	for _, e := range entries {
		sum += e.info.Size()
	}
	return (sum + totalBlockSz - 1) / totalBlockSz
}

func sortEntries(entries []entry) {
	sort.Slice(entries, func(i, j int) bool {
		ni, nj := entries[i].name, entries[j].name
		if ni == dirDot {
			return true
		}
		if nj == dirDot {
			return false
		}
		if ni == dirDotDot {
			return true
		}
		if nj == dirDotDot {
			return false
		}
		return strings.Compare(ni, nj) < 0
	})
}
