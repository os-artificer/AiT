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

package ollama

import (
	"fmt"
	"io"

	"github.com/os-artificer/AiT/internal/aissh/ai/ollama/disable"
	"github.com/os-artificer/AiT/internal/aissh/ai/ollama/setup"
	"github.com/os-artificer/AiT/internal/aissh/ai/ollama/start"
	"github.com/os-artificer/AiT/internal/aissh/ai/ollama/status"
	"github.com/os-artificer/AiT/internal/aissh/ai/ollama/stop"
	"github.com/os-artificer/AiT/internal/aissh/ai/ollama/subregistry"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	pkgsubregistry "github.com/os-artificer/AiT/internal/pkg/subregistry"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

var (
	ollamaIntroMain       = fmt.Sprintf("Manage local Ollama for tiered AI routing (under %s).", toolname.CLI("ai"))
	ollamaIntroSubcmdHelp = fmt.Sprintf("Run '%s <subcommand> --help' for flags and examples.", toolname.CLI("ai ollama"))
	ollamaUsageMain       = toolname.CLI("ai ollama")
	ollamaExampleStop     = toolname.CLI("ai ollama stop")
)

// WriteAIOllamaHelp prints aitssh ai ollama --help to w.
func WriteAIOllamaHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(ollamaIntroMain, ollamaIntroSubcmdHelp); err != nil {
		return err
	}
	if err := b.WriteUsage(ollamaUsageMain); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionSubcommands, pkgsubregistry.FormatSubcommandLines(subregistry.Entries())); err != nil {
		return err
	}
	if err := b.WriteExamples(
		pkghelpdoc.ExampleAIOllamaStatus,
		pkghelpdoc.ExampleAIOllamaSetupYes,
		pkghelpdoc.ExampleAIOllamaStartWait,
		ollamaExampleStop,
		pkghelpdoc.ExampleAIOllamaDisableStop,
	); err != nil {
		return err
	}
	return b.Finish("ollama")
}

// WriteOllamaStatusHelp prints aitssh ai ollama status --help to w.
func WriteOllamaStatusHelp(w io.Writer) error { return status.WriteHelp(w) }

// WriteOllamaSetupHelp prints aitssh ai ollama setup --help to w.
func WriteOllamaSetupHelp(w io.Writer) error { return setup.WriteHelp(w) }

// WriteOllamaStartHelp prints aitssh ai ollama start --help to w.
func WriteOllamaStartHelp(w io.Writer) error { return start.WriteHelp(w) }

// WriteOllamaStopHelp prints aitssh ai ollama stop --help to w.
func WriteOllamaStopHelp(w io.Writer) error { return stop.WriteHelp(w) }

// WriteOllamaDisableHelp prints aitssh ai ollama disable --help to w.
func WriteOllamaDisableHelp(w io.Writer) error { return disable.WriteHelp(w) }
