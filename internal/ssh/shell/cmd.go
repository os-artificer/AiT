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

package shell

import (
	"context"
	"os"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/ssh/registry"

	"github.com/spf13/cobra"
)

// Name is the top-level subcommand name for aissh shell.
const Name = "shell"

// NewCmd builds the `aissh shell` cobra command.
func NewCmd(runWithMgr registry.RunWithMgr) *cobra.Command {
	cmd := &cobra.Command{
		Use:   Name,
		Short: "Interactive AiT shell (AiT> prompt)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pkghelpdoc.HasHelpArg(args) {
				return WriteShellHelp(cmd.OutOrStdout())
			}
			return runWithMgr(func(ctx context.Context, mgr *session.Manager) error {
				return RunREPL(ctx, mgr, os.Stdin, os.Stdout)
			})
		},
	}
	pkghelpdoc.Bind(cmd, WriteShellHelp)
	return cmd
}
