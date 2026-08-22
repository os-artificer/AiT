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
	"runtime"
	"strings"
)

// IsSafeName reports whether name is a single path segment safe for AppRoot/bin lookup.
func IsSafeName(name string) bool {
	return isSafeBinName(name)
}

// CandidateNames returns OS-specific executable name candidates for name (Windows adds .exe/.cmd/.bat).
func CandidateNames(name string) []string {
	return binCandidateNames(name)
}

func isSafeBinName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	if strings.ContainsAny(name, `/\`) {
		return false
	}
	return true
}

func binCandidateNames(name string) []string {
	if runtime.GOOS == "windows" {
		return []string{name, name + winExtExe, name + winExtCmd, name + winExtBat}
	}
	return []string{name}
}

func resolveExecutable(path, binDir string) (bool, string) {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false, ""
	}

	if !fileModeExecutable(info.Mode()) {
		return false, ""
	}

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false, ""
	}
	if !pathUnderDir(resolved, binDir) {
		return false, ""
	}
	return true, path
}

func pathUnderDir(path, dir string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	absPath, err = filepath.EvalSymlinks(absPath)
	if err != nil {
		return false
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	absDir, err = filepath.EvalSymlinks(absDir)
	if err != nil {
		return false
	}

	rel, err := filepath.Rel(absDir, absPath)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}
