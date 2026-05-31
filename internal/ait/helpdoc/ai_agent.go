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

package helpdoc

import "io"

const (
	aiAgentIntroMain = "Send one message in agent mode: the AI may execute shell commands after confirmation."
	aiAgentIntroSSH  = "Does not enter interactive AI mode. Inside SSH use !ait ai agent <message> or !! agent <message>."
	aiAgentUsage     = "ait ai agent [message...]"
	aiAgentInputBody = `  message...   Prompt text (joined with spaces)
  (no args)    Read prompt from stdin until EOF (e.g. echo "question" | ait ai agent)
`
	aiAgentExampleQuote = `ait ai agent "find large files in /var/log"`
	aiAgentExampleModel = "ait ai agent --model minimax check disk usage"
	aiAgentExampleStdin = `echo "list running containers" | ait ai agent`
	aiAgentExampleSSH   = "!ait ai agent how do I find large files"
)

// WriteAIAgentHelp prints ait ai agent --help to w.
func WriteAIAgentHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(aiAgentIntroMain, aiAgentIntroSSH); err != nil {
		return err
	}
	if err := b.WriteUsage(aiAgentUsage); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionInput, aiAgentInputBody); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionFlags, aiAgentHelpFlagsBody); err != nil {
		return err
	}
	if err := b.WriteExamples(
		aiAgentExampleQuote,
		aiAgentExampleModel,
		aiAgentExampleStdin,
		aiAgentExampleSSH,
	); err != nil {
		return err
	}
	return b.Finish("agent")
}
