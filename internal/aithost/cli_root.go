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

package aithost

import (
	"fmt"
	"io"
	"os"

	"github.com/os-artificer/AiT/internal/aithost/registry"
	"github.com/os-artificer/AiT/internal/aithost/top"
	"github.com/os-artificer/AiT/internal/aithost/version"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"

	"github.com/spf13/cobra"
)

// Execute runs the top-level aithost CLI with args on os.Stdout and os.Stderr.
func Execute(args []string) error {
	return ExecuteWriters(args, os.Stdout, os.Stderr)
}

// ExecuteWriters runs the top-level aithost CLI with custom stdout and stderr.
func ExecuteWriters(args []string, stdout, stderr io.Writer) error {
	if len(args) > 0 && pkghelpdoc.HelpRequested(args) {
		return WriteRootHelp(stdout)
	}
	if len(args) == 1 && (args[0] == "-V" || args[0] == "--version") {
		_, err := fmt.Fprintln(stdout, version.Version())
		return err
	}
	if len(args) == 0 {
		ctx, cancel := CommandContext()
		defer cancel()
		return top.Run(ctx, stdout, stderr, top.Options{})
	}
	if registry.IsRegistered(args[0]) {
		return executeCobraRoot(args, stdout, stderr)
	}
	return errhint.WithTryMsg(
		fmt.Sprintf("unknown command %q", args[0]),
		"Unrecognized aithost command.",
		toolname.AIHost+" --help",
		toolname.AIHost+" top",
		toolname.AIHost+" snapshot",
	)
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           toolname.AIHost,
		Short:         "Live host CPU, memory, disk, network, and process monitor",
		Long:          fmt.Sprintf("Run '%s --help' for usage and examples.", toolname.AIHost),
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
