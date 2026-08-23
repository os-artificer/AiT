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
	"fmt"
	"io"
	"strings"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

func writeManageHelp(w io.Writer, topic string) error {
	spec, ok := manageHelpTopics[topic]
	if !ok {
		return fmt.Errorf("unknown manage help topic %q", topic)
	}

	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(spec.desc); err != nil {
		return err
	}
	if err := b.WriteUsage(spec.usage); err != nil {
		return err
	}
	examples := make([]string, len(spec.lines))
	copy(examples, spec.lines)
	if err := b.WriteExamples(examples...); err != nil {
		return err
	}
	return b.Finish(topic)
}

func formatManageCommands() string {
	var b strings.Builder
	order := []string{cmdInstall, cmdUpdate, cmdUpgrade}
	for _, name := range order {
		spec := manageHelpTopics[name]
		fmt.Fprintf(&b, "  %-12s %s\n", name, firstSentence(spec.desc))
	}
	return b.String()
}

func firstSentence(s string) string {
	if idx := strings.Index(s, "."); idx >= 0 {
		return strings.TrimSpace(s[:idx+1])
	}
	return s
}
