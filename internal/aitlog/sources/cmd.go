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

// Package sources implements `aitlog sources`, which reports the log sources
// on this machine and which backend would serve each one.
package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/logsrc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"

	"github.com/spf13/cobra"
)

// Name is the registered subcommand name.
const Name = "sources"

// New builds the `aitlog sources` command.
func New() *cobra.Command {
	var jsonOut bool
	var backend string
	cmd := &cobra.Command{
		Use:           Name + " [flags]",
		Short:         "List available log sources and backends",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if pkghelpdoc.HasHelpArg(args) {
				return WriteHelp(c.OutOrStdout())
			}
			return runSources(c.OutOrStdout(), jsonOut, backend)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON array")
	cmd.Flags().StringVar(&backend, "backend", "auto", "auto|native|builtin")
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}

func runSources(w io.Writer, jsonOut bool, backend string) error {
	q := logsrc.Query{Backend: logsrc.Backend(backend)}
	if q.Backend == "" {
		q.Backend = logsrc.BackendAuto
	}
	r := logsrc.DefaultResolver()
	caps := r.Probe(q)
	if jsonOut {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(caps)
	}
	var b strings.Builder
	cur := ""
	for _, c := range caps {
		if c.Source != cur {
			if cur != "" {
				b.WriteByte('\n')
			}
			fmt.Fprintf(&b, "%s:\n", c.Source)
			cur = c.Source
		}
		mark := " "
		if c.Selected {
			mark = "*"
		}
		status := "unavailable"
		if c.Available {
			status = "available"
		}
		fmt.Fprintf(&b, "  %s %-8s %-28s %s", mark, c.Backend, c.Provider, status)
		if c.Reason != "" {
			fmt.Fprintf(&b, " (%s)", c.Reason)
		}
		b.WriteByte('\n')
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// WriteHelp prints aitlog sources --help to w.
func WriteHelp(w io.Writer) error {
	body := pkghelpdoc.NewBody(w)
	if err := body.WriteIntro(
		"List logical log sources and which backends are available.",
		"A leading * marks the layer that would be selected for --backend.",
	); err != nil {
		return err
	}
	if err := body.WriteUsage(toolname.AILog + " sources [flags]"); err != nil {
		return err
	}
	if err := body.WriteExamples(
		toolname.AILog+" sources",
		toolname.AILog+" sources --json",
		toolname.AILog+" sources --backend builtin",
	); err != nil {
		return err
	}
	return body.Finish(Name)
}
