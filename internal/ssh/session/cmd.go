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

// Package session implements the `aissh session` Cobra command (list/load/suspend/switch/delete).
package session

import (
	"context"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	pkgsession "github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/toolname"

	"github.com/spf13/cobra"
)

// RunWithMgr opens a session.Manager and runs fn under a signal-cancellable context.
type RunWithMgr func(fn func(ctx context.Context, mgr *pkgsession.Manager) error) error

// New builds the `aissh session` cobra command tree.
// runWithMgr opens a session.Manager and runs fn under a signal-cancellable context.
func New(runWithMgr RunWithMgr) *cobra.Command {
	var doList, doSuspend, verbose, tmux, force bool
	var loadTarget, switchTarget, deleteTarget, password string

	cmd := &cobra.Command{
		Use:           Name,
		Short:         "Manage saved sessions",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if pkghelpdoc.HasHelpArg(args) {
				return WriteSessionHelp(cmd.OutOrStdout())
			}
			act, err := detectSessionAction(cmd, doList, doSuspend, loadTarget, switchTarget, deleteTarget)
			if err != nil {
				return err
			}

			return runWithMgr(func(ctx context.Context, mgr *pkgsession.Manager) error {
				return execSessionAction(ctx, mgr, cmd, act, args, verbose, tmux, password, force)
			})
		},
	}

	cmd.Flags().BoolVar(&doList, flagNameList, false, "list saved sessions")
	cmd.Flags().BoolVarP(&verbose, flagNameVerbose, "v", false, "with --list, include FULL_ID column")
	cmd.Flags().StringVar(&loadTarget, flagNameLoad, "", "load session by id or alias (omit value to pick)")
	cmd.Flags().BoolVar(&doSuspend, flagNameSuspend, false,
		"suspend current session (prefer "+toolname.MetaLine("suspend")+" in SSH)")
	cmd.Flags().StringVar(&switchTarget, flagNameSwitch, "", "switch to session id or alias (omit value to pick)")
	cmd.Flags().StringVar(&deleteTarget, flagNameDelete, "", "delete session id, alias, or all")
	if f := cmd.Flags().Lookup(flagNameSwitch); f != nil {
		f.NoOptDefVal = "-"
	}
	if f := cmd.Flags().Lookup(flagNameLoad); f != nil {
		f.NoOptDefVal = "-"
	}
	cmd.Flags().BoolVar(&tmux, flagNameTmux, false, "use remote tmux for load or switch")
	cmd.Flags().StringVar(&password, flagNamePassword, "", "SSH password when session auth has expired")
	cmd.Flags().BoolVar(&force, flagNameForce, false, "force switch or skip delete confirmation")

	pkghelpdoc.Bind(cmd, WriteSessionHelp)
	return cmd
}
