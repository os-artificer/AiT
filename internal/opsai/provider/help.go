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

	"github.com/os-artificer/AiT/internal/opsai/provider/subregistry"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	pkgsubregistry "github.com/os-artificer/AiT/internal/pkg/subregistry"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

const introMain = "Enable or disable AI provider routing without editing config.yaml by hand."

var (
	introOllamaHint = fmt.Sprintf("For Ollama install, models directory, and tiered setup, prefer: %s", toolname.AICLI("ollama setup"))
	usageLine       = toolname.AICLI("provider")
)

// WriteHelp prints aissh ai provider --help to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(introMain, introOllamaHint); err != nil {
		return err
	}
	if err := b.WriteUsage(usageLine); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionSubcommands, pkgsubregistry.FormatSubcommandLines(subregistry.Entries())); err != nil {
		return err
	}
	if err := b.WriteExamples(
		pkghelpdoc.ExampleAIProviderEnableDeepseek,
		pkghelpdoc.ExampleAIProviderDisableMinimax,
		pkghelpdoc.ExampleAIProviderEnableOllamaYes,
	); err != nil {
		return err
	}
	return b.Finish("provider")
}
