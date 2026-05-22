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
	"io"

	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/ssh/args"

	"github.com/spf13/cobra"
)

type transferCmdSpec struct {
	name  string
	short string
	op    args.TransferOp
}

func newTransferCmd(spec transferCmdSpec) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   spec.name + " [flags] <paths...>",
		Short:                 spec.short,
		DisableFlagsInUseLine: true,
		DisableFlagParsing:    true,
		SilenceErrors:         true,
	}
	write := func(w io.Writer) error {
		return helpdoc.WriteTransferHelp(w, spec.name)
	}
	bindCustomArgvHelp(cmd, write, func(ctx context.Context, mgr *session.Manager, argv []string) error {
		return runTransferArgv(ctx, mgr, argv, spec.op)
	})
	return cmd
}

func newGetCmd() *cobra.Command {
	return newTransferCmd(transferCmdSpec{
		name:  "get",
		short: "Download a file from a remote host (OpenSSH-compatible)",
		op:    args.OpGet,
	})
}

func newPutCmd() *cobra.Command {
	return newTransferCmd(transferCmdSpec{
		name:  "put",
		short: "Upload a file to a remote host (OpenSSH-compatible)",
		op:    args.OpPut,
	})
}
