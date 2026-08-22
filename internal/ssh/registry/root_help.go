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

package registry

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

// RootHelpEntries converts registered commands to root help entries.
func RootHelpEntries() []pkghelpdoc.RootCommandEntry {
	reg := Entries()
	out := make([]pkghelpdoc.RootCommandEntry, len(reg))
	for i, e := range reg {
		out[i] = pkghelpdoc.RootCommandEntry{Name: e.Name, Short: e.Short}
	}
	return out
}

// WriteRootHelp prints aissh --help from registered top-level commands.
func WriteRootHelp(w io.Writer) error {
	return pkghelpdoc.WriteRootHelp(w, RootHelpEntries())
}
