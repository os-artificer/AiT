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

package status

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

const (
	introMain  = "Read-only status for Ollama: config flags, binary on PATH, API reachability, and pulled models."
	introNote  = "Does not change config or start/stop the service."
	outputBody = `  ai.tiered.enabled
  ai.providers.ollama.enabled
  base_url, model, models_dir (OLLAMA_MODELS)
  binary, api
  detail (when unreachable)
  pulled_models (when API is up)
`
)

var (
	usageLine     = toolname.AICLI("ollama status")
	exampleStatus = toolname.AICLI("ollama status")
)

// WriteHelp prints ai ollama status --help to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(introMain, introNote); err != nil {
		return err
	}
	if err := b.WriteUsage(usageLine); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionOutput, outputBody); err != nil {
		return err
	}
	if err := b.WriteExamples(exampleStatus); err != nil {
		return err
	}
	return b.Finish("status")
}
