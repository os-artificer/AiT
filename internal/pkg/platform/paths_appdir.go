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
	"strings"
)

const (
	envHome    = "AIT_HOME"
	envDataDir = "AIT_DATA_DIR"
)

// AppRoot returns the unified AiT runtime directory (config, DB, scrollback, history, keyring).
// When AIT_HOME or AIT_DATA_DIR is set, that path is used; otherwise the OS default data home.
// Creates the directory with mode 0700 when missing.
func AppRoot() (string, error) {
	if root := resolveHomeOverride(); root != "" {
		if err := os.MkdirAll(root, 0o700); err != nil {
			return "", err
		}
		return root, nil
	}

	return defaultAppRoot()
}

// DataDir returns AppRoot (all runtime files live under one directory).
func DataDir() (string, error) {
	return AppRoot()
}

// ConfigDir returns AppRoot (config.yaml lives beside other runtime files).
func ConfigDir() (string, error) {
	return AppRoot()
}

// ConfigPath returns the path to config.yaml under AppRoot.
func ConfigPath() (string, error) {
	dir, err := AppRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

// LegacyConfigPath returns the pre-unification Linux config.yaml path, or "" when not applicable.
// Used only to migrate config on first Load.
func LegacyConfigPath() (string, error) {
	dir, err := legacyConfigDir()
	if err != nil || dir == "" {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

func legacyConfigDir() (string, error) {
	if runtime.GOOS != "linux" {
		return "", nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, appName), nil
	}
	return filepath.Join(home, ".config", appName), nil
}

func resolveHomeOverride() string {
	if v := strings.TrimSpace(os.Getenv(envHome)); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv(envDataDir)); v != "" {
		return v
	}
	return ""
}

func defaultAppRoot() (string, error) {
	if err := maybeMigrateLegacyDataRoot(); err != nil {
		return "", err
	}

	base, err := defaultAppRootPath()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(base, 0o700); err != nil {
		return "", err
	}
	return base, nil
}
