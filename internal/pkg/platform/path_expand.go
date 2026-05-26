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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrEmptyPath is returned by ExpandUserAbsPath when the trimmed input is empty.
// Callers wrap it with their own user-facing wording.
var ErrEmptyPath = errors.New("empty path")

// ExpandUserAbsPath expands a leading "~"/"~/" prefix using the current user's
// home directory and returns the cleaned absolute path.
//
// raw is trimmed of surrounding whitespace before expansion. An empty input
// yields ErrEmptyPath; callers wrap it with their own user-facing wording.
// Returns the original os.UserHomeDir, filepath.Abs, or filepath.Clean errors
// when the home directory or filesystem path resolution fails, plus
// fmt.Errorf("invalid path %q", raw) when raw begins with "~" but does not
// match "~" or "~/...".
func ExpandUserAbsPath(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrEmptyPath
	}
	if strings.HasPrefix(raw, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		switch {
		case raw == "~":
			raw = home
		case strings.HasPrefix(raw, "~/"):
			raw = filepath.Join(home, strings.TrimPrefix(raw, "~/"))
		default:
			return "", fmt.Errorf("invalid path %q", raw)
		}
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}
