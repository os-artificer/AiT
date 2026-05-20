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

// Package ait implements the AiT CLI application.
package ait

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

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
		newLoginCmd(),
		newListCmd(),
		newLoadCmd(),
		newSuspendCmd(),
		newSyncCmd(),
		newConfigCmd(),
		newSessionCmd(),
	)
	return root
}

// Run is the application entry invoked from cmd/ait.
func Run(args []string) error {
	root := newRootCmd()
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
