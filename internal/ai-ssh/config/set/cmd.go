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

package set

import (
	"fmt"

	"github.com/os-artificer/AiT/internal/ai-ssh/config/shared"
	pkgconfig "github.com/os-artificer/AiT/internal/pkg/config"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"

	"github.com/spf13/cobra"
)

const name = "set"

// New builds the `ai-ssh config set` cobra command.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "set <key> <value>",
		Short:         "Set config value",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]
			if provider, ok := shared.ParseAPIKeyConfigKey(key); ok {
				if err := shared.SetProviderAPIKey(provider, value); err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), shared.APIKeyStoredMessage(provider))
				return nil
			}

			cfg, err := pkgconfig.Load()
			if err != nil {
				return err
			}
			if err := cfg.SetKey(key, value); err != nil {
				return err
			}
			return pkgconfig.Save(cfg)
		},
	}
	pkghelpdoc.Bind(cmd, WriteHelp)
	return cmd
}
