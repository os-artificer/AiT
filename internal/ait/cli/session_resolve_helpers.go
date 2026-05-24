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

package cli

import "github.com/os-artificer/AiT/internal/pkg/errhint"

// resolveLoadTarget returns the session id or alias from --load and optional positional args.
// An empty result means interactive pick. Returns an error when extra positional args are present.
func resolveLoadTarget(loadTarget string, args []string) (string, error) {
	if loadTarget == "-" {
		loadTarget = ""
	}
	if loadTarget != "" {
		return loadTarget, nil
	}
	if len(args) == 1 {
		return args[0], nil
	}
	if len(args) > 1 {
		return "", errhint.WithTryMsg("unexpected arguments with --load",
			"Pass the session id or alias as the --load value or as a single positional argument.",
			"ait session --load <alias>",
			"ait session --load",
			"ait session --help",
		)
	}
	return "", nil
}
