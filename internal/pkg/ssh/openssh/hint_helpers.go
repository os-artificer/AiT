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

package openssh

import (
	"fmt"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
)

func destinationRequiredErr() error {
	return errhint.WithTryMsg("destination required (e.g. user@host or ssh config Host alias)",
		"Add a host destination after login options.",
		tryLoginExample, tryLoginSave, tryLoginHelp,
	)
}

// TmuxWithRemoteCommandErr reports invalid --tmux with a one-shot remote command.
func TmuxWithRemoteCommandErr() error {
	return errhint.WithTryMsg("cannot use --tmux with a remote command",
		"Tmux keeps an interactive shell; one-shot commands cannot use --tmux.",
		tryLoginExample,
		"ait login user@host uptime",
		tryLoginHelp,
	)
}

func transferPathCountErr(cmdLabel string) error {
	tries := []string{tryGetExample, tryGetHelp}
	msg := "need 1 to 3 path arguments for ait get (remote path required; local optional)"
	if cmdLabel == "put" {
		tries = []string{tryPutExample, tryPutHelp}
		msg = "need 1 to 3 path arguments for ait put (local path required; remote optional)"
	}
	return errhint.WithTryMsg(msg,
		fmt.Sprintf("Specify paths for ait %s.", cmdLabel),
		tries...,
	)
}

func sessionFlagRequiresValueErr() error {
	return errhint.WithTryMsg("--session requires a value",
		"Pass a saved session id or alias.",
		"ait session --list",
		"ait get --session prod <paths...>",
		tryGetHelp,
	)
}

func unknownFlagErr(flag string, cmdLabel string) error {
	return errhint.WithTryMsg(fmt.Sprintf("unknown flag %q", flag),
		fmt.Sprintf("See supported flags for ait %s.", cmdLabel),
		fmt.Sprintf("ait %s --help", cmdLabel),
	)
}
