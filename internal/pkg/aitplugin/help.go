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
	"fmt"
	"io"
	"strings"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

// WriteHelp prints Usage / Flags / Examples for m to w.
func WriteHelp(w io.Writer, m Manifest) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(m.Short); err != nil {
		return err
	}
	usage := m.Usage
	if usage == "" {
		usage = m.Name + " [args...]"
	}
	if err := b.WriteUsage(usage); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionFlags, formatFlags(m.Flags)); err != nil {
		return err
	}
	if err := b.WriteExamples(m.Examples...); err != nil {
		return err
	}
	return b.Finish(m.Name)
}

func formatFlags(flags []Flag) string {
	if len(flags) == 0 {
		return ""
	}
	var b strings.Builder
	for _, f := range flags {
		label := "--" + f.Name
		if f.Short != "" {
			label = "-" + f.Short + ", " + label
		}
		if f.Type != flagTypeBool {
			label += " " + f.Type
		}
		fmt.Fprintf(&b, "  %-22s %s\n", label, f.Help)
	}
	return b.String()
}
