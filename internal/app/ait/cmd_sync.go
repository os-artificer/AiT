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
	"github.com/os-artificer/AiT/internal/pkg/sshclient"
	syncpkg "github.com/os-artificer/AiT/internal/pkg/sync"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	var sessionID string
	cmd := &cobra.Command{
		Use:   "sync <src> <dst>",
		Short: "Sync files across hosts (rsync/scp/sftp)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := newManager()
			if err != nil {
				return err
			}
			defer mgr.Store.Close()
			ctx, cancel := commandContext()
			defer cancel()
			opts := syncpkg.Options{Src: args[0], Dst: args[1], AI: mgr.AI}
			if sessionID != "" {
				ssh, err := mgr.SSHOptionsForSession(sessionID, "")
				if err != nil {
					return err
				}
				opts.SSH = ssh
			} else {
				opts.SSH = sshclient.Options{UseAgent: true}
			}
			return syncpkg.Run(ctx, opts)
		},
	}
	cmd.Flags().StringVar(&sessionID, "session", "", "use saved session connection")
	return cmd
}
