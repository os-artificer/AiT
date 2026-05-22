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

	"github.com/spf13/cobra"
)

type helpWriter func(io.Writer) error

// bindHelp attaches login-style --help and SilenceUsage to cmd.
func bindHelp(cmd *cobra.Command, write helpWriter) {
	cmd.SilenceUsage = true
	if cmd.Long == "" {
		cmd.Long = helpdoc.LongHint(cmd.Name())
	}
	cmd.SetHelpFunc(func(c *cobra.Command, _ []string) {
		_ = write(c.OutOrStdout())
	})
}

// bindCustomArgvHelp is for DisableFlagParsing commands (login, get, put).
func bindCustomArgvHelp(
	cmd *cobra.Command,
	write helpWriter,
	run func(context.Context, *session.Manager, []string) error,
) {
	bindHelp(cmd, write)
	cmd.SilenceErrors = true
	_ = cmd.Flags().MarkHidden("help")
	cmd.RunE = func(c *cobra.Command, argv []string) error {
		if helpdoc.HelpRequested(argv) {
			return write(c.OutOrStdout())
		}
		return runWithManager(func(ctx context.Context, mgr *session.Manager) error {
			return run(ctx, mgr, argv)
		})
	}
}
