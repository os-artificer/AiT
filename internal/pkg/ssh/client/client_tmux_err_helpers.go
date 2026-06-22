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

package client

import (
	"errors"
	"fmt"
	"io"

	"github.com/os-artificer/AiT/internal/pkg/clihints"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/tmux"

	"golang.org/x/crypto/ssh"
)

const (
	errTmuxAttachFailed    = "tmux attach failed"
	errTmuxPaneSetupFailed = "tmux pane setup failed"
	errTmuxSessionNotFound = "tmux session not found"

	hintTmuxAttachExplain          = "Could not attach to the remote tmux session; it may be missing or the server rejected attach."
	hintTmuxPaneSetupExplain       = "Could not run AiT setup in the detached tmux pane before attach."
	hintTmuxSessionNotFoundExplain = "Remote tmux session is missing after setup; tmux may not be installed or the session was removed."
)

func tmuxAttachTryLines(tmuxName string) []string {
	return []string{
		clihints.TryConnectTmuxUserHost,
		clihints.TryConnectTmuxDevhost,
		fmt.Sprintf("ssh user@host 'tmux list-sessions | grep %s'", tmuxName),
		clihints.TrySessionList,
	}
}

// wrapTmuxPreAttachErr wraps detached-pane setup failures with a multi-line hint.
func wrapTmuxPreAttachErr(tmuxName string, cause error) error {
	if cause == nil {
		return nil
	}
	return errhint.WithTry(fmt.Errorf("%s: %w", errTmuxPaneSetupFailed, cause),
		hintTmuxPaneSetupExplain,
		append(tmuxAttachTryLines(tmuxName), "ssh user@host 'which tmux'")...,
	)
}

// wrapTmuxSessionNotFoundErr reports that the expected remote tmux session is absent.
func wrapTmuxSessionNotFoundErr(tmuxName string, cause error) error {
	var wrapped error
	if cause != nil {
		wrapped = fmt.Errorf("%s %q: %w", errTmuxSessionNotFound, tmuxName, cause)
	} else {
		wrapped = fmt.Errorf("%s %q", errTmuxSessionNotFound, tmuxName)
	}
	return errhint.WithTry(wrapped,
		hintTmuxSessionNotFoundExplain,
		append(tmuxAttachTryLines(tmuxName), clihints.TryConnectTmuxSaveDevhost)...,
	)
}

// wrapTmuxAttachExitErr wraps SSH session wait failures from tmux attach with context and Try lines.
func wrapTmuxAttachExitErr(tmuxName string, code int, cause error) error {
	return errhint.WithTry(fmt.Errorf("%s (exit %d): %w", errTmuxAttachFailed, code, cause),
		hintTmuxAttachExplain,
		tmuxAttachTryLines(tmuxName)...,
	)
}

// wrapInteractiveWaitErr adds tmux-specific hints when the interactive SSH session exits non-zero.
func wrapInteractiveWaitErr(tmuxEnabled bool, tmuxName string, err error) error {
	if err == nil || errors.Is(err, io.EOF) {
		return err
	}
	var exitErr *ssh.ExitError
	if tmuxEnabled && errors.As(err, &exitErr) {
		return wrapTmuxAttachExitErr(tmuxName, exitErr.ExitStatus(), err)
	}
	return err
}

// ensureTmuxSessionReady verifies the remote tmux session exists (direct SSH only).
func (c *Client) ensureTmuxSessionReady(stdin io.Writer, name string) error {
	if c.ChainMode() || c.conn == nil {
		return nil
	}

	ok, err := tmux.HasSession(c.conn, name)
	if err != nil {
		return wrapTmuxSessionNotFoundErr(name, err)
	}
	if ok {
		return nil
	}

	if err := c.tmuxEnsure(stdin, name); err != nil {
		return wrapTmuxSessionNotFoundErr(name, err)
	}

	ok, err = tmux.HasSession(c.conn, name)
	if err != nil {
		return wrapTmuxSessionNotFoundErr(name, err)
	}
	if !ok {
		return wrapTmuxSessionNotFoundErr(name, nil)
	}
	return nil
}
