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

package platform

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var warnLegacyDataConflictOnce sync.Once

const (
	legacyConflictIntro  = "AiT: legacy mixed-case directory was not migrated; using ait directory only."
	legacyConflictAction = "move or merge files from legacy into current, then remove the legacy directory"
	legacyConflictHint   = "set AIT_HOME to pick one directory explicitly"
	legacyConflictNone   = "none"
)

func warnLegacyDataConflict(legacyRoot, newRoot string) {
	warnLegacyDataConflictOnce.Do(func() {
		writeLegacyDataConflictWarn(os.Stderr, legacyRoot, newRoot)
	})
}

func writeLegacyDataConflictWarn(w io.Writer, legacyRoot, newRoot string) {
	_, _ = fmt.Fprintf(w, "%s\r\n", legacyConflictIntro)
	_, _ = fmt.Fprintf(w, "  legacy:  %s (%s)\r\n", formatDisplayPath(legacyRoot), formatToolDataList(legacyRoot))
	_, _ = fmt.Fprintf(w, "  current: %s (%s)\r\n", formatDisplayPath(newRoot), formatToolDataList(newRoot))
	_, _ = fmt.Fprintf(w, "  action:  %s\r\n", legacyConflictAction)
	_, _ = fmt.Fprintf(w, "  hint:    %s\r\n", legacyConflictHint)
}

func formatToolDataList(root string) string {
	names := listToolDataFiles(root)
	if len(names) == 0 {
		return legacyConflictNone
	}
	return strings.Join(names, ", ")
}

func listToolDataFiles(root string) []string {
	var names []string
	if _, err := os.Stat(filepath.Join(root, dbFile)); err == nil {
		names = append(names, dbFile)
	}
	if _, err := os.Stat(filepath.Join(root, ConfigFileName)); err == nil {
		names = append(names, ConfigFileName)
	}
	return names
}

func pathsSameDir(a, b string) bool {
	if a == b {
		return true
	}

	sa, errA := os.Stat(a)
	if errA != nil {
		return false
	}
	sb, errB := os.Stat(b)
	if errB != nil {
		return false
	}
	return os.SameFile(sa, sb)
}

func formatDisplayPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	rel, err := filepath.Rel(home, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return path
	}
	return filepath.Join("~", rel)
}
