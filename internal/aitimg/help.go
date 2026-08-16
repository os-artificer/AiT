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

package aitimg

import (
	"fmt"
	"io"
	"strings"

	"github.com/os-artificer/AiT/internal/aitimg/extract"
	"github.com/os-artificer/AiT/internal/aitimg/registry"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// WriteRootHelp prints aitimg --help to w (default extract usage + registered Commands).
func WriteRootHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := extract.WriteIntroUsageArgsFlags(b); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionCommands, formatRootCommands(registry.RootHelpEntries())); err != nil {
		return err
	}

	examples := append(extract.DefaultExamples(), toolname.AIImg+" version")
	if err := b.WriteExamples(examples...); err != nil {
		return err
	}
	return b.Finish(toolname.AIImg)
}

func formatRootCommands(entries []pkghelpdoc.RootCommandEntry) string {
	var b strings.Builder
	for _, sc := range entries {
		fmt.Fprintf(&b, "  %-10s %s\n", sc.Name, sc.Short)
	}
	return b.String()
}
