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

package ai

import (
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"

	"github.com/spf13/cobra"
)

// skipAIExperimentalRequire reports whether PersistentPreRunE should skip the gate
// so --help / -h / parent positional help still work when experimental AI is disabled.
// Bare "help" in ask/agent message args must not skip (gate stays enforced).
// Only inspects Cobra command args and help flags — never os.Args.
func skipAIExperimentalRequire(cmd *cobra.Command, args []string) bool {
	if helpFlagChanged(cmd) {
		return true
	}
	for _, a := range args {
		if a == pkghelpdoc.HelpArgH || a == pkghelpdoc.HelpArgHelp {
			return true
		}
	}
	// Bare "help" only on aissh ai parent (positional help), never on ask/agent messages.
	return cmd.Name() == Name && len(args) == 1 && args[0] == pkghelpdoc.HelpArgCmd
}

func helpFlagChanged(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if f := c.Flags().Lookup("help"); f != nil && f.Changed {
			return true
		}
	}
	return false
}
