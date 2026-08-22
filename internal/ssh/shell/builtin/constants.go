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
	"errors"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/replcmd"
	pkgsession "github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/ssh/client/meta"
	"github.com/os-artificer/AiT/internal/ssh/session"
)

// AiT shell REPL built-in command names.
const (
	CmdExit    = replcmd.Exit
	CmdQuit    = replcmd.Quit
	CmdClear   = replcmd.Clear
	CmdHistory = replcmd.History
	CmdAI      = "ai" // matches ai.Name (shell cannot import aissh/ai: import cycle)
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

// ErrLoginBlockedActive is returned when SSH connect is attempted while a session is active (R4).
var ErrLoginBlockedActive = errors.New(errLoginBlockedActive)

func loginBlockedWhenActiveErr() error {
	return LoginBlockedWhenActiveErr()
}

// LoginBlockedWhenActiveErr reports that SSH connect is blocked while a session is active (R4).
func LoginBlockedWhenActiveErr() error {
	return errhint.WithTry(ErrLoginBlockedActive,
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
