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
	"fmt"

	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	var (
		alias    string
		host     string
		port     int
		user     string
		identity string
		password string
		tmux     bool
	)
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to a host and save session",
		RunE: func(cmd *cobra.Command, args []string) error {
			if host == "" {
				return fmt.Errorf("--host is required")
			}
			mgr, err := newManager()
			if err != nil {
				return err
			}
			defer mgr.Store.Close()
			ctx, cancel := commandContext()
			defer cancel()
			return mgr.Login(ctx, session.LoginOpts{
				Alias: alias, Host: host, Port: port, User: user,
				IdentityPath: identity, Password: password, Tmux: tmux,
			})
		},
	}
	cmd.Flags().StringVar(&alias, "alias", "", "session alias")
	cmd.Flags().StringVar(&host, "host", "", "remote host (required)")
	cmd.Flags().IntVar(&port, "port", 22, "SSH port")
	cmd.Flags().StringVar(&user, "user", "", "SSH user")
	cmd.Flags().StringVar(&identity, "identity", "", "private key path")
	cmd.Flags().StringVar(&password, "password", "", "SSH password")
	cmd.Flags().BoolVar(&tmux, "tmux", false, "use remote tmux for session persistence")
	return cmd
}
