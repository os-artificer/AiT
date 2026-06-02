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

package disable

import (
	"fmt"

	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"

	"github.com/spf13/cobra"
)

const (
	name           = "disable"
	msgDisabledFmt = "disabled provider %s\n"
)

// New builds the `ait ai provider disable` cobra command.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "disable <provider>",
		Short:         "Disable an AI provider (no requests will be sent to it)",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if !config.IsKnownProvider(id) {
				return unknownProviderErr(id)
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if err := cfg.SetProviderEnabled(id, false); err != nil {
				return err
			}
			if err := config.Save(cfg); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), msgDisabledFmt, id)
			return nil
		},
	}
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}

func unknownProviderErr(id string) error {
	return errhint.WithTryMsg(fmt.Sprintf("unknown provider %q", id),
		"Use deepseek, minimax, or ollama.",
		"ait ai provider enable deepseek",
		"ait ai provider enable minimax",
		"ait ai provider enable ollama",
		"ait ai provider --help",
	)
}
