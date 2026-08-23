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

package logsrc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// nativeProbeTools maps a layer Provider prefix to the executable that layer
// needs on PATH.
var nativeProbeTools = map[string]string{
	"journalctl": "journalctl",
	"log show":   "log",
	"wevtutil":   "wevtutil",
}

// ChainNames returns the chain source names in stable alphabetical order so
// that `ailog sources` output does not depend on map iteration order.
func ChainNames() []string {
	chains := Chains()
	names := make([]string, 0, len(chains))
	for name := range chains {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// openLayer builds the layer's Source and opens a cursor for q.
func openLayer(ctx context.Context, r *Resolver, layer Layer, q Query) (Source, Cursor, error) {
	src, err := layer.Open(r, q)
	if err != nil {
		return nil, nil, err
	}
	cur, err := openCursor(ctx, src, q)
	if err != nil {
		return nil, nil, err
	}
	return src, cur, nil
}

// openCursor opens src, preferring a newest-first cursor for last-N queries
// when src implements ReverseOpener.
func openCursor(ctx context.Context, src Source, q Query) (Cursor, error) {
	if !q.Head && !q.Follow && src.SupportsReverse() {
		if ro, ok := src.(ReverseOpener); ok {
			return ro.OpenReverse(ctx, q)
		}
	}
	return src.Open(ctx, q)
}

func describeSkip(layer Layer) string {
	reason := layer.Reason
	if reason == "" {
		reason = "unavailable"
	}
	return layer.Provider + " (" + reason + ")"
}

func fallbackTrigger(tried []string) string {
	if len(tried) == 0 {
		return "native tool unavailable"
	}
	return "falling back after: " + tried[len(tried)-1]
}

func validateBackend(b Backend) error {
	switch b {
	case BackendAuto, BackendNative, BackendBuiltin:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrBadBackend, b)
	}
}

func layerMatchesBackend(layer Layer, b Backend) bool {
	switch b {
	case BackendAuto:
		return true
	case BackendNative:
		return layer.Kind == LayerNative
	case BackendBuiltin:
		return layer.Kind == LayerBuiltin
	default:
		return false
	}
}

// isPathLike reports whether s names a file rather than a logical source: it
// contains a separator or glob metacharacter, is absolute, or exists on disk.
func isPathLike(s string) bool {
	if s == "" {
		return false
	}
	if strings.ContainsAny(s, `/\`) || strings.ContainsAny(s, "*?[") {
		return true
	}
	if filepath.IsAbs(s) {
		return true
	}
	_, err := os.Stat(s)
	return err == nil
}

// probeLayer reports whether layer can serve source under backend b, without
// opening a cursor: native layers only need their tool on PATH, built-in file
// layers need at least one existing file.
func probeLayer(r *Resolver, layer Layer, source string, b Backend) (bool, string) {
	if !layerMatchesBackend(layer, b) {
		reason := layer.Reason
		if reason == "" {
			reason = "filtered by --backend"
		}
		return false, reason
	}
	if !layer.Available {
		return false, layer.Reason
	}
	if tool, ok := probeTool(layer.Provider); ok {
		if _, err := r.lookPath(tool); err != nil {
			return false, err.Error()
		}
		return true, ""
	}
	if layer.Kind == LayerBuiltin && layer.Open != nil {
		return probeBuiltin(r, layer, source)
	}
	return layer.Available, layer.Reason
}

func probeTool(provider string) (string, bool) {
	for prefix, tool := range nativeProbeTools {
		if strings.HasPrefix(provider, prefix) {
			return tool, true
		}
	}
	return "", false
}

func probeBuiltin(r *Resolver, layer Layer, source string) (bool, string) {
	src, err := layer.Open(r, Query{Source: source})
	if err != nil {
		return false, err.Error()
	}
	fs, ok := src.(*FileSource)
	if !ok {
		return true, ""
	}
	paths, err := ResolveLogFiles(fs.Paths)
	if err != nil {
		return false, err.Error()
	}
	if len(paths) == 0 {
		return false, "no files"
	}
	return true, ""
}
