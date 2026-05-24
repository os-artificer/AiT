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

// Try lines for session load/pick hint blocks (shared by load and interactive pick errors).
const (
	tryLoadByAlias     = "ait session --load <alias>"
	tryLoadByFullID    = "ait session --list -v    then ait session --load <FULL_ID>"
	tryLoadInteractive = "ait session --load       pick by # interactively"
	trySessionHelp     = "ait session --help"
	trySuspendInSSH    = "!ait suspend"
	tryLoadFromLocal   = "ait session --load [<id|alias>]"
	trySwitchForce     = "ait session --switch <id|alias> --force"
	tryLoginSave       = "ait login --save <destination>"
	tryLoginAlias      = "ait login --alias <name> <destination>"
)
