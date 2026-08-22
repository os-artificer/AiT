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
	"bytes"
	"context"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/os-artificer/AiT/internal/pkg/shellext"
)

func defaultInstallDirs() []string {
	exe, err := os.Executable()
	if err != nil {
		return nil
	}
	dirs := make([]string, 0, 2)
	dir := filepath.Dir(exe)
	dirs = append(dirs, dir)

	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return uniqueInstallDirs(dirs)
	}
	resolvedDir := filepath.Dir(resolved)
	if resolvedDir != dir {
		dirs = append(dirs, resolvedDir)
	}
	return uniqueInstallDirs(dirs)
}

// uniqueInstallDirs keeps the first occurrence of each install directory after
// resolving symlinks, so the same plugins/ tree is never scanned twice.
func uniqueInstallDirs(dirs []string) []string {
	seen := make(map[string]struct{}, len(dirs))
	out := make([]string, 0, len(dirs))
	for _, d := range dirs {
		key := d
		if abs, err := filepath.Abs(d); err == nil {
			key = abs
		}
		if resolved, err := filepath.EvalSymlinks(key); err == nil {
			key = resolved
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, d)
	}
	return out
}

func discoverPluginDir(ctx context.Context, dir, dirName string) (Plugin, bool) {
	jsonPath := filepath.Join(dir, pluginJSONName)
	if st, err := os.Stat(jsonPath); err == nil && !st.IsDir() {
		return discoverFromJSON(dir, dirName, jsonPath)
	}
	return discoverFromManifestCmd(ctx, dir, dirName)
}

func discoverFromJSON(dir, dirName, jsonPath string) (Plugin, bool) {
	m, err := LoadFile(jsonPath)
	if err != nil {
		warnf("ait: skip plugins/%s: %v", dirName, err)
		return Plugin{}, false
	}

	if m.Binary != dirName {
		warnf("ait: plugins/%s: directory name differs from binary %q; using binary", dirName, m.Binary)
	}

	path, ok := resolveExecutableInDir(dir, m.Binary)
	if !ok {
		look, _ := pathHooksSnapshot()
		path, err = look(m.Binary)
		if err != nil || path == "" {
			warnf("ait: skip plugins/%s: executable %q not found", dirName, m.Binary)
			return Plugin{}, false
		}
	}

	return Plugin{Manifest: m, Path: path}, true
}

func discoverFromManifestCmd(ctx context.Context, dir, dirName string) (Plugin, bool) {
	exe, ok := resolveExecutableInDir(dir, dirName)
	if !ok {
		warnf("ait: skip plugins/%s: no plugin.json and no executable", dirName)
		return Plugin{}, false
	}
	_, run := pathHooksSnapshot()
	data, err := run(ctx, exe)
	if err != nil {
		warnf("ait: skip plugins/%s: ait-plugin-manifest: %v", dirName, err)
		return Plugin{}, false
	}
	m, err := LoadBytes(data)
	if err != nil {
		warnf("ait: skip plugins/%s: %v", dirName, err)
		return Plugin{}, false
	}
	if m.Binary != dirName {
		warnf("ait: skip plugins/%s: manifest binary %q does not match directory", dirName, m.Binary)
		return Plugin{}, false
	}
	return Plugin{Manifest: m, Path: exe}, true
}

func resolveExecutableInDir(dir, binary string) (string, bool) {
	if !shellext.IsSafeName(binary) {
		return "", false
	}
	for _, cand := range shellext.CandidateNames(binary) {
		path := filepath.Join(dir, cand)
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		if !fileModeExecutable(info.Mode()) {
			continue
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return path, true
		}
		return abs, true
	}
	return "", false
}

func fileModeExecutable(mode fs.FileMode) bool {
	if runtime.GOOS == "windows" {
		return true
	}
	return mode&0o111 != 0
}

func runManifestCommand(ctx context.Context, exe string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, manifestTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, exe, ManifestArg)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return stdout.Bytes(), nil
}
