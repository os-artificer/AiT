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

package builtin

import (
	"context"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

// Try runs a built-in REPL command when words[0] matches the registry.
// Returns handled false when the first word is not a built-in.
func Try(ctx context.Context, env Env, words []string) (exitREPL, handled bool, err error) {
	if len(words) == 0 {
		return false, false, nil
	}

	switch words[0] {
	case pkghelpdoc.HelpArgCmd, pkghelpdoc.HelpArgH, pkghelpdoc.HelpArgHelp:
		exit, runErr := runHelp(ctx, env, words[1:])
		return exit, true, runErr
	}

	cmd, ok := commands[words[0]]
	if !ok {
		return false, false, nil
	}

	if cmd.blockedWhenActive != nil && env.Mgr != nil && env.Mgr.HasActiveSession() {
		return false, true, cmd.blockedWhenActive()
	}

	exit, runErr := cmd.run(ctx, env, words[1:])
	return exit, true, runErr
}
