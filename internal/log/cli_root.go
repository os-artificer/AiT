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

package ailog

import (
	"fmt"
	"io"
	"os"

	"github.com/os-artificer/AiT/internal/log/registry"
	"github.com/os-artificer/AiT/internal/log/view"
	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"

	"github.com/spf13/cobra"
)

// Execute runs the top-level ailog CLI with args on os.Stdout and os.Stderr.
func Execute(args []string) error {
	return ExecuteWriters(args, os.Stdout, os.Stderr)
}

// ExecuteWriters runs the top-level ailog CLI with custom stdout and stderr.
func ExecuteWriters(args []string, stdout, stderr io.Writer) error {
	if len(args) == 1 && args[0] == aitplugin.ManifestArg {
		return aitplugin.WriteManifest(stdout, toolname.AILog)
	}
	if len(args) == 0 || pkghelpdoc.HelpRequested(args) {
		return WriteRootHelp(stdout)
	}
	if registry.IsRegistered(args[0]) {
		return executeCobraRoot(args, stdout, stderr)
	}
	ctx, cancel := CommandContext()
	defer cancel()
	return view.Run(ctx, stdout, stderr, args)
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           toolname.AILog,
		Short:         "View and search system logs",
		Long:          fmt.Sprintf("Run '%s --help' for usage and examples.", toolname.AILog),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	registry.AttachRoot(root)
	root.SetHelpFunc(func(c *cobra.Command, _ []string) {
		_ = WriteRootHelp(c.OutOrStdout())
	})
	return root
}

func executeCobraRoot(args []string, out, errOut io.Writer) error {
	ctx, cancel := CommandContext()
	defer cancel()
	root := newRootCmd()
	root.SetOut(out)
	root.SetErr(errOut)
	root.SetArgs(args)
	root.SetContext(ctx)
	root.SilenceUsage = true
	root.SilenceErrors = true
	return root.Execute()
}
