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

	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/session"

	"github.com/spf13/cobra"
)

func newLoadCmd() *cobra.Command {
	var tmux bool
	var password string
	cmd := &cobra.Command{
		Use:   "load <id|alias>",
		Short: "Load a saved session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithManager(func(ctx context.Context, mgr *session.Manager) error {
				return mgr.Load(ctx, args[0], tmux, password)
			})
		},
	}
	cmd.Flags().BoolVar(&tmux, flagNameTmux, false, "attach remote tmux session")
	cmd.Flags().StringVar(&password, flagNamePassword, "", "SSH password (if auth expired)")
	bindHelp(cmd, helpdoc.WriteLoadHelp)
	return cmd
}
