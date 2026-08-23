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

// Package sources implements `aihost sources`.
package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/hoststats"

	"github.com/spf13/cobra"
)

// Name is the registered subcommand name.
const Name = "sources"

// New builds the `aihost sources` command.
func New() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:           Name + " [flags]",
		Short:         "List metrics collectors",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if pkghelpdoc.HasHelpArg(args) {
				return WriteHelp(c.OutOrStdout())
			}
			return runSources(c.OutOrStdout(), jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON array")
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}

func runSources(w io.Writer, jsonOut bool) error {
	srcs := hoststats.ListSources()
	if jsonOut {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(srcs)
	}
	var b strings.Builder
	for _, s := range srcs {
		fmt.Fprintf(&b, "%-6s %-22s %-10s %s\n", s.ID, s.Name, s.Status, s.Notes)
	}
	_, err := io.WriteString(w, b.String())
	return err
}
