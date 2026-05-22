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

package cli

import (
	"io"
	"os"

	"github.com/os-artificer/AiT/internal/ait/helpdoc"
	"github.com/os-artificer/AiT/internal/ait/meta"
	"github.com/os-artificer/AiT/internal/ait/shell"

	"github.com/spf13/cobra"
)

func init() {
	meta.SetExecutorWriters(ExecuteWriters)
	shell.SetCLIExecutor(ExecuteWriters)
	shell.SetAIModeRunner(runLocalAIModeFromShell)
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "ait",
		Short:         "AiT — AI Terminal (SSH client)",
		Long:          "AiT is an AI-enhanced SSH client with session management, optional tmux persistence, and Deepseek assistant.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(
		newShellCmd(),
		newAICmd(),
		newLoginCmd(),
		newListCmd(),
		newLoadCmd(),
		newSuspendCmd(),
		newGetCmd(),
		newPutCmd(),
		newConfigCmd(),
		newSessionCmd(),
	)
	bindHelp(root, helpdoc.WriteRootHelp)
	return root
}

// Execute runs the Cobra root with args on os.Stdout/os.Stderr defaults.
func Execute(args []string) error {
	return ExecuteWriters(args, os.Stdout, os.Stderr)
}

// ExecuteWriters runs the Cobra root with custom stdout/stderr (e.g. ait shell REPL).
func ExecuteWriters(args []string, out, errOut io.Writer) error {
	root := newRootCmd()
	root.SetOut(out)
	root.SetErr(errOut)
	root.SetArgs(args)
	root.SilenceUsage = true
	root.SilenceErrors = true

	return root.Execute()
}
