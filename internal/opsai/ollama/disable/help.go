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
	"fmt"
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

const (
	introMain = "Set ai.providers.ollama.enabled to false so AiT stops routing to Ollama."
	flagsBody = `  --stop    Also run ollama stop after disabling
`
)

var (
	introAlias         = fmt.Sprintf("Same config effect as: %s. Does not turn off ai.tiered.enabled.", toolname.AICLI("provider disable ollama"))
	usageLine          = toolname.AICLI("ollama disable [flags]")
	exampleDisable     = toolname.AICLI("ollama disable")
	exampleDisableStop = toolname.AICLI("ollama disable --stop")
)

// WriteHelp prints ai ollama disable --help to w.
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
