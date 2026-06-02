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

package disable

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

const (
	introMain  = "Set ai.providers.ollama.enabled to false so AiT stops routing to Ollama."
	introAlias = "Same config effect as: ait ai provider disable ollama. Does not turn off ai.tiered.enabled."
	usageLine  = "ait ai ollama disable [flags]"
	flagsBody  = `  --stop    Also run ollama stop after disabling
`
	exampleDisable     = "ait ai ollama disable"
	exampleDisableStop = "ait ai ollama disable --stop"
)

// WriteHelp prints ait ai ollama disable --help to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(introMain, introAlias); err != nil {
		return err
	}
	if err := b.WriteUsage(usageLine); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionFlags, flagsBody); err != nil {
		return err
	}
	if err := b.WriteExamples(exampleDisable, exampleDisableStop); err != nil {
		return err
	}
	return b.Finish("disable")
}
