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

package ask

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

const (
	helpFlagsBody = "  --model   Override cloud provider (deepseek|minimax); inherited from ait ai\n"
	introMain     = "Send a single message to the configured AI provider and print the reply."
	introSSH      = "Does not enter interactive AI mode. Inside SSH use !ait ai ask <message> or !! ask <message>."
	usageLine     = "ait ai ask [message...]"
	inputBody     = `  message...   Prompt text (joined with spaces)
  (no args)    Read prompt from stdin until EOF (e.g. echo "question" | ait ai ask)
`
	exampleQuote = `ait ai ask "explain this error"`
	exampleModel = "ait ai ask --model minimax summarize logs"
	exampleStdin = `echo "hello" | ait ai ask`
	exampleSSH   = "!ait ai ask how do I find large files"
)

// WriteHelp prints ait ai ask --help to w.
func WriteHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(introMain, introSSH); err != nil {
		return err
	}
	if err := b.WriteUsage(usageLine); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionInput, inputBody); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionFlags, helpFlagsBody); err != nil {
		return err
	}
	if err := b.WriteExamples(exampleQuote, exampleModel, exampleStdin, exampleSSH); err != nil {
		return err
	}
	return b.Finish("ask")
}
