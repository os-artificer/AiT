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
	"fmt"
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

var introSSH = "Does not enter interactive AI mode. Use " + toolname.AICLI("connect user@host") +
	" for remote Agent, or run from the ait> REPL."

const introMain = "Send one message in agent mode: the AI may execute shell commands after confirmation."

var (
	usageLine    = toolname.AICLI("agent [message...]")
	inputBody    = fmt.Sprintf("  message...   Prompt text (joined with spaces)\n  (no args)    Read prompt from stdin until EOF (e.g. echo \"question\" | %s)\n", toolname.AICLI("agent"))
	exampleQuote = toolname.AICLI(`agent "find large files in /var/log"`)
	exampleModel = toolname.AICLI("agent --model minimax check disk usage")
	exampleStdin = fmt.Sprintf(`echo "list running containers" | %s`, toolname.AICLI("agent"))
)

// WriteHelp prints aissh ai agent --help to w.
func WriteHelp(w io.Writer) error {
	return pkghelpdoc.WriteAIChatSubcommandHelp(w, "agent", introMain, introSSH, usageLine, inputBody,
		exampleQuote, exampleModel, exampleStdin, pkghelpdoc.ExampleAIAgentLocal)
}
