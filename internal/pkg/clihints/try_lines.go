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

// Package clihints holds shared errhint Try-block example command lines.
package clihints

// Session and login Try lines (single source for internal/ait and internal/pkg/ssh/client).
const (
	TrySessionList          = "ait session --list"
	TrySessionLoadAlias     = "ait session --load <alias>"
	TrySessionLoad          = "ait session --load"
	TrySessionLoadOptional  = "ait session --load [<id|alias>]"
	TrySessionLoadFullID    = "ait session --list -v    then ait session --load <FULL_ID>"
	TrySessionLoadPick      = "ait session --load       pick by # interactively"
	TrySessionSwitchID      = "ait session --switch <id|alias>"
	TrySessionSwitch        = "ait session --switch"
	TrySessionSwitchForce   = "ait session --switch <id|alias> --force"
	TrySessionHelp          = "ait session --help"
	TryLoginSave            = "ait login --save <destination>"
	TryLoginAlias           = "ait login --alias <name> <destination>"
	TryLoginHelp            = "ait login --help"
	TryLoginTmuxUserHost    = "ait login --tmux user@host"
	TryLoginTmuxDevhost     = "ait login --tmux devhost"
	TryLoginTmuxSaveDevhost = "ait login --tmux --save devhost"

	// OpenSSH parse / transfer Try lines (concrete examples for get/put/login argv hints).
	TryLoginExample      = "ait login user@host"
	TryLoginSaveUserHost = "ait login --save user@host"
	TryLoginRemoteCmd    = "ait login user@host uptime"
	TryGetHelp           = "ait get --help"
	TryPutHelp           = "ait put --help"
	TryGetExample        = "ait get user@host:/remote/path ./local"
	TryPutExample        = "ait put ./local user@host:/remote/path"
	TryGetSessionProd    = "ait get --session prod <paths...>"

	// Shell REPL builtin shorthand (maps to session --list).
	TryShellList = "ait list"
)
