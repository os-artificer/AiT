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
	"fmt"
	"slices"

	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// LongHint returns the standard one-line Cobra Long field for cmdName.
func LongHint(cmdName string) string {
	if cmdName == toolname.AISsh {
		return fmt.Sprintf(longHintRoot, toolname.AISsh)
	}
	return fmt.Sprintf(longHintSub, toolname.CLI(cmdName))
}

// LongHintChatOnce returns Cobra Long for one-shot ai ask/agent (stdin note preserved).
func LongHintChatOnce(subcmd string) string {
	return fmt.Sprintf(longHintChatOnce, toolname.CLI(subcmd))
}

// HelpRequested reports whether argv is only -h, --help, or help.
func HelpRequested(argv []string) bool {
	return len(argv) == 1 && isHelpArg(argv[0])
}

// isHelpArg returns true when arg equals -h, --help, or help.
func isHelpArg(arg string) bool {
	return arg == HelpArgH || arg == HelpArgHelp || arg == HelpArgCmd
}

// HasHelpArg reports whether args contain a bare "help" positional argument
// (not a flag value like --load help or --switch help).
func HasHelpArg(args []string) bool {
	return slices.Contains(args, HelpArgCmd)
}

// HelpFooter returns the standard trailing Flags section for cmdName.
func HelpFooter(cmdName string) string {
	return fmt.Sprintf(helpFooterFlagsPrefix, cmdName)
}
