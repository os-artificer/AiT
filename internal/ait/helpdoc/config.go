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

const configIntro = "Read and write AiT YAML configuration (session TTL, AI providers, API keys)."
const configSubcommands = `  set <key> <value>   set a configuration key
  get <key>           print a configuration value
`

const (
	configSetIntro     = "Set a configuration key in the AiT YAML config file."
	configSetUsage     = "ait config set <key> <value>"
	configSetArguments = "  key     configuration key (see Keys below)\n  value   new value\n"
	configGetIntro     = "Print a configuration value."
	configGetUsage     = "ait config get <key>"
	configGetArguments = "  key     configuration key (see Keys below)\n"
)

const configKeys = `  session.auth_ttl              duration for session unlock (e.g. 168h, 7d)
  ai.default                    default AI provider (deepseek|minimax)
  ai.fallback                   comma-separated failover order (optional)
  ai.providers.<id>.base_url    provider API base URL
  ai.providers.<id>.model       model name
  ai.providers.<id>.api_style   openai or anthropic (minimax only for anthropic)
  ai.<provider>.api_key         API key in OS keyring (recommended)
  deepseek.api_key              alias for ai.deepseek.api_key
  minimax.api_key               alias for ai.minimax.api_key
  deepseek.base_url / .model    aliases for ai.providers.deepseek.*
`

// WriteConfigHelp prints ait config --help to w.
func WriteConfigHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(configIntro); err != nil {
		return err
	}
	if err := b.WriteUsage("ait config"); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionSubcommands, configSubcommands); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionKeys, configKeys); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait config get session.auth_ttl",
		"ait config set deepseek.model deepseek-chat",
	); err != nil {
		return err
	}
	return b.Finish("config")
}

// WriteConfigSetHelp prints ait config set --help to w.
func WriteConfigSetHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(configSetIntro); err != nil {
		return err
	}
	if err := b.WriteUsage(configSetUsage); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionArguments, configSetArguments); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionKeys, configKeys); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait config set session.auth_ttl 336h",
		"ait config set ai.deepseek.api_key <key>",
	); err != nil {
		return err
	}
	return b.Finish("set")
}

// WriteConfigGetHelp prints ait config get --help to w.
func WriteConfigGetHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(configGetIntro); err != nil {
		return err
	}
	if err := b.WriteUsage(configGetUsage); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionArguments, configGetArguments); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionKeys, configKeys); err != nil {
		return err
	}
	if err := b.WriteExamples("ait config get deepseek.model"); err != nil {
		return err
	}
	return b.Finish("get")
}
