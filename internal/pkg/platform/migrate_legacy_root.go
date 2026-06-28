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
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

func maybeMigrateLegacyDataRoot() error {
	if resolveHomeOverride() != "" {
		return nil
	}

	newRoot, err := defaultAppRootPath()
	if err != nil {
		return err
	}
	legacyRoot, err := legacyDataRootPath()
	if err != nil {
		return err
	}
	if legacyRoot == "" || legacyRoot == newRoot {
		return nil
	}
	if pathsSameDir(legacyRoot, newRoot) {
		return nil
	}

	legacyHas, err := dirHasContent(legacyRoot)
	if err != nil || !legacyHas {
		return err
	}

	newHasToolData, err := rootHasToolData(newRoot)
	if err != nil {
		return err
	}
	if newHasToolData {
		warnLegacyDataConflict(legacyRoot, newRoot)
		return nil
	}

	if _, err := os.Stat(newRoot); os.IsNotExist(err) {
		return os.Rename(legacyRoot, newRoot)
	}

	if err := copyDir(legacyRoot, newRoot); err != nil {
		return err
	}
	if err := verifyLegacyCopied(legacyRoot, newRoot); err != nil {
		return err
	}
	return os.RemoveAll(legacyRoot)
}

func legacyDataRootPath() (string, error) {
	return dataRootForName(legacyMixedCaseAppName)
}

func defaultAppRootPath() (string, error) {
	return dataRootForName(appName)
}

func dataRootForName(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var base string
	switch runtimeGOOS() {
	case "windows":
		base = os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
		}
		base = filepath.Join(base, name)
	case "darwin":
		base = filepath.Join(home, "Library", "Application Support", name)
	default:
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			base = filepath.Join(xdg, name)
		} else {
			base = filepath.Join(home, ".local", "share", name)
		}
	}
	return base, nil
}

func dirHasContent(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return len(entries) > 0, nil
}

func rootHasToolData(root string) (bool, error) {
	return len(listToolDataFiles(root)) > 0, nil
}

func verifyLegacyCopied(legacyRoot, newRoot string) error {
	return filepath.WalkDir(legacyRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(legacyRoot, path)
		if err != nil {
			return err
		}
		dst := filepath.Join(newRoot, rel)
		if _, err := os.Stat(dst); err != nil {
			return err
		}
		return nil
	})
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		return os.WriteFile(target, data, d.Type().Perm())
	})
}

// runtimeGOOS is overridden in tests.
var runtimeGOOS = func() string { return runtime.GOOS }
