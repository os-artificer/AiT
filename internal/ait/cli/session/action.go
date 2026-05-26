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
	"github.com/os-artificer/AiT/internal/pkg/errhint"

	"github.com/spf13/cobra"
)

const (
	errSessionActionRequired  = "specify one session action: --list, --load, --switch, --suspend, or --delete"
	errSessionActionExclusive = "only one session action flag may be set"
)

type sessionActionKind int

const (
	sessionActionNone sessionActionKind = iota
	sessionActionList
	sessionActionLoad
	sessionActionSwitch
	sessionActionSuspend
	sessionActionDelete
)

type sessionAction struct {
	kind   sessionActionKind
	target string
}

func detectSessionAction(
	cmd *cobra.Command,
	doList, doSuspend bool,
	loadTarget, switchTarget, deleteTarget string,
) (sessionAction, error) {
	var act sessionAction
	n := 0

	if doList {
		n++
		act.kind = sessionActionList
	}
	if cmd.Flags().Changed(flagNameLoad) {
		n++
		act.kind = sessionActionLoad
		act.target = loadTarget
		if act.target == "-" {
			act.target = ""
		}
	}
	if cmd.Flags().Changed(flagNameSwitch) {
		n++
		act.kind = sessionActionSwitch
		act.target = switchTarget
		if act.target == "-" {
			act.target = ""
		}
	}
	if doSuspend {
		n++
		act.kind = sessionActionSuspend
	}
	if deleteTarget != "" {
		n++
		act.kind = sessionActionDelete
		act.target = deleteTarget
	}

	if n == 0 {
		return act, sessionActionRequiredErr()
	}
	if n > 1 {
		return act, sessionActionExclusiveErr()
	}

	return act, nil
}

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

func resolveSwitchTarget(cmd *cobra.Command, switchTarget string, args []string) (string, error) {
	if switchTarget != "" {
		return switchTarget, nil
	}
	if len(args) == 1 {
		return args[0], nil
	}
	if len(args) > 1 {
		return "", unexpectedSwitchArgsErr()
	}
	return "", nil
}

func sessionActionRequiredErr() error {
	return errhint.WithTryMsg(errSessionActionRequired,
		"Specify exactly one session action.",
		"ait session --list",
		"ait session --load [<id|alias>]",
		"ait session --switch [id|alias]",
		"ait session --help",
	)
}

func sessionActionExclusiveErr() error {
	return errhint.WithTryMsg(errSessionActionExclusive,
		"Use only one of --list, --load, --switch, --suspend, or --delete per invocation.",
		"ait session --help",
	)
}

func unexpectedSwitchArgsErr() error {
	return errhint.WithTryMsg("unexpected arguments with --switch",
		"Pass the session id or alias as the --switch value or as a single positional argument.",
		"ait session --switch <id|alias>",
		"ait session --switch",
		"ait session --help",
	)
}
