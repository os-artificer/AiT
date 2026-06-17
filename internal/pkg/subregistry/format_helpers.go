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

package subregistry

import (
	"fmt"
	"strings"
)

// FormatSubcommandLines renders registry entries as helpdoc subcommand section lines.
func FormatSubcommandLines(entries []Entry) string {
	var b strings.Builder
	for _, e := range entries {
		fmt.Fprintf(&b, "  %s   %s\n", e.Name, e.Short)
	}
	return b.String()
}

// JoinChildNames returns child subcommand names joined with sep (e.g. "enable|disable").
func JoinChildNames(entries []Entry, sep string) string {
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name
	}
	return strings.Join(names, sep)
}
