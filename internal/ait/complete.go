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

package ait

import (
	"context"
	"sort"

	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
)

// completeREPLLine is the ait TTY Completer: first-token completion against
// discovered plugins and reserved builtins.
func completeREPLLine(line string, cursor int) (newLine string, newCursor int, display string) {
	names := candidateCommandNames(context.Background())
	return completeFirstToken(names, line, cursor)
}

func candidateCommandNames(ctx context.Context) []string {
	seen := make(map[string]struct{})
	var names []string

	for _, p := range aitplugin.Discover(ctx) {
		name := p.Manifest.Name
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	for _, name := range aitplugin.ReservedCommandNames() {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
