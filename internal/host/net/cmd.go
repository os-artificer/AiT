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

// Package net implements `aihost net`.
package net

import (
	"github.com/os-artificer/AiT/internal/host/shared"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"

	"github.com/spf13/cobra"
)

// Name is the registered subcommand name.
const Name = "net"

// New builds the `aihost net` command.
func New() *cobra.Command {
	var f netFlags
	cmd := &cobra.Command{
		Use:           Name + " [flags]",
		Short:         "Show network traffic or connections",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if pkghelpdoc.HasHelpArg(args) {
				return WriteHelp(c.OutOrStdout())
			}
			if err := f.validate(c); err != nil {
				return err
			}
			if f.Conn {
				return RunConnections(c.Context(), c.OutOrStdout(), c, f)
			}
			return shared.RunSection(c.Context(), c.OutOrStdout(), Name, f.Flags)
		},
	}
	bindNetFlags(cmd, &f)
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
