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

package view

import (
	"fmt"

	"github.com/os-artificer/AiT/internal/aitlog/queryflags"
	"github.com/os-artificer/AiT/internal/aitlog/version"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"

	"github.com/spf13/cobra"
)

func newCmd() *cobra.Command {
	var f queryflags.Flags
	cmd := &cobra.Command{
		Use:           toolname.AILog + " [flags] [keyword...]",
		Short:         "View and search system logs",
		Long:          fmt.Sprintf("Run '%s --help' for usage and examples.", toolname.AILog),
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if f.ShowVersion {
				_, err := fmt.Fprintln(c.OutOrStdout(), version.Version())
				return err
			}
			return runView(c.Context(), c.OutOrStdout(), c.ErrOrStderr(), f, args)
		},
	}
	queryflags.Bind(cmd, &f)
	queryflags.BindVersion(cmd, &f)
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
