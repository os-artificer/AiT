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

package aish

import (
	"io"
	"os"

	"github.com/os-artificer/AiT/internal/ai-ssh/registry"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"

	"github.com/spf13/cobra"
)

// newRootCmd builds the Cobra root command with registry-attached subcommands and shared help binding.
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           toolname.AISsh,
		Short:         "AiT ai-ssh — AI-enhanced SSH client",
		Long:          pkghelpdoc.LongHint(toolname.AISsh),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	registry.AttachRoot(root, registry.Deps{
		RunWithMgr:     runWithManager,
		ExecuteWriters: ExecuteWriters,
	})

	pkghelpdoc.Bind(root, registry.WriteRootHelp)
	return root
}

// Execute runs the Cobra root with args on os.Stdout and os.Stderr.
// Returns Cobra execution errors (including subcommand failures); does not set os.Exit.
func Execute(args []string) error {
	return ExecuteWriters(args, os.Stdout, os.Stderr)
}

// ExecuteWriters runs the Cobra root with custom stdout and stderr (e.g. ai-ssh shell REPL or meta !ai-ssh).
// args are passed to the root command; returns Cobra execution errors without calling os.Exit.
func ExecuteWriters(args []string, out, errOut io.Writer) error {
	root := newRootCmd()
	root.SetOut(out)
	root.SetErr(errOut)
	root.SetArgs(args)
	root.SilenceUsage = true
	root.SilenceErrors = true

	return root.Execute()
}
