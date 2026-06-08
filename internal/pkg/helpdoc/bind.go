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

package helpdoc

import (
	"io"

	"github.com/spf13/cobra"
)

// Writer renders --help output for one Cobra command into w.
type Writer func(w io.Writer) error

// Bind attaches login-style --help and SilenceUsage to cmd.
// write renders the help body; the default Long is set from LongHint(cmd.Name()).
func Bind(cmd *cobra.Command, write Writer) {
	cmd.SilenceUsage = true
	if cmd.Long == "" {
		cmd.Long = LongHint(cmd.Name())
	}

	cmd.SetHelpFunc(func(c *cobra.Command, _ []string) {
		_ = write(c.OutOrStdout())
	})
}

// BindCustomArgv is for DisableFlagParsing commands (login, get, put).
// run receives raw argv (after a --help check) and the resolved cobra command.
func BindCustomArgv(cmd *cobra.Command, write Writer, run func(c *cobra.Command, argv []string) error) {
	Bind(cmd, write)
	cmd.SilenceErrors = true
	_ = cmd.Flags().MarkHidden("help")

	cmd.RunE = func(c *cobra.Command, argv []string) error {
		if HelpRequested(argv) {
			return write(c.OutOrStdout())
		}
		return run(c, argv)
	}
}
