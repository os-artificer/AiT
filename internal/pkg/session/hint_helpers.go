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

import (
	"fmt"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/store"
)

const (
	errTmuxSetupFailed   = "tmux setup failed"
	errTmuxRequiresTTY   = "tmux requires an interactive terminal"
	hintTmuxSetupExplain = "Remote tmux is missing or not on PATH for this connection."
	hintTmuxTTYExplain   = "Tmux attach needs a TTY on stdin."
)

func invalidSelectionErr() error {
	return errhint.WithTryMsg(errInvalidSelection,
		hintInvalidSelectionExplain,
		tryLoadInteractive, trySessionHelp,
	)
}

func pickCancelledErr() error {
	return errhint.WithTry(ErrPickCancelled,
		hintPickCancelledExplain,
		tryLoadByAlias,
		tryLoadInteractive,
		trySessionHelp,
	)
}

// TmuxRequiresTTYErr reports that --tmux cannot run without a TTY on stdin.
func TmuxRequiresTTYErr() error {
	return errhint.WithTryMsg(errTmuxRequiresTTY,
		hintTmuxTTYExplain,
		TryConnectTmuxUserHost,
		TryConnectTmuxDevhost,
		tryConnectHelp,
	)
}

func tmuxEnsureErr(sess *store.Session, cause error) error {
	hostLabel := sessionHostLabel(sess)
	return errhint.WithTry(fmt.Errorf("%s: %w", errTmuxSetupFailed, cause),
		hintTmuxSetupExplain,
		fmt.Sprintf("ssh %s 'which tmux'", hostLabel),
		"sudo apt install tmux    # Debian/Ubuntu",
		"brew install tmux        # macOS",
		tryConnectHelp,
	)
}

func wrapLoadDialErr(sess *store.Session, cause error) error {
	tries := []string{
		tryLoadByAlias,
		tryLoadByFullID,
		trySessionHelp,
	}
	if sess.ProxyJump != "" || sess.LoginDestination != "" {
		dest := sessionDialDestination(sess)
		tries = append([]string{fmt.Sprintf("ssh -G %s | grep -i proxy", dest)}, tries...)
	}
	msg := cause.Error()
	if strings.Contains(strings.ToLower(msg), "connection refused") ||
		strings.Contains(strings.ToLower(msg), "no route") {
		return errhint.WithTry(cause,
			"Could not reach the saved host; check ProxyJump and network path.",
			tries...,
		)
	}
	return cause
}
