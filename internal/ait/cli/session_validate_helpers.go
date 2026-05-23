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

import (
	"fmt"

	"github.com/spf13/cobra"
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
	if loadTarget != "" {
		n++
		act.kind = sessionActionLoad
		act.target = loadTarget
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
		return act, fmt.Errorf("%s", errSessionActionRequired)
	}
	if n > 1 {
		return act, fmt.Errorf("%s", errSessionActionExclusive)
	}

	return act, nil
}

func resolveSwitchTarget(cmd *cobra.Command, switchTarget string, args []string) (string, error) {
	if switchTarget != "" {
		return switchTarget, nil
	}
	if len(args) == 1 {
		return args[0], nil
	}
	if len(args) > 1 {
		return "", fmt.Errorf("unexpected arguments with --switch")
	}
	return "", nil
}
