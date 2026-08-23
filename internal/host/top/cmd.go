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

package top

import (
	"time"

	"github.com/os-artificer/AiT/internal/host/shared"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"

	"github.com/spf13/cobra"
)

// Name is the registered subcommand name.
const Name = "top"

// New builds the `aihost top` command.
func New() *cobra.Command {
	var f shared.Flags
	cmd := &cobra.Command{
		Use:           Name + " [flags]",
		Short:         "Live multi-panel host dashboard",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if pkghelpdoc.HasHelpArg(args) {
				return WriteHelp(c.OutOrStdout())
			}
			pq, err := f.ProcessQuery()
			if err != nil {
				return err
			}
			interval := f.Interval
			if interval <= 0 {
				interval = time.Second
			}
			return Run(c.Context(), c.OutOrStdout(), c.ErrOrStderr(), Options{
				Interval: interval,
				Limit:    pq.Limit,
				Sort:     pq.Sort,
			})
		},
	}
	shared.BindJSONInterval(cmd, &f)
	shared.BindProcessQueryLimit(cmd, &f, 0)
	_ = cmd.Flags().MarkHidden("json")
	_ = cmd.Flags().MarkHidden("watch")
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
