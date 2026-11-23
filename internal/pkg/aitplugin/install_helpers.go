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

package aitplugin

import (
	"context"
	"errors"
	"fmt"
	"sort"
)

// ErrInstallRootUnavailable indicates PrimaryInstallDir could not resolve a writable install root.
var ErrInstallRootUnavailable = errors.New("ait install root unavailable")

// IsManageCommand reports whether name is install, update, or upgrade.
func IsManageCommand(name string) bool {
	_, ok := manageCommandNames[name]
	return ok
}

// OfficialPluginNames returns sorted official plugin command names (host, img, keep, log, ssh, vid).
// The returned slice is a new copy.
func OfficialPluginNames() []string {
	names := make([]string, 0, len(BuiltinBinaryNames))
	for _, b := range BuiltinBinaryNames {
		names = append(names, b.Name)
	}
	sort.Strings(names)
	return names
}

// IsOfficialPlugin reports whether name is a shipped official plugin command name (legacy ai-prefixed names accepted).
func IsOfficialPlugin(name string) bool {
	if canon, ok := LegacyPluginNames[name]; ok {
		name = canon
	}
	for _, b := range BuiltinBinaryNames {
		if b.Name == name {
			return true
		}
	}
	return false
}

// InstalledOfficialPlugins returns official plugin names currently visible to Discover.
func InstalledOfficialPlugins(ctx context.Context) []string {
	installed := make(map[string]struct{})
	for _, p := range Discover(ctx) {
		if IsOfficialPlugin(p.Manifest.Name) {
			installed[p.Manifest.Name] = struct{}{}
		}
	}
	names := make([]string, 0, len(installed))
	for name := range installed {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// PrimaryInstallDir returns the directory where ait and plugins should be installed.
// When the executable path resolves through a symlink, the resolved directory is used.
func PrimaryInstallDir() (string, error) {
	dirs := installDirsSnapshot()
	if len(dirs) == 0 {
		return "", fmt.Errorf("%w: %v", ErrInstallRootUnavailable, "cannot resolve executable path")
	}
	if len(dirs) > 1 {
		return dirs[len(dirs)-1], nil
	}
	return dirs[0], nil
}

// PluginCommandName maps a plugin command, legacy name, or binary basename to the manifest command name.
func PluginCommandName(nameOrBinary string) string {
	if canon, ok := LegacyPluginNames[nameOrBinary]; ok {
		return canon
	}
	for _, b := range BuiltinBinaryNames {
		if b.Name == nameOrBinary || b.Binary == nameOrBinary {
			return b.Name
		}
	}
	return nameOrBinary
}

// PluginBinary returns the shipped binary basename for an official plugin command or legacy name.
// Unknown names are returned unchanged.
func PluginBinary(name string) string {
	cmd := PluginCommandName(name)
	for _, b := range BuiltinBinaryNames {
		if b.Name == cmd {
			return b.Binary
		}
	}
	if name == cmd {
		return name
	}
	return name
}
