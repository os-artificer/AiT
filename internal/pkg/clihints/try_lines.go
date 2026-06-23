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

// Package clihints holds shared errhint Try-block example command lines for the ai-ssh CLI.
package clihints

// Session and login Try lines (single source for internal/aissh and internal/pkg/ssh/client).
const (
	TrySessionList            = "ai-ssh session --list"
	TrySessionLoadAlias       = "ai-ssh session --load <alias>"
	TrySessionLoad            = "ai-ssh session --load"
	TrySessionLoadOptional    = "ai-ssh session --load [<id|alias>]"
	TrySessionLoadFullID      = "ai-ssh session --list -v    then ai-ssh session --load <FULL_ID>"
	TrySessionLoadPick        = "ai-ssh session --load       pick by # interactively"
	TrySessionSwitchID        = "ai-ssh session --switch <id|alias>"
	TrySessionSwitch          = "ai-ssh session --switch"
	TrySessionSwitchForce     = "ai-ssh session --switch <id|alias> --force"
	TrySessionHelp            = "ai-ssh session --help"
	TryConnectSave            = "ai-ssh --save <destination>"
	TryConnectAlias           = "ai-ssh --alias <name> <destination>"
	TryConnectHelp            = "ai-ssh --help"
	TryConnectTmuxUserHost    = "ai-ssh --tmux user@host"
	TryConnectTmuxDevhost     = "ai-ssh --tmux devhost"
	TryConnectTmuxSaveDevhost = "ai-ssh --tmux --save devhost"

	// OpenSSH parse / transfer Try lines (concrete examples for get/put/connect argv hints).
	TryConnectExample      = "ai-ssh user@host"
	TryConnectSaveUserHost = "ai-ssh --save user@host"
	TryConnectRemoteCmd    = "ai-ssh user@host uptime"
	TryGetHelp             = "ai-ssh get --help"
	TryPutHelp             = "ai-ssh put --help"
	TryGetExample          = "ai-ssh get user@host:/remote/path ./local"
	TryPutExample          = "ai-ssh put ./local user@host:/remote/path"
	TryGetSessionProd      = "ai-ssh get --session prod <paths...>"

	// Shell REPL builtin shorthand (maps to session --list).
	TryShellList = "ai-ssh list"
)
