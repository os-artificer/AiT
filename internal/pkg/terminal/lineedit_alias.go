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

package terminal

import (
	"bufio"
	"io"

	"github.com/os-artificer/AiT/internal/pkg/terminal/lineedit"
)

// LineEditOptions configures ReadLine behavior in raw TTY mode (alias of lineedit.Options).
type LineEditOptions = lineedit.Options

// ReadLine reads one edited line from br while stdin is in raw mode.
// See lineedit.ReadLine for full behavior, including Ctrl+R reverse search and Ctrl+L redraw.
func ReadLine(br *bufio.Reader, out io.Writer, opt LineEditOptions) (string, error) {
	return lineedit.ReadLine(br, out, opt)
}

// ReadLineRaw reads one edited line without history recall or recording (legacy wrapper).
func ReadLineRaw(br *bufio.Reader, out io.Writer, prompt string) (string, error) {
	return lineedit.ReadLineRaw(br, out, prompt)
}
