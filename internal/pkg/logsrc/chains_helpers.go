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
	"fmt"
	"os"
	"path/filepath"
)

// builtinFiles builds a pure-Go text-file layer. Paths may contain globs and
// are filtered to those that exist when the layer is opened, so a chain can
// list alternatives (syslog and messages) in one layer.
func builtinFiles(label string, paths []string) Layer {
	return Layer{
		Kind:      LayerBuiltin,
		Provider:  label,
		Available: true,
		Open: func(r *Resolver, q Query) (Source, error) {
			existing := existingPaths(paths)
			if len(existing) == 0 {
				return nil, fmtErrNoSource(label)
			}
			return &FileSource{Paths: existing, Label: label}, nil
		},
	}
}

// deferredLayer builds an always-unavailable placeholder so that `ailog
// sources` can name the store it cannot read yet.
func deferredLayer(provider, reason string) Layer {
	return Layer{Kind: LayerDeferred, Provider: provider, Available: false, Reason: reason}
}

func fmtErrTool(name string) error {
	return fmt.Errorf("%w: %s", ErrToolMissing, name)
}

func fmtErrNoSource(label string) error {
	return fmt.Errorf("%w: %s", ErrNoSource, label)
}

func existingPaths(paths []string) []string {
	var out []string
	for _, p := range paths {
		matches, err := filepath.Glob(p)
		if err != nil {
			continue
		}
		if len(matches) == 0 {
			if _, err := os.Stat(p); err == nil {
				matches = []string{p}
			}
		}
		out = append(out, matches...)
	}
	return out
}
