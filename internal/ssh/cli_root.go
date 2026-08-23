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

package aissh

import (
	"context"
	"io"
	"os"

	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
	"github.com/os-artificer/AiT/internal/pkg/cliroute"
	"github.com/os-artificer/AiT/internal/pkg/config"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
	"github.com/os-artificer/AiT/internal/ssh/connect"
	"github.com/os-artificer/AiT/internal/ssh/registry"

	"github.com/spf13/cobra"
)

// Execute runs the Cobra root with args on os.Stdout and os.Stderr.
// Returns Cobra execution errors (including subcommand failures); does not set os.Exit.
func Execute(args []string) error {
	return ExecuteWriters(args, os.Stdout, os.Stderr)
}

// ExecuteWriters runs the top-level aissh CLI with custom stdout and stderr.
// Empty argv and -h/--help print root help; ssh-style argv run connect; else Cobra subcommands.
func ExecuteWriters(args []string, out, errOut io.Writer) error {
	if len(args) == 1 && args[0] == aitplugin.ManifestArg {
		return aitplugin.WriteManifest(out, toolname.AISsh)
	}

	args, saw := config.StripAIExperimentalFlag(args)
	if saw {
		config.SetAIExperimentalCLI(true)
	}

	if len(args) == 0 || pkghelpdoc.HelpRequested(args) {
		return registry.WriteRootHelp(out)
	}
	if cliroute.IsSSHConnectArgs(args) {
		return runWithManager(func(ctx context.Context, mgr *session.Manager) error {
			return connect.Run(ctx, mgr, args)
		})
	}
	return executeCobraRoot(args, out, errOut)
}

// ExecuteCobraOnly runs registered subcommands without the top-level SSH default route.
// Used from shell REPL, meta !ait, and local AI Agent ait_cli injection.
func ExecuteCobraOnly(args []string, out, errOut io.Writer) error {
	// Strip without flipping CLI gate (nested argv must not enable mid-process).
	args, _ = config.StripAIExperimentalFlag(args)

	if len(args) == 0 || pkghelpdoc.HelpRequested(args) {
		return registry.WriteRootHelp(out)
	}
	return executeCobraRoot(args, out, errOut)
}

// newRootCmd builds the Cobra root command with registry-attached subcommands and shared help binding.
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           toolname.AISsh,
		Short:         "AiT aissh — SSH client (AI features experimental, off by default)",
		Long:          pkghelpdoc.LongHint(toolname.AISsh),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	registry.AttachRoot(root, registry.Deps{
		RunWithMgr:       runWithManager,
		ExecuteWriters:   ExecuteWriters,
		ExecuteCobraOnly: ExecuteCobraOnly,
	})

	root.SetHelpFunc(func(c *cobra.Command, _ []string) {
		_ = registry.WriteRootHelp(c.OutOrStdout())
	})
	return root
}

func executeCobraRoot(args []string, out, errOut io.Writer) error {
	root := newRootCmd()
	root.SetOut(out)
	root.SetErr(errOut)
	root.SetArgs(args)
	root.SilenceUsage = true
	root.SilenceErrors = true
	return root.Execute()
}
