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

import (
	"io"

	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

const sectionTTYKeysCommon = `  Ctrl+A, Home     move cursor to start of line
  Ctrl+E, End      move cursor to end of line
  Left, Right      move cursor one character (UTF-8 aware)
  Up, Down         recall previous / next line from loaded input history (incl. prior sessions)
  Ctrl+R           reverse incremental search through input history
  Backspace, Del   delete character before cursor
  Ctrl+U           delete entire line
  Ctrl+K           delete from cursor to end of line
  Ctrl+W           delete word before cursor
  Ctrl+L           clear screen and redraw prompt (line kept)
  Ctrl+C           cancel current line (REPL continues)
  Ctrl+D           exit REPL when line is empty (EOF)
  Tab              insert indentation
`

const sectionTTYKeysShellNote = "  clear            clear terminal screen (command)\n"

var sectionTTYKeysAINote = `  clear            clear AI chat history (not the same as Ctrl+L)
  In SSH remote shell, meta line input (!ait / !!) supports the same editing keys as local REPL.
  In ` + toolname.MetaLine("local") + ` mode, !! is passed to the local shell for bash/zsh history expansion.
`

// TTYKeysSurface selects shell vs AI key help notes.
type TTYKeysSurface int

const (
	TTYKeysShell TTYKeysSurface = iota
	TTYKeysAI
)

// WriteTTYKeysHelp writes the Keys section for local TTY REPL input.
func WriteTTYKeysHelp(w io.Writer, surface TTYKeysSurface) error {
	body := sectionTTYKeysCommon
	if surface == TTYKeysShell {
		body += sectionTTYKeysShellNote
	} else {
		body += sectionTTYKeysAINote
	}
	b := NewBody(w)
	return b.WriteSection(SectionKeys, body)
}
