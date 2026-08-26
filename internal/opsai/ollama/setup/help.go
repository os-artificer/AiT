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

package setup

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

const (
	introMain    = "Enable ai.tiered local-first routing and configure providers.ollama."
	introModels  = "Prompts for a models directory (OLLAMA_MODELS) unless --models-dir or config already set."
	introInstall = "When ollama is not in PATH, you are prompted [y/N] before installing (brew/winget/install.sh)."
	introDecline = "Declining install still saves tiered config; use local Ollama when you install it manually."
	flagsBody    = `  --model string       Ollama model to pull (default llama3.2)
  --base-url string    Ollama API base (default http://127.0.0.1:11434/v1)
  --models-dir string  Model storage directory (OLLAMA_MODELS); interactive when unset
  -y, --yes            Allow install without prompting
  --no-install         Never install Ollama
  --no-pull            Skip ollama pull
  --dry-run            Show planned steps only
`
)

var (
	usageLine        = toolname.AICLI("ollama setup [flags]")
	exampleSetup     = toolname.AICLI("ollama setup")
	exampleSetupYes  = toolname.AICLI("ollama setup --yes")
	exampleNoInstall = toolname.AICLI("ollama setup --no-install")
)

// WriteHelp prints ai ollama setup --help to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(introMain, introModels, introInstall, introDecline); err != nil {
		return err
	}
	if err := b.WriteUsage(usageLine); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionFlags, flagsBody); err != nil {
		return err
	}
	if err := b.WriteExamples(exampleSetup, exampleSetupYes, exampleNoInstall); err != nil {
		return err
	}
	return b.Finish("setup")
}
