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

const (
	exampleAIOllamaStatus      = "ait ai ollama status"
	exampleAIOllamaSetupYes    = "ait ai ollama setup --yes"
	exampleAIOllamaStartWait   = "ait ai ollama start --wait"
	exampleAIOllamaDisableStop = "ait ai ollama disable --stop"
)

const ollamaSubcommands = `  status            show tiered/enable flags and API reachability
  setup [flags]     enable tiered routing; install/start/pull Ollama
  start [flags]     start ollama serve in the background
  stop              stop local ollama serve
  disable [flags]   set ai.providers.ollama.enabled false (optional --stop)
`

const (
	ollamaIntroMain        = "Manage local Ollama for tiered AI routing (under ait ai)."
	ollamaIntroSubcmdHelp  = "Run 'ait ai ollama <subcommand> --help' for flags and examples."
	ollamaUsageMain        = "ait ai ollama"
	ollamaExampleStartWait = "ait ai ollama start --wait"
	ollamaExampleStop      = "ait ai ollama stop"
	ollamaExampleDisable   = "ait ai ollama disable"
	ollamaStatusIntroMain  = "Read-only status for Ollama: config flags, binary on PATH, API reachability, and pulled models."
	ollamaStatusIntroNote  = "Does not change config or start/stop the service."
	ollamaStatusUsage      = "ait ai ollama status"
	ollamaStatusOutputBody = `  ai.tiered.enabled
  ai.providers.ollama.enabled
  base_url, model, models_dir (OLLAMA_MODELS)
  binary, api
  detail (when unreachable)
  pulled_models (when API is up)
`
	ollamaSetupIntroMain    = "Enable ai.tiered local-first routing and configure providers.ollama."
	ollamaSetupIntroModels  = "Prompts for a models directory (OLLAMA_MODELS) unless --models-dir or config already set."
	ollamaSetupIntroInstall = "When ollama is not in PATH, you are prompted [y/N] before installing (brew/winget/install.sh)."
	ollamaSetupIntroDecline = "Declining install still saves tiered config; use local Ollama when you install it manually."
	ollamaSetupUsage        = "ait ai ollama setup [flags]"
	ollamaSetupFlagsBody    = `  --model string       Ollama model to pull (default llama3.2)
  --base-url string    Ollama API base (default http://127.0.0.1:11434/v1)
  --models-dir string  Model storage directory (OLLAMA_MODELS); interactive when unset
  -y, --yes            Allow install without prompting
  --no-install         Never install Ollama
  --no-pull            Skip ollama pull
  --dry-run            Show planned steps only
`
	ollamaSetupExample       = "ait ai ollama setup"
	ollamaSetupExampleNoInst = "ait ai ollama setup --no-install"
	ollamaStartIntroMain     = "Start ollama serve in the background. Does not install Ollama or change config."
	ollamaStartIntroAPI      = "If the API is already reachable, prints a message and exits successfully."
	ollamaStartUsage         = "ait ai ollama start [flags]"
	ollamaStartFlagsBody     = `  --wait               Poll until /api/tags responds
  --base-url string    API base to probe (default from config)
`
	ollamaStartExample      = "ait ai ollama start"
	ollamaStopIntroMain     = "Stop local ollama serve processes. Does not disable Ollama in config."
	ollamaStopIntroWindows  = "On Windows, may not quit the Ollama desktop app; close it manually if needed."
	ollamaStopUsage         = "ait ai ollama stop"
	ollamaDisableIntroMain  = "Set ai.providers.ollama.enabled to false so AiT stops routing to Ollama."
	ollamaDisableIntroAlias = "Same config effect as: ait ai provider disable ollama. Does not turn off ai.tiered.enabled."
	ollamaDisableUsage      = "ait ai ollama disable [flags]"
	ollamaDisableFlagsBody  = `  --stop    Also run ollama stop after disabling
`
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
		exampleAIOllamaStatus,
		exampleAIOllamaSetupYes,
		ollamaExampleStartWait,
		ollamaExampleStop,
		exampleAIOllamaDisableStop,
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
