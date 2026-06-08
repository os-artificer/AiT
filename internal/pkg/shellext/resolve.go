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

package shellext

import (
	"os"
	"path/filepath"

	"github.com/os-artificer/AiT/internal/pkg/platform"
)

// ResolveAppBin looks up name under <AppRoot>/bin.
// Returns the resolved path and true when an executable file is found.
func ResolveAppBin(name string) (string, bool) {
	if !isSafeBinName(name) {
		return "", false
	}

	binDir, err := platform.BinDir()
	if err != nil {
		return "", false
	}

	if _, err := os.Stat(binDir); err != nil {
		return "", false
	}

	for _, candidate := range binCandidateNames(name) {
		path := filepath.Join(binDir, candidate)
		if ok, resolved := resolveExecutable(path, binDir); ok {
			return resolved, true
		}
	}
	return "", false
}
