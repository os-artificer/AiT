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

package cli

// AI — provider and ollama CLI strings reused across commands.
const (
	errUnknownProviderFmt  = "unknown provider %q (use deepseek, minimax, or ollama)"
	flagNameBaseURL        = "base-url"
	flagNameModelsDir      = "models-dir"
	flagNameNoInstall      = "no-install"
	flagNameNoPull         = "no-pull"
	flagNameYes            = "yes"
	flagShorthandYes       = "y"
	msgEnabledProviderFmt  = "enabled provider %s\n"
	msgDisabledProviderFmt = "disabled provider %s\n"
)

// Config — dotted key fragments used when resolving API keys.
const (
	configKeyPrefixAI     = "ai."
	configKeySuffixAPIKey = ".api_key"
)

// Session — flag names shared by load and session switch.
const (
	flagNameTmux     = "tmux"
	flagNamePassword = "password"
	flagNameForce    = "force"
)
