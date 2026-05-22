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
	aiChatIntroMain = "Send a single message to the configured AI provider and print the reply."
	aiChatIntroSSH  = "Does not enter interactive AI mode. Inside SSH use !ait ai chat <message> or !! chat <message>."
	aiChatUsage     = "ait ai chat [message...]"
	aiChatInputBody = `  message...   Prompt text (joined with spaces)
  (no args)    Read prompt from stdin until EOF (e.g. echo "question" | ait ai chat)
`
	aiChatExampleQuote = `ait ai chat "explain this error"`
	aiChatExampleModel = "ait ai chat --model minimax summarize logs"
	aiChatExampleStdin = `echo "hello" | ait ai chat`
	aiChatExampleSSH   = "!ait ai chat how do I find large files"
)

// WriteAIChatHelp prints ait ai chat --help to w.
func WriteAIChatHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(aiChatIntroMain, aiChatIntroSSH); err != nil {
		return err
	}
	if err := b.WriteUsage(aiChatUsage); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionInput, aiChatInputBody); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionFlags, aiChatHelpFlagsBody); err != nil {
		return err
	}
	if err := b.WriteExamples(
		aiChatExampleQuote,
		aiChatExampleModel,
		aiChatExampleStdin,
		aiChatExampleSSH,
	); err != nil {
		return err
	}
	return b.Finish("chat")
}
