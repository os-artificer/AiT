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

package terminal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var scrollbackBaseName = regexp.MustCompile(`^[a-zA-Z0-9-]+` + regexp.QuoteMeta(ScrollbackSuffix) + `$`)

// ScrollbackFileName returns the canonical scrollback sidecar basename for sessionID
// (sessionID + ScrollbackSuffix).
func ScrollbackFileName(sessionID string) string {
	return sessionID + ScrollbackSuffix
}

// ResolveScrollbackPath maps a stored basename (or legacy full path) to a file
// under dir and rejects paths outside that directory.
// Returns fmt.Errorf for empty stored, invalid basename, or path traversal attempts.
func ResolveScrollbackPath(dir, stored string) (string, error) {
	if stored == "" {
		return "", fmt.Errorf("empty scrollback path")
	}
	base := filepath.Base(stored)
	if !scrollbackBaseName.MatchString(base) {
		return "", fmt.Errorf("invalid scrollback filename %q", base)
	}

	dirAbs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	full := filepath.Join(dirAbs, base)
	fullAbs, err := filepath.Abs(full)
	if err != nil {
		return "", err
	}
	if fullAbs != dirAbs && !strings.HasPrefix(fullAbs, dirAbs+string(os.PathSeparator)) {
		return "", fmt.Errorf("scrollback path escapes data directory")
	}
	return fullAbs, nil
}
