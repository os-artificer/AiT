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

const appName = "ait"

// DataDir returns the per-user AiT data directory, creating it with mode 0700 when missing.
// Returns the path, or os.UserHomeDir / MkdirAll errors.
func DataDir() (string, error) {
	return appDir(appDirData)
}

// ConfigDir returns the per-user AiT config directory, creating it with mode 0700 when missing.
// Returns the path, or os.UserHomeDir / MkdirAll errors.
func ConfigDir() (string, error) {
	return appDir(appDirConfig)
}

// DBPath returns the default SQLite database file path under DataDir.
// Returns the joined path, or DataDir errors.
func DBPath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ait.db"), nil
}

// ScrollbackDir returns the scrollback sidecar directory under DataDir, created with mode 0700.
// Returns the path, or DataDir / MkdirAll errors.
func ScrollbackDir() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}

	p := filepath.Join(dir, "scrollback")
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
	return filepath.Join(dir, "repl_history"), nil
}
