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

package search

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// WriteHelp prints ailog search --help to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(
		"Search system logs. Requires at least one keyword, --expr, or --regex.",
		"Same filters as the default route; use this when a keyword would collide with a subcommand (e.g. version).",
		"Shell-quote --expr so | and ! are not interpreted by the shell.",
	); err != nil {
		return err
	}
	if err := b.WriteUsage(toolname.AILog + " search [flags] [keyword...]"); err != nil {
		return err
	}
	if err := b.WriteExamples(
		toolname.AILog+" search error",
		toolname.AILog+" search version",
		toolname.AILog+` search --expr 'error|fail' --since 1h`,
		toolname.AILog+" search -e 'fail(ed|ure)' --since 1h",
	); err != nil {
		return err
	}
	return b.Finish(Name)
}
