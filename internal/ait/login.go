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

import (
	"context"
	"errors"

	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/sshargs"
	"github.com/os-artificer/AiT/internal/pkg/sshclient"
)

func runLogin(ctx context.Context, argv []string) error {
	spec, err := sshargs.Parse(argv)
	if err != nil {
		return err
	}
	resolved, err := sshargs.Resolve(spec)
	if err != nil {
		return err
	}
	mgr, err := newManager()
	if err != nil {
		return err
	}
	defer mgr.Store.Close()
	opts := loginOptsFromResolved(resolved)
	err = mgr.Login(ctx, opts)
	if re, ok := err.(*sshclient.RemoteExitError); ok {
		return exitError(re.Code)
	}
	return err
}

func loginOptsFromResolved(r sshargs.Resolved) session.LoginOpts {
	return session.LoginOpts{
		Alias:                 r.Alias,
		Host:                  r.Host,
		Port:                  r.Port,
		User:                  r.User,
		IdentityPath:          r.IdentityPath,
		IdentityExplicit:      r.IdentityExplicit,
		Password:              r.Password,
		Tmux:                  r.Tmux,
		RemoteCommand:         append([]string(nil), r.RemoteCommand...),
		ProxyJump:             append([]string(nil), r.ProxyJump...),
		ProxyCommand:          r.ProxyCommand,
		KnownHostsPath:        r.KnownHostsPath,
		StrictHostKeyChecking: r.StrictHostKeyChecking,
		ConnectTimeout:        r.ConnectTimeout,
		BatchMode:             r.BatchMode,
	}
}

// exitError wraps an exit code for main.
type exitError int

func (e exitError) Error() string {
	return ""
}

// ExitCode returns the process exit code if err is from a remote command.
func ExitCode(err error) (int, bool) {
	var ee exitError
	if errors.As(err, &ee) {
		return int(ee), true
	}
	var re *sshclient.RemoteExitError
	if errors.As(err, &re) {
		return re.Code, true
	}
	return 0, false
}
