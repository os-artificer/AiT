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

package validate

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

const (
	introMain = "Check config.yaml for YAML syntax and configuration value errors. Does not create a missing file."
	checks    = `  YAML syntax (line hints from the parser)
  session.auth_ttl          valid duration
  ai.default / ai.fallback  known cloud providers
  ai.providers.*            base_url, api_style, model, models_dir
  ai.tiered.*               tiered routing fields
`
)

var usageLine = toolname.CLI("config validate")

// WriteHelp prints aissh config validate --help to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(introMain); err != nil {
		return err
	}
	if err := b.WriteUsage(usageLine); err != nil {
		return err
	}
	if err := b.WriteSection("Checks:", checks); err != nil {
		return err
	}
	if err := b.WriteExamples(toolname.CLI("config validate")); err != nil {
		return err
	}
	return b.Finish("validate")
}
