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

package shell

import (
	"github.com/os-artificer/AiT/internal/ssh/shell/builtin"
)

// Commands — AiT shell REPL built-in verbs (re-exported for ai/mode and meta).
const (
	ReplCmdExit    = builtin.CmdExit
	ReplCmdQuit    = builtin.CmdQuit
	ReplCmdClear   = builtin.CmdClear
	ReplCmdHistory = builtin.CmdHistory
	ReplCmdLoad    = builtin.CmdLoad
	ReplCmdList    = builtin.CmdList
	ReplCmdLL      = builtin.CmdLL
	ReplCmdSuspend = builtin.CmdSuspend
)

// IO — error prefix and runner messages.
const (
	errShellCLIExecutorNotConfig    = "AiT: shell CLI executor not configured"
	errShellExternalRunnerNotConfig = "AiT: shell external runner not configured"
)
