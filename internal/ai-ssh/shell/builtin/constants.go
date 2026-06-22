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

package builtin

import (
	"github.com/os-artificer/AiT/internal/ai-ssh/login"
	"github.com/os-artificer/AiT/internal/ai-ssh/session"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	pkgsession "github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/ssh/client/meta"
)

// AiT shell REPL built-in command names.
const (
	CmdExit    = "exit"
	CmdQuit    = "quit"
	CmdClear   = "clear"
	CmdHistory = "history"
	CmdAI      = "ai" // matches ai.Name (shell cannot import ai: cycle via mode)
	CmdLogin   = login.Name
	CmdLoad    = "load"
	CmdList    = session.ReplAliasList
	CmdLL      = "ll"
	CmdSuspend = "suspend"
)

const (
	errAIModeRunnerNotConfig = "AiT: AI mode runner not configured"
	errLoginBlockedActive    = "login is unavailable while a session is active"
	errLoadBlockedActive     = "load is unavailable while a session is active"
	hintLoginBlockedExplain  = "Suspend the current SSH session before starting another login from the local shell."
	hintLoadBlockedExplain   = "Suspend the current SSH session before loading another from the local shell."
)

func loginBlockedWhenActiveErr() error {
	return errhint.WithTryMsg(errLoginBlockedActive,
		hintLoginBlockedExplain,
		meta.LineSuspend,
		pkgsession.TrySessionSwitchID,
		pkgsession.TrySessionHelp,
	)
}

func loadBlockedWhenActiveErr() error {
	return errhint.WithTryMsg(errLoadBlockedActive,
		hintLoadBlockedExplain,
		meta.LineSuspend,
		pkgsession.TrySessionSwitchID,
		pkgsession.TrySessionHelp,
	)
}
