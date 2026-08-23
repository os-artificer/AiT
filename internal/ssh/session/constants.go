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

const (
	// Name is the top-level `aissh session` subcommand name.
	Name = "session"
	// FlagList is the long --list flag passed to the session executor.
	FlagList = "--list"
	// ReplAliasList is the shell/meta REPL alias for session --list.
	ReplAliasList = "list"
)

const (
	flagNameList     = ReplAliasList
	flagNameLoad     = "load"
	flagNameSuspend  = "suspend"
	flagNameSwitch   = "switch"
	flagNameDelete   = "delete"
	flagNameTmux     = "tmux"
	flagNamePassword = "password"
	flagNameForce    = "force"
	flagNameVerbose  = "verbose"
)
