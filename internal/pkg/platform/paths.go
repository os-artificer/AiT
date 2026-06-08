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
	"os"
	"path/filepath"
)

// DBPath returns the default SQLite database file path under AppRoot (DataDir).
// Returns the joined path, or DataDir errors.
func DBPath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, dbFile), nil
}

// ScrollbackDir returns the scrollback sidecar directory under DataDir, created with mode 0700.
// Returns the path, or DataDir / MkdirAll errors.
func ScrollbackDir() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}

	p := filepath.Join(dir, scrollbackDir)
	if err := os.MkdirAll(p, 0o700); err != nil {
		return "", err
	}
	return p, nil
}

// HistoryPath returns the REPL input history file under DataDir.
func HistoryPath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, historyBasename), nil
}

// BinDir returns the optional user extension directory under AppRoot (<AppRoot>/bin).
// The directory is not created automatically; callers add executables as needed.
func BinDir() (string, error) {
	dir, err := AppRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, binDirName), nil
}
