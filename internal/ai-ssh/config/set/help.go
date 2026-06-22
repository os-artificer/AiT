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

package set

import (
	"io"

	"github.com/os-artificer/AiT/internal/ai-ssh/config/shared"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

const (
	introMain = "Set a configuration key in the AiT YAML config file."
	usageLine = "ai-ssh config set <key> <value>"
	arguments = "  key     configuration key (see Keys below)\n  value   new value\n"
)

// WriteHelp prints ai-ssh config set --help to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(introMain); err != nil {
		return err
	}
	if err := b.WriteUsage(usageLine); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionArguments, arguments); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionKeys, shared.KeysHelp); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ai-ssh config set session.auth_ttl 336h",
		"ai-ssh config set ai.deepseek.api_key <key>",
	); err != nil {
		return err
	}
	return b.Finish("set")
}
