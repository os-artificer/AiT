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

	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/session"

	"github.com/spf13/cobra"
)

func newSessionCmd() *cobra.Command {
	var doList, doSuspend, verbose, tmux, force bool
	var loadTarget, switchTarget, deleteTarget, password string

	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage saved sessions",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			act, err := detectSessionAction(cmd, doList, doSuspend, loadTarget, switchTarget, deleteTarget)
			if err != nil {
				return err
			}

			return runWithManager(func(ctx context.Context, mgr *session.Manager) error {
				switch act.kind {
				case sessionActionList:
					return mgr.List(cmd.OutOrStdout(), verbose)
				case sessionActionLoad:
					target, err := resolveLoadTarget(act.target, args)
					if err != nil {
						return err
					}
					if target == "" {
						if err := session.CheckInteractiveStdin(); err != nil {
							return err
						}
						target, err = mgr.PickSession()
						if err != nil {
							return err
						}
					}
					return mgr.Load(ctx, target, tmux, password)
				case sessionActionSuspend:
					return mgr.SuspendCurrent()
				case sessionActionDelete:
					if act.target == "all" {
						return mgr.DeleteAll(force)
					}
					return mgr.Delete(act.target, force)
				case sessionActionSwitch:
					target, err := resolveSwitchTarget(cmd, act.target, args)
					if err != nil {
						return err
					}
					if target == "" {
						if err := session.CheckInteractiveStdin(); err != nil {
							return err
						}
						target, err = mgr.PickSession()
						if err != nil {
							return err
						}
					}
					return mgr.Switch(ctx, target, tmux, password, force)
				default:
					return fmt.Errorf("internal: unknown session action")
				}
			})
		},
	}

	cmd.Flags().BoolVar(&doList, flagNameList, false, "list saved sessions")
	cmd.Flags().BoolVarP(&verbose, flagNameVerbose, "v", false, "with --list, include FULL_ID column")
	cmd.Flags().StringVar(&loadTarget, flagNameLoad, "", "load session by id or alias (omit value to pick)")
	cmd.Flags().BoolVar(&doSuspend, flagNameSuspend, false, "suspend current session (prefer !ait suspend in SSH)")
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

	bindHelp(cmd, helpdoc.WriteSessionHelp)
	return cmd
}
