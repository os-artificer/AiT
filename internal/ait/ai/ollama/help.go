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
	"io"

	"github.com/os-artificer/AiT/internal/ait/ai/ollama/disable"
	"github.com/os-artificer/AiT/internal/ait/ai/ollama/setup"
	"github.com/os-artificer/AiT/internal/ait/ai/ollama/start"
	"github.com/os-artificer/AiT/internal/ait/ai/ollama/status"
	"github.com/os-artificer/AiT/internal/ait/ai/ollama/stop"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

const ollamaSubcommands = `  status            show tiered/enable flags and API reachability
  setup [flags]     enable tiered routing; install/start/pull Ollama
  start [flags]     start ollama serve in the background
  stop              stop local ollama serve
  disable [flags]   set ai.providers.ollama.enabled false (optional --stop)
`

const (
	ollamaIntroMain       = "Manage local Ollama for tiered AI routing (under ait ai)."
	ollamaIntroSubcmdHelp = "Run 'ait ai ollama <subcommand> --help' for flags and examples."
	ollamaUsageMain       = "ait ai ollama"
	ollamaExampleStop     = "ait ai ollama stop"
)

// WriteAIOllamaHelp prints ait ai ollama --help to w.
func WriteAIOllamaHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(ollamaIntroMain, ollamaIntroSubcmdHelp); err != nil {
		return err
	}
	if err := b.WriteUsage(ollamaUsageMain); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionSubcommands, ollamaSubcommands); err != nil {
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

// WriteOllamaStatusHelp prints ait ai ollama status --help to w.
func WriteOllamaStatusHelp(w io.Writer) error { return status.WriteHelp(w) }

// WriteOllamaSetupHelp prints ait ai ollama setup --help to w.
func WriteOllamaSetupHelp(w io.Writer) error { return setup.WriteHelp(w) }

// WriteOllamaStartHelp prints ait ai ollama start --help to w.
func WriteOllamaStartHelp(w io.Writer) error { return start.WriteHelp(w) }

// WriteOllamaStopHelp prints ait ai ollama stop --help to w.
func WriteOllamaStopHelp(w io.Writer) error { return stop.WriteHelp(w) }

// WriteOllamaDisableHelp prints ait ai ollama disable --help to w.
func WriteOllamaDisableHelp(w io.Writer) error { return disable.WriteHelp(w) }
