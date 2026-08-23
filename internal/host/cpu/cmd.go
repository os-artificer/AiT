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

// Package cpu implements `aihost cpu`.
package cpu

import (
	"github.com/os-artificer/AiT/internal/host/shared"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"

	"github.com/spf13/cobra"
)

// Name is the registered subcommand name.
const Name = "cpu"

// New builds the `aihost cpu` command.
func New() *cobra.Command {
	var f shared.Flags
	cmd := &cobra.Command{
		Use:           Name + " [flags]",
		Short:         "Show CPU usage",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if pkghelpdoc.HasHelpArg(args) {
				return WriteHelp(c.OutOrStdout())
			}
			return shared.RunSection(c.Context(), c.OutOrStdout(), Name, f)
		},
	}
	shared.BindJSONInterval(cmd, &f)
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
