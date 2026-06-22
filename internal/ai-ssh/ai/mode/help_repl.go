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

package mode

import (
	"io"

	"github.com/os-artificer/AiT/internal/pkg/ai"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

// WriteREPLHelp prints in-AI-mode help for the given scope.
func WriteREPLHelp(w io.Writer, scope ai.Scope) error {
	b := pkghelpdoc.NewBody(w)
	intro := pkghelpdoc.AIIntroLocal
	cmds := pkghelpdoc.AIReplCommands + pkghelpdoc.AIReplCommandsLocalExtra
	examples := "mode\nrun ls\nai-ssh list\napprove each\n"
	if scope == ai.ScopeRemote {
		intro = pkghelpdoc.AIIntroRemote
		cmds = pkghelpdoc.AIReplCommands
		examples = "mode agent\nrun uptime\napprove session\n"
	}
	if err := b.WriteIntro(intro); err != nil {
		return err
	}
	if err := b.WriteUsage("help | mode [ask|agent] | approve [each|session] | model [provider] | run <cmd> | history | clear | exit"); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionCommands, cmds); err != nil {
		return err
	}
	if err := pkghelpdoc.WriteTTYKeysHelp(w, pkghelpdoc.TTYKeysAI); err != nil {
		return err
	}
	if err := b.WriteExamples(examples); err != nil {
		return err
	}
	return b.Finish("help")
}
