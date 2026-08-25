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
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// WriteRootHelp prints ait --help to w (triggers plugin discovery).
// Commands lists discovered plugins first (by order, then name), then reserved builtins.
// When no plugins are found, writes a no-plugins notice and Try block (AP9).
func WriteRootHelp(w io.Writer) error {
	plugins := aitplugin.Discover(context.Background())

	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(
		"AiT toolbox interactive shell.",
		"Run without arguments to open the REPL. Plugin tools are discovered under <install-dir>/plugins/.",
	); err != nil {
		return err
	}
	if err := b.WriteUsage(toolname.AIT); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionCommands, formatRootCommands(plugins)); err != nil {
		return err
	}

	if len(plugins) == 0 {
		if _, err := fmt.Fprintf(w, "%s\n\n%s\n\n", noPluginsMsg, errhint.FormatTryBlock(
			"make build",
			"make install",
			toolname.AIT+" (then) install",
			toolname.AIT,
		)); err != nil {
			return err
		}
	}

	if err := b.WriteExamples(
		toolname.AIT,
		toolname.AIT+" (then) help",
		toolname.AIT+" (then) install",
		toolname.AIT+" (then) agent",
		toolname.AIT+" (then) aissh user@host.example.com",
		toolname.AIT+" (then) ailog search error",
	); err != nil {
		return err
	}
	return b.Finish(toolname.AIT)
}

func formatRootCommands(plugins []aitplugin.Plugin) string {
	var b strings.Builder
	for _, p := range plugins {
		fmt.Fprintf(&b, "  %-12s %s\n", p.Manifest.Name, p.Manifest.Short)
	}
	fmt.Fprintf(&b, "  %-12s %s\n", cmdHelp, "show this help or help <plugin>")
	fmt.Fprintf(&b, "  %-12s %s\n", aitplugin.CmdAgent, "ops agent task loop (experimental)")
	fmt.Fprintf(&b, "  %-12s %s\n", cmdVersion, "print ait version")
	fmt.Fprintf(&b, "  %-12s %s\n", cmdHistory, "list or clear input history")
	fmt.Fprintf(&b, "  %-12s %s\n", cmdClear, "clear terminal screen")
	fmt.Fprint(&b, formatManageCommands())
	fmt.Fprintf(&b, "  %-12s %s\n", cmdExit+"/"+cmdQuit, "leave the REPL")
	return b.String()
}

func writePluginHelp(w io.Writer, name string) error {
	if _, ok := manageHelpTopics[name]; ok {
		return writeManageHelp(w, name)
	}

	p, ok := aitplugin.Lookup(name)
	if !ok {
		return errhint.WithTryMsg(
			helpPluginNotFound+": "+name,
			"Use help with no arguments to list discovered plugins.",
			toolname.AIT,
			cmdHelp,
		)
	}
	return aitplugin.WriteHelp(w, p.Manifest)
}
