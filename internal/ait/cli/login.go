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

	"github.com/os-artificer/AiT/internal/ait/login"
	"github.com/os-artificer/AiT/internal/pkg/session"

	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "login [flags] destination [command [argument ...]]",
		Short:                 "Log in to a host (OpenSSH-compatible) and save session",
		Long:                  "Connect to a remote host using OpenSSH-compatible flags. Run 'ait login --help' for flag details and examples.",
		DisableFlagsInUseLine: true,
		DisableFlagParsing:    true,
		SilenceUsage:          true,
		SilenceErrors:         true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if login.HelpRequested(args) {
				return login.WriteHelp(cmd.OutOrStdout())
			}

			return runWithManager(func(ctx context.Context, mgr *session.Manager) error {
				return login.Run(ctx, mgr, args)
			})
		},
	}
	cmd.SetHelpFunc(func(c *cobra.Command, _ []string) {
		_ = login.WriteHelp(c.OutOrStdout())
	})
	_ = cmd.Flags().MarkHidden("help")
	return cmd
}
