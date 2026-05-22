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
	"os"

	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	"github.com/os-artificer/AiT/internal/ait/shell"
	"github.com/os-artificer/AiT/internal/pkg/session"

	"github.com/spf13/cobra"
)

func newShellCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Interactive AiT shell (AiT> prompt)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithManager(func(ctx context.Context, mgr *session.Manager) error {
				return shell.RunREPL(ctx, mgr, os.Stdin, os.Stdout)
			})
		},
	}
	bindHelp(cmd, helpdoc.WriteShellHelp)
	return cmd
}
