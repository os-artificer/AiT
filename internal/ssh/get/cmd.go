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
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/ssh/registry"

	"github.com/spf13/cobra"
)

const (
	// Name is the top-level subcommand name.
	Name = "get"
)

// New builds the `aissh get` cobra command.
func New(runWithMgr registry.RunWithMgr) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   Name + " [flags] <paths...>",
		Short:                 "Download a file from a remote host (OpenSSH-compatible)",
		DisableFlagsInUseLine: true,
		DisableFlagParsing:    true,
		SilenceUsage:          true,
		SilenceErrors:         true,
	}
	write := func(w io.Writer) error {
		return pkghelpdoc.WriteTransferHelp(w, Name)
	}
	pkghelpdoc.BindCustomArgv(cmd, write, func(_ *cobra.Command, argv []string) error {
		return runWithMgr(func(ctx context.Context, mgr *session.Manager) error {
			return Run(ctx, mgr, argv)
		})
	})
	return cmd
}
