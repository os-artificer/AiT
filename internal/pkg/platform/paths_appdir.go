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
	"runtime"
)

type appDirKind int

const (
	appDirData appDirKind = iota
	appDirConfig
)

func appDir(kind appDirKind) (string, error) {
	var base string
	switch runtime.GOOS {
	case "windows":
		base = os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
		}
		base = filepath.Join(base, appName)
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, "Library", "Application Support", appName)
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		switch kind {
		case appDirData:
			if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
				base = filepath.Join(xdg, appName)
			} else {
				base = filepath.Join(home, ".local", "share", appName)
			}
		case appDirConfig:
			if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
				base = filepath.Join(xdg, appName)
			} else {
				base = filepath.Join(home, ".config", appName)
			}
		}
	}

	if err := os.MkdirAll(base, 0o700); err != nil {
		return "", err
	}
	return base, nil
}
