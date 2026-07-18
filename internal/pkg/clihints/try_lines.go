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

// Package clihints holds shared errhint Try-block example command lines for the aitssh CLI.
package clihints

import "github.com/os-artificer/AiT/internal/pkg/toolname"

// Session and login Try lines (single source for internal/aissh and internal/pkg/ssh/client).
var (
	TrySessionList            = toolname.CLI("session --list")
	TrySessionLoadAlias       = toolname.CLI("session --load <alias>")
	TrySessionLoad            = toolname.CLI("session --load")
	TrySessionLoadOptional    = toolname.CLI("session --load [<id|alias>]")
	TrySessionLoadFullID      = toolname.CLI("session --list -v") + "    then " + toolname.CLI("session --load <FULL_ID>")
	TrySessionLoadPick        = toolname.CLI("session --load") + "       pick by # interactively"
	TrySessionSwitchID        = toolname.CLI("session --switch <id|alias>")
	TrySessionSwitch          = toolname.CLI("session --switch")
	TrySessionSwitchForce     = toolname.CLI("session --switch <id|alias> --force")
	TrySessionHelp            = toolname.CLI("session --help")
	TryConnectSave            = toolname.CLI("--save <destination>")
	TryConnectAlias           = toolname.CLI("--alias <name> <destination>")
	TryConnectHelp            = toolname.CLI("--help")
	TryConnectTmuxUserHost    = toolname.CLI("--tmux user@host")
	TryConnectTmuxDevhost     = toolname.CLI("--tmux devhost")
	TryConnectTmuxSaveDevhost = toolname.CLI("--tmux --save devhost")

	// OpenSSH parse / transfer Try lines (concrete examples for get/put/connect argv hints).
	TryConnectExample      = toolname.CLI("user@host")
	TryConnectSaveUserHost = toolname.CLI("--save user@host")
	TryConnectRemoteCmd    = toolname.CLI("user@host uptime")
	TryGetHelp             = toolname.CLI("get --help")
	TryPutHelp             = toolname.CLI("put --help")
	TryGetExample          = toolname.CLI("get user@host:/remote/path ./local")
	TryPutExample          = toolname.CLI("put ./local user@host:/remote/path")
	TryGetSessionProd      = toolname.CLI("get --session prod <paths...>")

	// Shell REPL builtin shorthand (maps to session --list).
	TryShellList = toolname.CLI("list")

	// AI experimental gate (EX*): enable for current process / persist / env.
	TryAIExperimentalCLI    = toolname.CLI("--ai-experimental ai")
	TryAIExperimentalConfig = toolname.CLI("config set ai.experimental.enabled true")
	TryAIExperimentalEnv    = "AIT_AI_EXPERIMENTAL=1 " + toolname.CLI("ai")
)
