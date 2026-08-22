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

package ait

import "github.com/os-artificer/AiT/internal/pkg/aitplugin"

const (
	replPrompt      = "ait> "
	historyBasename = "ait_repl_history"

	cmdExit    = aitplugin.CmdExit
	cmdQuit    = aitplugin.CmdQuit
	cmdHelp    = aitplugin.CmdHelp
	cmdVersion = aitplugin.CmdVersion
	cmdHistory = aitplugin.CmdHistory
	cmdClear   = aitplugin.CmdClear
	cmdInstall = aitplugin.CmdInstall
	cmdUpdate  = aitplugin.CmdUpdate
	cmdUpgrade = aitplugin.CmdUpgrade

	argVersionShort     = "-V"
	argVersionLong      = "--version"
	interactiveOnlyMsg  = "ait is interactive only"
	interactiveOnlyExpl = "Open the REPL with no arguments. One-shot plugin argv is not supported."
	unknownCmdMsg       = "unknown command"
	noPluginsMsg        = "no plugins found"
	helpPluginNotFound  = "unknown help topic"
	unclosedQuoteExpl   = "Close the quote or escape the line before running the command."
)
