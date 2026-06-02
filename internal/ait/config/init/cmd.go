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

package init

import (
	"fmt"
	"strings"

	pkgconfig "github.com/os-artificer/AiT/internal/pkg/config"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"

	"github.com/spf13/cobra"
)

const name = "init"

// New builds the `ait config init` cobra command.
func New() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:           name,
		Short:         "Create or update config.yaml with defaults for missing keys",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			res, err := pkgconfig.EnsureFile(dryRun)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			switch {
			case res.Created && !dryRun:
				fmt.Fprintf(out, "Created %s\n", res.Path)
			case res.Created && dryRun:
				fmt.Fprintf(out, "Would create %s\n", res.Path)
			case res.Updated:
				if dryRun {
					fmt.Fprintf(out, "Would update %s (added: %s)\n", res.Path, strings.Join(res.AddedKeys, ", "))
				} else if len(res.AddedKeys) > 0 {
					fmt.Fprintf(out, "Updated %s (added: %s)\n", res.Path, strings.Join(res.AddedKeys, ", "))
				} else {
					fmt.Fprintf(out, "Updated %s\n", res.Path)
				}
			default:
				fmt.Fprintf(out, "Config at %s is up to date\n", res.Path)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print changes without writing config.yaml")
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
