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

const (
	providerIntroMain         = "Enable or disable AI provider routing without editing config.yaml by hand."
	providerIntroOllamaHint   = "For Ollama install, models directory, and tiered setup, prefer: ait ai ollama setup"
	providerUsageMain         = "ait ai provider"
	providerEnableIntroMain   = "Enable routing to a provider. Cloud providers need an API key in config or keyring."
	providerEnableIntroOllama = "For ollama, also runs availability checks (install [y/N] unless --yes / --no-install)."
	providerEnableUsage       = "ait ai provider enable <provider> [flags]"
	providerEnableProviders   = "  deepseek, minimax, ollama\n"
	providerEnableFlagsOllama = `  -y, --yes            Allow install without prompting
  --no-install         Skip Ollama install
  --no-pull            Skip ollama pull
  --model, --base-url, --models-dir   Passed to Ollama ensure logic
`
	providerDisableIntro  = "Disable routing to a provider; no Chat/Assist traffic is sent to it."
	providerDisableUsage  = "ait ai provider disable <provider>"
	providerDisableOllama = "ait ai provider disable ollama"
)

// WriteAIProviderHelp prints ait ai provider --help to w.
func WriteAIProviderHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(providerIntroMain, providerIntroOllamaHint); err != nil {
		return err
	}
	if err := b.WriteUsage(providerUsageMain); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionSubcommands, providerSubcommands); err != nil {
		return err
	}
	if err := b.WriteExamples(
		exampleAIProviderEnableDeepseek,
		exampleAIProviderDisableMinimax,
		exampleAIProviderEnableOllamaYes,
	); err != nil {
		return err
	}
	return b.Finish("provider")
}

// WriteProviderEnableHelp prints ait ai provider enable --help to w.
func WriteProviderEnableHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(providerEnableIntroMain, providerEnableIntroOllama); err != nil {
		return err
	}
	if err := b.WriteUsage(providerEnableUsage); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionProviders, providerEnableProviders); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionFlagsOllama, providerEnableFlagsOllama); err != nil {
		return err
	}
	if err := b.WriteExamples(
		exampleAIProviderEnableDeepseek,
		exampleAIProviderEnableOllamaYes,
	); err != nil {
		return err
	}
	return b.Finish("enable")
}

// WriteProviderDisableHelp prints ait ai provider disable --help to w.
func WriteProviderDisableHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(providerDisableIntro); err != nil {
		return err
	}
	if err := b.WriteUsage(providerDisableUsage); err != nil {
		return err
	}
	if err := b.WriteExamples(exampleAIProviderDisableMinimax, providerDisableOllama); err != nil {
		return err
	}
	return b.Finish("disable")
}
