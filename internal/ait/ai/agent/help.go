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

package agent

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

const (
	helpFlagsBody = "  --model   Override cloud provider (deepseek|minimax); inherited from ait ai\n"
	introMain     = "Send one message in agent mode: the AI may execute shell commands after confirmation."
	introSSH      = "Does not enter interactive AI mode. Inside SSH use !ait ai agent <message> or !! agent <message>."
	usageLine     = "ait ai agent [message...]"
	inputBody     = `  message...   Prompt text (joined with spaces)
  (no args)    Read prompt from stdin until EOF (e.g. echo "question" | ait ai agent)
`
	exampleQuote = `ait ai agent "find large files in /var/log"`
	exampleModel = "ait ai agent --model minimax check disk usage"
	exampleStdin = `echo "list running containers" | ait ai agent`
	exampleSSH   = "!ait ai agent how do I find large files"
)

// WriteHelp prints ait ai agent --help to w.
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
	return b.Finish("agent")
}
