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

package helpdoc

import "io"

const ollamaSubcommands = `  status            show tiered/enable flags and API reachability
  setup [flags]     enable tiered routing; install/start/pull Ollama
  start [flags]     start ollama serve in the background
  stop              stop local ollama serve
  disable [flags]   set ai.providers.ollama.enabled false (optional --stop)
`

// WriteAIOllamaHelp prints ait ai ollama --help to w.
func WriteAIOllamaHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(
		"Manage local Ollama for tiered AI routing (under ait ai).",
		"Run 'ait ai ollama <subcommand> --help' for flags and examples.",
	); err != nil {
		return err
	}
	if err := b.WriteUsage("ait ai ollama"); err != nil {
		return err
	}
	if err := b.WriteSection("Subcommands", ollamaSubcommands); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait ai ollama status",
		"ait ai ollama setup --yes",
		"ait ai ollama start --wait",
		"ait ai ollama stop",
		"ait ai ollama disable --stop",
	); err != nil {
		return err
	}
	return b.Finish("ollama")
}

// WriteOllamaStatusHelp prints ait ai ollama status --help to w.
func WriteOllamaStatusHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(
		"Read-only status for Ollama: config flags, binary on PATH, API reachability, and pulled models.",
		"Does not change config or start/stop the service.",
	); err != nil {
		return err
	}
	if err := b.WriteUsage("ait ai ollama status"); err != nil {
		return err
	}
	if err := b.WriteSection("Output", `  ai.tiered.enabled
  ai.providers.ollama.enabled
  base_url, model, models_dir (OLLAMA_MODELS)
  binary, api
  detail (when unreachable)
  pulled_models (when API is up)
`); err != nil {
		return err
	}
	if err := b.WriteExamples("ait ai ollama status"); err != nil {
		return err
	}
	return b.Finish("status")
}

// WriteOllamaSetupHelp prints ait ai ollama setup --help to w.
func WriteOllamaSetupHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(
		"Enable ai.tiered local-first routing and configure providers.ollama.",
		"Prompts for a models directory (OLLAMA_MODELS) unless --models-dir or config already set.",
		"When ollama is not in PATH, you are prompted [y/N] before installing (brew/winget/install.sh).",
		"Declining install still saves tiered config; use local Ollama when you install it manually.",
	); err != nil {
		return err
	}
	if err := b.WriteUsage("ait ai ollama setup [flags]"); err != nil {
		return err
	}
	if err := b.WriteSection("Flags", `  --model string       Ollama model to pull (default llama3.2)
  --base-url string    Ollama API base (default http://127.0.0.1:11434/v1)
  --models-dir string  Model storage directory (OLLAMA_MODELS); interactive when unset
  -y, --yes            Allow install without prompting
  --no-install         Never install Ollama
  --no-pull            Skip ollama pull
  --dry-run            Show planned steps only
`); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait ai ollama setup",
		"ait ai ollama setup --yes",
		"ait ai ollama setup --no-install",
	); err != nil {
		return err
	}
	return b.Finish("setup")
}

// WriteOllamaStartHelp prints ait ai ollama start --help to w.
func WriteOllamaStartHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(
		"Start ollama serve in the background. Does not install Ollama or change config.",
		"If the API is already reachable, prints a message and exits successfully.",
	); err != nil {
		return err
	}
	if err := b.WriteUsage("ait ai ollama start [flags]"); err != nil {
		return err
	}
	if err := b.WriteSection("Flags", `  --wait               Poll until /api/tags responds
  --base-url string    API base to probe (default from config)
`); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait ai ollama start",
		"ait ai ollama start --wait",
	); err != nil {
		return err
	}
	return b.Finish("start")
}

// WriteOllamaStopHelp prints ait ai ollama stop --help to w.
func WriteOllamaStopHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(
		"Stop local ollama serve processes. Does not disable Ollama in config.",
		"On Windows, may not quit the Ollama desktop app; close it manually if needed.",
	); err != nil {
		return err
	}
	if err := b.WriteUsage("ait ai ollama stop"); err != nil {
		return err
	}
	if err := b.WriteExamples("ait ai ollama stop"); err != nil {
		return err
	}
	return b.Finish("stop")
}

// WriteOllamaDisableHelp prints ait ai ollama disable --help to w.
func WriteOllamaDisableHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(
		"Set ai.providers.ollama.enabled to false so AiT stops routing to Ollama.",
		"Same config effect as: ait ai provider disable ollama. Does not turn off ai.tiered.enabled.",
	); err != nil {
		return err
	}
	if err := b.WriteUsage("ait ai ollama disable [flags]"); err != nil {
		return err
	}
	if err := b.WriteSection("Flags", `  --stop    Also run ollama stop after disabling
`); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait ai ollama disable",
		"ait ai ollama disable --stop",
	); err != nil {
		return err
	}
	return b.Finish("disable")
}
