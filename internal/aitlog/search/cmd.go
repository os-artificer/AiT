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

// Package search implements `aitlog search`, the explicit keyword-search form
// of the default view route.
package search

import (
	"github.com/os-artificer/AiT/internal/aitlog/queryflags"
	"github.com/os-artificer/AiT/internal/aitlog/view"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"

	"github.com/spf13/cobra"
)

// Name is the registered subcommand name.
const Name = "search"

// New builds the `aitlog search` command.
func New() *cobra.Command {
	var f queryflags.Flags
	cmd := &cobra.Command{
		Use:           Name + " [flags] [keyword...]",
		Short:         "Search system logs (keywords, --expr, or -e)",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if pkghelpdoc.HasHelpArg(args) {
				return WriteHelp(c.OutOrStdout())
			}
			if len(args) == 0 && f.Regex == "" && f.Expr == "" {
				return errhint.WithTryMsg(
					"search requires a keyword",
					"Provide at least one keyword, --expr, or -e/--regex.",
					toolname.AILog+" search error",
					toolname.AILog+` search --expr 'error|fail'`,
					toolname.AILog+" search version",
				)
			}
			return view.RunSearch(c.Context(), c.OutOrStdout(), c.ErrOrStderr(), f, args)
		},
	}
	queryflags.Bind(cmd, &f)
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
