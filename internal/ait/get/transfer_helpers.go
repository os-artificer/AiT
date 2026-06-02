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

package get

import (
	"context"

	"github.com/os-artificer/AiT/internal/pkg/session"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/ssh/resolve"
	pkgtransfer "github.com/os-artificer/AiT/internal/pkg/transfer"
)

func runTransferArgv(ctx context.Context, mgr *session.Manager, argv []string) error {
	spec, err := ParseTransfer(argv)
	if err != nil {
		return err
	}
	client, err := dialForTransfer(ctx, mgr, spec)
	if err != nil {
		return err
	}
	defer client.Close()

	return pkgtransfer.Get(ctx, client.Conn(), spec.RemotePath, spec.LocalPath)
}

func dialForTransfer(ctx context.Context, mgr *session.Manager, spec TransferSpec) (*sshclient.Client, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if spec.Session != "" {
		sshOpts, err := mgr.SSHOptionsForSession(spec.Session, spec.Conn.Password)
		if err != nil {
			return nil, err
		}
		return mgr.DialSSH(sshOpts, spec.Conn.Password, sshOpts.Host)
	}

	resolved, err := resolve.Resolve(spec.Conn)
	if err != nil {
		return nil, err
	}
	hostLabel := resolved.User + "@" + resolved.Host
	if resolved.User == "" {
		hostLabel = resolved.Host
	}
	return mgr.DialSSH(resolved.ClientOptions(), spec.Conn.Password, hostLabel)
}
