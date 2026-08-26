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

package start

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

const (
	introMain = "Start ollama serve in the background. Does not install Ollama or change config."
	introAPI  = "If the API is already reachable, prints a message and exits successfully."
	flagsBody = `  --wait               Poll until /api/tags responds
  --base-url string    API base to probe (default from config)
`
)

var (
	usageLine        = toolname.AICLI("ollama start [flags]")
	exampleStart     = toolname.AICLI("ollama start")
	exampleStartWait = toolname.AICLI("ollama start --wait")
)

// WriteHelp prints ai ollama start --help to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(introMain, introAPI); err != nil {
		return err
	}
	if err := b.WriteUsage(usageLine); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionFlags, flagsBody); err != nil {
		return err
	}
	if err := b.WriteExamples(exampleStart, exampleStartWait); err != nil {
		return err
	}
	return b.Finish("start")
}
