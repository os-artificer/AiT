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

package view

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// WriteHelp prints aitlog default-route --help to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := WriteIntroUsageArgsFlags(b); err != nil {
		return err
	}
	if err := b.WriteExamples(DefaultExamples()...); err != nil {
		return err
	}
	return b.Finish(toolname.AILog)
}

// WriteIntroUsageArgsFlags writes shared Intro/Usage/Arguments/Flags for root help.
func WriteIntroUsageArgsFlags(b *pkghelpdoc.Body) error {
	if err := b.WriteIntro(
		"View and search system logs on Linux, macOS, and Windows.",
		"Uses native tools when available and falls back to built-in readers so all features keep working without external binaries.",
		"Keyword matching is case-insensitive AND; use search when a keyword collides with a subcommand name (e.g. version).",
	); err != nil {
		return err
	}
	if err := b.WriteUsage(toolname.AILog + " [flags] [keyword...]"); err != nil {
		return err
	}
	if err := b.WriteSection("Arguments", "  [keyword...]   optional keywords (AND match) against log messages\n"); err != nil {
		return err
	}
	return b.WriteSection("Flags", ""+
		"  -e, --regex string         match with regular expression\n"+
		"      --case-sensitive       case-sensitive matching\n"+
		"  -n, --lines int            last N matching records (default 1000)\n"+
		"      --head                 take first N matches instead of last N\n"+
		"      --since string         not older than (15m, 2h, today, RFC3339)\n"+
		"      --until string         not newer than\n"+
		"  -p, --priority string      minimum priority (emerg…debug)\n"+
		"  -u, --unit string          filter by unit/service\n"+
		"  -S, --source string        source name or file path/glob\n"+
		"  -f, --follow               follow new entries\n"+
		"      --json                 NDJSON output\n"+
		"      --raw                  include original line in JSON\n"+
		"      --utc                  timestamps in UTC\n"+
		"      --no-pager             disable interactive pager\n"+
		"      --backend string       auto|native|builtin (default auto)\n"+
		"      --quiet                suppress notices\n"+
		"  -V, --version              print version and exit\n")
}

// DefaultExamples returns example lines for root and view help.
func DefaultExamples() []string {
	return []string{
		toolname.AILog + " -n 50",
		toolname.AILog + " error --since 1h",
		toolname.AILog + " -e 'fail(ed|ure)' -n 100",
		toolname.AILog + " --backend builtin -S /var/log/system.log -n 20",
		toolname.AILog + " --json --no-pager -n 5",
		toolname.AILog + " sources",
	}
}
