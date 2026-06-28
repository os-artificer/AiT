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
	"fmt"
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

var introSSH = "Does not enter interactive AI mode. Inside SSH use " + toolname.MetaLine("ai ask <message>") +
	" or !! ask <message>."

const introMain = "Send a single message to the configured AI provider and print the reply."

var (
	usageLine    = toolname.CLI("ai ask [message...]")
	inputBody    = fmt.Sprintf("  message...   Prompt text (joined with spaces)\n  (no args)    Read prompt from stdin until EOF (e.g. echo \"question\" | %s)\n", toolname.CLI("ai ask"))
	exampleQuote = toolname.CLI(`ai ask "explain this error"`)
	exampleModel = toolname.CLI("ai ask --model minimax summarize logs")
	exampleStdin = fmt.Sprintf(`echo "hello" | %s`, toolname.CLI("ai ask"))
)

// WriteHelp prints aitssh ai ask --help to w.
func WriteHelp(w io.Writer) error {
	return pkghelpdoc.WriteAIChatSubcommandHelp(w, "ask", introMain, introSSH, usageLine, inputBody,
		exampleQuote, exampleModel, exampleStdin, pkghelpdoc.ExampleAIAskSSH)
}
