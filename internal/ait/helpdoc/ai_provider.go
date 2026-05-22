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

const providerSubcommands = `  enable <provider>   set ai.providers.<id>.enabled true
  disable <provider>  set ai.providers.<id>.enabled false
`

// WriteAIProviderHelp prints ait ai provider --help to w.
func WriteAIProviderHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(
		"Enable or disable AI provider routing without editing config.yaml by hand.",
		"For Ollama install, models directory, and tiered setup, prefer: ait ai ollama setup",
	); err != nil {
		return err
	}
	if err := b.WriteUsage("ait ai provider"); err != nil {
		return err
	}
	if err := b.WriteSection("Subcommands", providerSubcommands); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait ai provider enable deepseek",
		"ait ai provider disable minimax",
		"ait ai provider enable ollama --yes",
	); err != nil {
		return err
	}
	return b.Finish("provider")
}

// WriteProviderEnableHelp prints ait ai provider enable --help to w.
func WriteProviderEnableHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(
		"Enable routing to a provider. Cloud providers need an API key in config or keyring.",
		"For ollama, also runs availability checks (install [y/N] unless --yes / --no-install).",
	); err != nil {
		return err
	}
	if err := b.WriteUsage("ait ai provider enable <provider> [flags]"); err != nil {
		return err
	}
	if err := b.WriteSection("Providers", "  deepseek, minimax, ollama\n"); err != nil {
		return err
	}
	if err := b.WriteSection("Flags (ollama)", `  -y, --yes            Allow install without prompting
  --no-install         Skip Ollama install
  --no-pull            Skip ollama pull
  --model, --base-url, --models-dir   Passed to Ollama ensure logic
`); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait ai provider enable deepseek",
		"ait ai provider enable ollama --yes",
	); err != nil {
		return err
	}
	return b.Finish("enable")
}

// WriteProviderDisableHelp prints ait ai provider disable --help to w.
func WriteProviderDisableHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro("Disable routing to a provider; no Chat/Assist traffic is sent to it."); err != nil {
		return err
	}
	if err := b.WriteUsage("ait ai provider disable <provider>"); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait ai provider disable minimax",
		"ait ai provider disable ollama",
	); err != nil {
		return err
	}
	return b.Finish("disable")
}
