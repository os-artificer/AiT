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

package aivid

import (
	"fmt"
	"io"
	"os"

	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
	"github.com/os-artificer/AiT/internal/vid/registry"

	"github.com/spf13/cobra"
)

// Execute runs the top-level aitvid CLI with args on os.Stdout and os.Stderr.
func Execute(args []string) error {
	return ExecuteWriters(args, os.Stdout, os.Stderr)
}

// ExecuteWriters runs the top-level aitvid CLI with custom stdout and stderr.
func ExecuteWriters(args []string, stdout, stderr io.Writer) error {
	if len(args) == 1 && args[0] == aitplugin.ManifestArg {
		return aitplugin.WriteManifest(stdout, toolname.AITVid)
	}
	if len(args) == 0 || pkghelpdoc.HelpRequested(args) {
		return WriteRootHelp(stdout)
	}
	return executeCobraRoot(args, stdout, stderr)
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           toolname.VidCLI(""),
		Short:         "Search video content (OCR, objects, save hit frames)",
		Long:          fmt.Sprintf("Run '%s --help' for usage and examples.", toolname.VidCLI("")),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	registry.AttachRoot(root, registry.Deps{})
	root.SetHelpFunc(func(c *cobra.Command, _ []string) {
		_ = WriteRootHelp(c.OutOrStdout())
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
