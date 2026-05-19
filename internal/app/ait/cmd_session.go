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

package ait

import (
	"github.com/spf13/cobra"
)

func newSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage saved sessions",
	}
	cmd.AddCommand(newSessionSwitchCmd(), newSessionDeleteCmd())
	return cmd
}

func newSessionSwitchCmd() *cobra.Command {
	var tmux bool
	var password string
	var force bool
	cmd := &cobra.Command{
		Use:   "switch [id|alias]",
		Short: "Switch to another session",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := newManager()
			if err != nil {
				return err
			}
			defer mgr.Store.Close()
			target := ""
			if len(args) > 0 {
				target = args[0]
			} else {
				target, err = mgr.PickSession()
				if err != nil {
					return err
				}
			}
			ctx, cancel := commandContext()
			defer cancel()
			return mgr.Switch(ctx, target, tmux, password, force)
		},
	}
	cmd.Flags().BoolVar(&tmux, "tmux", false, "use tmux for target session")
	cmd.Flags().StringVar(&password, "password", "", "SSH password if auth expired")
	cmd.Flags().BoolVar(&force, "force", false, "switch even if current session is still marked active")
	return cmd
}

func newSessionDeleteCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <id|alias|all>",
		Short: "Delete one session or all sessions (ait session delete all)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := newManager()
			if err != nil {
				return err
			}
			defer mgr.Store.Close()
			if args[0] == "all" {
				return mgr.DeleteAll(force)
			}
			return mgr.Delete(args[0], force)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "skip confirmation")
	return cmd
}
