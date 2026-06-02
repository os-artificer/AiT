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

package enable

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

const (
	introMain     = "Enable routing to a provider. Cloud providers need an API key in config or keyring."
	introOllama   = "For ollama, also runs availability checks (install [y/N] unless --yes / --no-install)."
	usageLine     = "ait ai provider enable <provider> [flags]"
	providersBody = "  deepseek, minimax, ollama\n"
	flagsOllama   = `  -y, --yes            Allow install without prompting
  --no-install         Skip Ollama install
  --no-pull            Skip ollama pull
  --model, --base-url, --models-dir   Passed to Ollama ensure logic
`
	exampleDeepseek = "ait ai provider enable deepseek"
	exampleOllama   = "ait ai provider enable ollama --yes"
)

// WriteHelp prints ait ai provider enable --help to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(introMain, introOllama); err != nil {
		return err
	}
	if err := b.WriteUsage(usageLine); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionProviders, providersBody); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionFlagsOllama, flagsOllama); err != nil {
		return err
	}
	if err := b.WriteExamples(exampleDeepseek, exampleOllama); err != nil {
		return err
	}
	return b.Finish("enable")
}
