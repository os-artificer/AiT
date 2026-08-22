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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/os-artificer/AiT/internal/pkg/platform"
	"github.com/os-artificer/AiT/internal/pkg/shellext"
)

var (
	discoverMu    sync.Mutex
	discoverOnce  sync.Once
	cachedPlugins []Plugin

	installDirsFn           = defaultInstallDirs
	lookPathFn              = platform.LookPath
	warnWriter    io.Writer = os.Stderr
	runManifestFn           = runManifestCommand
)

// Discover returns plugins found under InstallDir/plugins and PATH-filled builtins.
// The first call in the process populates a cache; later calls return a copy of the same list.
// The caller's ctx is not used for discovery I/O (avoids a cancelled first call poisoning the Once cache);
// manifest probes use an internal timeout.
func Discover(ctx context.Context) []Plugin {
	_ = ctx

	discoverOnce.Do(func() {
		plugins := discoverAll(context.Background())
		discoverMu.Lock()
		cachedPlugins = plugins
		discoverMu.Unlock()
	})

	discoverMu.Lock()
	defer discoverMu.Unlock()
	out := make([]Plugin, len(cachedPlugins))
	copy(out, cachedPlugins)
	return out
}

// Lookup returns the cached plugin with the given short name, if any.
func Lookup(name string) (Plugin, bool) {
	for _, p := range Discover(context.Background()) {
		if p.Manifest.Name == name {
			return p, true
		}
	}
	return Plugin{}, false
}

// InvalidateDiscover clears the Discover cache so the next Discover rescans install dirs.
func InvalidateDiscover() {
	resetDiscoverCache()
}

// ResetDiscoverForTest clears the Discover cache (tests only).
func ResetDiscoverForTest() {
	resetDiscoverCache()
}

func resetDiscoverCache() {
	discoverMu.Lock()
	defer discoverMu.Unlock()
	discoverOnce = sync.Once{}
	cachedPlugins = nil
}

// SetInstallDirsForTest overrides InstallDir discovery (tests only). Call ResetDiscoverForTest after.
func SetInstallDirsForTest(dirs []string) {
	discoverMu.Lock()
	defer discoverMu.Unlock()
	installDirsFn = func() []string { return append([]string(nil), dirs...) }
}

// SetLookPathForTest overrides PATH lookup (tests only).
func SetLookPathForTest(fn func(string) (string, error)) {
	discoverMu.Lock()
	defer discoverMu.Unlock()
	if fn == nil {
		lookPathFn = platform.LookPath
		return
	}
	lookPathFn = fn
}

// SetWarnWriterForTest sets where discovery warnings are written (tests only).
func SetWarnWriterForTest(w io.Writer) {
	discoverMu.Lock()
	defer discoverMu.Unlock()
	if w == nil {
		warnWriter = os.Stderr
		return
	}
	warnWriter = w
}

// SetRunManifestForTest overrides the ait-plugin-manifest runner (tests only).
func SetRunManifestForTest(fn func(context.Context, string) ([]byte, error)) {
	discoverMu.Lock()
	defer discoverMu.Unlock()
	if fn == nil {
		runManifestFn = runManifestCommand
		return
	}
	runManifestFn = fn
}

// RestoreDiscoverHooksForTest resets injectable discover hooks (tests only).
func RestoreDiscoverHooksForTest() {
	discoverMu.Lock()
	defer discoverMu.Unlock()
	installDirsFn = defaultInstallDirs
	lookPathFn = platform.LookPath
	warnWriter = os.Stderr
	runManifestFn = runManifestCommand
}

func discoverAll(ctx context.Context) []Plugin {
	byName := make(map[string]Plugin)

	for _, installDir := range installDirsSnapshot() {
		scanPluginsRoot(ctx, filepath.Join(installDir, pluginsDirName), byName)
	}
	fillBuiltinPATH(ctx, byName)

	out := make([]Plugin, 0, len(byName))
	for _, p := range byName {
		out = append(out, p)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Manifest.Order != out[j].Manifest.Order {
			return out[i].Manifest.Order < out[j].Manifest.Order
		}
		return out[i].Manifest.Name < out[j].Manifest.Name
	})
	return out
}

func installDirsSnapshot() []string {
	discoverMu.Lock()
	fn := installDirsFn
	discoverMu.Unlock()
	return fn()
}

func scanPluginsRoot(ctx context.Context, pluginsRoot string, byName map[string]Plugin) {
	entries, err := os.ReadDir(pluginsRoot)
	if err != nil {
		return
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, dirName := range names {
		if !shellext.IsSafeName(dirName) {
			warnf("ait: skip plugins/%s: unsafe directory name", dirName)
			continue
		}
		p, ok := discoverPluginDir(ctx, filepath.Join(pluginsRoot, dirName), dirName)
		if !ok {
			continue
		}
		registerPlugin(byName, p)
	}
}

func fillBuiltinPATH(ctx context.Context, byName map[string]Plugin) {
	for _, b := range BuiltinBinaryNames {
		if _, exists := byName[b.Name]; exists {
			continue
		}

		look, run := pathHooksSnapshot()
		exe, err := look(b.Binary)
		if err != nil || exe == "" {
			continue
		}

		data, err := run(ctx, exe)
		if err != nil {
			warnf("ait: skip PATH %s: manifest: %v", b.Binary, err)
			continue
		}

		m, err := LoadBytes(data)
		if err != nil {
			warnf("ait: skip PATH %s: %v", b.Binary, err)
			continue
		}
		if m.Name != b.Name || m.Binary != b.Binary {
			warnf("ait: skip PATH %s: manifest name/binary mismatch (got %q/%q, want %q/%q)",
				b.Binary, m.Name, m.Binary, b.Name, b.Binary)
			continue
		}

		registerPlugin(byName, Plugin{Manifest: m, Path: exe})
	}
}

func pathHooksSnapshot() (look func(string) (string, error), run func(context.Context, string) ([]byte, error)) {
	discoverMu.Lock()
	defer discoverMu.Unlock()
	return lookPathFn, runManifestFn
}

func registerPlugin(byName map[string]Plugin, p Plugin) {
	if prev, exists := byName[p.Manifest.Name]; exists {
		warnf("ait: skip duplicate plugin name %q (keeping %s, ignoring %s)",
			p.Manifest.Name, prev.Path, p.Path)
		return
	}
	byName[p.Manifest.Name] = p
}

func warnf(format string, args ...any) {
	discoverMu.Lock()
	w := warnWriter
	discoverMu.Unlock()
	if w == nil {
		return
	}
	_, _ = fmt.Fprintf(w, format+"\n", args...)
}
