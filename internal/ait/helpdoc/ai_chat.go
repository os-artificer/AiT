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

// WriteAIChatHelp prints ait ai chat --help to w.
func WriteAIChatHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(
		"Send a single message to the configured AI provider and print the reply.",
		"Does not enter interactive AI mode. Inside SSH use !ait ai chat <message> or !! chat <message>.",
	); err != nil {
		return err
	}
	if err := b.WriteUsage("ait ai chat [message...]"); err != nil {
		return err
	}
	if err := b.WriteSection("Input", `  message...   Prompt text (joined with spaces)
  (no args)    Read prompt from stdin until EOF (e.g. echo "question" | ait ai chat)
`); err != nil {
		return err
	}
	if err := b.WriteSection("Flags", "  --model   Override cloud provider (deepseek|minimax); inherited from ait ai\n"); err != nil {
		return err
	}
	if err := b.WriteExamples(
		`ait ai chat "explain this error"`,
		"ait ai chat --model minimax summarize logs",
		`echo "hello" | ait ai chat`,
		"!ait ai chat how do I find large files",
	); err != nil {
		return err
	}
	return b.Finish("chat")
}
