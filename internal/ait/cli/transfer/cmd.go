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

// Package transfer builds the `ait get` and `ait put` Cobra commands.
package transfer

import (
	"context"
	"io"

	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/session"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
	sshparse "github.com/os-artificer/AiT/internal/pkg/ssh/parse"
	"github.com/os-artificer/AiT/internal/pkg/ssh/resolve"
	pkgtransfer "github.com/os-artificer/AiT/internal/pkg/transfer"

	"github.com/spf13/cobra"
)

// RunWithMgr opens a session.Manager and runs fn under a signal-cancellable context.
type RunWithMgr func(fn func(ctx context.Context, mgr *session.Manager) error) error

// NewGet builds the `ait get` cobra command.
func NewGet(runWithMgr RunWithMgr) *cobra.Command {
	return newCmd(runWithMgr, "get", "Download a file from a remote host (OpenSSH-compatible)", sshparse.OpGet)
}

// NewPut builds the `ait put` cobra command.
func NewPut(runWithMgr RunWithMgr) *cobra.Command {
	return newCmd(runWithMgr, "put", "Upload a file to a remote host (OpenSSH-compatible)", sshparse.OpPut)
}

func newCmd(runWithMgr RunWithMgr, name, short string, op sshparse.TransferOp) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   name + " [flags] <paths...>",
		Short:                 short,
		DisableFlagsInUseLine: true,
		DisableFlagParsing:    true,
		SilenceUsage:          true,
		SilenceErrors:         true,
	}
	write := func(w io.Writer) error {
		return helpdoc.WriteTransferHelp(w, name)
	}
	helpdoc.BindCustomArgv(cmd, write, func(_ *cobra.Command, argv []string) error {
		return runWithMgr(func(ctx context.Context, mgr *session.Manager) error {
			return runTransferArgv(ctx, mgr, argv, op)
		})
	})
	return cmd
}

// runTransferArgv parses argv like ait login and runs get or put per op.
func runTransferArgv(ctx context.Context, mgr *session.Manager, argv []string, op sshparse.TransferOp) error {
	spec, err := sshparse.ParseTransfer(argv, op)
	if err != nil {
		return err
	}
	client, err := dialForTransfer(ctx, mgr, spec)
	if err != nil {
		return err
	}
	defer client.Close()

	if op == sshparse.OpPut {
		return pkgtransfer.Put(ctx, client.Conn(), spec.LocalPath, spec.RemotePath)
	}
	return pkgtransfer.Get(ctx, client.Conn(), spec.RemotePath, spec.LocalPath)
}

func dialForTransfer(ctx context.Context, mgr *session.Manager, spec sshparse.TransferSpec) (*sshclient.Client, error) {
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
