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

package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/ssh/args"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/transfer"
)

// runGetArgv parses argv like ait login and downloads one remote file.
func runGetArgv(ctx context.Context, mgr *session.Manager, argv []string) error {
	spec, err := args.ParseTransfer(argv, args.OpGet)
	if err != nil {
		return err
	}
	client, err := dialForTransfer(ctx, mgr, spec)
	if err != nil {
		return err
	}
	defer client.Close()

	return transfer.Get(ctx, client.Conn(), spec.RemotePath, spec.LocalPath)
}

// runPutArgv parses argv like ait login and uploads one local file.
func runPutArgv(ctx context.Context, mgr *session.Manager, argv []string) error {
	spec, err := args.ParseTransfer(argv, args.OpPut)
	if err != nil {
		return err
	}
	client, err := dialForTransfer(ctx, mgr, spec)
	if err != nil {
		return err
	}
	defer client.Close()

	return transfer.Put(ctx, client.Conn(), spec.LocalPath, spec.RemotePath)
}

func dialForTransfer(ctx context.Context, mgr *session.Manager, spec args.TransferSpec) (*sshclient.Client, error) {
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

	resolved, err := args.Resolve(spec.Conn)
	if err != nil {
		return nil, err
	}
	hostLabel := resolved.User + "@" + resolved.Host
	if resolved.User == "" {
		hostLabel = resolved.Host
	}
	return mgr.DialSSH(resolved.ClientOptions(), spec.Conn.Password, hostLabel)
}

func transferHelpRequested(argv []string) bool {
	return len(argv) == 1 && (argv[0] == "-h" || argv[0] == "--help")
}

func writeTransferHelp(w io.Writer, cmd string) error {
	const tpl = `%s — download or upload a single file (OpenSSH-compatible flags).

Usage:
  ait %s [flags] <paths...>

Paths (order-independent):
  user@host:/remote/file  and  local/path     (two arguments)
  destination  remote/path  local/path       (three arguments)
  with --session: remote/path and local/path on the saved session host

Flags match ait login: -p, -pPORT, -l, -i, -iPATH, -F, -J, -o, --password, --host, --user, --port, --identity
  --session NAME   use a saved session (remote path may omit user@host)

Run 'ait login --help' for full OpenSSH flag details.

Examples:
  ait get -p22 -i~/.ssh/id ./app.log user@deploy:/var/log/app.log
  ait get deploy@10.0.0.5 /var/log/app.log ./app.log
  ait get ./app.log /var/log/app.log --session prod
  ait put -l root ./bin/app user@host:/usr/local/bin/app

`
	_, err := fmt.Fprintf(w, tpl, cmd, cmd)
	return err
}
