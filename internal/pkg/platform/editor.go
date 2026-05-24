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
	"os"
	"strings"
)

var (
	// ErrNoEditor is returned when VISUAL, EDITOR, and the OS default are all unavailable.
	ErrNoEditor = errors.New("no editor configured")
)

const (
	envVisual = "VISUAL"
	envEditor = "EDITOR"
)

// ResolveEditor returns the editor command from VISUAL, EDITOR, or the OS default.
// Returns errhint with install hints when no editor is available.
func ResolveEditor() (string, error) {
	if v := strings.TrimSpace(os.Getenv(envVisual)); v != "" {
		return v, nil
	}
	if v := strings.TrimSpace(os.Getenv(envEditor)); v != "" {
		return v, nil
	}
	return defaultEditor()
}

func editorNotFoundErr() error {
	return ErrNoEditor
}

// shellQuote wraps s in single quotes for Unix sh -c (handles spaces).
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
