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

import "sort"

// ReservedCommandNames returns the reserved ait REPL builtin names in sorted order.
// Names match reservedNames (help, exit, quit, version, history, clear).
// The returned slice is a new copy; callers may modify it.
func ReservedCommandNames() []string {
	names := make([]string, 0, len(reservedNames))
	for name := range reservedNames {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
