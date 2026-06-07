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

package login

import (
	"io"
	"strings"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

const helpShort = `  login flags (see "ait login --help" for full detail):
    -p port  -l user  -i identity  -F config  -J jump  -o key=value
    --save  --alias name  --tmux  --password P
    --host H  --user U  --port N  --identity P  (legacy, same as ssh flags above)
`

// HelpRequested reports whether argv is only -h, --help, or help.
func HelpRequested(argv []string) bool {
	return pkghelpdoc.HelpRequested(argv)
}

// WriteHelp prints full login usage, flags, and examples to w.
func WriteHelp(w io.Writer) error {
	_, err := io.WriteString(w, HelpFull)
	return err
}

// WriteHelpShort prints a compact flag list for the ait shell REPL to w.
func WriteHelpShort(w io.Writer) error {
	_, err := io.WriteString(w, helpShort)
	return err
}

// helpContains checks substrings for tests.
func helpContains(substrs ...string) bool {
	for _, s := range substrs {
		if !strings.Contains(HelpFull, s) {
			return false
		}
	}
	return true
}
