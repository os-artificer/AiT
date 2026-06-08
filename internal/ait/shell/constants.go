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
	"github.com/os-artificer/AiT/internal/ait/login"
	"github.com/os-artificer/AiT/internal/ait/session"
)

// Commands — AiT shell REPL built-in verbs (shared core verbs exported for ai/mode).
const (
	ReplCmdExit    = "exit"
	ReplCmdQuit    = "quit"
	ReplCmdClear   = "clear"
	ReplCmdHistory = "history"
	ReplCmdAI      = "ai" // matches ai.Name (shell cannot import ai: cycle via mode)
	ReplCmdLogin   = login.Name
	ReplCmdLoad    = "load"
	ReplCmdList    = session.ReplAliasList
	ReplCmdSuspend = "suspend"
)

// IO — error prefix and runner messages.
const (
	errAIModeRunnerNotConfig     = "AiT: AI mode runner not configured"
	errShellCLIExecutorNotConfig = "AiT: shell CLI executor not configured"
	errLoginBlockedActive        = "AiT: login is unavailable while a session is active. Use suspend first or session --switch."
	errLoadBlockedActive         = "AiT: load is unavailable while a session is active. Use suspend first or session --switch."
)
