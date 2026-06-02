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

package provider

import (
	"fmt"
	"io"
	"strings"

	"github.com/os-artificer/AiT/internal/ait/ai/provider/subregistry"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

const (
	introMain       = "Enable or disable AI provider routing without editing config.yaml by hand."
	introOllamaHint = "For Ollama install, models directory, and tiered setup, prefer: ait ai ollama setup"
	usageLine       = "ait ai provider"
	exampleEnable   = "ait ai provider enable deepseek"
	exampleDisable  = "ait ai provider disable minimax"
	exampleOllama   = "ait ai provider enable ollama --yes"
)

// WriteHelp prints ait ai provider --help to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(introMain, introOllamaHint); err != nil {
		return err
	}
	if err := b.WriteUsage(usageLine); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionSubcommands, formatSubcommands()); err != nil {
		return err
	}
	if err := b.WriteExamples(exampleEnable, exampleDisable, exampleOllama); err != nil {
		return err
	}
	return b.Finish("provider")
}

func formatSubcommands() string {
	var b strings.Builder
	for _, e := range subregistry.Entries() {
		fmt.Fprintf(&b, "  %s   %s\n", e.Name, e.Short)
	}
	return b.String()
}
