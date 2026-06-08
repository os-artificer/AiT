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
	"github.com/os-artificer/AiT/internal/pkg/errhint"
)

func loadUsageErr() error {
	return errhint.WithTryMsg("usage: load [<id|alias>] [--tmux] [--password P]",
		"Load a saved session or pick interactively with no arguments.",
		"load",
		"load prod",
		"list",
	)
}

func passwordRequiresValueErr() error {
	return errhint.WithTryMsg("--password requires a value",
		"Pass the SSH password after --password.",
		"load prod --password secret",
	)
}

func unknownLoadFlagErr(flag string) error {
	return errhint.WithTryMsg("unknown load flag "+flag,
		"Supported load flags are --tmux and --password.",
		"load prod --tmux",
		"help",
	)
}

// parseLoadArgs parses load arguments: load [<id|alias>] [--tmux] [--password P].
// An empty argv is valid; the caller runs interactive pick (load with no arguments).
func parseLoadArgs(argv []string) (idOrAlias, password string, tmux bool, err error) {
	if len(argv) < 1 {
		return "", "", false, loadUsageErr()
	}
	idOrAlias = argv[0]

	for i := 1; i < len(argv); i++ {
		switch argv[i] {
		case "--tmux":
			tmux = true
		case "--password":
			i++
			if i >= len(argv) {
				return "", "", false, passwordRequiresValueErr()
			}
			password = argv[i]
		default:
			return "", "", false, unknownLoadFlagErr(argv[i])
		}
	}
	return idOrAlias, password, tmux, nil
}
