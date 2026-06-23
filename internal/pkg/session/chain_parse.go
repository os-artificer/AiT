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

package session

import (
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func chainPasswordRequiresValueErr() error {
	return errhint.WithTryMsg("--password requires a value",
		"Pass the SSH password after --password.",
		toolname.MetaLine("session --switch prod --password secret"),
	)
}

func chainUnknownFlagErr(flag string) error {
	return errhint.WithTryMsg("unknown flag "+flag,
		"Supported flags are --tmux and --password.",
		toolname.MetaLine("load prod --tmux"),
		toolname.MetaLine("session --switch prod --tmux"),
	)
}

func chainUnexpectedSwitchArgsErr() error {
	return errhint.WithTryMsg("unexpected arguments with --switch",
		"Pass the session id or alias as a single argument.",
		toolname.MetaLine("session --switch <id|alias>"),
		toolname.MetaLine("session --switch"),
	)
}

// ParseChainLoadArgs parses argv after "!ait load " (without the load keyword).
// needPick is true when no target id/alias was given and the caller should prompt.
func ParseChainLoadArgs(argv []string) (target, password string, tmux, needPick bool, err error) {
	if len(argv) == 0 {
		return "", "", false, true, nil
	}

	i := 0
	if !strings.HasPrefix(argv[0], "-") {
		target = argv[0]
		i = 1
	}

	for i < len(argv) {
		switch argv[i] {
		case "--tmux":
			tmux = true
			i++
		case "--password":
			i++
			if i >= len(argv) {
				return "", "", false, false, chainPasswordRequiresValueErr()
			}
			password = argv[i]
			i++
		default:
			if strings.HasPrefix(argv[i], "-") {
				return "", "", false, false, chainUnknownFlagErr(argv[i])
			}
			if target != "" {
				return "", "", false, false, errhint.WithTryMsg("unexpected arguments with load",
					"Pass the session id or alias as the first argument.",
					toolname.MetaLine("load <id|alias>"),
					toolname.MetaLine("load"),
				)
			}
			target = argv[i]
			i++
		}
	}

	if target == "" {
		needPick = true
	}
	return target, password, tmux, needPick, nil
}

// ParseChainSwitchArgs parses argv after "!ait session --switch ".
func ParseChainSwitchArgs(argv []string) (target, password string, tmux, needPick bool, err error) {
	if len(argv) == 0 {
		return "", "", false, true, nil
	}

	i := 0
	for i < len(argv) {
		switch argv[i] {
		case "--tmux":
			tmux = true
			i++
		case "--password":
			i++
			if i >= len(argv) {
				return "", "", false, false, chainPasswordRequiresValueErr()
			}
			password = argv[i]
			i++
		default:
			if strings.HasPrefix(argv[i], "-") {
				return "", "", false, false, chainUnknownFlagErr(argv[i])
			}
			if target != "" {
				return "", "", false, false, chainUnexpectedSwitchArgsErr()
			}
			target = argv[i]
			i++
		}
	}

	if target == "" {
		needPick = true
	}
	return target, password, tmux, needPick, nil
}
